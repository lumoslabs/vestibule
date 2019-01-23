package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/joho/godotenv"
	"go.mozilla.org/sops/decrypt"
)

var handlers = make(map[string]sopsFileHandler)

type sopsFileHandler func(*sopsFile, *envv) error

func init() {
	handlers[".yaml"] = decryptYAML
	handlers[".yml"] = decryptYAML
	handlers[".json"] = decryptJSON
	handlers[".env"] = decryptEnv
	handlers[".sh"] = decryptEnv
}

func getHandler(s string) sopsFileHandler {
	if h, ok := handlers[filepath.Ext(s)]; ok {
		return h
	}
	return decryptBin
}

func sopsDecrypt(c *sopsConfig, e *envv) error {
	for _, f := range c.Files {
		h := getHandler(f.EncryptedPath)
		if er := h(&f, e); er != nil {
			return er
		}
	}
	return nil
}

func decryptJSON(f *sopsFile, e *envv) error {
	data, er := decrypt.File(f.EncryptedPath, "json")
	if er != nil {
		return er
	}

	if f.DecryptedPath != "" {
		return writeClearTextFile(f, data)
	}

	var env map[string]string
	if er := yaml.Unmarshal(data, env); er != nil {
		return er
	}

	e.Add(env)
	return nil
}

func decryptYAML(f *sopsFile, e *envv) error {
	data, er := decrypt.File(f.EncryptedPath, "yaml")
	if er != nil {
		return er
	}

	if f.DecryptedPath != "" {
		return writeClearTextFile(f, data)
	}

	var env map[string]string
	if er := yaml.Unmarshal(data, env); er != nil {
		return er
	}

	e.Add(env)
	return nil
}

func decryptEnv(f *sopsFile, e *envv) error {
	data, er := decrypt.File(f.EncryptedPath, "dotenv")
	if er != nil {
		return er
	}

	if f.DecryptedPath != "" {
		return writeClearTextFile(f, data)
	}

	var env map[string]string
	if env, er := godotenv.Unmarshal(string(data)); er == nil {
		e.Add(env)
	}

	e.Add(env)
	return nil
}

func decryptBin(f *sopsFile, e *envv) error {
	data, er := decrypt.File(f.EncryptedPath, "")
	if er != nil {
		return er
	}
	return writeClearTextFile(f, data)
}

func writeClearTextFile(f *sopsFile, data []byte) error {
	var (
		out  = f.DecryptedPath
		mode = f.DecryptedMode
	)

	if out == "" {
		out = f.EncryptedPath
	}
	if mode == 0 {
		mode = os.FileMode(0700)
	}

	return ioutil.WriteFile(out, data, mode)
}
