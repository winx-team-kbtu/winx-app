package profileclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteUserData_SendsAuthenticatedDelete(t *testing.T) {
	var method, path, key string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path, key = r.Method, r.URL.Path, r.Header.Get("x-api-key")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	require.NoError(t, New(srv.URL+"/", "secret-key", time.Second).DeleteUserData(context.Background(), 42))

	assert.Equal(t, http.MethodDelete, method)
	assert.Equal(t, "/internal/users/42", path)
	assert.Equal(t, "secret-key", key)
}

func TestDeleteUserData_NonSuccessStatusIsAnError(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError, http.StatusBadGateway} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"message":"nope"}`))
		}))

		err := New(srv.URL, "k", time.Second).DeleteUserData(context.Background(), 1)
		srv.Close()

		require.Error(t, err, "status %d must not count as success", status)
		assert.Contains(t, err.Error(), "nope")
	}
}

func TestDeleteUserData_UnreachableServiceIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	require.Error(t, New(url, "k", 200*time.Millisecond).DeleteUserData(context.Background(), 1))
}

func TestDeleteUserData_TimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	require.Error(t, New(srv.URL, "k", 50*time.Millisecond).DeleteUserData(context.Background(), 1))
}
