package sops

import (
	"os"

	"github.com/lumoslabs/vestibule/pkg/environ"
)

type SopsProvider struct {
	Files []*EncryptedFile `env:"SOPS_FILES" envSeparator:":"`
}

type EncryptedFile struct {
	Path, Ext     string
	OutputPath    string
	OutputMode    os.FileMode
	UnmarshalFunc func([]byte, interface{}) error
}

type EncryptedFileHandler func(*EncryptedFile, environ.Environ) error
