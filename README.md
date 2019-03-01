# vestibule

[![Github PR check contexts](https://img.shields.io/github/status/contexts/pulls/lumoslabs/vestibule/1.svg)]
[![Github liscense](https://img.shields.io/github/license/lumoslabs/vestibule.svg)]
[![Github release](https://img.shields.io/github/release-pre/lumoslabs/vestibule.svg)]

A [gosu](https://github.com/tianon/gosu) port which will load secrets from various secrets backends into the environment before exec'ing your baby. [Vault](https://www.vaultproject.io) and / or [Sops](https://github.com/mozilla/sops)

## Providers

Enable providers by setting `VEST_PROVIDERS` in the environment before running `vest`

Available providers:

  * [`vault`](https://www.vaultproject.io)
  * [`sops`](https://github.com/mozilla/sops)
  * [`ejson`](https://github.com/Shopify/ejson)
  * plain old `.env` files

## Usage

```
Usage: vest user-spec command [args]
   eg: vest myuser bash
       vest nobody:root bash -c 'whoami && id'
       vest 1000:1 id

  Environment Variables:

    VAULT_PROVIDERS=provider1,...
      Comma separated list of enabled providers. By default only vault is enabled.

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

    EJSON_FILES=/path/to/file1:...
    EJSON_KEYS=pubkey;privkey:...
      If EJSON_FILES is set, will iterate over each file (colon separated), attempting to decrypt using keys
      from EJSON_KEYS. If EJSON_FILES is not set, will look for any .ejson files in CWD. Cleartext decrypted
      json will be parsed into a map[string]string and injected into Environ.

    DOTENV_FILES=/path/to/file1:...
      if DOTENV_FILES is set, will iterate over each file, parse and inject into Environ. If DOTENV_FILES is
      not set, will look for any .env files in CWD.
```
