package vault

import (
	"bytes"
	"encoding/json"
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

	kubernetesTokenFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

func init() {
	environ.RegisterProvider(Name, New)
}

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
		keyParts := strings.Split(key.Path, "/")
		if len(keyParts) < 2 {
			log.Debugf("Ignoring invalid vault KV key. key=%s", key.Path)
			continue
		}

		tail := len(keyParts) - 1
		for i := 0; i <= tail; i++ {
			if keyParts[i] == "" {
				keyParts = append(keyParts[:i], keyParts[i+1:tail+1]...)
				tail--
				i--
			}
		}

		// Add 'data' as the second path element if it does not exist
		if keyParts[1] != "data" {
			keyParts = append(keyParts, "")
			copy(keyParts[2:], keyParts[1:])
			keyParts[1] = "data"
		}

		reqPath := strings.Join(keyParts, "/")
		reqData := make(map[string][]string)
		if key.Version != nil {
			reqData["version"] = []string{strconv.Itoa(*(key.Version))}
		}

		log.Debugf("Fetching KVv2 secret from vault. key=%s data=%#v", reqPath, reqData)
		response, er := c.Logical().ReadWithData(reqPath, reqData)

		if er != nil || response == nil {
			log.Debugf("Failed to get KVv2 secret from vault, trying KVv1. key=%s err=%v", reqPath, er)

			reqPath = strings.Join(append(keyParts[:1], keyParts[2:]...), "/")
			response, er := c.Logical().Read(reqPath)
			if er != nil || response == nil {
				log.Debugf("Failed to get KVv1 secret from vault. key=%s err=%v", reqPath, er)
				continue
			}
		}

		var secrets map[string]interface{}
		if meta, data := response.Data["metadata"], response.Data["data"]; meta != nil && data != nil {
			var ok bool
			if secrets, ok = data.(map[string]interface{}); !ok {
				log.Debugf("Unexpected response from vault: %#v")
				return ErrVaultUnexpectedResponse
			}
		} else {
			secrets = response.Data
		}

		env := make(map[string]string)
		for k, v := range secrets {
			if s, ok := v.(string); ok {
				env[k] = s
			}
		}

		e.SafeMerge(env)
	}

	if c.IamRole != "" {

		// attempt to get aws creds from vault
		// only looks for sts roles
		reqPath := strings.TrimSpace(strings.Trim(c.AwsPath, "/")) + "/sts/" + strings.TrimSpace(c.IamRole)
		if creds, er := c.getAwsCreds(reqPath); er == nil {

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
					log.Debugf("Failed writing shared aws credentials file. file=%s err=%v", c.AwsCredFile, er)
				}
			}

			e.SafeMerge(creds)
		} else {
			log.Debugf("Failed to get aws creds from vault. path=%s err=%v", reqPath, er)
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
		return map[string]string(nil), ErrVaultEmptyResponse
	}

	data := make(map[string]string, len(iam.Data))
	for key, value := range iam.Data {
		if v, ok := value.(string); ok {
			data[key] = v
		}
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
		keyParts := strings.SplitN(k, VaultKeySeparator, 2)
		key := KVKey{Path: strings.TrimLeft(keyParts[0], "/"), Version: nil}

		keys = append(keys, &key)

		if len(keyParts) == 1 {
			continue
		}

		v, er := strconv.Atoi(keyParts[1])
		if er != nil {
			return nil, er
		}

		key.Version = &v
	}

	return keys, nil
}

func getKubernetesSAToken() ([]byte, error) {
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 && len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		return []byte(nil), ErrNotInKubernetes
	}
	if _, er := fs.Stat(kubernetesTokenFilePath); er != nil {
		return []byte(nil), ErrNotInKubernetes
	}

	return ioutil.ReadFile(kubernetesTokenFilePath)
}
