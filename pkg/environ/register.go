package environ

import (
	"fmt"
)

var providers map[string]ProviderFactory

func RegisterProvider(name string, fn ProviderFactory) {
	if providers == nil {
		providers = make(map[string]ProviderFactory)
	}
	providers[name] = fn
}

func GetProvider(name string) (Provider, error) {
	fn, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("Unregistered provider %s", name)
	}

	return fn()
}
