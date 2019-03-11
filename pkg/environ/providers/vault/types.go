package vault

import "github.com/hashicorp/vault/api"

// Client is an environ.Provider and github.com/hashicorp/vault/api.Client which will get the requested keys
type Client struct {
	*api.Client
	AuthMethod  string `env:"VAULT_AUTH_METHOD" envDefault:"kubernetes"`
	AuthPath    string `env:"VAULT_AUTH_PATH"`
	AuthData    string `env:"VAULT_AUTH_DATA" envDefault:"{}"`
	AppRole     string `env:"VAULT_APP_ROLE"`
	IamRole     string `env:"VAULT_IAM_ROLE"`
	AwsPath     string `env:"VAULT_AWS_PATH" envDefault:"aws"`
	AwsCredFile string `env:"AWS_SHARED_CREDENTIALS_FILE" envDefault:"/var/aws/credentials"`
	Keys        KVKeys `env:"VAULT_KV_KEYS"`
}

// KVKeys is an alias for []*KVKey. Needed for caarlos0/env to support parsing.
type KVKeys []*KVKey

// KVKey is a kv ver2 key in Vault
type KVKey struct {
	Path    string
	Version *int
}
