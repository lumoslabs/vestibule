package environ

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"gopkg.in/yaml.v2"

	"github.com/joho/godotenv"
)

// 1 or more non-word characters
const regex = "[^0-9A-Za-z_]+"

// New returns a new blank Environ instance
func New() *Environ {
	return &Environ{
		m:          make(map[string]string),
		re:         regexp.MustCompile(regex),
		marshaller: json.Marshal,
	}
}

// NewFromEnv returns a new Environ instance populated from os.Environ
func NewFromEnv() *Environ {
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
	defer e.Unlock()

	for k, v := range m {
		e.m[k] = v
	}
}

// SafeMerge takes a map[string]string and adds it to this Environ without overwriting keys
func (e *Environ) SafeMerge(m map[string]string) {
	e.Lock()
	defer e.Unlock()

	for k, v := range m {
		if _, ok := e.m[k]; !ok {
			e.m[k] = v
		}
	}
}

// SafeAppend takes a slice in the form of os.Environ() - '=' delimited - and appends it to Environ without overwriting keys.
func (e *Environ) SafeAppend(s []string) {
	e.Lock()
	defer e.Unlock()

	for _, item := range s {
		bits := strings.SplitN(item, "=", 2)
		if len(bits) == 2 {
			if _, ok := e.m[bits[0]]; !ok {
				e.m[bits[0]] = bits[1]
			}
		}
	}
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

// Map returns a copy of the underlying map[string]string
func (e *Environ) Map() map[string]string {
	e.RLock()
	defer e.RUnlock()

	dup := make(map[string]string, len(e.m))
	for k, v := range e.m {
		key := strings.ToUpper(e.re.ReplaceAllString(k, "_"))
		dup[key] = v
	}

	return dup
}

// String returns a stringified representation of this Environ
func (e *Environ) String() string {
	return fmt.Sprintf("%#q", e.Slice())
}

// SetMarshaller sets the marshalling function for the Environ object.
func (e *Environ) SetMarshaller(m string) {
	switch strings.ToLower(m) {
	case "yaml", "yml":
		e.marshaller = yaml.Marshal
	case "env", "dotenv":
		e.marshaller = marshalDotEnv
	case "toml":
		e.marshaller = marshalToml
	default:
		e.marshaller = json.Marshal
	}
}

// Write writes the marshalled byte slice of the underlying map to the given io.Writer
func (e *Environ) Write(w io.Writer) error {
	e.RLock()
	defer e.RUnlock()

	out, er := e.marshaller(e.Map())
	if er != nil {
		return er
	}

	_, er = w.Write(out)
	return er
}

func marshalDotEnv(in interface{}) ([]byte, error) {
	inTyped, ok := in.(map[string]string)
	if !ok {
		return []byte(nil), fmt.Errorf("Invalid input type: %T", inTyped)
	}

	out, er := godotenv.Marshal(inTyped)
	return []byte(out), er
}

func marshalToml(in interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if er := toml.NewEncoder(buf).Encode(in); er != nil {
		return []byte(nil), er
	}
	return buf.Bytes(), nil
}
