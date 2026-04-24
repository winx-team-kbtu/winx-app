package auth

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

func (m *MockClient) Login(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) Register(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) Refresh(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) Check(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) Logout(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) ForgotPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) ResetPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) ChangePassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) VerifyPin(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, body, contentType, headers)
	return args.Get(0).(Response), args.Error(1)
}

func TestNewService(t *testing.T) {
	mockClient := new(MockClient)
	service := NewService(mockClient)
	assert.NotNil(t, service)
}

func TestService_Login(t *testing.T) {
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
			name:        "successful login",
			body:        []byte(`{"email":"test@example.com","password":"password123"}`),
			contentType: "application/json",
			headers:     map[string]string{},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"access_token":"token123","token_type":"Bearer"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:        "login with invalid credentials",
			body:        []byte(`{"email":"wrong@example.com","password":"wrong"}`),
			contentType: "application/json",
			headers:     map[string]string{},
			mockResp: Response{
				StatusCode:  http.StatusUnauthorized,
				ContentType: "application/json",
				Body:        []byte(`{"error":"invalid credentials"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("Login", mock.Anything, tt.body, tt.contentType, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.Login(context.Background(), tt.body, tt.contentType, tt.headers)

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

func TestService_Register(t *testing.T) {
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
			name:        "successful registration",
			body:        []byte(`{"email":"new@example.com","password":"StrongPass123!"}`),
			contentType: "application/json",
			headers:     map[string]string{},
			mockResp: Response{
				StatusCode:  http.StatusCreated,
				ContentType: "application/json",
				Body:        []byte(`{"id":1,"email":"new@example.com"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("Register", mock.Anything, tt.body, tt.contentType, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.Register(context.Background(), tt.body, tt.contentType, tt.headers)

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

func TestService_Refresh(t *testing.T) {
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
			name:        "successful refresh",
			body:        []byte(`{"refresh_token":"valid-refresh-token"}`),
			contentType: "application/json",
			headers:     map[string]string{},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"access_token":"new-token","refresh_token":"new-refresh"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("Refresh", mock.Anything, tt.body, tt.contentType, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.Refresh(context.Background(), tt.body, tt.contentType, tt.headers)

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

func TestService_ForgotPassword(t *testing.T) {
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
			name:        "successful forgot password",
			body:        []byte(`{"email":"user@example.com"}`),
			contentType: "application/json",
			headers:     map[string]string{},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"message":"reset email sent"}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("ForgotPassword", mock.Anything, tt.body, tt.contentType, tt.headers).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.ForgotPassword(context.Background(), tt.body, tt.contentType, tt.headers)

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
