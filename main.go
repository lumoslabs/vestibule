package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"text/template"
)

const version = "0.0.1"

func init() {
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()
}

func getVersion() string {
	return fmt.Sprintf(`%s (%s on %s/%s; %s)`, version, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
}

func usage() string {
	t := template.Must(template.New("usage").Parse(`
Usage: {{ .Self }} user-spec command [args]
   eg: {{ .Self }} myuser bash
       {{ .Self }} nobody:root bash -c 'whoami && id'
       {{ .Self }} 1000:1 id

  Environment Variables:

    SOPS_FILES=/path/to/file[;/path/to/output[;mode]]:...
      If SOPS_FILES is set, will iterate over each file (colon separated), attempting to decrypt with Sops.
      The decrypted cleartext file can be optionally written out to a separate location (with optional filemode)
      or will be parsed into a map[string]string and injected into Environ
    
    VAULT_KEYS=/path/to/key[@version]:...
      If VAULT_KEYS is set, will iterate over each key (colon separated), attempting to get the secret from Vault.
      Secrets are pulled at the optional version or latest, then injected into Environ. If running in Kubernetes,
      the Pod's ServiceAccount token will automatically be looked up and used for Vault authentication.
    
    VAULT_*
      All vault client configuration environment variables are respected.
      More information at https://www.vaultproject.io/docs/commands/#environment-variables

{{ .Self }} version: {{ .Version }}
{{ .Self }} license: GPL-3 (full text at https://github.com/lumoslabs/vestibule)
`))
	var b bytes.Buffer
	template.Must(t, t.Execute(&b, struct {
		Self    string
		Version string
	}{
		Self:    filepath.Base(os.Args[0]),
		Version: getVersion(),
	}))
	return strings.TrimSpace(b.String()) + "\n"
}

func main() {
	log.SetFlags(0) // no timestamps on our logs

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "--help", "-h", "-?":
			fmt.Println(usage())
			os.Exit(0)
		case "--version", "-v":
			fmt.Println(getVersion())
			os.Exit(0)
		}
	}
	if len(os.Args) <= 2 {
		log.Println(usage())
		os.Exit(1)
	}

	e := envv(os.Environ())
	c, _ := newConfig()
	sopsDecrypt(c.sopsConfig, &e)
	if v, er := newVaultClient(); er == nil {
		v.getKeys(c.vaultConfig, &e)
	}

	// clear HOME so that SetupUser will set it
	os.Unsetenv("HOME")

	if err := SetupUser(os.Args[1]); err != nil {
		log.Fatalf("error: failed switching to %q: %v", os.Args[1], err)
	}

	name, err := exec.LookPath(os.Args[2])
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if err = syscall.Exec(name, os.Args[2:], e); err != nil {
		log.Fatalf("error: exec failed: %v", err)
	}
}
