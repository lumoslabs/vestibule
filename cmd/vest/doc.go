package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

var (
	version      = "dev"
	commit, date string
)

func appVersion() string {
	return fmt.Sprintf(
		`%s (%s on %s/%s; %s)`,
		strings.Join([]string{version, commit, date}, "/"),
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Compiler,
	)
}

func usage() string {
	t := template.Must(template.New("usage").Parse(`
Usage: {{ .Self }} user-spec command [args]
   eg: {{ .Self }} myuser bash
       {{ .Self }} nobody:root bash -c 'whoami && id'
       {{ .Self }} 1000:1 id

  Environment Variables:

    VEST_USER=user[:group]
      The user [and group] to run the command under. Overrides commandline if set.

    VEST_PROVIDERS=provider1,...
      Comma separated list of enabled providers. By default only vault is enabled.

    VEST_OUTPUT_FILE=/path/to/file
      If set, will write gathered secrets from enabled providers to the specified file. On error, does nothing.

    VEST_OUTPUT_FORMAT=<json|yaml|yml|env|dotenv|toml>
      The format of the output file. Default is json.

    SOPS_FILES=/path/to/file[;/path/to/output[;mode]]:...
      If SOPS_FILES is set, will iterate over each file (colon separated), attempting to decrypt with Sops.
      The decrypted cleartext file can be optionally written out to a separate location (with optional filemode)
      or will be parsed into a map[string]string and injected into Environ
    
    VAULT_KV_KEYS=/path/to/key[@version]:...
      If VAULT_KEYS is set, will iterate over each key (colon separated), attempting to get the secret from Vault.
      Secrets are pulled at the optional version or latest, then injected into Environ. If running in Kubernetes,
      the Pod's ServiceAccount token will automatically be looked up and used for Vault authentication.
    
    VAULT_AUTH_METHOD=kubernetes
      Authentication method for vault. Default is kubernetes.

    VAULT_APP_ROLE
      App role name to use with the kubernetes authentication method.

    VAULT_IAM_ROLE
      IAM role to request from vault. If returns credentials, the access key and secret key will be injected into
      the process environment using the standard environment variables and a credentials file will be written to
      the path from AWS_SHARED_CREDENTIALS_FILE (by default "/var/aws/credentials")

    VAULT_AWS_PATH
      Mountpoint for the vault AWS secret engine. Defaults to "aws".

    VAULT_AUTH_PATH
      Authentication path for vault authentication - e.g. okta/login/:user. Overrides VAULT_AUTH_METHOD if set.
    
    VAULT_AUTH_DATA
      Data payload to send with authentication request. JSON object.
    
    VAULT_*
      All vault client configuration environment variables are respected.
      More information at https://www.vaultproject.io/docs/commands/#environment-variables
    
    EJSON_FILES=/path/to/file1:...
    EJSON_KEYS=pubkey;privkey:...
      If EJSON_FILES is set, will iterate over each file (colon separated), attempting to decrypt using keys
      from EJSON_KEYS. If EJSON_FILES is not set, will look for any .ejson files in CWD. Cleartext decrypted
      json will be parsed into a map[string]string and injected into Environ.
    
    DOTENV_FILES=/path/to/file1:...
      if DOTENV_FILES is set, will iterate over each file, parse and inject into Environ. If DOTENV_FILES is
      not set, will look for any .env files in CWD.

{{ .Self }} version: {{ .Version }}
{{ .Self }} license: GPL-3 (full text at https://github.com/lumoslabs/vestibule)
`))
	var b bytes.Buffer
	template.Must(t, t.Execute(&b, struct {
		Self    string
		Version string
	}{
		Self:    filepath.Base(os.Args[0]),
		Version: appVersion(),
	}))
	return strings.TrimSpace(b.String()) + "\n"
}
