package vault

import (
	"errors"

	"github.com/hashicorp/vault/api"
)

const (
	EnvAwsAccessKeyId        = "AWS_ACCESS_KEY_ID"
	EnvAwsProfile            = "AWS_PROFILE"
	EnvAwsSecretAccessKey    = "AWS_SECRET_ACCESS_KEY"
	EnvAwsSessionToken       = "AWS_SESSION_TOKEN"
	EnvAwsSharedCredFile     = "AWS_SHARED_CREDENTIALS_FILE"
	EnvGoogleCredFile        = "GOOGLE_CREDENTIALS_FILE"
	EnvGoogleToken           = "GCP_TOKEN"
	EnvKubernetesServiceHost = "KUBERNETES_SERVICE_HOST"
	EnvKubernetesServicePort = "KUBERNETES_SERVICE_PORT"
	EnvVaultAppJWT           = "VAULT_APP_JWT"
	EnvVaultAppRole          = "VAULT_APP_ROLE"
	EnvVaultAppSecret        = "VAULT_APP_SECRET"
	EnvVaultAuthData         = "VAULT_AUTH_DATA"
	EnvVaultAuthMethod       = "VAULT_AUTH_METHOD"
	EnvVaultAuthPath         = "VAULT_AUTH_PATH"
	EnvVaultAwsPath          = "VAULT_AWS_PATH"
	EnvVaultAwsRole          = "VAULT_AWS_ROLE"
	EnvVaultIamRole          = "VAULT_IAM_ROLE"
	EnvVaultGcpCredType      = "VAULT_GCP_CRED_TYPE"
	EnvVaultGcpPath          = "VAULT_GCP_PATH"
	EnvVaultGcpRole          = "VAULT_GCP_ROLE"
	EnvVaultKeys             = "VAULT_KV_KEYS"
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
		EnvVaultAwsRole: `Name of the aws role to generate credentials against. If credentials are returned, the access key and secret key will be injected into
the process environment using the standard environment variables and a credentials file will be written to
the path from AWS_SHARED_CREDENTIALS_FILE (by default "/var/run/aws/credentials")`,
		"VAULT_*":            "All vault client configuration environment variables are respected. More information at https://www.vaultproject.io/docs/commands/#environment-variables",
		EnvVaultIamRole:      "[DEPRECATED] Name of the aws role to generate credentials against.",
		EnvAwsProfile:        `AWS profile to use in the shared credentials file. Defaults to "default"`,
		EnvAwsSharedCredFile: `Path to the AWS shared credentials file to write credentials to. Defaults to "/var/run/aws/credentials"`,
		EnvGoogleCredFile:    `Path to the GCP service account credentials file to create. Defaults to "/var/run/gcp/creds.json"`,
		EnvVaultAppJWT:       "The jwt for use with OIDC/JWT authentication",
		EnvVaultAppRole:      "Either the role id for AppRole authentication, or the role name fo Kubernetes authentication.",
		EnvVaultAppSecret:    "The secret id for use with AppRole authentication",
		EnvVaultAuthData:     "Data payload to send with authentication request. JSON object.",
		EnvVaultAuthMethod:   `Authentication method for vault. Default is "kubernetes".`,
		EnvVaultAuthPath:     "Authentication path for vault authentication - e.g. okta/login/:user. Overrides VAULT_AUTH_METHOD if set.",
		EnvVaultAwsPath:      `Mountpoint for the vault AWS secret engine. Defaults to "aws".`,
		EnvVaultGcpCredType:  "GCP credential type to generate. Defaults to key. Accepted values are [token key]",
		EnvVaultGcpPath:      `Mountpoint for the vault GCP secret engine. Defaults to "gcp".`,
		EnvVaultGcpRole:      "Name of the GCP role in vault to generate credentials against.",
	}
)

// Client is an environ.Provider and github.com/hashicorp/vault/api.Client which will get the requested keys
type Client struct {
	*api.Client
	AuthMethod  string              `env:"VAULT_AUTH_METHOD"`
	AuthPath    string              `env:"VAULT_AUTH_PATH"`
	AuthData    *RedactableAuthData `env:"VAULT_AUTH_DATA"`
	AppRole     string              `env:"VAULT_APP_ROLE"`
	AppSecret   string              `env:"VAULT_APP_SECRET"`
	AppJWT      string              `env:"VAULT_APP_JWT"`
	AwsRole     string              `env:"VAULT_AWS_ROLE"`
	IamRole     string              `env:"VAULT_IAM_ROLE"`
	AwsPath     string              `env:"VAULT_AWS_PATH" envDefault:"aws"`
	AwsCredFile string              `env:"AWS_SHARED_CREDENTIALS_FILE" envDefault:"/var/run/aws/credentials"`
	AwsProfile  string              `env:"AWS_PROFILE" envDefault:"default"`
	GcpPath     string              `env:"VAULT_GCP_PATH" envDefault:"gcp"`
	GcpRole     string              `env:"VAULT_GCP_ROLE"`
	GcpCredType string              `env:"VAULT_GCP_CRED_TYPE" envDefault:"key"`
	GcpCredFile string              `env:"GOOGLE_CREDENTIALS_FILE" envDefault:"/var/run/gcp/creds.json"`
	Keys        []KVKey             `env:"VAULT_KV_KEYS" envSeparator:":"`
}

// KVKeys is an alias for []*KVKey. Needed for caarlos0/env to support parsing.
type KVKeys []KVKey

// KVKey is a kv ver2 key in Vault
type KVKey struct {
	Path    string
	Version *int
}

type EnvKey struct {
	name string
	host string
	path string
	key  string
	data map[string][]string
}

type RedactableAuthData struct {
	data map[string]string
}
