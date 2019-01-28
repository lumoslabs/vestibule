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

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"go.mozilla.org/sops/decrypt"
	yaml "gopkg.in/yaml.v2"

	"github.com/lumoslabs/vestibule/pkg/environ"
)

const (
	SopsProviderName       = "sops"
	EncryptedFileSeparator = ";"
	DefaultOutputMode      = os.FileMode(0700)
)

func init() {
	environ.RegisterProvider(SopsProviderName, NewSopsProvider)
}

func NewSopsProvider() (environ.Provider, error) {
	var (
		sp = &SopsProvider{}
		p  = env.CustomParsers{reflect.TypeOf(EncryptedFile{}): encryptedFileParser}
	)
	return sp, env.ParseWithFuncs(sp, p)
}

func (sp *SopsProvider) AddToEnviron(e *environ.Environ) error {
	for _, f := range sp.Files {
		data, er := f.Decrypt()
		if er != nil {
			return er
		}

		if f.OutputPath != "" {
			if er := f.Write(data); er != nil {
				return er
			}
		} else {
			if env, er := f.Decode(data); er == nil {
				e.Append(env)
			}
		}
	}
	return nil
}

func (ef *EncryptedFile) Decrypt() ([]byte, error) {
	return decrypt.File(ef.Path, ef.Ext)
}

func (ef *EncryptedFile) Decode(data []byte) (map[string]string, error) {
	env := make(map[string]string)
	if er := ef.UnmarshalFunc(data, env); er != nil {
		return nil, er
	}
	return env, nil
}

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
