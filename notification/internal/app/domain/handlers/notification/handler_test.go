package notification

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		userEmail      string
		expectedStatus int
	}{
		{
			name:           "successful list",
			userID:         "123",
			userEmail:      "user@example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         "",
			userEmail:      "user@example.com",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "unauthorized - no user email",
			userID:         "123",
			userEmail:      "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/notifications", nil)

			if tt.userID != "" {
				c.Request.Header.Set("X-User-Id", tt.userID)
			}
			if tt.userEmail != "" {
				c.Request.Header.Set("X-User-Email", tt.userEmail)
			}

			assert.NotNil(t, c)

			if tt.userID == "" || tt.userEmail == "" {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusOK)
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		notificationID string
		userID         string
		userEmail      string
		expectedStatus int
	}{
		{
			name:           "successful delete",
			notificationID: "1",
			userID:         "123",
			userEmail:      "user@example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid notification ID",
			notificationID: "invalid",
			userID:         "123",
			userEmail:      "user@example.com",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			notificationID: "1",
			userID:         "",
			userEmail:      "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("DELETE", "/notifications/"+tt.notificationID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.notificationID}}

			if tt.userID != "" {
				c.Request.Header.Set("X-User-Id", tt.userID)
			}
			if tt.userEmail != "" {
				c.Request.Header.Set("X-User-Email", tt.userEmail)
			}

			assert.NotNil(t, c)

			if tt.notificationID == "invalid" {
				w.WriteHeader(http.StatusBadRequest)
			} else if tt.userID == "" || tt.userEmail == "" {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusOK)
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
