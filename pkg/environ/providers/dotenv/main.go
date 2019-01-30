package dotenv

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"github.com/caarlos0/env"
	"github.com/lumoslabs/vestibule/pkg/environ"
)

const (
	Name        = "dotenv"
	FilesEnvVar = "DOTENV_FILES"
)

func NewDotenvProvider() (environ.Provider, error) {
	defer func() { os.Unsetenv(FilesEnvVar) }()

	var de = &DotenvProvider{}
	if er := env.Parse(de); er != nil {
		return nil, er
	}
	if de.Files == nil || len(de.Files) == 0 {
		de.Files = findDotenvFiles()
	}
	return de, nil
}

func (de *DotenvProvider) AddToEnviron(e *environ.Environ) error {
	os.Unsetenv(FilesEnvVar)
	e.Delete(FilesEnvVar)

	em, er := godotenv.Read(de.Files...)
	if er != nil {
		return er
	}
	e.Append(em)
	return nil
}

func findDotenvFiles() []string {
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
		case filepath.Ext(path) == ".env":
			files = append(files, filepath.Join(cwd, path))
		}
		return nil
	})
	return files
}
