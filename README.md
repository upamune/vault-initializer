# vault-initializer [![CircleCI](https://circleci.com/gh/upamune/vault-initializer/tree/master.svg?style=svg)](https://circleci.com/gh/upamune/vault-initializer/tree/master) [![Docker Repository on Quay](https://quay.io/repository/upamune/vault-initializer/status "Docker Repository on Quay")](https://quay.io/repository/upamune/vault-initializer) ![Go Report Card](https://goreportcard.com/badge/github.com/upamune/vault-initializer) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![logo](https://i.gyazo.com/90a9c2c4da924ae3f644fd1431bd7317.png)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fupamune%2Fvault-initializer.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fupamune%2Fvault-initializer?ref=badge_shield)

This is a port of [vault-init](https://github.com/kelseyhightower/vault-init) to AWS/GCP.

## Usage

The `vault-initializer` service is designed to be run alongside a Vault server and communicate over local host.


## Configuration

The vault-initializer service supports the following environment variables for configuration:

* `CHECK_INTERVAL` - The time in seconds between Vault health checks. (300s)
* `VAULT_ADDR` - Address of Vault service. (https://127.0.0.1:8200)
* `KMS_KEY_ID` - The Google Cloud KMS or AWS KMS key ID used to encrypt and decrypt the vault master key and root token.
* `REGION` - Region of AWS KMS/S3 or GCP KMS/GCS.
* `S3_BUCKET_NAME`  - The AWS Storage Bucket where the vault master key and root token is stored.
* `GCS_BUCKET_NAME` - The Google Cloud Storage Bucket where the vault master key and root token is stored.


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fupamune%2Fvault-initializer.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fupamune%2Fvault-initializer?ref=badge_large)