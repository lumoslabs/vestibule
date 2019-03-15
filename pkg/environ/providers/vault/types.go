package vault

import (
	"errors"

	"github.com/hashicorp/vault/api"
)

var (
	// ErrVaultEmptyResponse is returned when vault respondes with no data
	ErrVaultEmptyResponse = errors.New("no data returned from vault")
	// ErrVaultUnexpectedResponse is returned when vault does not respond with the expected data
	ErrVaultUnexpectedResponse = errors.New("unexpected response from vault")
	// ErrNotInKubernetes is returned when vestibule is not running in a kubernetes cluster
	ErrNotInKubernetes = errors.New("not running in kubernetes cluster")
)

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
