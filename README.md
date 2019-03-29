# vestibule

![Github PR check contexts](https://img.shields.io/github/status/contexts/pulls/lumoslabs/vestibule/1.svg)
![Github liscense](https://img.shields.io/github/license/lumoslabs/vestibule.svg)
![Github release](https://img.shields.io/github/release-pre/lumoslabs/vestibule.svg)

A [gosu](https://github.com/tianon/gosu) port which will load secrets from various secrets backends into the environment before exec'ing your baby.

## Providers

Enable providers by setting `VEST_PROVIDERS` in the environment before running `vest`

Available providers:

* [`vault`](https://www.vaultproject.io) (enabled by default)
* [`sops`](https://github.com/mozilla/sops)
* [`ejson`](https://github.com/Shopify/ejson)
* plain old `.env` files

## Images

Alpine and Ubuntu based docker images are available at [`quay.io/lumoslabs/vestibule`](https://quay.io/repository/lumoslabs/vestibule?tag=latest&tab=tags)

## Building

This project uses [goreleaser](https://goreleaser.com/) for building and publishing.
Install instructions for goreleaser are [here](https://goreleaser.com/install/).

The handy Makefile here provides targets:

* `snapshot`: Use goreleaser to make an unpublished snapshot build
* `release`: Use goreleaser to cut and publish a real release
* `test`: Run go tests
* `test-race`: Run go tests with `-race`
* `test-memory`: Run go tests with `-msan`
* `test-all`: Run all tests
* `linux`: Use `go build` to build `vestibule` for `linux`
* `darwin`: Use `go build` to build `vestibule` for `darwin`

## Usage

    Usage: vest user-spec command [args]
      eg: vest myuser bash
          vest nobody:root bash -c 'whoami && id'
          vest 1000:1 id

      Environment Variables:

        VEST_DEBUG
          Enable debug logging.

        VEST_PROVIDERS
          Comma separated list of enabled providers. By default only Vault is
          enabled. Available providers: [dotenv ejson vault sops]

        VEST_UPCASE_VAR_NAMES
          Upcase environment variable names gathered from secret providers.

        VEST_USER
          The user [and group] to run the command as. Overrides commandline if set.
          e.g. VEST_USER=user[:group]

        VEST_VERBOSE
          Enable verbose logging.

        VAULT_*
          All vault client configuration environment variables are respected. More
          information at
          https://www.vaultproject.io/docs/commands/#environment-variables

        VAULT_APP_ROLE
          App role name to use with the kubernetes authentication method.

        VAULT_AUTH_DATA
          Data payload to send with authentication request. JSON object.

        VAULT_AUTH_METHOD
          Authentication method for vault. Default is "kubernetes".

        VAULT_AUTH_PATH
          Authentication path for vault authentication - e.g. okta/login/:user.
          Overrides VAULT_AUTH_METHOD if set.

        VAULT_AWS_PATH
          Mountpoint for the vault AWS secret engine. Defaults to "aws".

        VAULT_IAM_ROLE
          IAM role to request from vault. If returns credentials, the access key and
          secret key will be injected into the process environment using the
          standard environment variables and a credentials file will be written to
          the path from AWS_SHARED_CREDENTIALS_FILE (by default
          "/var/aws/credentials")

        VAULT_KV_KEYS
          If VAULT_KV_KEYS is set, will iterate over each key (colon separated),
          attempting to get the secret from Vault. Secrets are pulled at the
          optional version or latest, then injected into Environ. If running in
          Kubernetes, the Pod's ServiceAccount token will automatically be looked up
          and used for Vault authentication. e.g.
          VAULT_KV_KEYS=/path/to/key1[@version]:/path/to/key2[@version]:...

        DOTENV_FILES
          if DOTENV_FILES is set, will iterate over each file, parse and inject into
          Environ. If DOTENV_FILES is not set, will look for any .env files in CWD.
          e.g. DOTENV_FILES=/path/to/file1:/path/to/file2:...

        EJSON_FILES
          If EJSON_FILES is set, will iterate over each file (colon separated),
          attempting to decrypt using keys from EJSON_KEYS. If EJSON_FILES is not
          set, will look for any .ejson files in CWD. Cleartext decrypted json will
          be parsed into a map[string]string and injected into Environ. e.g.
          EJSON_FILES=/path/to/file1:/path/to/file2:...

        EJSON_KEYS
          Colon separated list of public/private ejson keys. Public/private keys
          separated by semicolon. e.g.
          EJSON_KEYS=pubkey1;privkey1:pubkey2;privkey2:...

        SOPS_FILES
          If SOPS_FILES is set, will iterate over each file (colon separated),
          attempting to decrypt with Sops. The decrypted cleartext file can be
          optionally written out to a separate location (with optional filemode) or
          will be parsed into a map[string]string and injected into Environ e.g.
          SOPS_FILES=/path/to/file[;/path/to/output[;mode]]:...

## Writing to a file

Sometimes you just need credentials to be on disk, amirite?

If so, you can run `bule` to write gathered secrets to a given file in a given format.
All provider environment variables from `vest` are also applicable with `bule`

    e.g. VAULT_KV_KEYS=secret/db-creds bule /var/secrets/db-creds.json

    usage: bule [<flags>] <file>

    Write secrets to a file! What could go wrong?

    Flags:
      -h, --help                Show context-sensitive help (also try --help-long and --help-man).
      -D, --debug               Debug output
      -v, --verbose             Verbose output
      -F, --format=json         Format of the output file. Available formats: [dotenv env json toml yaml yml]
      -p, --provider=vault ...  Secret provider. Can be used multiple times. Available providers: [dotenv ejson vault sops]
          --upcase-var-names    Upcase environment variable names gathered from secret providers.
          --version             Show application version.

    Args:
      <file>  Path of output file
