package middleware

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestApiKeyMiddleware_Exists(t *testing.T) {
	middleware := ApiKey()
	assert.NotNil(t, middleware)
	assert.IsType(t, gin.HandlerFunc(nil), middleware)
}

func TestRecoveryWithLogger_Exists(t *testing.T) {
	middleware := RecoveryWithLogger()
	assert.NotNil(t, middleware)
	assert.IsType(t, gin.HandlerFunc(nil), middleware)
}

func TestRequestLogger_Exists(t *testing.T) {
	middleware := RequestLogger()
	assert.NotNil(t, middleware)
	assert.IsType(t, gin.HandlerFunc(nil), middleware)
}

func TestNewAuthUserMiddleware_Exists(t *testing.T) {
	assert.NotNil(t, NewAuthUserMiddleware)
}
