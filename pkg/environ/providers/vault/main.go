package vault

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-ini/ini"

	"github.com/spf13/afero"

	"github.com/caarlos0/env"
	"github.com/hashicorp/vault/api"
	"github.com/lumoslabs/vestibule/pkg/environ"
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

	awsCredentialsFileFmt   = "[default]\naws_access_key_id=%s\naws_secret_access_key=%s\n"
	kubernetesTokenFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
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
	vaultConfig := api.DefaultConfig()
	vaultConfig.Timeout = time.Second * 5
	vaultConfig.HttpClient.Timeout = time.Second * 5
	vaultConfig.MaxRetries = 1
	vc, er := api.NewClient(vaultConfig)
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

	if token, er := getKubernetesSAToken(); er == nil && len(token) > 0 {
		data = make(map[string]interface{})
		data["role"] = c.AppRole
		data["jwt"] = string(token)
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
			log.Debugf("Ignoring invalid vault KV key. key=%s", key.Path)
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
		log.Debugf("Fetching KVv2 secret from vault. key=%s data=%#v", p, d)
		s, er := c.Logical().ReadWithData(p, d)
		if er != nil || s == nil {
			log.Debugf("Failed to get KVv2 secret from vault, trying KVv1. key=%s err=%v", p, er)
			p = strings.Join(append(bits[:1], bits[2:]...), "/")
			s, er := c.Logical().Read(p)
			if er != nil || s == nil {
				log.Debugf("Failed to get KVv1 secret from vault. key=%s err=%v", p, er)
				continue
			}
		}

		data, ok := s.Data["data"].(map[string]interface{})
		if !ok {
			log.Debugf("Unexpected response from vault: %#v")
			return VaultUnexpectedResponseErr
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
		// attempt to get aws creds from vault
		// only looks for sts roles
		path := strings.TrimSpace(strings.Trim(c.AwsPath, "/")) + "/sts/" + strings.TrimSpace(c.IamRole)
		creds, er := c.getAwsCreds(path)

		if er != nil || len(creds) > 0 {

			// attempt to write received creds to file
			if er := fs.MkdirAll(filepath.Dir(c.AwsCredFile), 0755); er == nil {
				if f, er := fs.Create(c.AwsCredFile); er == nil {
					content := ini.Empty()
					section, _ := content.NewSection("default")
					for key, value := range creds {
						section.NewKey(strings.ToLower(key), value)
					}
					buf := new(bytes.Buffer)
					content.WriteTo(buf)
					f.Write(buf.Bytes())
					f.Close()
					creds["AWS_SHARED_CREDENTIALS_FILE"] = c.AwsCredFile
				} else {
					log.Debugf("Failed writing shared aws credentials file. file=%s error=%v", c.AwsCredFile, er)
				}
			}

			e.SafeMerge(creds)
		} else {
			log.Debugf("Failed to get aws creds from vault. path=%s err=%v", path, er)
		}
	}
	return nil
}

func (c *Client) getAwsCreds(path string) (map[string]string, error) {
	log.Debugf("Requesting aws credentials from vault. path=%s", path)
	iam, er := c.Logical().Read(path)
	if er != nil {
		return map[string]string(nil), er
	}
	if iam == nil {
		return map[string]string(nil), VaultEmptyResponseErr
	}

	data := make(map[string]string, len(iam.Data))
	for key, value := range iam.Data {
		if v, ok := value.(string); ok {
			data[key] = v
		}
	}
	securityToken, ok := iam.Data["security_token"].(string)
	if !ok {
		return creds, errors.New("vault did not return a security token")
	}

	return map[string]string{
		"AWS_ACCESS_KEY_ID":     data["access_key"],
		"AWS_SECRET_ACCESS_KEY": data["secret_key"],
		"AWS_SESSION_TOKEN":     data["security_token"],
	}, nil
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

func getKubernetesSAToken() ([]byte, error) {
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 && len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		return []byte(nil), NotInKubernetesErr
	}
	if _, er := os.Stat(kubernetesTokenFilePath); er != nil {
		return []byte(nil), NotInKubernetesErr
	}

	return ioutil.ReadFile(kubernetesTokenFilePath)
}
