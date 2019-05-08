package ejson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Shopify/ejson"
	ejJson "github.com/Shopify/ejson/json"
	env "github.com/caarlos0/env/v5"
	"github.com/lumoslabs/vestibule/pkg/environ"
)

const (
	// Name is the Provider name
	Name = "ejson"

	// KeyPairEnvSeparator is the separator between key pairs in the KeysEnvVar
	KeyPairEnvSeparator = ":"

	// KeyPairSeparator is the separator betwen a public and private key in the KeysEnvVar
	KeyPairSeparator = ";"

	// FilesEnvVar is the environment variable holding the ejson file list
	FilesEnvVar = "EJSON_FILES"

	// KeysEnvVar is the environment variable holding the ejson public / private key pairs
	KeysEnvVar = "EJSON_KEYS"
)

func init() {
	environ.RegisterProvider(Name, New)
}

// New returns a new Decoder instance or an error if configuring failed.
// If no ejson files are listed in FilesEnvVar, then will search in CWD for .ejson files
func New() (environ.Provider, error) {
	defer func() {
		os.Unsetenv(FilesEnvVar)
		os.Unsetenv(KeysEnvVar)
	}()

	var d = new(Decoder)
	p := env.CustomParsers{reflect.TypeOf(KeyPairMap{}): keyPairMapParser}
	if er := env.ParseWithFuncs(d, p); er != nil {
		return nil, er
	}
	if d.Files == nil || len(d.Files) == 0 {
		d.Files = findEjsonFiles()
	}
	return d, nil
}

func keyPairMapParser(s string) (interface{}, error) {
	var (
		kpm   = make(map[string]string)
		items = strings.Split(s, KeyPairEnvSeparator)
	)

	for _, item := range items {
		bits := strings.SplitN(item, KeyPairSeparator, 2)
		if len(bits) == 2 {
			kpm[bits[0]] = bits[1]
		}
	}
	return KeyPairMap(kpm), nil
}

// AddToEnviron uses github.com/Shopify/ejson to decrypt the given ejson files using the provided key pairs. Cleartext
// file data is decoded into map[string]string objects and merged with the provided environ.Environ
func (d *Decoder) AddToEnviron(e *environ.Environ) error {
	e.Delete(FilesEnvVar)
	e.Delete(KeysEnvVar)
	for _, f := range d.Files {
		privkey, er := matchPrivateKey(f, d.KeyPairs)
		if er != nil {
			return er
		}

		clear, er := ejson.DecryptFile(f, os.TempDir(), privkey)
		if er != nil {
			return er
		}

		env := make(map[string]string)
		if er := json.Unmarshal(clear, &env); er != nil {
			return er
		}
		delete(env, ejJson.PublicKeyField)
		envv := make(map[string]string, len(env))
		for k, v := range env {
			envv[strings.TrimLeft(k, "_")] = v
		}
		e.SafeMerge(envv)
	}
	return nil
}

func matchPrivateKey(path string, kpm KeyPairMap) (string, error) {
	data, er := ioutil.ReadFile(path)
	if er != nil {
		return "", er
	}

	doc := make(map[string]string)
	if er := json.Unmarshal(data, &doc); er != nil {
		return "", er
	}

	pubkey := doc[ejJson.PublicKeyField]
	privkey, ok := kpm[pubkey]
	if !ok {
		return "", fmt.Errorf("Unknown public key %s", pubkey)
	}
	return privkey, nil
}

func findEjsonFiles() []string {
	files := make([]string, 0)

	cwd, er := os.Getwd()
	if er != nil {
		return files
	}

	filepath.Walk(cwd, func(path string, info os.FileInfo, er error) error {
		switch {
		case er != nil:
			return er
		case path != cwd && info.IsDir():
			return filepath.SkipDir
		case filepath.Ext(path) == ".ejson":
			files = append(files, filepath.Join(cwd, path))
		}
		return nil
	})
	return files
}
