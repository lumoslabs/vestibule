package ejson

type EjsonProvider struct {
	Files    []string   `env:"EJSON_FILES" envSeparator:":"`
	KeyPairs KeyPairMap `env:"EJSON_KEYS"`
}

type KeyPairMap map[string]string
