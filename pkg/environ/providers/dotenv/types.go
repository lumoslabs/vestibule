package dotenv

// Parser is an github.com/lumoslabs/vestibule/pkg/environ.Provider which accepts a list of dotenv files and, using github.com/joho/godotenv,
// parses them and adds the result to an environ.Environ object
type Parser struct {
	Files []string `env:"DOTENV_FILES" envSeparator:":"`
}
