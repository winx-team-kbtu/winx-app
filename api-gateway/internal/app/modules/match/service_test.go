package match

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) SwipeLeft(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) SwipeRight(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) List(ctx context.Context, headers map[string]string) (Response, error) {
	args := m.Called(ctx, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) Delete(ctx context.Context, id string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, id, headers)
	return args.Get(0).(Response), args.Error(1)
}

func TestService_SwipeLeft(t *testing.T) {
	tests := []struct {
		name        string
		body        []byte
		contentType string
		headers     map[string]string
		mockResp    Response
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful swipe left",
			body:        []byte(`{"target_user_id":123}`),
			contentType: "application/json",
			headers:     map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"message":"swipe recorded"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("SwipeLeft", mock.Anything, tt.body, tt.contentType, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.SwipeLeft(context.Background(), tt.body, tt.contentType, tt.headers)

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

func TestService_SwipeRight(t *testing.T) {
	tests := []struct {
		name        string
		body        []byte
		contentType string
		headers     map[string]string
		mockResp    Response
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful swipe right",
			body:        []byte(`{"target_user_id":123}`),
			contentType: "application/json",
			headers:     map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"message":"swipe recorded"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:        "match created",
			body:        []byte(`{"target_user_id":456}`),
			contentType: "application/json",
			headers:     map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusCreated,
				ContentType: "application/json",
				Body:        []byte(`{"message":"match created","match_id":789}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("SwipeRight", mock.Anything, tt.body, tt.contentType, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.SwipeRight(context.Background(), tt.body, tt.contentType, tt.headers)

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

func TestService_List(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		mockResp Response
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful list matches",
			headers: map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"matches":[{"id":1,"name":"Jane"}]}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("List", mock.Anything, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.List(context.Background(), tt.headers)

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

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		headers  map[string]string
		mockResp Response
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful delete match",
			id:      "123",
			headers: map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"message":"match deleted"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "match not found",
			id:      "999",
			headers: map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusNotFound,
				ContentType: "application/json",
				Body:        []byte(`{"error":"match not found"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("Delete", mock.Anything, tt.id, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.Delete(context.Background(), tt.id, tt.headers)

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
