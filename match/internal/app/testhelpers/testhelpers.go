package testhelpers

import (
	"bytes"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func SetupTestContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return c, w
}

func SetupTestContextWithBody(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func SetupAuthContext(c *gin.Context, userID, email string) {
	if userID != "" {
		c.Request.Header.Set("X-User-Id", userID)
	}
	if email != "" {
		c.Request.Header.Set("X-User-Email", email)
	}
}
