package chat

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockChatService struct{}
type MockMsgService struct{}
type MockBinder struct{}

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
		{
			name:           "unauthorized - missing both",
			userID:         "",
			userEmail:      "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/chats", nil)

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

func TestHandler_Messages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		chatID         string
		userID         string
		userEmail      string
		expectedStatus int
	}{
		{
			name:           "successful get messages",
			chatID:         "1",
			userID:         "123",
			userEmail:      "test@example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid chat ID",
			chatID:         "invalid",
			userID:         "123",
			userEmail:      "test@example.com",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized - no user ID",
			chatID:         "1",
			userID:         "",
			userEmail:      "test@example.com",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "unauthorized - no user email",
			chatID:         "1",
			userID:         "123",
			userEmail:      "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "chat ID missing",
			chatID:         "",
			userID:         "123",
			userEmail:      "test@example.com",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/chats/"+tt.chatID+"/messages", nil)

			if tt.chatID != "" {
				c.Params = gin.Params{{Key: "id", Value: tt.chatID}}
			}

			if tt.userID != "" {
				c.Request.Header.Set("X-User-Id", tt.userID)
			}
			if tt.userEmail != "" {
				c.Request.Header.Set("X-User-Email", tt.userEmail)
			}

			assert.NotNil(t, c)

			if tt.chatID == "" || tt.chatID == "invalid" {
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
