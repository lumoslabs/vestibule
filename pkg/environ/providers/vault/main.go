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
	Name              = "vault"
	VaultKeySeparator = "@"
	KeysEnvVar        = "VAULT_KEYS"
)

func NewVaultProvider() (environ.Provider, error) {
	vc, er := api.NewClient(api.DefaultConfig())
	if er != nil {
		return nil, er
	}

	if c, er := rest.InClusterConfig(); er == nil {
		vc.SetToken(c.BearerToken)
	}

	v := &VaultProvider{c: vc}
	p := env.CustomParsers{reflect.TypeOf(VaultKey{}): vaultKeyParser}
	if er := env.ParseWithFuncs(v, p); er != nil {
		return nil, er
	}
	os.Unsetenv(KeysEnvVar)
	return v, nil
}

func (v *VaultProvider) AddToEnviron(e *environ.Environ) error {
	e.Delete(KeysEnvVar)
	for _, key := range v.Keys {
		data := make(map[string][]string)
		data["version"] = []string{"0"}
		if key.Version != 0 {
			data["version"] = []string{"1"}
		}

		s, er := v.c.Logical().ReadWithData(key.Path, data)
		if er != nil {
			return er
		}

		env := make(map[string]string)
		for k, v := range s.Data {
			if s, ok := v.(string); ok {
				env[k] = s
			}
		}
		e.Append(env)
	}
	return nil
}

func vaultKeyParser(s string) (interface{}, error) {
	var (
		bits = strings.Split(s, VaultKeySeparator)
		k    = VaultKey{Path: bits[0]}
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
