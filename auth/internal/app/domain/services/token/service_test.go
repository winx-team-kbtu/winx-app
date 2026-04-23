package token

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/stretchr/testify/assert"
)

func TestTokenService_IssueToken(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]string
		mockServer func() *httptest.Server
		wantErr    bool
	}{
		{
			name: "successful token issuance",
			params: map[string]string{
				"grant_type": "password",
				"username":   "test@example.com",
				"password":   "password123",
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`))
				}))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestTokenService_ValidateToken(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		mockSetup func(*OAuthServerMock)
		wantErr   bool
	}{
		{
			name:  "valid token",
			token: "Bearer valid-token",
			mockSetup: func(mock *OAuthServerMock) {
				mock.ValidateBearerTokenFn = func(r *http.Request) (oauth2.TokenInfo, error) {
					return &TokenInfoMock{AccessToken: "valid-token", UserID: "1"}, nil
				}
			},
			wantErr: false,
		},
		{
			name:  "invalid token",
			token: "Bearer invalid-token",
			mockSetup: func(mock *OAuthServerMock) {
				mock.ValidateBearerTokenFn = func(r *http.Request) (oauth2.TokenInfo, error) {
					return nil, assert.AnError
				}
			},
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := &OAuthServerMock{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockServer)
			}

			service := NewService(mockServer)
			_, err := service.ValidateToken(context.Background(), tt.token)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
