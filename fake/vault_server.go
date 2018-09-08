package fake

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

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

type VaultServer struct {
	secretThreshold    int
	unsealedMap        map[string]struct{}
	keys               []string
	encodedKeys        []string
	rootToken          string
	alreadyInitialised bool
	alreadyUnsealed    bool
	r                  *rand.Rand
}

func NewVaultServer() *VaultServer {
	return &VaultServer{
		unsealedMap: map[string]struct{}{},
		alreadyInitialised: false,
		alreadyUnsealed:    false,
		r:                  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *VaultServer) generateKeys(n int) []string {
	keys := make([]string, 0, n)

	for i := 0; i < n; i++ {
		h := sha256.Sum256([]byte(strconv.FormatInt(s.r.Int63(), 10)))
		keys = append(keys, fmt.Sprintf("%x", h))
	}

	return keys
}

func (s *VaultServer) unsealHandler(w http.ResponseWriter, r *http.Request) {
	if !s.alreadyInitialised || s.alreadyUnsealed {
		w.WriteHeader(http.StatusBadRequest)
		return
	}


	req := &UnsealRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid json body: %v", err)))
		return
	}
	defer r.Body.Close()

	isFound := false
	for _, k := range s.encodedKeys {
		if req.Key == k {
			s.unsealedMap[k] = struct{}{}
			isFound = true
		}
	}
	if !isFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(s.unsealedMap) >= s.secretThreshold {
		s.alreadyUnsealed = true
	}

	res := &UnsealResponse{
		Sealed:   !s.alreadyUnsealed,
	}
	if err := json.NewEncoder(w).Encode(res); err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *VaultServer) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if !s.alreadyInitialised {
		w.WriteHeader(http.StatusNotImplemented)
		return
	} else if !s.alreadyUnsealed {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (s *VaultServer) initHandler(w http.ResponseWriter, r *http.Request) {
	if s.alreadyInitialised {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("already initialized"))
		return
	}

	req := &InitRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json body"))
		return
	}
	defer r.Body.Close()

	keys := s.generateKeys(req.SecretShares)
	s.keys = keys
	b64Keys := make([]string, 0, req.SecretShares)
	for _, k := range keys {
		b64Keys = append(b64Keys, base64.StdEncoding.EncodeToString([]byte(k)))
	}
	s.encodedKeys = b64Keys

	rootToken := s.generateKeys(1)[0]
	res := &InitResponse{
		Keys:       s.keys,
		KeysBase64: s.encodedKeys,
		RootToken:  rootToken,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to encode json: %v", err)))
		return
	}

	s.alreadyInitialised = true
}

func (s *VaultServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v1/sys/health":
		s.healthCheckHandler(w, r)
		return
	case "/v1/sys/init":
		s.initHandler(w, r)
		return
	case "/v1/sys/unseal":
		s.unsealHandler(w, r)
		return
	default:
		panic(fmt.Sprintf("unexpected url parh: %s", r.URL.Path))
	}
}
