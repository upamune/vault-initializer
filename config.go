package main

import "time"

type Config struct {
	VaultAddr string `envconfig:"VAULT_ADDR" required:"true" default:"https://127.0.0.1:8200"`
	CheckInterval time.Duration `envconfig:"CHECK_INTERVAL" default:"300s"`

	KMSKeyID string `envconfig:"KMS_KEY_ID" required:"true"`

	// For GCP
	GCSBucketName string `envconfig:"GCS_BUCKET_NAME"`
	// For AWS
	S3BucketName string `envconfig:"S3_BUCKET_NAME"`
}

