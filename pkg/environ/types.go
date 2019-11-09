package environ

import (
	"errors"
	"net/url"
	"regexp"
	"sync"
)

var ErrUnknownProvider = errors.New("unknown secrets provider")

// Environ is a concurrency safe-ish map[string]string for holding environment variables
type Environ struct {
	mu         sync.RWMutex
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

// RawKey represents an environment variable key to be populated with secret data from a secrets provider
type RawKey struct {
	Name string
	URL  *url.URL
}



// Provider is a secrets provider able to inject variables into the environment
type Provider interface {
	AddToEnviron(*Environ) error
	AddKeysToEnviron(*Environ) error
	AddKey(RawKey)
}

// ProviderFactory is a func that returns a new Provider
type ProviderFactory func() (Provider, error)

type unregisteredProviderError struct {
	provider string
}

type marshaller func(in interface{}) ([]byte, error)
