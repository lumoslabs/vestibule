package ejson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/Shopify/ejson"
	ejJson "github.com/Shopify/ejson/json"
	"github.com/caarlos0/env"
	"github.com/lumoslabs/vestibule/pkg/environ"
)

const (
	EjsonProviderName   = "ejson"
	KeyPairEnvSeparator = ":"
	KeyPairSeparator    = ";"
)

func NewEjsonProvider() (environ.Provider, error) {
	var ep = new(EjsonProvider)
	p := env.CustomParsers{reflect.TypeOf(KeyPairMap{}): keyPairMapParser}
	if er := env.ParseWithFuncs(ep, p); er != nil {
		return nil, er
	}
	os.Unsetenv("EJSON_KEYS")
	return ep, nil
}

func keyPairMapParser(s string) (interface{}, error) {
	var (
		kpm   = make(map[string]string)
		items = strings.Split(s, KeyPairEnvSeparator)
	)

	for _, item := range items {
		bits := strings.Split(item, KeyPairSeparator)
		if len(bits) == 2 {
			kpm[bits[0]] = bits[1]
		}
	}
	return KeyPairMap(kpm), nil
}

func (ep *EjsonProvider) AddToEnviron(e *environ.Environ) error {
	e.Delete("EJSON_FILES")
	e.Delete("EJSON_KEYS")
	for _, f := range ep.Files {
		privkey, er := matchPrivateKey(f, ep.KeyPairs)
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
		e.Append(envv)
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
