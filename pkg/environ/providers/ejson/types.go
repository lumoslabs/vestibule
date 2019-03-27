package ejson

// EnvVars is a map of known vonfiguration environment variables and their usage descriptions
var EnvVars = map[string]string{
	"EJSON_FILES": `If EJSON_FILES is set, will iterate over each file (colon separated), attempting to decrypt using keys
from EJSON_KEYS. If EJSON_FILES is not set, will look for any .ejson files in CWD. Cleartext decrypted
json will be parsed into a map[string]string and injected into Environ.
e.g. EJSON_FILES=/path/to/file1:/path/to/file2:...`,
	"EJSON_KEYS": `Colon separated list of public/private ejson keys. Public/private keys separated by semicolon.
e.g. EJSON_KEYS=pubkey1;privkey1:pubkey2;privkey2:...`,
}

// Decoder is an github.com/lumoslabs/vestibule/pkg/environ.Provider which accepts a list of ejson files and public/private key pairs.
// Using these and github.com/Shopify/ejson it decodes the files and adds them to a github.com/lumoslabs/vestibule/pkg/environ.Environ
type Decoder struct {
	Files    []string   `env:"EJSON_FILES" envSeparator:":"`
	KeyPairs KeyPairMap `env:"EJSON_KEYS"`
}

// KeyPairMap is a map[string]string that holds a map of ejson public / private key pairs
type KeyPairMap map[string]string
