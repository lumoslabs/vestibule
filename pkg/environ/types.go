package environ

import (
	"net/url"
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

// Key represents an environment variable key to be populated with secret data from a secrets provider
type Key struct {
	name string
	host string
	path string
	key  string
	data url.Values
}

// Provider is a secrets provider able to inject variables into the environment
type Provider interface {
  AddToEnviron(*Environ) error
  AddKeysToEnviron(*Environ) error
  AddKey(Key)
}

// ProviderFactory is a func that returns a new Provider
type ProviderFactory func() (Provider, error)

type unregisteredProviderError struct {
	provider string
}

type marshaller func(in interface{}) ([]byte, error)
