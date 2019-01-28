package vault

import "github.com/hashicorp/vault/api"

type VaultProvider struct {
	Keys []*VaultKey `env:"VAULT_KEYS" envSeparator:":"`
	c    *api.Client
}

type VaultKey struct {
	Path    string
	Version int
}
