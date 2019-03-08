package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"

	"github.com/lumoslabs/vestibule/pkg/environ/providers/dotenv"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/ejson"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/sops"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/vault"

	"github.com/caarlos0/env"

	"github.com/lumoslabs/vestibule/pkg/environ"
)

type config struct {
	User      string   `env:"VEST_USER"`
	Providers []string `env:"VEST_PROVIDERS" envSeparator:"," envDefault:"vault"`
}

func init() {
	runtime.LockOSThread()
}

func main() {
	log.SetFlags(0) // no timestamps on our logs

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

	environ.RegisterProvider(dotenv.Name, dotenv.New)
	environ.RegisterProvider(ejson.Name, ejson.New)
	environ.RegisterProvider(vault.Name, vault.New)
	environ.RegisterProvider(sops.Name, sops.New)

	var (
		e  = environ.NewFromEnv()
		c  = new(config)
		wg sync.WaitGroup
	)

	env.Parse(c)
	for _, name := range c.Providers {
		p, er := environ.GetProvider(name)
		if er != nil {
			log.Println(er)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if er := p.AddToEnviron(e); er != nil {
				log.Println(er)
			}
		}()
	}
	wg.Wait()

	if name, er := exec.LookPath(os.Args[1]); er != nil {
		os.Unsetenv("HOME")
		e.Delete("HOME")

		usr := os.Args[1]
		if c.User != "" {
			usr = c.User
		}
		if er := SetupUser(usr, e); er != nil {
			log.Fatalf("error: failed switching to %q: %v", usr, er)
		}

		name, er = exec.LookPath(os.Args[2])
		if er != nil {
			log.Fatalf("error: %v", er)
		}

		if er = syscall.Exec(name, os.Args[2:], e.Slice()); er != nil {
			log.Fatalf("error: exec failed: %v", er)
		}
	} else {
		if c.User != "" {
			os.Unsetenv("HOME")
			e.Delete("HOME")

			if er := SetupUser(c.User, e); er != nil {
				log.Fatalf("error: failed switching to %q: %v", c.User, er)
			}
		}

		if er = syscall.Exec(name, os.Args[1:], e.Slice()); er != nil {
			log.Fatalf("error: exec failed: %v", er)
		}
	}
}
