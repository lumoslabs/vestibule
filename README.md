# vestibule

A [gosu](https://github.com/tianon/gosu) port which will load secrets from [Vault](https://www.vaultproject.io) and / or [Sops](https://github.com/mozilla/sops) before exec'ing your baby.

## Usage

```
Usage: vest user-spec command [args]
   eg: vest myuser bash
       vest nobody:root bash -c 'whoami && id'
       vest 1000:1 id

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

    EJSON_FILES=/path/to/file1:...
    EJSON_KEYS=pubkey;privkey:...
      If EJSON_FILES is set, will iterate over each file (colon separated), attempting to decrypt using keys
      from EJSON_KEYS. Cleartext decrypted json will be parsed into a map[string]string and injected into Environ.

vest version: 0.0.1 (go1.11.4 on linux/amd64; gc)
vest license: GPL-3 (full text at https://github.com/lumoslabs/vestibule)
```