package dotenv

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	env "github.com/caarlos0/env/v5"
	"github.com/lumoslabs/vestibule/pkg/environ"
)

const (
	// Name is the name of this Provider
	Name = "dotenv"

	// FilesEnvVar is the environment variable which lists the files to load
	FilesEnvVar = "DOTENV_FILES"
)

func init() {
	environ.RegisterProvider(Name, New)
}

// New returns a new Parser as an environ.Provider or an error if configuring failed
// if no files are listed, will search in CWD for any .env files.
func New() (environ.Provider, error) {
	defer func() { os.Unsetenv(FilesEnvVar) }()

	var de = &Parser{}
	if er := env.Parse(de); er != nil {
		return nil, er
	}
	if de.Files == nil || len(de.Files) == 0 {
		de.Files = findDotenvFiles()
	}
	return de, nil
}

// AddToEnviron uses godotenv to read all specified dotenv files and add them to the environ.Environ without overwriting
func (p *Parser) AddToEnviron(e *environ.Environ) error {
	os.Unsetenv(FilesEnvVar)
	e.Delete(FilesEnvVar)

	em, er := godotenv.Read(p.Files...)
	if er != nil {
		return er
	}
	e.SafeMerge(em)
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
