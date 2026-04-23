package user

import (
	"context"
	"testing"

	"auth/internal/app/domain/core/dto/services/user"
	"auth/internal/app/models/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, input interface{}) (models.User, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, input interface{}) (bool, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, input interface{}) (models.User, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(models.User), args.Error(1)
}

func TestUserService_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     user.CreateDTO
		mockSetup func(*MockUserRepository)
		wantErr   bool
	}{
		{
			name: "successful user creation",
			input: user.CreateDTO{
				Email:    "test@example.com",
				Password: "StrongPass123!",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("Create", mock.Anything, mock.Anything).Return(models.User{
					ID:    1,
					Email: "test@example.com",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			input: user.CreateDTO{
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("Create", mock.Anything, mock.Anything).Return(models.User{}, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

		})
	}
}

func TestUserService_GetByEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		mockUser  models.User
		mockError error
		wantErr   bool
	}{
		{
			name:  "existing user",
			email: "test@example.com",
			mockUser: models.User{
				ID:    1,
				Email: "test@example.com",
			},
			mockError: nil,
			wantErr:   false,
		},
		{
			name:      "non-existent user",
			email:     "nonexistent@example.com",
			mockUser:  models.User{},
			mockError: gorm.ErrRecordNotFound,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
