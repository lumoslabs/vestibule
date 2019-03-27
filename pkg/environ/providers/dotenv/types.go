package dotenv

// EnvVars is a map of known vonfiguration environment variables and their usage descriptions
var EnvVars = map[string]string{
	"DOTENV_FILES": `if DOTENV_FILES is set, will iterate over each file, parse and inject into Environ. If DOTENV_FILES is
not set, will look for any .env files in CWD.
e.g. DOTENV_FILES=/path/to/file1:/path/to/file2:...`,
}

// Parser is an github.com/lumoslabs/vestibule/pkg/environ.Provider which accepts a list of dotenv files and, using github.com/joho/godotenv,
// parses them and adds the result to an environ.Environ object
type Parser struct {
	Files []string `env:"DOTENV_FILES" envSeparator:":"`
}
