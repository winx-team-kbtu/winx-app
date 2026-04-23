package chat

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) List(ctx context.Context, query url.Values, headers map[string]string) (Response, error) {
	args := m.Called(ctx, query, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) Messages(ctx context.Context, chatID string, query url.Values, headers map[string]string) (Response, error) {
	args := m.Called(ctx, chatID, query, headers)
	return args.Get(0).(Response), args.Error(1)
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name     string
		query    url.Values
		headers  map[string]string
		mockResp Response
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful list chats",
			query:   url.Values{"limit": []string{"10"}},
			headers: map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"chats":[{"id":1,"name":"Chat 1"}]}`),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "empty chats list",
			query:   url.Values{},
			headers: map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"chats":[]}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("List", mock.Anything, tt.query, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.List(context.Background(), tt.query, tt.headers)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResp.StatusCode, resp.StatusCode)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestService_Messages(t *testing.T) {
	tests := []struct {
		name     string
		chatID   string
		query    url.Values
		headers  map[string]string
		mockResp Response
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get messages",
			chatID:  "123",
			query:   url.Values{"limit": []string{"50"}},
			headers: map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"messages":[{"id":1,"text":"Hello"}]}`),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "chat not found",
			chatID:  "999",
			query:   url.Values{},
			headers: map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusNotFound,
				ContentType: "application/json",
				Body:        []byte(`{"error":"chat not found"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("Messages", mock.Anything, tt.chatID, tt.query, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.Messages(context.Background(), tt.chatID, tt.query, tt.headers)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResp.StatusCode, resp.StatusCode)
			}
			mockClient.AssertExpectations(t)
		})
	}
}
