package sops

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"go.mozilla.org/sops"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"go.mozilla.org/sops/decrypt"
	yaml "gopkg.in/yaml.v2"

	"github.com/lumoslabs/vestibule/pkg/environ"
)

const (
	// Name is the environ.Provider name
	Name = "sops"

	// EncryptedFileSeparator is the separator between attributes of the encrypted files in FilesEnvVar
	EncryptedFileSeparator = ";"

	// DefaultOutputMode is the default FileMode of generated files
	DefaultOutputMode = os.FileMode(0700)

	//FilesEnvVar is the environment variable holding the list of encrypted files
	FilesEnvVar = "SOPS_FILES"
)

func init() {
	environ.RegisterProvider(Name, New)
}

// New returns a Decoder object as an environ.Environ or an error if configuring failed.
func New() (environ.Provider, error) {
	defer func() { os.Unsetenv(FilesEnvVar) }()
	var (
		d = &Decoder{}
		p = env.CustomParsers{reflect.TypeOf(EncryptedFile{}): encryptedFileParser}
	)
	return d, env.ParseWithFuncs(d, p)
}

// AddToEnviron uses go.mozilla.org/sops/decrypt to decrypt the file, then either unmarshals the result into
// a map[string]string and merges that into an environ.Environ object, or writes the cleartext out to the given
// output path if set
func (d *Decoder) AddToEnviron(e *environ.Environ) error {
	os.Unsetenv(FilesEnvVar)
	e.Delete(FilesEnvVar)
	for _, f := range d.Files {
		data, er := f.Decrypt()
		if er != nil {
			return er
		}

		if f.OutputPath != "" {
			if er := f.Write(data); er != nil {
				return er
			}
		} else {
			if env, er := f.Unmarshal(data); er == nil {
				envv := make(map[string]string, len(env))
				for k, v := range env {
					envv[strings.TrimRight(k, sops.DefaultUnencryptedSuffix)] = v
				}
				e.SafeMerge(envv)
			}
		}
	}
	return nil
}

// Decrypt uses go.mozilla.org/sops/decrypt to decrypt an encrypted file
func (ef *EncryptedFile) Decrypt() ([]byte, error) {
	return decrypt.File(ef.Path, ef.Ext)
}

// Unmarshal uses the configured unmarshal function to unmarshal a decrypted file
func (ef *EncryptedFile) Unmarshal(data []byte) (map[string]string, error) {
	env := make(map[string]string)
	if er := ef.UnmarshalFunc(data, &env); er != nil {
		return nil, er
	}
	return env, nil
}

// Write writes out cleartext to the configured output path
func (ef *EncryptedFile) Write(data []byte) error {
	mode := DefaultOutputMode
	if ef.OutputMode > 0 {
		mode = ef.OutputMode
	}

	return ioutil.WriteFile(ef.OutputPath, data, mode)
}

func encryptedFileParser(s string) (interface{}, error) {
	var (
		bits = strings.Split(s, EncryptedFileSeparator)
		ef   = EncryptedFile{Path: bits[0]}
	)

	switch len(bits) {
	case 3:
		if v, er := strconv.ParseUint(bits[2], 10, 32); er == nil {
			ef.OutputMode = os.FileMode(v)
		}
		fallthrough
	case 2:
		ef.OutputPath = bits[1]
	}

	switch filepath.Ext(s) {
	case ".yaml", ".yml":
		ef.Ext = "yaml"
		ef.UnmarshalFunc = yaml.Unmarshal
	case ".json":
		ef.Ext = "json"
		ef.UnmarshalFunc = json.Unmarshal
	case ".dotenv", ".env", ".sh", ".bash":
		ef.Ext = "dotenv"
		ef.UnmarshalFunc = envUmarshalFunc
	default:
		return nil, fmt.Errorf("Unknown file type: %s", s)
	}

	return &ef, nil
}

func envUmarshalFunc(d []byte, m interface{}) error {
	m, er := godotenv.Unmarshal(string(d))
	return er
}
