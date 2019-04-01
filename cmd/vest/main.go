package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/lumoslabs/vestibule/pkg/environ/providers/dotenv"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/ejson"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/sops"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/vault"
	"github.com/opencontainers/runc/libcontainer/user"

	"github.com/caarlos0/env"

	"github.com/lumoslabs/vestibule/pkg/environ"
	logger "github.com/lumoslabs/vestibule/pkg/log"
)

var (
	envVars = map[string]string{
		"VEST_USER": `The user [and group] to run the command as. Overrides commandline if set.
e.g. VEST_USER=user[:group]`,
		"VEST_PROVIDERS": fmt.Sprintf(`Comma separated list of enabled providers. By default only Vault is enabled.
Available providers: %v`, secretProviders),
		"VEST_DEBUG":            "Enable debug logging.",
		"VEST_VERBOSE":          "Enable verbose logging.",
		"VEST_UPCASE_VAR_NAMES": "Upcase environment variable names gathered from secret providers. Default: true",
	}

	secretProviders = []string{
		dotenv.Name,
		ejson.Name,
		vault.Name,
		sops.Name,
	}
	secretProviderEnvVars = []map[string]string{
		envVars,
		vault.EnvVars,
		dotenv.EnvVars,
		ejson.EnvVars,
		sops.EnvVars,
	}
)

type config struct {
	User       string   `env:"VEST_USER"`
	Providers  []string `env:"VEST_PROVIDERS" envSeparator:"," envDefault:"vault"`
	Debug      bool     `env:"VEST_DEBUG"`
	Verbose    bool     `env:"VEST_VERBOSE"`
	UpcaseVars bool     `env:"VEST_UPCASE_VAR_NAMES" envDefault:"true"`
}

func init() {
	runtime.LockOSThread()
}

func main() {
	logLevel := "disabled"

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "--help", "-h", "-?":
			fmt.Println(usage())
			os.Exit(0)
		case "--version", "-v":
			fmt.Println(appVersion())
			os.Exit(0)
		}
	}

	conf := new(config)
	env.Parse(conf)
	if conf.Verbose {
		logLevel = "info"
	}
	if conf.Debug {
		logLevel = "debug"
	}

	log := newLogger(logLevel, os.Stderr)
	logger.SetLogger(log)
	if conf.Debug {
		log.Debugf("Config: %#v", conf)
	}

	secrets := environ.New()
	secrets.UpcaseKeys = conf.UpcaseVars
	secrets.Populate(conf.Providers)

	if name, er := exec.LookPath(os.Args[1]); er != nil {
		os.Unsetenv("HOME")
		secrets.Delete("HOME")

		u := os.Args[1]
		if conf.User != "" {
			u = conf.User
		}

		usr, er := getUser(u)
		if er != nil {
			log.Infof("error: unable to find %q: %v", u, er)
			os.Exit(1)
		}

		name, er = exec.LookPath(os.Args[2])
		if er != nil {
			log.Infof("error: %v", er)
			os.Exit(1)
		}

		if er := SetupUser(usr); er != nil {
			log.Infof("error: failed switching to %q: %v", u, er)
			os.Exit(1)
		}

		secrets.SafeAppend(os.Environ())
		if er = syscall.Exec(name, os.Args[2:], secrets.Slice()); er != nil {
			log.Infof("error: exec failed: %v", er)
			os.Exit(1)
		}
	} else {
		if conf.User != "" {
			os.Unsetenv("HOME")
			secrets.Delete("HOME")

			usr, er := getUser(conf.User)
			if er != nil {
				log.Infof("error: unable to find %q: %v", conf.User, er)
				os.Exit(1)
			}

			if er := SetupUser(usr); er != nil {
				log.Infof("error: failed switching to %q: %v", conf.User, er)
				os.Exit(1)
			}
		}

		secrets.SafeAppend(os.Environ())
		if er = syscall.Exec(name, os.Args[1:], secrets.Slice()); er != nil {
			log.Infof("error: exec failed: %v", er)
			os.Exit(1)
		}
	}
}

func getUser(usr string) (*user.ExecUser, error) {
	defaultExecUser := user.ExecUser{
		Uid:  syscall.Getuid(),
		Gid:  syscall.Getgid(),
		Home: "/",
	}
	passwdPath, err := user.GetPasswdPath()
	if err != nil {
		return nil, err
	}
	groupPath, err := user.GetGroupPath()
	if err != nil {
		return nil, err
	}

	return user.GetExecUserPath(usr, &defaultExecUser, passwdPath, groupPath)
}
