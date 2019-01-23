package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
)

const (
	vaultKeySeparator = "@"
	sopsKeySeparator  = ";"
)

type envv []string

type config struct {
	*vaultConfig
	*sopsConfig
}

type vaultConfig struct {
	Keys []vaultKey `env:"VAULT_KEYS" envSeparator:":"`
}

type sopsConfig struct {
	Files []sopsFile `env:"SOPS_FILES" envSeparator:":"`
}

type vaultKey struct {
	Path    string
	Version int
}

type sopsFile struct {
	EncryptedPath string
	DecryptedPath string
	DecryptedMode os.FileMode
}

func (e *envv) Add(m map[string]string) {
	for k, v := range m {
		*e = append(*e, fmt.Sprintf("%s=%s", k, v))
	}
}

func vaultKeyParser(s string) (interface{}, error) {
	var (
		bits = strings.Split(s, vaultKeySeparator)
		k    = vaultKey{Path: bits[0]}
	)

	if len(bits) == 1 {
		return k, nil
	}

	v, er := strconv.Atoi(bits[1])
	if er != nil {
		return nil, er
	}

	k.Version = v
	return k, nil
}

func sopsFileParser(s string) (interface{}, error) {
	var (
		bits = strings.Split(s, sopsKeySeparator)
		k    = sopsFile{EncryptedPath: bits[0]}
	)

	if len(bits) == 2 {
		k.DecryptedPath = bits[1]
	}
	if len(bits) == 3 {
		if v, er := strconv.ParseUint(bits[2], 10, 32); er == nil {
			k.DecryptedMode = os.FileMode(v)
		}
	}

	return k, nil
}

func newConfig() (*config, error) {
	var (
		c      = config{&vaultConfig{}, &sopsConfig{}}
		parser = env.CustomParsers{
			reflect.TypeOf(vaultKey{}): vaultKeyParser,
			reflect.TypeOf(sopsFile{}): sopsFileParser,
		}
	)

	if er := env.ParseWithFuncs(&c, parser); er != nil {
		return nil, er
	}
	return &c, nil
}
