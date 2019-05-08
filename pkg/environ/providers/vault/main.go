package vault

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ini/ini"

	"github.com/spf13/afero"

	util "github.com/Masterminds/goutils"
	env "github.com/caarlos0/env/v5"
	"github.com/hashicorp/vault/api"
	"github.com/lumoslabs/vestibule/pkg/environ"
	"github.com/lumoslabs/vestibule/pkg/log"
)

var fs = afero.NewOsFs()

const (
	// Name is the Provider name
	Name = "vault"

	// VaultKeysSeparator is the separator between vault keys in KeysEnvVar
	VaultKeysSeparator = ":"

	// VaultKeySeparator is the separator between key and version in KeysEnvVar
	VaultKeySeparator = "@"

	kubernetesTokenFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

	defaultClientTimeout = time.Second * 3
	defaultClientRetries = 1
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

	vcc := api.DefaultConfig()
	if util.IsBlank(os.Getenv(api.EnvVaultClientTimeout)) {
		vcc.Timeout = defaultClientTimeout
	}
	if util.IsBlank(os.Getenv(api.EnvVaultMaxRetries)) {
		vcc.MaxRetries = defaultClientRetries
	}

	vc, er := api.NewClient(vcc)

	if er != nil {
		return nil, er
	}

	v := &Client{Client: vc}
	p := env.CustomParsers{
		reflect.TypeOf(KVKey{}):             parseVaultKVKey,
		reflect.TypeOf(&RedactedAuthData{}): parseRedactedAuthData,
	}

	if er := env.ParseWithFuncs(v, p); er != nil {
		return nil, er
	}
	log.Debugf("Generated new vault client. client=%+v", v)

	if v.Token() == "" {
		if er := v.SetVaultToken(); er != nil {
			return nil, fmt.Errorf("vault login failed: %v", er)
		}
	}
	return v, nil
}

// SetVaultToken sets the AuthMethod and AuthPath if not already set and uses those to request a session token from vault
func (client *Client) SetVaultToken() error {
	client.SetLoginPath()

	data := make(map[string]interface{})
	switch {
	case !util.IsBlank(client.AppRole) && !util.IsBlank(client.AppSecret):
		log.Debugf("using approle method=%s path=%s", client.AuthMethod, client.AuthPath)
		data["role_id"] = client.AppRole
		data["secret_id"] = client.AppSecret
	case !util.IsBlank(client.AppRole) && !util.IsBlank(client.AppJWT):
		log.Debugf("using jwt method=%s path=%s", client.AuthMethod, client.AuthPath)
		data["role"] = client.AppRole
		data["jwt"] = client.AppJWT
	case client.AuthData != nil:
		data = client.AuthData.toGenericMap()
	default:
		if inCluster() {
			log.Debugf("using kubernetes method=%s path=%s", client.AuthMethod, client.AuthPath)
			kt, er := afero.ReadFile(fs, kubernetesTokenFilePath)
			if er != nil || len(kt) == 0 {
				return fmt.Errorf("failed to get k8s service account token: %v", er)
			}
			data["role"] = client.AppRole
			data["jwt"] = string(kt)
		} else {
			log.Debugf("using approle without secret_id method=%s path=%s", client.AuthMethod, client.AuthPath)
			// Assume we are using approle with only a role_id
			data["role_id"] = client.AppRole
		}
	}

	client.SetToken("token")
	if log.IsDebug() {
		d, _ := json.Marshal(redact(data))
		log.Debugf("Requesting session token from vault. path=%s data=%s", client.AuthPath, string(d))
	}
	auth, er := client.Logical().Write(client.AuthPath, data)
	if er != nil {
		return er
	}

	token, er := auth.TokenID()
	if er != nil {
		return er
	}
	if util.IsBlank(token) {
		return ErrVaultEmptyResponse
	}

	client.SetToken(token)
	return nil
}

// SetAuthMethod sets the AuthMethod if not already set
func (client *Client) SetAuthMethod() {
	switch {
	case !util.IsBlank(client.AuthMethod):
	case !util.IsBlank(client.AppSecret):
		client.AuthMethod = "approle"
	case !util.IsBlank(client.AppJWT):
		client.AuthMethod = "jwt"
	default:
		client.AuthMethod = "kubernetes"
	}
}

// SetLoginPath sets the api path to login with vault for the auth method
func (client *Client) SetLoginPath() {
	client.SetAuthMethod()

	if util.IsBlank(client.AuthPath) {
		client.AuthPath = fmt.Sprintf(`auth/%s/login`, client.AuthMethod)
	} else {
		parts := strings.Split(client.AuthPath, "/")
		if parts[0] != "auth" {
			parts = append([]string{"auth"}, parts...)
		}
		if len(parts) == 2 {
			parts = append(parts, "login")
		}
		client.AuthPath = strings.Join(parts, "/")
		client.AuthMethod = parts[1]
	}

	client.AuthPath = strings.TrimSpace(client.AuthPath)
}

// AddToEnviron iterates through the given []VaultKeys, decoding the data returned from each key into a map[string]string
// and merging it into the environ.Environ
func (client *Client) AddToEnviron(env *environ.Environ) error {
	for _, ev := range sensitiveEnvVars {
		env.Delete(ev)
	}

	var wg sync.WaitGroup

	for _, key := range client.Keys {
		wg.Add(1)
		go func(k KVKey) {
			defer wg.Done()
			if data, er := client.getKVData(k); er == nil {
				env.SafeMerge(data)
			} else {
				log.Debugf("Failed to get data for key. key=%s err=%v", k.Path, er)
			}
		}(key)
	}

	if !util.IsBlank(client.AwsRole) || !util.IsBlank(client.IamRole) {
		p := strings.TrimSpace(strings.Trim(client.AwsPath, "/")) + "/sts/"
		if util.IsBlank(client.AwsRole) {
			p = p + strings.TrimSpace(client.IamRole)
		} else {
			p = p + strings.TrimSpace(client.AwsRole)
		}

		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			// attempt to get aws creds from vault
			// only looks for sts roles
			creds, er := client.getAwsCreds(path)
			if er != nil {
				log.Debugf("Failed to get aws creds from vault. path=%s err=%v", path, er)
				return
			}

			if er := client.writeAwsSharedFile(
				creds[EnvAwsAccessKeyId],
				creds[EnvAwsSecretAccessKey],
				creds[EnvAwsSessionToken],
			); er != nil {
				log.Infof("Failed to write aws shared credentials file. file=%s err=%+v", client.AwsCredFile, er)
				return
			}

			creds[EnvAwsSharedCredFile] = client.AwsCredFile
			env.SafeMerge(creds)
		}(p)
	}

	if !util.IsBlank(client.GcpRole) {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			if client.GcpCredType != "key" && client.GcpCredType != "token" {
				return
			}

			creds, er := client.getGCPCreds(path)
			if er != nil {
				log.Infof("Failed to get gcp credentials from vault. path=%s err=%+v", path, er)
				return
			}

			switch client.GcpCredType {
			case "token":
				env.SafeMerge(map[string]string{EnvGoogleToken: creds["token"]})
			case "key":
				if er := client.writeGCPKeyFile(creds["private_key_data"]); er != nil {
					log.Infof("Failed to write gcp credentials file. path=%s err=%+v", client.GcpCredFile, er)
					return
				}
				env.SafeMerge(map[string]string{EnvGoogleCredFile: client.GcpCredFile})
			}
		}(strings.TrimSpace(strings.Trim(client.GcpPath, "/")) + "/" + client.GcpCredType + "/" + strings.TrimSpace(client.GcpRole))
	}

	wg.Wait()
	return nil
}

func (client *Client) getKVData(key KVKey) (map[string]string, error) {
	keyParts := strings.Split(key.Path, "/")
	if len(keyParts) < 2 {
		return nil, ErrInvalidKVKey
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
	response, er := client.Logical().ReadWithData(reqPath, reqData)

	if er != nil || response == nil {
		log.Debugf("Failed to get KVv2 secret from vault, trying KVv1. key=%s err=%v", reqPath, er)

		reqPath = strings.Join(append(keyParts[:1], keyParts[2:]...), "/")
		response, er := client.Logical().Read(reqPath)
		if er != nil {
			return nil, er
		}
		if response == nil {
			return nil, ErrUnexpectedVaultResponse
		}
	}

	var responseData map[string]interface{}
	if meta, data := response.Data["metadata"], response.Data["data"]; meta != nil && data != nil {
		var ok bool
		if responseData, ok = data.(map[string]interface{}); !ok {
			return nil, ErrUnexpectedVaultResponse
		}
	} else {
		responseData = response.Data
	}

	e := make(map[string]string)
	for k, v := range responseData {
		if s, ok := v.(string); ok {
			e[k] = s
		}
	}
	return e, nil
}

func (client *Client) getAwsCreds(path string) (map[string]string, error) {
	log.Debugf("Requesting aws credentials from vault. path=%s", path)
	iam, er := client.Logical().Read(path)
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
		EnvAwsAccessKeyId:     data["access_key"],
		EnvAwsSecretAccessKey: data["secret_key"],
		EnvAwsSessionToken:    data["security_token"],
	}, nil
}

func (client *Client) writeAwsSharedFile(accessKey, secretKey, sessionToken string) error {
	var file afero.File
	var content *ini.File

	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	content, er := ini.Load(client.AwsCredFile)
	if er != nil {
		content = ini.Empty()
	}

	section, _ := content.NewSection(client.AwsProfile)
	section.NewKey(strings.ToLower(EnvAwsAccessKeyId), accessKey)
	section.NewKey(strings.ToLower(EnvAwsSecretAccessKey), secretKey)
	if sessionToken != "" {
		section.NewKey(strings.ToLower(EnvAwsSessionToken), sessionToken)
	}
	buf := new(bytes.Buffer)
	content.WriteTo(buf)

	if er := fs.MkdirAll(filepath.Dir(client.AwsCredFile), 0755); er != nil {
		return er
	}

	f, er := fs.Open(client.AwsCredFile)
	if er != nil {
		f, er = fs.Create(client.AwsCredFile)
		if er != nil {
			return er
		}
	}

	f.Write(buf.Bytes())
	f.Close()
	return nil
}

func (client *Client) getGCPCreds(path string) (map[string]string, error) {
	log.Debugf("Requesting GCP credentials from vault. path=%s", path)
	resp, er := client.Logical().Read(path)
	if er != nil {
		return map[string]string(nil), er
	}
	if resp == nil {
		return map[string]string(nil), ErrVaultEmptyResponse
	}

	data := make(map[string]string, len(resp.Data))
	for key, value := range resp.Data {
		if v, ok := value.(string); ok {
			data[key] = v
		}
	}

	return data, nil
}

func (client *Client) writeGCPKeyFile(encoded string) error {
	creds, er := base64.StdEncoding.DecodeString(encoded)
	if er != nil {
		return er
	}

	if er := fs.MkdirAll(filepath.Dir(client.GcpCredFile), 0755); er != nil {
		return er
	}

	f, er := fs.Open(client.GcpCredFile)
	if er != nil {
		f, er = fs.Create(client.GcpCredFile)
		if er != nil {
			return er
		}
	}

	f.Write([]byte(creds))
	return f.Close()
}

func (vd *RedactedAuthData) toGenericMap() map[string]interface{} {
	gm := make(map[string]interface{}, len(vd.data))
	for k, v := range vd.data {
		gm[k] = v
	}
	return gm
}

func (vd *RedactedAuthData) String() string {
	redacted := redact(vd.toGenericMap())
	out, er := json.Marshal(redacted)
	if er != nil {
		return fmt.Sprintf("%v", redacted)
	}
	return string(out)
}

func parseVaultKVKeys(s string) (interface{}, error) {
	log.Debugf("parsing kv keys kvkey=%v", s)
	keys := KVKeys{}

	for _, k := range strings.Split(s, VaultKeysSeparator) {
		key, _ := parseVaultKVKey(k)
		keys = append(keys, key.(KVKey))
	}

	return keys, nil
}

func parseVaultKVKey(s string) (interface{}, error) {
	keyParts := strings.SplitN(s, VaultKeySeparator, 2)
	key := KVKey{Path: strings.TrimLeft(keyParts[0], "/"), Version: nil}

	if len(keyParts) == 1 {
		return key, nil
	}

	v, er := strconv.Atoi(keyParts[1])
	if er != nil {
		return nil, er
	}

	key.Version = &v
	return key, nil
}

func parseRedactedAuthData(s string) (interface{}, error) {
	data := make(map[string]string)
	if er := json.Unmarshal([]byte(s), &data); er != nil {
		return nil, fmt.Errorf("failed to parse VAULT_AUTH_DATA: %v", er)
	}
	return &RedactedAuthData{data: data}, nil
}

func redact(sensitive map[string]interface{}) map[string]string {
	clean := make(map[string]string, len(sensitive))
	for k, v := range sensitive {
		switch k {
		default:
			clean[k] = v.(string)
		case "jwt", "secret_id", "role_id", "password", "identity", "signature", "pkcs7", "token":
			clean[k] = "[REDACTED]"
		}
	}
	return clean
}

func inCluster() bool {
	_, er := fs.Stat(kubernetesTokenFilePath)
	return util.IsBlank(os.Getenv(EnvKubernetesServiceHost)) && util.IsBlank(os.Getenv(EnvKubernetesServicePort)) && er != nil
}
