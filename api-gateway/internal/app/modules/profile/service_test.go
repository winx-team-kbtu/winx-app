package profile

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

func (m *MockClient) GetMe(ctx context.Context, headers map[string]string) (Response, error) {
	args := m.Called(ctx, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) Store(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) GetPhotos(ctx context.Context, headers map[string]string) (Response, error) {
	args := m.Called(ctx, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) StorePhoto(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) ListInterests(ctx context.Context, headers map[string]string, query url.Values) (Response, error) {
	args := m.Called(ctx, headers, query)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) LookupLocationByIP(ctx context.Context, headers map[string]string) (Response, error) {
	args := m.Called(ctx, headers)
	return args.Get(0).(Response), args.Error(1)
}

func TestService_GetMe(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		mockResp Response
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get profile",
			headers: map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"id":1,"name":"John","email":"john@example.com"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "unauthorized",
			headers: map[string]string{},
			mockResp: Response{
				StatusCode:  http.StatusUnauthorized,
				ContentType: "application/json",
				Body:        []byte(`{"error":"unauthorized"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("GetMe", mock.Anything, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.GetMe(context.Background(), tt.headers)

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

func TestService_Store(t *testing.T) {
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
			name:        "successful create/update profile",
			body:        []byte(`{"name":"John","bio":"Software engineer"}`),
			contentType: "application/json",
			headers:     map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"id":1,"name":"John"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("Store", mock.Anything, tt.body, tt.contentType, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.Store(context.Background(), tt.body, tt.contentType, tt.headers)

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

func TestService_GetPhotos(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		mockResp Response
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get photos",
			headers: map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"photos":["url1","url2"]}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("GetPhotos", mock.Anything, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.GetPhotos(context.Background(), tt.headers)

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

func TestService_ListInterests(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		query    url.Values
		mockResp Response
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful list interests",
			headers: map[string]string{"Authorization": "Bearer token"},
			query:   url.Values{"limit": []string{"10"}},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"interests":[{"id":1,"name":"Music"}]}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("ListInterests", mock.Anything, tt.headers, tt.query).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.ListInterests(context.Background(), tt.headers, tt.query)

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

func TestService_LookupLocationByIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		mockResp Response
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful location lookup",
			headers: map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"country":"USA","city":"New York"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("LookupLocationByIP", mock.Anything, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.LookupLocationByIP(context.Background(), tt.headers)

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
