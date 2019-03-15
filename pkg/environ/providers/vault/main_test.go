package vault

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/lumoslabs/vestibule/pkg/environ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
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
      "data": {
        "foo": "bar"
      },
      "metadata": {
        "version": %d
      }
    }
  }`

	vaultAWSResponse = `
  {
    "data": {
      "access_key": "1234",
			"secret_key": "1234",
			"security_token": "1234"
    }
  }`
)

func testServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if testing.Verbose() {
		// 	fmt.Printf("vault GET %s\n", r.RequestURI)
		// }
		switch {
		default:
			http.Error(w, `{"errors":[]}`, http.StatusNotFound)
		case strings.HasPrefix(r.RequestURI, "/v1/auth"):
			fmt.Fprintln(w, vaultAuthResponse)
		case strings.HasPrefix(r.RequestURI, "/v1/kv/data"):
			var version int
			versions, ok := r.URL.Query()["version"]
			if !ok || len(versions[0]) < 1 {
				version = 0
			} else {
				version, _ = strconv.Atoi(versions[0])
			}

			fmt.Fprintf(w, fmt.Sprintf(vaultSecretDataResponse, version))
		case strings.HasPrefix(r.RequestURI, "/v1/aws/sts"):
			fmt.Fprintln(w, vaultAWSResponse)
		}
	}))
}

func TestNew(t *testing.T) {
	tt := []struct {
		kvKeys string
		token  string
	}{
		{"kv/foo/bar", "1234"},
		{"kv/foo/bar", ""},
	}

	ts := testServer()
	defer ts.Close()

	os.Setenv("VAULT_ADDR", ts.URL)
	defer func() {
		os.Unsetenv("VAULT_ADDR")
	}()

	for _, test := range tt {
		os.Setenv("VAULT_KV_KEYS", test.kvKeys)
		if test.token != "" {
			os.Setenv("VAULT_TOKEN", test.token)
		}

		c, er := New()
		require.NoError(t, er)
		token := c.(*Client).Token()
		if test.token != "" {
			assert.Equal(t, test.token, token)
		} else {
			assert.Equal(t, "1234-5678", token)
		}

		os.Unsetenv("VAULT_KV_KEYS")
		os.Unsetenv("VAULT_TOKEN")
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
		keys string
		iam  string
	}{
		{"kv/foo/bar", ""},
		{"/kv/data/foo/bar", ""},
		{"kv/foo/bar", "my-role"},
		{"kv/foo/bar@2", "my-role"},
	}

	fs = afero.NewMemMapFs()
	ts := testServer()
	defer ts.Close()

	os.Setenv("VAULT_ADDR", ts.URL)
	defer func() {
		os.Unsetenv("VAULT_ADDR")
	}()

	for _, test := range tt {
		os.Setenv("VAULT_KV_KEYS", test.keys)
		if test.iam != "" {
			os.Setenv("VAULT_IAM_ROLE", test.iam)
		}

		c, er := New()
		require.NoError(t, er)

		e := environ.New()
		c.AddToEnviron(e)
		if test.iam != "" {
			assert.Equal(t, 5, e.Len())
		} else {
			assert.Equal(t, 1, e.Len())
		}

		val, ok := e.Load("foo")
		assert.True(t, ok)
		assert.Equal(t, "bar", val)

		if test.iam != "" {
			ak, ok := e.Load("AWS_ACCESS_KEY_ID")
			assert.True(t, ok)
			assert.Equal(t, "1234", ak)

			content, _ := afero.ReadFile(fs, c.(*Client).AwsCredFile)
			assert.Equal(t, fmt.Sprintf(awsCredentialsFileFmt, "1234", "1234"), string(content))
		}

		os.Unsetenv("VAULT_KV_KEYS")
		os.Unsetenv("VAULT_IAM_ROLE")
	}
}
