package auth

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The full path gateway -> client -> upstream is exercised against a fake auth service.
func TestDeleteAccount_ProxiedAsAuthenticatedDeleteWithBody(t *testing.T) {
	var method, path, key, authz, body string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		method, path, key, authz, body = r.Method, r.URL.Path, r.Header.Get("x-api-key"), r.Header.Get("Authorization"), string(raw)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"deleted"}`))
	}))
	defer upstream.Close()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.DELETE("/account", NewHandler(NewService(NewClient(upstream.URL, "internal-key", time.Second))).DeleteAccount)

	req := httptest.NewRequest(http.MethodDelete, "/account", bytes.NewBufferString(`{"password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, http.MethodDelete, method)
	assert.Equal(t, "/account", path)
	assert.Equal(t, "internal-key", key)
	assert.Equal(t, "Bearer abc", authz, "the user's session must reach auth, which identifies the account from it")
	assert.JSONEq(t, `{"password":"secret"}`, body)
}

func TestDeleteAccount_UpstreamPendingStatusReachesTheClient(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"success":false,"message":"account deletion is pending"}`))
	}))
	defer upstream.Close()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.DELETE("/account", NewHandler(NewService(NewClient(upstream.URL, "k", time.Second))).DeleteAccount)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/account", bytes.NewBufferString(`{"password":"x"}`)))

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code, "a pending deletion must never be rewritten into a success")
	assert.Contains(t, rec.Body.String(), "pending")
}

func TestDeleteAccount_UnreachableAuthIs502(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.DELETE("/account", NewHandler(NewService(NewClient(url, "k", 200*time.Millisecond))).DeleteAccount)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/account", bytes.NewBufferString(`{}`)))

	assert.Equal(t, http.StatusBadGateway, rec.Code)
}
