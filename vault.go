package main

import (
	vApi "github.com/hashicorp/vault/api"
	"k8s.io/client-go/rest"
)

type vaultClient struct {
	*vApi.Client
}

func newVaultClient() (*vaultClient, error) {
	vc, er := vApi.NewClient(vApi.DefaultConfig())
	if er != nil {
		return nil, er
	}

	if c, er := rest.InClusterConfig(); er == nil {
		vc.SetToken(c.BearerToken)
	}
	return &vaultClient{vc}, er
}

func (v *vaultClient) getKeys(c *vaultConfig, e *envv) error {
	for _, key := range c.Keys {
		data := make(map[string][]string)
		data["version"] = []string{"0"}
		if key.Version != 0 {
			data["version"] = []string{"1"}
		}

		s, er := v.Logical().ReadWithData(key.Path, data)
		if er != nil {
			return er
		}

		var env map[string]string
		for k, v := range s.Data {
			if s, ok := v.(string); ok {
				env[k] = s
			}
		}
		e.Add(env)
	}
	return nil
}
