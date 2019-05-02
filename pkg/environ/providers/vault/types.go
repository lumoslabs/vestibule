package vault

import (
	"errors"

	"github.com/hashicorp/vault/api"
)

const (
	EnvVaultAuthMethod       = "VAULT_AUTH_METHOD"
	EnvVaultAuthPath         = "VAULT_AUTH_PATH"
	EnvVaultAuthData         = "VAULT_AUTH_DATA"
	EnvVaultAppRole          = "VAULT_APP_ROLE"
	EnvVaultAppSecret        = "VAULT_APP_SECRET"
	EnvVaultAppJWT           = "VAULT_APP_JWT"
	EnvVaultIAMRole          = "VAULT_IAM_ROLE"
	EnvVaultAWSPath          = "VAULT_AWS_PATH"
	EnvVaultKeys             = "VAULT_KV_KEYS"
	EnvKubernetesServiceHost = "KUBERNETES_SERVICE_HOST"
	EnvKubernetesServicePort = "KUBERNETES_SERVICE_PORT"
)

var (
	// ErrVaultEmptyResponse is returned when vault respondes with no data
	ErrVaultEmptyResponse = errors.New("no data returned from vault")
	// ErrVaultUnexpectedResponse is returned when vault does not respond with the expected data
	ErrVaultUnexpectedResponse = errors.New("unexpected response from vault")
	// ErrNotInKubernetes is returned when vestibule is not running in a kubernetes cluster
	ErrNotInKubernetes = errors.New("not running in kubernetes cluster")
	// ErrInvalidKVKey is returned when the given key is invalid
	ErrInvalidKVKey = errors.New("invalid vault KV key")
	// ErrUnexpectedVaultResponse is returned when vault returns something we cannot handle
	ErrUnexpectedVaultResponse = errors.New("unexpected response from vault")

	sensitiveEnvVars = []string{
		EnvVaultKeys,
		EnvVaultAuthData,
		EnvVaultAppSecret,
		EnvVaultAppRole,
		EnvVaultAppJWT,
	}

	// EnvVars is a map of known vonfiguration environment variables and their usage descriptions
	EnvVars = map[string]string{
		EnvVaultKeys: `If VAULT_KV_KEYS is set, will iterate over each key (colon separated), attempting to get the secret from Vault.
Secrets are pulled at the optional version or latest, then injected into Environ. If running in Kubernetes,
the Pod's ServiceAccount token will automatically be looked up and used for Vault authentication.
e.g. VAULT_KV_KEYS=/path/to/key1[@version]:/path/to/key2[@version]:...`,
		EnvVaultAuthMethod: `Authentication method for vault. Default is "kubernetes".`,
		EnvVaultAppRole:    "Either the role id for AppRole authentication, or the role name fo Kubernetes authentication.",
		EnvVaultAppSecret:  "The secret id for use with AppRole authentication",
		EnvVaultAppJWT:     "The jwt for use with OIDC/JWT authentication",
		EnvVaultIAMRole: `IAM role to request from vault. If returns credentials, the access key and secret key will be injected into
the process environment using the standard environment variables and a credentials file will be written to
the path from AWS_SHARED_CREDENTIALS_FILE (by default "/var/aws/credentials")`,
		EnvVaultAWSPath:  `Mountpoint for the vault AWS secret engine. Defaults to "aws".`,
		EnvVaultAuthPath: "Authentication path for vault authentication - e.g. okta/login/:user. Overrides VAULT_AUTH_METHOD if set.",
		EnvVaultAuthData: "Data payload to send with authentication request. JSON object.",
		"VAULT_*":        "All vault client configuration environment variables are respected. More information at https://www.vaultproject.io/docs/commands/#environment-variables",
	}
)

// Client is an environ.Provider and github.com/hashicorp/vault/api.Client which will get the requested keys
type Client struct {
	*api.Client
	AuthMethod  string `env:"VAULT_AUTH_METHOD"`
	AuthPath    string `env:"VAULT_AUTH_PATH"`
	AuthData    string `env:"VAULT_AUTH_DATA"`
	AppRole     string `env:"VAULT_APP_ROLE"`
	AppSecret   string `env:"VAULT_APP_SECRET"`
	AppJWT      string `env:"VAULT_APP_JWT"`
	IamRole     string `env:"VAULT_IAM_ROLE"`
	AwsPath     string `env:"VAULT_AWS_PATH" envDefault:"aws"`
	AwsCredFile string `env:"AWS_SHARED_CREDENTIALS_FILE" envDefault:"/var/run/aws/credentials"`
	Keys        KVKeys `env:"VAULT_KV_KEYS"`
}

// KVKeys is an alias for []*KVKey. Needed for caarlos0/env to support parsing.
type KVKeys []*KVKey

// KVKey is a kv ver2 key in Vault
type KVKey struct {
	Path    string
	Version *int
}
