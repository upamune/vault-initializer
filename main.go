package main

import (
	"context"
	"net/http"
	"time"

	gstorage "cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Starting the vault-initializer service...")

	config := &Config{}
	if err := envconfig.Process("", config); err != nil {
		log.Fatal().Msg("failed to process config")
	}

	log.Info().
		Str("vault_addr", config.VaultAddr).
		Dur("check_interval", config.CheckInterval).
		Str("kms_key_id", config.KMSKeyID).
		Str("gcs_bucket_name", config.GCSBucketName).
		Str("s3_bucket_name", config.S3BucketName).
		Msg("config")

	var (
		storage Storage
		kms     KMS
	)
	if config.S3BucketName != "" {
		sess := session.New()
		storage = NewS3(sess, config.S3BucketName)
		kms = NewAWSKMS(sess, config.KMSKeyID)
	} else if config.GCSBucketName != "" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sclient, err := gstorage.NewClient(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create GCS client")
		}
		storage = NewGCS(sclient, config.GCSBucketName)

		if s, err := NewGCPKMS(ctx, config.KMSKeyID); err != nil {
			log.Fatal().Err(err).Msg("failed to create GCPKMS client")
		} else {
			kms = s
		}
	} else {
		log.Fatal().Msg("both S3BucketName and GCSBucketName are empty")
	}

	vault := NewVault(config.VaultAddr, storage, kms)

	for {
		status, err := vault.HealthCheck()
		if err != nil {
			log.Error().
				Str("vault_addr", config.VaultAddr).
				Msg("failed to call health check endpoint of Vault")
			time.Sleep(config.CheckInterval)
			continue
		}

		switch status {
		case http.StatusOK:
			log.Info().Str("vault_addr", config.VaultAddr).Msg("Vault is initialized and unsealed")
		case http.StatusTooManyRequests:
			log.Info().Str("vault_addr", config.VaultAddr).Msg("Vault is unsealed and in standby mode")
		case http.StatusNotImplemented:
			log.Info().Str("vault_addr", config.VaultAddr).Msg("Vault is not initialized. Initializing and unsealing...")
			if err := vault.Initialize(); err != nil {
				log.Error().
					Err(err).
					Str("vault_addr", config.VaultAddr).
					Msg("failed to initialize Vault")
				time.Sleep(config.CheckInterval)
				continue
			}
			if err := vault.Unseal(); err != nil {
				log.Error().
					Err(err).
					Str("vault_addr", config.VaultAddr).
					Msg("failed to unseal Vault")
				time.Sleep(config.CheckInterval)
				continue
			}
		case http.StatusServiceUnavailable:
			log.Info().Str("vault_addr", config.VaultAddr).Msg("Vault is sealed. Unsealing...")
			if err := vault.Unseal(); err != nil {
				log.Error().
					Err(err).
					Str("vault_addr", config.VaultAddr).
					Msg("failed to unseal Vault")
				time.Sleep(config.CheckInterval)
				continue
			}
		default:
			log.Info().
				Str("vault_addr", config.VaultAddr).
				Int("status_code", status).
				Msg("Vault is in an unknown state.")
		}

		log.Info().Str("vault_addr", config.VaultAddr).Msgf("Next check in %s", config.CheckInterval)
		time.Sleep(config.CheckInterval)
	}
}
