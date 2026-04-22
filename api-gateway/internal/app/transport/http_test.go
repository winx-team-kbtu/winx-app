package transport

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestReadBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "successful body read",
			body:    `{"key":"value"}`,
			wantErr: false,
		},
		{
			name:    "empty body",
			body:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.body))

			body, err := ReadBody(c)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.body, string(body))
			}
		})
	}
}

func TestWriteJSONError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		statusCode int
		message    string
	}{
		{
			name:       "bad request error",
			statusCode: 400,
			message:    "invalid request",
		},
		{
			name:       "unauthorized error",
			statusCode: 401,
			message:    "unauthorized",
		},
		{
			name:       "server error",
			statusCode: 500,
			message:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			WriteJSONError(c, tt.statusCode, tt.message)

			assert.Equal(t, tt.statusCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.message)
			assert.Contains(t, w.Body.String(), "false")
		})
	}
}

func TestForwardHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		headers     map[string]string
		forwardKeys []string
		expected    map[string]string
	}{
		{
			name: "forward existing headers",
			headers: map[string]string{
				"Authorization": "Bearer token123",
				"X-Request-ID":  "req-123",
			},
			forwardKeys: []string{"Authorization", "X-Request-ID"},
			expected: map[string]string{
				"Authorization": "Bearer token123",
				"X-Request-ID":  "req-123",
			},
		},
		{
			name: "forward only specified headers",
			headers: map[string]string{
				"Authorization": "Bearer token123",
				"X-Request-ID":  "req-123",
				"User-Agent":    "test-agent",
			},
			forwardKeys: []string{"Authorization"},
			expected: map[string]string{
				"Authorization": "Bearer token123",
			},
		},
		{
			name:        "forward non-existent headers",
			headers:     map[string]string{},
			forwardKeys: []string{"Authorization"},
			expected:    map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)

			for k, v := range tt.headers {
				c.Request.Header.Set(k, v)
			}

			result := ForwardHeaders(c, tt.forwardKeys...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
