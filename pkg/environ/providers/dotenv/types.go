package dotenv

type DotenvProvider struct {
	Files []string `env:"DOTENV_FILES" envSeparator:":"`
}
