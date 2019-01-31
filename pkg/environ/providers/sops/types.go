package sops

import (
	"os"
)

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
