package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) ListByRecipient(ctx context.Context, recipient string) ([]interface{}, error) {
	args := m.Called(ctx, recipient)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockNotificationRepository) DeleteByIDAndRecipient(ctx context.Context, id int64, recipient string) (bool, error) {
	args := m.Called(ctx, id, recipient)
	return args.Bool(0), args.Error(1)
}

func TestNotificationService_ListByRecipient(t *testing.T) {
	tests := []struct {
		name      string
		recipient string
		mockSetup func(*MockNotificationRepository)
		wantErr   bool
	}{
		{
			name:      "successful list",
			recipient: "user@example.com",
			mockSetup: func(repo *MockNotificationRepository) {
				repo.On("ListByRecipient", mock.Anything, "user@example.com").
					Return([]interface{}{}, nil)
			},
			wantErr: false,
		},
		{
			name:      "repository error",
			recipient: "user@example.com",
			mockSetup: func(repo *MockNotificationRepository) {
				repo.On("ListByRecipient", mock.Anything, "user@example.com").
					Return([]interface{}{}, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockNotificationRepository)
			tt.mockSetup(mockRepo)

		})
	}
}

func TestNotificationService_DeleteByIDAndRecipient(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		recipient string
		mockSetup func(*MockNotificationRepository)
		wantErr   bool
	}{
		{
			name:      "successful delete",
			id:        1,
			recipient: "user@example.com",
			mockSetup: func(repo *MockNotificationRepository) {
				repo.On("DeleteByIDAndRecipient", mock.Anything, int64(1), "user@example.com").
					Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:      "notification not found",
			id:        999,
			recipient: "user@example.com",
			mockSetup: func(repo *MockNotificationRepository) {
				repo.On("DeleteByIDAndRecipient", mock.Anything, int64(999), "user@example.com").
					Return(false, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockNotificationRepository)
			tt.mockSetup(mockRepo)

		})
	}
}
