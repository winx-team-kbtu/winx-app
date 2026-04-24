package recommendation

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
			name:    "successful list recommendations",
			headers: map[string]string{"Authorization": "Bearer token"},
			query:   url.Values{"limit": []string{"20"}},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"recommendations":[{"id":1,"name":"Jane"}]}`),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "empty recommendations",
			headers: map[string]string{"Authorization": "Bearer token"},
			query:   url.Values{},
			mockResp: Response{
				StatusCode:  http.StatusOK,
				ContentType: "application/json",
				Body:        []byte(`{"recommendations":[]}`),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "unauthorized",
			headers: map[string]string{},
			query:   url.Values{},
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
