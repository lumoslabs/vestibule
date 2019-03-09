package environ

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// 1 or more non-word characters
const regex = "[^0-9A-Za-z_]+"

// NewEnviron returns a new blank Environ instance
func NewEnviron() *Environ {
	return &Environ{m: make(map[string]string), re: regexp.MustCompile(regex)}
}

// NewEnvironFromEnv returns a new Environ instance populated from os.Environ
func NewEnvironFromEnv() *Environ {
	e := make(map[string]string)
	for _, item := range os.Environ() {
		bits := strings.SplitN(item, "=", 2)
		if len(bits) == 2 {
			e[bits[0]] = bits[1]
		}
	}
	return &Environ{m: e, re: regexp.MustCompile(regex)}
}

// Merge takes a map[string]string and adds it to this Environ, overwriting any conflicting keys.
func (e *Environ) Merge(m map[string]string) {
	e.Lock()
	for k, v := range m {
		e.m[k] = v
	}
	e.Unlock()
}

// SafeMerge takes a map[string]string and adds it to this Environ without overwriting keys
func (e *Environ) SafeMerge(m map[string]string) {
	e.Lock()
	for k, v := range m {
		if _, ok := e.m[k]; !ok {
			e.m[k] = v
		}
	}
	e.Unlock()
}

// Set takes a key / value pair and adds it to this Environ
func (e *Environ) Set(k, v string) {
	e.Lock()
	defer e.Unlock()
	e.m[k] = v
}

// Load takes a key and returns the value if it exists or false
func (e *Environ) Load(k string) (v string, ok bool) {
	e.RLock()
	defer e.RUnlock()
	v, ok = e.m[k]
	return
}

// Delete takes a key and removes it from this Environ, returning the value
func (e *Environ) Delete(key string) (v string) {
	e.Lock()
	defer e.Unlock()

	v = e.m[key]
	delete(e.m, key)
	return
}

// Len returns the length of this Environ
func (e *Environ) Len() (l int) {
	e.RLock()
	defer e.RUnlock()
	l = len(e.m)
	return
}

// Slice returns a sorted []string of key / value pairs from this Environ instance
// suitable for use in palce of os.Environ()
func (e *Environ) Slice() []string {
	e.RLock()
	var s = make([]string, 0, e.Len())
	for k, v := range e.m {
		key := strings.ToUpper(e.re.ReplaceAllString(k, "_"))
		s = append(s, key+"="+v)
	}
	e.RUnlock()
	sort.Strings(s)
	return s
}

// String returns a stringified representation of this Environ
func (e *Environ) String() string {
	return fmt.Sprintf("%#q", e.Slice())
}
