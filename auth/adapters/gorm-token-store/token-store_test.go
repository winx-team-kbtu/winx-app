package gorm_token_store

import (
	"context"
	"testing"

	"github.com/go-oauth2/oauth2/v4"
	oauthModels "github.com/go-oauth2/oauth2/v4/models"
	"github.com/stretchr/testify/assert"
)

func TestNewGormTokenStore(t *testing.T) {
	store := &GormTokenStoreMock{}
	assert.NotNil(t, store)
}

func TestGormTokenStoreMock_Create(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*GormTokenStoreMock)
		wantErr   bool
	}{
		{
			name: "successful create",
			setupFunc: func(m *GormTokenStoreMock) {
				m.CreateFn = func(ctx context.Context, info oauth2.TokenInfo) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "create with error",
			setupFunc: func(m *GormTokenStoreMock) {
				m.CreateFn = func(ctx context.Context, info oauth2.TokenInfo) error {
					return assert.AnError
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &GormTokenStoreMock{}
			if tt.setupFunc != nil {
				tt.setupFunc(mock)
			}

			err := mock.Create(context.Background(), nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGormTokenStoreMock_GetByAccess(t *testing.T) {
	tests := []struct {
		name      string
		access    string
		setupFunc func(*GormTokenStoreMock)
		wantErr   bool
	}{
		{
			name:   "successful get by access",
			access: "valid-access-token",
			setupFunc: func(m *GormTokenStoreMock) {
				m.GetByAccessFn = func(ctx context.Context, access string) (oauth2.TokenInfo, error) {
					return &oauthModels.Token{}, nil
				}
			},
			wantErr: false,
		},
		{
			name:   "token not found",
			access: "invalid-token",
			setupFunc: func(m *GormTokenStoreMock) {
				m.GetByAccessFn = func(ctx context.Context, access string) (oauth2.TokenInfo, error) {
					return nil, assert.AnError
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &GormTokenStoreMock{}
			if tt.setupFunc != nil {
				tt.setupFunc(mock)
			}

			_, err := mock.GetByAccess(context.Background(), tt.access)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
