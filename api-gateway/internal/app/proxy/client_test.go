package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		body        []byte
		query       url.Values
		headers     map[string]string
		statusCode  int
		response    []byte
		contentType string
		wantErr     bool
	}{
		{
			name:        "successful GET request",
			method:      http.MethodGet,
			path:        "/test",
			body:        nil,
			query:       nil,
			headers:     map[string]string{},
			statusCode:  http.StatusOK,
			response:    []byte(`{"message":"success"}`),
			contentType: "application/json",
			wantErr:     false,
		},
		{
			name:        "successful POST request",
			method:      http.MethodPost,
			path:        "/create",
			body:        []byte(`{"name":"test"}`),
			query:       nil,
			headers:     map[string]string{},
			statusCode:  http.StatusCreated,
			response:    []byte(`{"id":1}`),
			contentType: "application/json",
			wantErr:     false,
		},
		{
			name:        "server error",
			method:      http.MethodGet,
			path:        "/error",
			body:        nil,
			query:       nil,
			headers:     map[string]string{},
			statusCode:  http.StatusInternalServerError,
			response:    []byte(`{"error":"server error"}`),
			contentType: "application/json",
			wantErr:     false,
		},
		{
			name:        "request with query parameters",
			method:      http.MethodGet,
			path:        "/search",
			body:        nil,
			query:       url.Values{"q": []string{"test"}, "limit": []string{"10"}},
			headers:     map[string]string{},
			statusCode:  http.StatusOK,
			response:    []byte(`{"results":[]}`),
			contentType: "application/json",
			wantErr:     false,
		},
		{
			name:        "request with custom headers",
			method:      http.MethodGet,
			path:        "/protected",
			body:        nil,
			query:       nil,
			headers:     map[string]string{"X-Custom": "value"},
			statusCode:  http.StatusOK,
			response:    []byte(`{"data":"protected"}`),
			contentType: "application/json",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify method
				assert.Equal(t, tt.method, r.Method)

				// Verify path
				assert.Equal(t, tt.path, r.URL.Path)

				// Verify query parameters if any
				if tt.query != nil {
					for key, values := range tt.query {
						assert.Equal(t, values, r.URL.Query()[key])
					}
				}

				// Verify custom headers
				for key, value := range tt.headers {
					assert.Equal(t, value, r.Header.Get(key))
				}

				// Set response headers
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.statusCode)
				w.Write(tt.response)
			}))
			defer server.Close()

			// Create client
			client := NewClient(server.URL, "test-api-key", 5*time.Second)

			// Create request
			req := Request{
				Method:      tt.method,
				Path:        tt.path,
				ContentType: "application/json",
				Body:        tt.body,
				Headers:     tt.headers,
				Query:       tt.query,
			}

			// Execute request
			resp, err := client.Do(context.Background(), req)

			// Assert results
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.statusCode, resp.StatusCode)
				assert.Equal(t, tt.contentType, resp.ContentType)
				assert.Equal(t, tt.response, resp.Body)
			}
		})
	}
}

func TestClient_Timeout(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create client with short timeout
	client := NewClient(server.URL, "test-key", 1*time.Second)

	req := Request{
		Method: http.MethodGet,
		Path:   "/slow",
	}

	_, err := client.Do(context.Background(), req)
	assert.Error(t, err)
}

func TestClient_WithAPIKey(t *testing.T) {
	expectedAPIKey := "test-api-key-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key header
		assert.Equal(t, expectedAPIKey, r.Header.Get("x-api-key"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, expectedAPIKey, 5*time.Second)

	req := Request{
		Method: http.MethodGet,
		Path:   "/secure",
	}

	resp, err := client.Do(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_EmptyAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API key header should not be sent when empty
		assert.Empty(t, r.Header.Get("x-api-key"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "", 5*time.Second)

	req := Request{
		Method: http.MethodGet,
		Path:   "/public",
	}

	resp, err := client.Do(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_WithQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		assert.Equal(t, "value1", r.URL.Query().Get("param1"))
		assert.Equal(t, "value2", r.URL.Query().Get("param2"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 5*time.Second)

	req := Request{
		Method: http.MethodGet,
		Path:   "/search",
		Query: url.Values{
			"param1": {"value1"},
			"param2": {"value2"},
			"limit":  {"10"},
		},
	}

	resp, err := client.Do(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_WithCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify custom headers
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 5*time.Second)

	req := Request{
		Method:      http.MethodPost,
		Path:        "/protected",
		ContentType: "application/json",
		Body:        []byte(`{"data":"test"}`),
		Headers: map[string]string{
			"Authorization":   "Bearer token123",
			"X-Custom-Header": "custom-value",
		},
	}

	resp, err := client.Do(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_InvalidURL(t *testing.T) {
	client := NewClient("http://invalid-url-that-does-not-exist:9999", "test-key", 1*time.Second)

	req := Request{
		Method: http.MethodGet,
		Path:   "/test",
	}

	_, err := client.Do(context.Background(), req)
	assert.Error(t, err)
}

func TestClient_EmptyMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Default method should be GET
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 5*time.Second)

	req := Request{
		Method: "", // Empty method should default to GET
		Path:   "/test",
	}

	resp, err := client.Do(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 5*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := Request{
		Method: http.MethodGet,
		Path:   "/test",
	}

	_, err := client.Do(ctx, req)
	assert.Error(t, err)
}
