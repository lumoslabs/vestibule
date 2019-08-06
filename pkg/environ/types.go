package environ

import (
	"regexp"
	"sync"
)

// Environ is a concurrency safe-ish map[string]string for holding environment variables
type Environ struct {
	sync.RWMutex
	m          map[string]string
	re         *regexp.Regexp
	marshaller marshaller

	// UpcaseKeys will force all Environ keys to be upcased
	UpcaseKeys bool
}

// Options contains flags for dealing with adding new keys to Environ
type Options struct {
	// Overwrite will allow new keys to overwrite existing keys if true
	Overwrite bool
}

// Provider is a secrets provider able to inject variables into the environment
type Provider interface {
	AddToEnviron(*Environ) error
}

// ProviderFactory is a func that returns a new Provider
type ProviderFactory func() (Provider, error)

type unregisteredProviderError struct {
	provider string
}

type marshaller func(in interface{}) ([]byte, error)
