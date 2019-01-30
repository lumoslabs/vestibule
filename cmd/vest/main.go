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

type Config struct {
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
			fmt.Println(getVersion())
			os.Exit(0)
		}
	}
	if len(os.Args) <= 2 {
		log.Println(usage())
		os.Exit(1)
	}

	environ.RegisterProvider(dotenv.Name, dotenv.NewDotenvProvider)
	environ.RegisterProvider(ejson.Name, ejson.NewEjsonProvider)
	environ.RegisterProvider(vault.Name, vault.NewVaultProvider)
	environ.RegisterProvider(sops.Name, sops.NewSopsProvider)

	var (
		e  = environ.NewEnvironFromEnv()
		c  = new(Config)
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

	os.Unsetenv("HOME")
	e.Delete("HOME")

	if err := SetupUser(os.Args[1], e); err != nil {
		log.Fatalf("error: failed switching to %q: %v", os.Args[1], err)
	}

	name, err := exec.LookPath(os.Args[2])
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if err = syscall.Exec(name, os.Args[2:], e.Slice()); err != nil {
		log.Fatalf("error: exec failed: %v", err)
	}
}
