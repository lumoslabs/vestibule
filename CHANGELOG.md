# Changelog

_Ordered oldest -> newest_

## 0.0.1

* initial release for testing

## 0.0.2

* Fix some bugs around user switching
* Add automatic file discovery to ejson and dotenv providers
* Comments!

## v0.0.3

* Gitlab-ci support
* Include vendored dependencies
* Goreleaser support

## v0.0.4

* Fix gitlab-ci.yml

## v0.1.0

*  Vault Provider: fix authentication
   *  Add new auth env vars: 
      *  `VAULT_AUTH_METHOD`
      *  `VAULT_AUTH_PATH`
      *  `VAULT_AUTH_DATA`
      *  `VAULT_APP_ROLE`
   *  Support actually retrieving a session token from Vault
*  Vault Provider: change kv env var from `VAULT_KEYS` to `VAULT_KV_KEYS`
*  Vault Provider: clean up kv paths

## v0.2.0

* Vault Provider: add request for aws credentials from vault
* Vault Provider: get aws creds if requested
* Add feature to have secrets written to a file
* Add logging!

## v0.2.1

* Vault Provider: reduce http.Client timeout to 5 seconds
* Vault Provider: reduce retries to 1
