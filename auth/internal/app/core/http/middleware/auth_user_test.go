package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestApiKeyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "valid API key",
			apiKey:         "valid-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid API key",
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing API key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			if tt.apiKey != "" {
				c.Request.Header.Set("x-api-key", tt.apiKey)
			}

		})
	}
}

func TestRecoveryWithLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("panic recovery", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		handler := RecoveryWithLogger()

		assert.NotNil(t, handler)
	})
}
