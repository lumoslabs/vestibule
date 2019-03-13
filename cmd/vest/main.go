package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"

	"github.com/lumoslabs/vestibule/pkg/environ/providers/dotenv"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/ejson"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/sops"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/vault"
	"github.com/opencontainers/runc/libcontainer/user"

	"github.com/caarlos0/env"

	"github.com/lumoslabs/vestibule/pkg/environ"
)

var log environ.Logger

type config struct {
	User      string   `env:"VEST_USER"`
	Providers []string `env:"VEST_PROVIDERS" envSeparator:"," envDefault:"vault"`
	OutFile   string   `env:"VEST_OUTPUT_FILE" envExpand:"true"`
	OutFmt    string   `env:"VEST_OUTPUT_FORMAT" envDefault:"json"`
	Debug     bool     `env:"VEST_DEBUG"`
	Verbose   bool     `env:"VEST_VERBOSE"`
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

	var (
		e  = environ.New()
		c  = new(config)
		wg sync.WaitGroup
	)

	env.Parse(c)
	if c.Verbose {
		logLevel = "info"
	}
	if c.Debug {
		logLevel = "debug"
	}

	log = environ.NewLogger(logLevel, os.Stderr)
	environ.SetLogger(log)
	environ.RegisterProvider(dotenv.Name, dotenv.New)
	environ.RegisterProvider(ejson.Name, ejson.New)
	environ.RegisterProvider(vault.Name, vault.New)
	environ.RegisterProvider(sops.Name, sops.New)

	for _, name := range c.Providers {
		p, er := environ.GetProvider(name)
		if er != nil {
			log.Infof("Skipping provider: %v", er)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if er := p.AddToEnviron(e); er != nil {
				log.Infof("Failed add secrets to Environ. provider=%s msg=%s", name, er.Error())
			}
		}()
	}
	wg.Wait()

	if c.OutFile != "" {
		if file, er := os.Create(c.OutFile); er == nil {
			e.SetMarshaller(c.OutFmt)
			e.Write(file)
			file.Close()
		} else {
			log.Infof("Failed to write secrets to file. file=%s msg=%s", c.OutFile, er.Error())
		}
	}

	if name, er := exec.LookPath(os.Args[1]); er != nil {
		os.Unsetenv("HOME")
		e.Delete("HOME")

		u := os.Args[1]
		if c.User != "" {
			u = c.User
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

		if c.OutFile != "" {
			os.Chown(c.OutFile, usr.Uid, usr.Gid)
		}

		if er := SetupUser(usr); er != nil {
			log.Infof("error: failed switching to %q: %v", u, er)
			os.Exit(1)
		}

		e.SafeAppend(os.Environ())
		if er = syscall.Exec(name, os.Args[2:], e.Slice()); er != nil {
			log.Infof("error: exec failed: %v", er)
			os.Exit(1)
		}
	} else {
		if c.User != "" {
			os.Unsetenv("HOME")
			e.Delete("HOME")

			usr, er := getUser(c.User)
			if er != nil {
				log.Infof("error: unable to find %q: %v", c.User, er)
				os.Exit(1)
			}

			if c.OutFile != "" {
				os.Chown(c.OutFile, usr.Uid, usr.Gid)
			}

			if er := SetupUser(usr); er != nil {
				log.Infof("error: failed switching to %q: %v", c.User, er)
				os.Exit(1)
			}
		}

		e.SafeAppend(os.Environ())
		if er = syscall.Exec(name, os.Args[1:], e.Slice()); er != nil {
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
