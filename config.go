package main

import "time"

type Config struct {
	VaultAddr string `envconfig:"VAULT_ADDR" required:"true"`
	CheckInterval time.Duration `envconfig:"CHECK_INTERVAL" default:"10s"`

	KMSKeyID string `envconfig:"KMS_KEY_ID"`

	// For GCP
	GCSBucketName string `envconfig:"GCS_BUCKET_NAME"`
	// For AWS
	S3BucketName string `envconfig:"S3_BUCKET_NAME"`
}

