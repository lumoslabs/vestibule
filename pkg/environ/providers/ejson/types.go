package ejson

// Decoder is an github.com/lumoslabs/vestibule/pkg/environ.Provider which accepts a list of ejson files and public/private key pairs.
// Using these and github.com/Shopify/ejson it decodes the files and adds them to a github.com/lumoslabs/vestibule/pkg/environ.Environ
type Decoder struct {
	Files    []string   `env:"EJSON_FILES" envSeparator:":"`
	KeyPairs KeyPairMap `env:"EJSON_KEYS"`
}

// KeyPairMap is a map[string]string that holds a map of ejson public / private key pairs
type KeyPairMap map[string]string
