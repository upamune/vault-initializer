package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awskms "github.com/aws/aws-sdk-go/service/kms"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudkms/v1"
)

type KMS interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(encryptedData []byte) ([]byte, error)
}

type GCPKMS struct {
	keyID string
	svc   *cloudkms.Service
}

func NewGCPKMS(ctx context.Context, keyID string) (*GCPKMS, error) {
	client, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloudkms")
	if err != nil {
		return nil, err
	}

	svc, err := cloudkms.New(client)
	if err != nil {
		return nil, err
	}
	svc.UserAgent = fmt.Sprintf("vault-initializer/0.x.0 (%s)", runtime.Version())

	return &GCPKMS{
		keyID: keyID,
		svc:   svc,
	}, nil
}

func (kms *GCPKMS) Encrypt(data []byte) ([]byte, error) {
	res, err := kms.svc.Projects.Locations.KeyRings.CryptoKeys.Encrypt(kms.keyID, &cloudkms.EncryptRequest{
		Plaintext: base64.StdEncoding.EncodeToString(data),
	}).Do()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encrypt by GCPKMS")
	}

	return []byte(res.Ciphertext), nil
}

func (kms *GCPKMS) Decrypt(encryptedData []byte) ([]byte, error) {
	res, err := kms.svc.Projects.Locations.KeyRings.CryptoKeys.Decrypt(kms.keyID, &cloudkms.DecryptRequest{
		Ciphertext: string(encryptedData),
	}).Do()
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt by GCPKMS")
	}

	dst, err := base64.StdEncoding.DecodeString(res.Plaintext)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode text: %s", res.Plaintext)
	}

	return dst, nil
}

type AWSKMS struct {
	keyID   string
	session *session.Session
}

func NewAWSKMS(sess *session.Session, keyID string) *AWSKMS {
	return &AWSKMS{
		keyID:   keyID,
		session: sess,
	}
}

func (kms *AWSKMS) Encrypt(data []byte) ([]byte, error) {
	svc := awskms.New(kms.session)

	res, err := svc.Encrypt(&awskms.EncryptInput{
		KeyId:     aws.String(kms.keyID),
		Plaintext: data,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt by AWSKMS")
	}

	return res.CiphertextBlob, nil
}

func (kms *AWSKMS) Decrypt(encryptedData []byte) ([]byte, error) {
	svc := awskms.New(kms.session)

	res, err := svc.Decrypt(&awskms.DecryptInput{
		CiphertextBlob: encryptedData,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt by AWSKMS")
	}

	return res.Plaintext, nil
}
