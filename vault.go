package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type Vault struct {
	Addr                string
	KeyStorage          Storage
	KMS                 KMS
	httpClient          *http.Client
	rootTokenObjectName string
	unsealObjectName    string
}

type InitRequest struct {
	SecretShares    int `json:"secret_shares"`
	SecretThreshold int `json:"secret_threshold"`
}

type InitResponse struct {
	Keys       []string `json:"keys"`
	KeysBase64 []string `json:"keys_base64"`
	RootToken  string   `json:"root_token"`
}

type UnsealRequest struct {
	Key   string `json:"key"`
	Reset bool   `json:"reset"`
}

type UnsealResponse struct {
	Sealed   bool `json:"sealed"`
	T        int  `json:"t"`
	N        int  `json:"n"`
	Progress int  `json:"progress"`
}

func NewVault(addr string, storage Storage, kms KMS) *Vault {
	return &Vault{
		Addr:       addr,
		KeyStorage: storage,
		KMS:        kms,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		rootTokenObjectName: "root-token.enc",
		unsealObjectName:    "unseal-keys.json.enc",
	}
}

func (v *Vault) HealthCheck() (int, error) {
	endpoint := fmt.Sprintf("%s/v1/sys/health", v.Addr)
	res, err := v.httpClient.Head(endpoint)
	if err != nil {
		return 0, err
	}
	if res != nil && res.Body != nil {
		res.Body.Close()
	}

	return res.StatusCode, nil
}

func (v *Vault) Initialize() error {
	initRequest := &InitRequest{
		SecretShares:    5,
		SecretThreshold: 3,
	}

	data, err := json.Marshal(initRequest)
	if err != nil {
		return err
	}
	r := bytes.NewReader(data)

	endpoint := fmt.Sprintf("%s/v1/sys/init", v.Addr)
	req, err := http.NewRequest(http.MethodPut, endpoint, r)
	if err != nil {
		return err
	}

	res, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("failed to initialize request to vault")
	}

	initResponse := &InitResponse{}
	if err := json.NewDecoder(res.Body).Decode(initResponse); err != nil {
		return err
	}

	erootToken, err := v.KMS.Encrypt([]byte(initResponse.RootToken))
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(initResponse); err != nil {
		return err
	}

	b := buf.Bytes()
	d := make([]byte, base64.StdEncoding.EncodedLen(len(b)))

	base64.StdEncoding.Encode(d, b)
	eunsealKeys, err := v.KMS.Encrypt([]byte(d))
	if err != nil {
		return err
	}

	if err := v.KeyStorage.Put(v.rootTokenObjectName, erootToken); err != nil {
		return err
	}

	if err := v.KeyStorage.Put(v.unsealObjectName, eunsealKeys); err != nil {
		return err
	}

	return nil
}

func (v *Vault) Unseal() error {
	edata, err := v.KeyStorage.Get(v.unsealObjectName)
	if err != nil {
		return err
	}

	b64data, err := v.KMS.Decrypt(edata)
	if err != nil {
		return err
	}

	initResponse := &InitResponse{}
	data := make([]byte, base64.StdEncoding.DecodedLen(len(b64data)))
	if _, err := base64.StdEncoding.Decode(data, b64data); err != nil {
		return err
	}

	if err := json.Unmarshal(data, initResponse); err != nil {
		return err
	}

	for _, key := range initResponse.KeysBase64 {
		done, err := v.unsealOne(key)
		if err != nil {
			return err
		}

		if done {
			return nil
		}
	}

	return errors.New("failed to unseal the Vault")
}

func (v *Vault) unsealOne(key string) (bool, error) {
	b := bytes.NewBuffer([]byte(""))
	if err := json.NewEncoder(b).Encode(&UnsealRequest{Key: key}); err != nil {
		return false, err
	}

	endpoint := v.Addr + "/v1/sys/unseal"
	req, err := http.NewRequest(http.MethodPut, endpoint, b)
	if err != nil {
		return false, err
	}

	res, err := v.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	unsealResponse := &UnsealResponse{}
	if err := json.NewDecoder(res.Body).Decode(unsealResponse); err != nil {
		return false, err
	}

	if !unsealResponse.Sealed {
		return true, nil
	}

	return false, nil
}
