package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/upamune/vault-initializer/fake"
)

func TestVault(t *testing.T) {
	t.Run("not initialize yet", func(t *testing.T) {
		ts := httptest.NewServer(fake.NewVaultServer())
		defer ts.Close()
		vault := NewVault(ts.URL, &fake.Storage{}, &fake.KMS{})

		code, err := vault.HealthCheck()
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotImplemented, code)
	})

	t.Run("initialized, but not sealed", func(t *testing.T) {
		ts := httptest.NewServer(fake.NewVaultServer())
		defer ts.Close()
		vault := NewVault(ts.URL, &fake.Storage{}, &fake.KMS{})

		err := vault.Initialize()
		require.NoError(t, err)

		code, err := vault.HealthCheck()
		require.NoError(t, err)

		assert.Equal(t, http.StatusServiceUnavailable, code)
	})

	t.Run("initialized, and unsealed", func(t *testing.T) {
		ts := httptest.NewServer(fake.NewVaultServer())
		defer ts.Close()
		vault := NewVault(ts.URL, &fake.Storage{}, &fake.KMS{})

		err := vault.Initialize()
		require.NoError(t, err)

		err = vault.Unseal()
		require.NoError(t, err)

		code, err := vault.HealthCheck()
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, code)
	})
}
