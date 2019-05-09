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

## v0.2.2

* Vault Provider: add AWS_SESSION_TOKEN to aws creds gathered
* fix checksum mismatch for hashicorp/go-rootcerts
* Vault Provider: use go-ini for credential file creation

## v0.2.3

* Use ubuntu:bionic instead of debian:stretch for docker image

## v1.0.0

* Release!
* `bule`
  * Handles secrets writing
* Docker
  * Add bule to the entrypoint.sh
  * Add jq

## v1.0.1

* Actually add `bule` to the Docker entrypoint.sh
* Actually add jq
* Split logging out to a simple internal package

## v1.1.0

* Logging works as intended now e.g. levels are fixed
* In image entrypoint.sh, `bule` writes secrets to /var/run/vestibule/secrets
* Default path for aws credentials file is `/var/run/aws/credentials`
* Case of variable names is configurable, defaults to upcasing them
* Add verbose logging for `bule`

## v1.1.2

* Pass through timeouts to the vault client

## v1.2.0

* Refactor vault provider
  * Handle setting defaults on the vault client better
  * Better logging with redacting sensitive items
  * Better tests
* Add first class support for Vault jwt login
* Add first class support for Vault approle login
* VAULT_IAM_* -> VAULT_AWS_*
* Add support for generating GCP service account key
* Add homebrew release

## v1.2.2

* Fix homebrew tap

## v1.2.3

* Allow for approle with only role_id

## v1.2.4

* Reenable VAULT_IAM_ROLE, but mark it deprecated
* Add debugging around VAULT_AUTH_DATA

## v1.2.5

* Fix issue with vault kv keys not being split

## v1.2.6

* Fix vault kv keys race condition

## v1.2.7

* Use RedactableAuthData for auth data so we always redact when loggin

## v1.2.8

* Fix kubernetes in cluster method

## v1.2.9

* Actually fix kubernetes in cluster method
