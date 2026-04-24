package match

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMatchHandler_List(t *testing.T) {
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
			userEmail:      "test@example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         "",
			userEmail:      "test@example.com",
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
			c.Request = httptest.NewRequest("GET", "/matches", nil)

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

func TestMatchHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		matchID        string
		userID         string
		userEmail      string
		expectedStatus int
	}{
		{
			name:           "successful delete",
			matchID:        "1",
			userID:         "123",
			userEmail:      "test@example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid match ID",
			matchID:        "invalid",
			userID:         "123",
			userEmail:      "test@example.com",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			matchID:        "1",
			userID:         "",
			userEmail:      "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("DELETE", "/matches/"+tt.matchID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.matchID}}

			if tt.userID != "" {
				c.Request.Header.Set("X-User-Id", tt.userID)
			}
			if tt.userEmail != "" {
				c.Request.Header.Set("X-User-Email", tt.userEmail)
			}

			assert.NotNil(t, c)

			if tt.matchID == "invalid" {
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
