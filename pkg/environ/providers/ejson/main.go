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

func init() {
	environ.RegisterProvider(EjsonProviderName, NewEjsonProvider)
}

func NewEjsonProvider() (environ.Provider, error) {
	var ep = new(EjsonProvider)
	p := env.CustomParsers{reflect.TypeOf(KeyPairMap{}): keyPairMapParser}
	if er := env.ParseWithFuncs(ep, p); er != nil {
		return nil, er
	}
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
	os.Unsetenv("EJSON_KEYS")
	for _, f := range ep.Files {
		data, er := ioutil.ReadFile(f)
		if er != nil {
			return er
		}

		pubkey, er := ejJson.ExtractPublicKey(data)
		if er != nil {
			return er
		}

		privkey, ok := ep.KeyPairs[string(pubkey[:])]
		if !ok {
			return fmt.Errorf("Unknown public key %v", string(pubkey[:]))
		}

		clear, er := ejson.DecryptFile(f, os.TempDir(), privkey)
		if er != nil {
			return er
		}

		env := make(map[string]string)
		if er := json.Unmarshal(clear, env); er != nil {
			return er
		}
		e.Append(env)
	}
	return nil
}
