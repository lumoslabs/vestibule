package vault

import "github.com/hashicorp/vault/api"

// Client is an environ.Provider and github.com/hashicorp/vault/api.Client which will get the requested keys
type Client struct {
	*api.Client
	Keys []*vaultKey `env:"VAULT_KEYS" envSeparator:":"`
}

type vaultKey struct {
	Path    string
	Version int
}
