package notification

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

func (m *MockClient) List(ctx context.Context, headers map[string]string, query url.Values) (Response, error) {
	args := m.Called(ctx, headers, query)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockClient) Delete(ctx context.Context, id string, headers map[string]string) (Response, error) {
	args := m.Called(ctx, id, headers)
	return args.Get(0).(Response), args.Error(1)
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		query    url.Values
		mockResp Response
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful list notifications",
			headers: map[string]string{"Authorization": "Bearer token"},
			query:   url.Values{"limit": []string{"10"}},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"notifications":[{"id":1,"message":"Welcome"}]}`),
			},
			mockErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			mockClient.On("List", mock.Anything, tt.headers, tt.query).
				Return(tt.mockResp, tt.mockErr)

			service := NewService(mockClient)
			resp, err := service.List(context.Background(), tt.headers, tt.query)

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
			name:    "successful delete notification",
			id:      "123",
			headers: map[string]string{"Authorization": "Bearer token"},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"message":"deleted"}`),
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
