package environ

import (
	"fmt"
	"os"
	"strings"
)

func NewEnviron() *Environ {
	return &Environ{m: make(map[string]string)}
}

func NewEnvironFromEnv() *Environ {
	e := make(map[string]string)
	for _, item := range os.Environ() {
		bits := strings.Split(item, "=")
		if len(bits) == 2 {
			e[bits[0]] = bits[1]
		}
	}
	return &Environ{m: e}
}

func (e *Environ) Append(m map[string]string) {
	e.Lock()
	for k, v := range m {
		e.m[k] = v
	}
	e.Unlock()
}

func (e *Environ) Slice() []string {
	var s = make([]string, 0)
	e.RLock()
	for k, v := range e.m {
		s = append(s, k+"="+v)
	}
	e.RUnlock()
	return s
}

func (e *Environ) String() string {
	return fmt.Sprintf("%#q", e.Slice())
}
