package environ

import (
	"fmt"

	"github.com/lumoslabs/vestibule/pkg/log"
)

var providers map[string]ProviderFactory

// RegisterProvider adds the named Provider's factory function to the map of known Providers
func RegisterProvider(name string, fn ProviderFactory) {
	log.Debugf("Registering provider. name=%s", name)
	if providers == nil {
		providers = make(map[string]ProviderFactory)
	}
	providers[name] = fn
}

// GetProvider returns a new instance of the named Provider or an unregistered provider error
func GetProvider(name string) (Provider, error) {
	fn, ok := providers[name]
	if !ok {
		return nil, newUnregisteredProviderError(name)
	}

	return fn()
}

func newUnregisteredProviderError(name string) *unregisteredProviderError {
	return &unregisteredProviderError{name}
}

func (e *unregisteredProviderError) Error() string {
	return fmt.Sprintf("unregistered provider %s", e.provider)
}
