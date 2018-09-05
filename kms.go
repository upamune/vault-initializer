package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awskms "github.com/aws/aws-sdk-go/service/kms"
	"github.com/pkg/errors"
)

type KMS interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(encryptedData []byte) ([]byte, error)
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
