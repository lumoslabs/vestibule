package sops

import (
	"os"
)

// EnvVars is a map of known vonfiguration environment variables and their usage descriptions
var EnvVars = map[string]string{
	"SOPS_FILES": `If SOPS_FILES is set, will iterate over each file (colon separated), attempting to decrypt with Sops.
The decrypted cleartext file can be optionally written out to a separate location (with optional filemode)
or will be parsed into a map[string]string and injected into Environ
e.g. SOPS_FILES=/path/to/file[;/path/to/output[;mode]]:...`,
}

// Decoder is an environ.Provider which accepts a list of files encrypted with github.com/mozilla/sops
type Decoder struct {
	Files []*EncryptedFile `env:"SOPS_FILES" envSeparator:":"`
}

// EncryptedFile is a file that has been encrypted with github.com/mozilla/sops
type EncryptedFile struct {
	Path, Ext     string
	OutputPath    string
	OutputMode    os.FileMode
	UnmarshalFunc func([]byte, interface{}) error
}
