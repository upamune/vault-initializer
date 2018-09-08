package fake

import "strings"

type KMS struct {
	prefix string
}

func (kms *KMS) Encrypt(data []byte) ([]byte, error) {
	return []byte(kms.prefix + string(data)), nil
}

func (kms *KMS) Decrypt(encryptedData []byte) ([]byte, error) {
	return []byte(strings.TrimPrefix(string(encryptedData), kms.prefix)), nil
}
