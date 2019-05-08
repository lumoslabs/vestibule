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

	// {"private_key":"private-key","private_key_id":"private-key-id"}
	vaultGcpKeyResponse = `
    {
      "data": {
        "private_key_data": "eyJwcml2YXRlX2tleSI6InByaXZhdGUta2V5IiwicHJpdmF0ZV9rZXlfaWQiOiJwcml2YXRlLWtleS1pZCJ9Cg=="
      }
    }
  `

	vaultGcpKeyResponseFail = `
    {
      "data": {
        "private_key_data": ""
      }
    }
  `
)

func testServer(dbg bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if testing.Verbose() && dbg {
			log.Debugf("mock-vault %s %s", r.Method, r.RequestURI)
		}
		switch {
		default:
			http.Error(w, `{"errors":[]}`, http.StatusNotFound)
		case strings.HasPrefix(r.RequestURI, "/v1/auth"):
			if testing.Verbose() && dbg {
				data, _ := ioutil.ReadAll(r.Body)
				log.Debugf("mock-vault\t\tdata=%s\n", string(data))
			}
			fmt.Fprintln(w, vaultAuthResponse)
		case strings.HasPrefix(r.RequestURI, "/v1/secrets/data"):
			if strings.Contains(r.RequestURI, VaultKeysSeparator) {
				http.Error(w, `{"errors":["kv keys not split"]}`, http.StatusNotFound)
				return
			}
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
		case strings.HasPrefix(r.RequestURI, "/v1/gcp/key"):
			if strings.Contains(r.RequestURI, "fail") {
				fmt.Fprintln(w, vaultGcpKeyResponseFail)
			} else {
				fmt.Fprintln(w, vaultGcpKeyResponse)
			}
		}
	}))
}

func TestNew(t *testing.T) {
	tt := []struct {
		envv   map[string]string
		method string
		path   string
	}{
		{map[string]string{"VAULT_TOKEN": "1234"}, "", ""},
		{map[string]string{EnvVaultAppRole: "test", EnvKubernetesServiceHost: "host", EnvKubernetesServicePort: "port"}, "kubernetes", "auth/kubernetes/login"},
		{map[string]string{EnvVaultAppRole: "test", EnvVaultAppSecret: "secret"}, "approle", "auth/approle/login"},
		{map[string]string{EnvVaultAppRole: "test", EnvVaultAppJWT: "secret"}, "jwt", "auth/jwt/login"},
		{map[string]string{EnvVaultAppRole: "test", EnvVaultAppJWT: "secret", EnvVaultAuthPath: "my-jwt"}, "my-jwt", "auth/my-jwt/login"},
		{map[string]string{EnvVaultAuthPath: "okta/login/foo", EnvVaultAuthData: `{"password":"password"}`}, "okta", "auth/okta/login/foo"},
	}

	if testing.Verbose() && os.Getenv("CI_DEBUG_TRACE") == "true" {
		log.SetLogger(log.NewDebugLogger())
	}

	currEnv := os.Environ()
	ts := testServer(os.Getenv("CI_DEBUG_TRACE") == "true")

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
		fs = afero.NewMemMapFs()
		afero.WriteFile(fs, kubernetesTokenFilePath, []byte("jwt"), 0644)

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
		keys, er := parseVaultKVKeys(test.keys)
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
		{map[string]string{EnvVaultKeys: "secrets/foo/bar/0"}, 1},
		{map[string]string{EnvVaultKeys: "secrets/foo/bar/1"}, 2},
		{map[string]string{EnvVaultKeys: "secrets/foo/bar/baz/2"}, 3},
		{map[string]string{EnvVaultKeys: "secrets/foo/bar/3:/secrets/foo/bar/baz/3"}, 4},
		{map[string]string{EnvVaultKeys: "secrets/foo/bar@2"}, 1},
		{map[string]string{EnvVaultAwsRole: "test"}, 4},
		{map[string]string{EnvVaultGcpRole: "test"}, 1},
		{map[string]string{EnvVaultGcpRole: "fail"}, 1},
	}

	if testing.Verbose() && os.Getenv("CI_DEBUG_TRACE") == "true" {
		log.SetLogger(log.NewDebugLogger())
	}

	currEnv := os.Environ()
	ts := testServer(os.Getenv("CI_DEBUG_TRACE") == "true")

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
		fs = afero.NewMemMapFs()
		afero.WriteFile(fs, kubernetesTokenFilePath, []byte("jwt"), 0644)

		os.Setenv("VAULT_ADDR", ts.URL)
		os.Setenv(EnvVaultAuthData, "{}")
		for k, v := range test.envv {
			os.Setenv(k, v)
		}

		c, er := New()
		require.NoErrorf(t, er, `%d: vars=%v`, i, test.envv)

		e := environ.New()
		c.AddToEnviron(e)
		assert.Equalf(t, test.envLen, e.Len(), `%d: vars=%v env=%v`, i, test.envv, e)

		if _, ok := test.envv[EnvVaultKeys]; ok {
			val, ok := e.Load("0")
			assert.True(t, ok)
			assert.Equalf(t, "data", val, `%d: vars=%v env=%v`, i, test.envv, e)
		}

		if _, ok := test.envv[EnvVaultAwsRole]; ok {
			ak, ok := e.Load("AWS_ACCESS_KEY_ID")
			assert.True(t, ok)
			assert.Equalf(t, "aws-access-key", ak, `%d: vars=%v env=%v`, i, test.envv, e)

			content, er := afero.ReadFile(fs, c.(*Client).AwsCredFile)
			require.NoErrorf(t, er, `%d: vars=%v env=%v`, i, test.envv, e)
			data, er := ini.Load(content)
			require.NoErrorf(t, er, `%d: vars=%v env=%v`, i, test.envv, e)

			assert.Equalf(t, "aws-access-key", data.Section("default").Key("aws_access_key_id").String(), `%d: vars=%v env=%v`, i, test.envv, e)
			assert.Equalf(t, "aws-secret-key", data.Section("default").Key("aws_secret_access_key").String(), `%d: vars=%v env=%v`, i, test.envv, e)
			assert.Equalf(t, "aws-session-token", data.Section("default").Key("aws_session_token").String(), `%d: vars=%v env=%v`, i, test.envv, e)
		}

		if role, ok := test.envv[EnvVaultGcpRole]; ok {
			content, er := afero.ReadFile(fs, c.(*Client).GcpCredFile)
			require.NoErrorf(t, er, `%d: vars=%v env=%v`, i, test.envv, e)
			var data map[string]string
			if role == "fail" {
				er := json.Unmarshal(content, &data)
				assert.Error(t, er, `%d: vars=%v env=%v content=%s`, i, test.envv, e, string(content))
				t.Logf("expected gcp error: %+v", er)
			} else {
				require.NoErrorf(t, json.Unmarshal(content, &data), `%d: vars=%v env=%v content=%s`, i, test.envv, e, string(content))

				assert.Equalf(t, "private-key", data["private_key"], `%d: vars=%v env=%v`, i, test.envv, e)
				assert.Equalf(t, "private-key-id", data["private_key_id"], `%d: vars=%v env=%v`, i, test.envv, e)
			}
		}
	}
}
