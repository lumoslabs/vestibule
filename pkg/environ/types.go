package environ

import "sync"

type Environ struct {
	sync.RWMutex
	m map[string]string
}

type Provider interface {
	AddToEnviron(*Environ) error
}

type ProviderFactory func() (Provider, error)
