package environ

import (
	"regexp"
	"sync"
)

// Environ is a concurrency safe-ish map[string]string for holding environment variables
type Environ struct {
	sync.RWMutex
	m  map[string]string
	re *regexp.Regexp
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
