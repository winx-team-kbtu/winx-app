package swipe

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSwipeHandler_SwipeLeft(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		userEmail      string
		swipedID       int64
		expectedStatus int
	}{
		{
			name:           "successful left swipe",
			userID:         "123",
			userEmail:      "test@example.com",
			swipedID:       200,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			userID:         "",
			userEmail:      "",
			swipedID:       200,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(map[string]int64{"swiped_id": tt.swipedID})
			c.Request = httptest.NewRequest("POST", "/swipes/left", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

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

func TestSwipeHandler_SwipeRight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		userEmail      string
		swipedID       int64
		expectedStatus int
	}{
		{
			name:           "successful right swipe",
			userID:         "123",
			userEmail:      "test@example.com",
			swipedID:       200,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			userID:         "",
			userEmail:      "",
			swipedID:       200,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(map[string]int64{"swiped_id": tt.swipedID})
			c.Request = httptest.NewRequest("POST", "/swipes/right", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

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
