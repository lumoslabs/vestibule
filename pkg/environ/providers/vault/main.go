package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/afero"

	"github.com/caarlos0/env"
	"github.com/hashicorp/vault/api"
	"github.com/lumoslabs/vestibule/pkg/environ"
	"k8s.io/client-go/rest"
)

var (
	fs  = afero.NewOsFs()
	log = environ.GetLogger()

	sensitiveEnvVars = []string{
		"VAULT_KV_KEYS",
		"VAULT_AUTH_DATA",
	}
)

const (
	// Name is the Provider name
	Name = "vault"

	// VaultKeysSeparator is the separator between vault keys in KeysEnvVar
	VaultKeysSeparator = ":"

	// VaultKeySeparator is the separator between key and version in KeysEnvVar
	VaultKeySeparator = "@"

	awsCredentialsFileFmt = "[default]\naws_access_key_id=%s\naws_secret_access_key=%s\n"
)

// New returns a Client as an environ.Provider or an error if configuring failed. If running in a Kubernetes
// cluster and not provided a token, will use the service account token.
func New() (environ.Provider, error) {
	defer func() {
		for _, ev := range sensitiveEnvVars {
			os.Unsetenv(ev)
		}
	}()

	log.Debugf("Creating vault api client. addr=%v", os.Getenv("VAULT_ADDR"))
	vc, er := api.NewClient(api.DefaultConfig())
	if er != nil {
		return nil, er
	}

	v := &Client{Client: vc}
	p := env.CustomParsers{reflect.TypeOf(KVKeys{}): vaultKeyParser}
	if er := env.ParseWithFuncs(v, p); er != nil {
		return nil, er
	}

	if v.Token() == "" {
		if er := v.setVaultToken(); er != nil {
			return nil, fmt.Errorf("Vault login failed: %v", er)
		}
	}
	return v, nil
}

func (c *Client) setVaultToken() error {
	var data map[string]interface{}

	if kc, er := rest.InClusterConfig(); er == nil {
		data = make(map[string]interface{})
		data["role"] = c.AppRole
		data["jwt"] = kc.BearerToken
	} else {
		var d interface{}
		if er := json.Unmarshal([]byte(c.AuthData), &d); er != nil {
			return er
		}
		data = d.(map[string]interface{})
	}

	var path string
	if c.AuthPath != "" {
		path = fmt.Sprintf("auth/%s", strings.TrimPrefix(c.AuthPath, "auth/"))
	} else {
		path = fmt.Sprintf("auth/%s/login", c.AuthMethod)
	}

	c.SetToken("token")
	log.Debugf("Requesting session token from vault. path=%s data=%#v", path, data)
	auth, er := c.Logical().Write(path, data)
	if er != nil {
		return er
	}

	c.SetToken(auth.Auth.ClientToken)
	return nil
}

// AddToEnviron iterates through the given []VaultKeys, decoding the data returned from each key into a map[string]string
// and merging it into the environ.Environ
func (c *Client) AddToEnviron(e *environ.Environ) error {
	for _, ev := range sensitiveEnvVars {
		e.Delete(ev)
	}
	for _, key := range c.Keys {
		bits := strings.Split(key.Path, "/")
		if len(bits) < 2 {
			continue
		}

		tail := len(bits) - 1
		for i := 0; i <= tail; i++ {
			if bits[i] == "" {
				bits = append(bits[:i], bits[i+1:tail+1]...)
				tail--
				i--
			}
		}

		// Add 'data' as the second path element if it does not exist
		if bits[1] != "data" {
			bits = append(bits, "")
			copy(bits[2:], bits[1:])
			bits[1] = "data"
		}

		p := strings.Join(bits, "/")
		d := make(map[string][]string)
		if key.Version != nil {
			d["version"] = []string{strconv.Itoa(*(key.Version))}
		}
		s, er := c.Logical().ReadWithData(p, d)
		if er != nil || s == nil {
			log.Debugf("Failed to get kv2 secret from vault, trying kv1. key=%s err=%s", p, er.Error())
			p = strings.Join(append(bits[:1], bits[2:]...), "/")
			s, er := c.Logical().ReadWithData(p, d)
			if er != nil || s == nil {
				log.Infof("Failed to get kv secret from vault. key=%s err=%s", p, er.Error())
				continue
			}
		}

		var data map[string]interface{}
		kvData := s.Data
		_, metaOK := kvData["metadata"]
		_, dataOK := kvData["data"]
		if metaOK && dataOK {
			kv2Data, ok := kvData["data"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("Unexpected response from Vault: %T %#v", kvData, kvData)
			}
			data = kv2Data
		} else {
			data = kvData
		}

		env := make(map[string]string)
		for k, v := range data {
			if s, ok := v.(string); ok {
				env[k] = s
			}
		}

		e.SafeMerge(env)
	}

	if c.IamRole != "" {
		path := strings.TrimSpace(strings.Trim(c.AwsPath, "/")) + "/sts/" + strings.TrimSpace(c.IamRole)
		log.Debugf("Requesting aws credentials from vault. path=%s", path)
		iam, er := c.Logical().Read(path)
		if er != nil {
			return er
		}

		accessKey, ok := iam.Data["access_key"].(string)
		if !ok {
			return fmt.Errorf("Unexpected response from Vault. path=%s", path)
		}
		secretKey, ok := iam.Data["secret_key"].(string)
		if !ok {
			return fmt.Errorf("Unexpected response from Vault. path=%s", path)
		}

		creds := map[string]string{
			"AWS_ACCESS_KEY_ID":     accessKey,
			"AWS_SECRET_ACCESS_KEY": secretKey,
		}

		if er := fs.MkdirAll(filepath.Dir(c.AwsCredFile), 0755); er == nil {
			if f, er := fs.Create(c.AwsCredFile); er == nil {
				f.WriteString(fmt.Sprintf(awsCredentialsFileFmt, accessKey, secretKey))
				f.Close()
				creds["AWS_SHARED_CREDENTIALS_FILE"] = c.AwsCredFile
			} else {
				log.Infof("Failed writing shared aws credentials file. file=%s error=%v", c.AwsCredFile, er)
			}
		}

		e.SafeMerge(creds)
	}
	return nil
}

func vaultKeyParser(s string) (interface{}, error) {

	keys := KVKeys{}

	for _, k := range strings.Split(s, VaultKeysSeparator) {
		bits := strings.SplitN(k, VaultKeySeparator, 2)
		key := KVKey{Path: strings.TrimLeft(bits[0], "/"), Version: nil}

		keys = append(keys, &key)

		if len(bits) == 1 {
			continue
		}

		v, er := strconv.Atoi(bits[1])
		if er != nil {
			return nil, er
		}

		key.Version = &v
	}

	return keys, nil
}
