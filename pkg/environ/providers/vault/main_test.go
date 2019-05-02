package vault

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-ini/ini"

	"github.com/spf13/afero"

	"github.com/lumoslabs/vestibule/pkg/environ"
	"github.com/lumoslabs/vestibule/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	kvKey      = "kv/foo/bar"
	vaultToken = "1234-5678"

	vaultAuthResponse = `
  {
    "lease_duration": 100,
    "renewable": true,
    "auth": {
      "client_token": "1234-5678",
      "accessor": "1234-5678",
      "policies": ["default"]
    }
  }`

	vaultSecretDataResponse = `
  {
    "data": {
      "data": %s,
      "metadata": {
        "version": %d
      }
    }
  }`

	vaultAWSResponse = `
  {
    "data": {
      "access_key": "aws-access-key",
      "secret_key": "aws-secret-key",
      "security_token": "aws-session-token"
    }
  }`
)

func testServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if testing.Verbose() {
			fmt.Printf("vault GET %s\n", r.RequestURI)
		}
		switch {
		default:
			http.Error(w, `{"errors":[]}`, http.StatusNotFound)
		case strings.HasPrefix(r.RequestURI, "/v1/auth"):
			if testing.Verbose() {
				data, _ := ioutil.ReadAll(r.Body)
				fmt.Printf("vault DATA %s\n", string(data))
			}
			fmt.Fprintln(w, vaultAuthResponse)
		case strings.HasPrefix(r.RequestURI, "/v1/kv/data"):
			var version int
			versions, ok := r.URL.Query()["version"]
			if !ok || len(versions[0]) < 1 {
				version = 0
			} else {
				version, _ = strconv.Atoi(versions[0])
			}

			n, er := strconv.Atoi(filepath.Base(r.RequestURI))
			if er != nil {
				n = 0
			}

			d := make(map[string]string)
			for i := 0; i <= n; i++ {
				d[fmt.Sprint(i)] = "data"
			}
			data, _ := json.Marshal(d)
			fmt.Fprintf(w, fmt.Sprintf(vaultSecretDataResponse, string(data), version))
		case strings.HasPrefix(r.RequestURI, "/v1/aws/sts"):
			fmt.Fprintln(w, vaultAWSResponse)
		}
	}))
}

func TestNew(t *testing.T) {
	tt := []struct {
		envv   map[string]string
		method string
		path   string
	}{
		{map[string]string{"VAULT_TOKEN": "1234"}, "kubernetes", "auth/kubernetes/login"},
		{map[string]string{EnvVaultAppRole: "test", EnvKubernetesServiceHost: "host", EnvKubernetesServicePort: "port"}, "kubernetes", "auth/kubernetes/login"},
		{map[string]string{EnvVaultAppRole: "test", EnvVaultAppSecret: "secret"}, "approle", "auth/approle/login"},
		{map[string]string{EnvVaultAppRole: "test", EnvVaultAppJWT: "secret"}, "jwt", "auth/jwt/login"},
		{map[string]string{EnvVaultAppRole: "test", EnvVaultAppJWT: "secret", EnvVaultAuthPath: "my-jwt"}, "my-jwt", "auth/my-jwt/login"},
		{map[string]string{EnvVaultAuthPath: "okta/login/foo", EnvVaultAuthData: `{"password":"password"}`}, "okta", "auth/okta/login/foo"},
	}

	// if testing.Verbose() {
	// 	log.SetLogger(log.NewDebugLogger())
	// }

	fs = afero.NewMemMapFs()
	currEnv := os.Environ()
	ts := testServer()
	afero.WriteFile(fs, kubernetesTokenFilePath, []byte("token"), 0644)

	defer func() {
		ts.Close()
		os.Clearenv()
		log.SetLogger(log.NewNilLogger())
		for _, item := range currEnv {
			if parts := strings.Split(item, "="); len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}()

	for i, test := range tt {
		os.Clearenv()
		os.Setenv("VAULT_ADDR", ts.URL)
		for k, v := range test.envv {
			os.Setenv(k, v)
		}

		c, er := New()
		require.NoError(t, er)
		token := c.(*Client).Token()

		if tok, ok := test.envv["VAULT_TOKEN"]; ok {
			assert.Equalf(t, tok, token, `%d: %#v`, i, c)
		} else {
			assert.Equalf(t, vaultToken, token, `%d: %#v`, i, c)
		}

		assert.Equalf(t, test.method, c.(*Client).AuthMethod, `%d: %#v`, i, c)
		assert.Equalf(t, test.path, c.(*Client).AuthPath, `%d: %#v`, i, c)
	}
}

func TestKeyParser(t *testing.T) {
	tt := []struct {
		keys string
		len  int
		ver  bool
	}{
		{"kv/foo/bar", 1, false},
		{"/kv/foo/bar:kv/bif/baz", 2, false},
		{"/kv//foo/bar@1:kv/bif/baz@2", 2, true},
	}

	for _, test := range tt {
		keys, er := vaultKeyParser(test.keys)
		require.NoError(t, er)

		switch keys := keys.(type) {
		default:
			t.Errorf("Wrong type for parsed key: %T", keys)
		case KVKeys:
			assert.Equal(t, test.len, len(keys))
			if test.ver {
				for i, k := range keys {
					assert.Equal(t, i+1, *(k.Version))
				}
			}
		}
	}
}

func TestAddToEnviron(t *testing.T) {
	tt := []struct {
		envv   map[string]string
		envLen int
	}{
		{map[string]string{EnvVaultKeys: kvKey + "/0"}, 1},
		{map[string]string{EnvVaultKeys: "/" + kvKey + "/1"}, 2},
		{map[string]string{EnvVaultKeys: kvKey + "/baz/2"}, 3},
		{map[string]string{EnvVaultKeys: kvKey + "/3:" + kvKey + "/baz/3"}, 4},
		{map[string]string{EnvVaultKeys: kvKey + "@2"}, 1},
		{map[string]string{EnvVaultIAMRole: "test"}, 4},
	}

	// if testing.Verbose() {
	// 	log.SetLogger(log.NewDebugLogger())
	// }

	fs = afero.NewMemMapFs()
	currEnv := os.Environ()
	ts := testServer()
	afero.WriteFile(fs, kubernetesTokenFilePath, []byte("jwt"), 0644)

	defer func() {
		ts.Close()
		os.Clearenv()
		log.SetLogger(log.NewNilLogger())
		for _, item := range currEnv {
			if parts := strings.Split(item, "="); len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}()

	for i, test := range tt {
		os.Clearenv()
		os.Setenv("VAULT_ADDR", ts.URL)
		os.Setenv(EnvVaultAuthData, "{}")
		for k, v := range test.envv {
			os.Setenv(k, v)
		}

		c, er := New()
		require.NoError(t, er)

		e := environ.New()
		c.AddToEnviron(e)
		assert.Equalf(t, test.envLen, e.Len(), `%d: env=%v`, i, e)

		if _, ok := test.envv[EnvVaultKeys]; ok {
			val, ok := e.Load("0")
			assert.True(t, ok)
			assert.Equalf(t, "data", val, `%d: env=%v`, i, e)
		}

		if _, ok := test.envv[EnvVaultIAMRole]; ok {
			ak, ok := e.Load("AWS_ACCESS_KEY_ID")
			assert.True(t, ok)
			assert.Equalf(t, "aws-access-key", ak, `%d: env=%v`, i, e)

			content, _ := afero.ReadFile(fs, c.(*Client).AwsCredFile)
			data, _ := ini.Load(content)
			assert.Equalf(t, "aws-access-key", data.Section("default").Key("aws_access_key_id").String(), `%d: env=%v`, i, e)
			assert.Equalf(t, "aws-secret-key", data.Section("default").Key("aws_secret_access_key").String(), `%d: env=%v`, i, e)
			assert.Equalf(t, "aws-session-token", data.Section("default").Key("aws_session_token").String(), `%d: env=%v`, i, e)
		}
	}
}
