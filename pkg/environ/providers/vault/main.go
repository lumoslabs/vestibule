package vault

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
	"github.com/hashicorp/vault/api"
	"github.com/lumoslabs/vestibule/pkg/environ"
	"k8s.io/client-go/rest"
)

const (
	// Name is the Provider name
	Name = "vault"

	// VaultKeySeparator is the separator between key and version in KeysEnvVar
	VaultKeySeparator = "@"

	// KeysEnvVar is the environment variable holding the keys to lookup
	KeysEnvVar = "VAULT_KEYS"
)

// New returns a Client as an environ.Provider or an error if configuring failed. If running in a Kubernetes
// cluster and not provided a token, will use the service account token.
func New() (environ.Provider, error) {
	defer func() { os.Unsetenv(KeysEnvVar) }()
	vc, er := api.NewClient(api.DefaultConfig())
	if er != nil {
		return nil, er
	}

	if c, er := rest.InClusterConfig(); er == nil && vc.Token() == "" {
		vc.SetToken(c.BearerToken)
	}

	v := &Client{Client: vc}
	p := env.CustomParsers{reflect.TypeOf(vaultKey{}): vaultKeyParser}
	if er := env.ParseWithFuncs(v, p); er != nil {
		return nil, er
	}
	return v, nil
}

// AddToEnviron iterates through the given []VaultKeys, decoding the data returned from each key into a map[string]string
// and merging it into the environ.Environ
func (c *Client) AddToEnviron(e *environ.Environ) error {
	e.Delete(KeysEnvVar)
	for _, key := range c.Keys {
		s, er := c.Logical().ReadWithData(key.Path, map[string][]string{
			"version": []string{strconv.Itoa(key.Version)},
		})
		if er != nil {
			return er
		}

		env := make(map[string]string)
		for k, v := range s.Data {
			if s, ok := v.(string); ok {
				env[k] = s
			}
		}
		e.SafeMerge(env)
	}
	return nil
}

func vaultKeyParser(s string) (interface{}, error) {
	var (
		bits = strings.SplitN(s, VaultKeySeparator, 2)
		k    = vaultKey{Path: bits[0]}
	)

	if len(bits) == 1 {
		return k, nil
	}

	v, er := strconv.Atoi(bits[1])
	if er != nil {
		return nil, er
	}

	k.Version = v
	return &k, nil
}
