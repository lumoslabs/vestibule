package environ

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func NewEnviron() *Environ {
	return &Environ{m: make(map[string]string)}
}

func NewEnvironFromEnv() *Environ {
	e := make(map[string]string)
	for _, item := range os.Environ() {
		bits := strings.SplitN(item, "=", 2)
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

func (e *Environ) Set(k, v string) {
	e.Lock()
	defer e.Unlock()
	e.m[k] = v
}

func (e *Environ) Load(k string) (v string, ok bool) {
	e.RLock()
	defer e.RUnlock()
	v, ok = e.m[k]
	return
}

func (e *Environ) Delete(key string) {
	e.Lock()
	delete(e.m, key)
	e.Unlock()
}

func (e *Environ) Slice() []string {
	var s = make([]string, 0)
	e.RLock()
	for k, v := range e.m {
		s = append(s, k+"="+v)
	}
	e.RUnlock()
	sort.Strings(s)
	return s
}

func (e *Environ) String() string {
	return fmt.Sprintf("%#q", e.Slice())
}
