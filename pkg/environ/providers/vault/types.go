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

	sensitiveEnvVars = []string{
		"VAULT_KV_KEYS",
		"VAULT_AUTH_DATA",
	}

	// EnvVars is a map of known vonfiguration environment variables and their usage descriptions
	EnvVars = map[string]string{
		"VAULT_KV_KEYS": `If VAULT_KV_KEYS is set, will iterate over each key (colon separated), attempting to get the secret from Vault.
Secrets are pulled at the optional version or latest, then injected into Environ. If running in Kubernetes,
the Pod's ServiceAccount token will automatically be looked up and used for Vault authentication.
e.g. VAULT_KV_KEYS=/path/to/key1[@version]:/path/to/key2[@version]:...`,
		"VAULT_AUTH_METHOD": `Authentication method for vault. Default is "kubernetes".`,
		"VAULT_APP_ROLE":    "App role name to use with the kubernetes authentication method.",
		"VAULT_IAM_ROLE": `IAM role to request from vault. If returns credentials, the access key and secret key will be injected into
the process environment using the standard environment variables and a credentials file will be written to
the path from AWS_SHARED_CREDENTIALS_FILE (by default "/var/aws/credentials")`,
		"VAULT_AWS_PATH":  `Mountpoint for the vault AWS secret engine. Defaults to "aws".`,
		"VAULT_AUTH_PATH": "Authentication path for vault authentication - e.g. okta/login/:user. Overrides VAULT_AUTH_METHOD if set.",
		"VAULT_AUTH_DATA": "Data payload to send with authentication request. JSON object.",
		"VAULT_*":         "All vault client configuration environment variables are respected. More information at https://www.vaultproject.io/docs/commands/#environment-variables",
	}
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
