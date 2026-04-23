package chat

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) Create(ctx context.Context, input interface{}) (interface{}, error) {
	args := m.Called(ctx, input)
	return args.Get(0), args.Error(1)
}

func (m *MockChatRepository) FindByMatchID(ctx context.Context, matchID int64) (interface{}, error) {
	args := m.Called(ctx, matchID)
	return args.Get(0), args.Error(1)
}

func (m *MockChatRepository) ListByUserID(ctx context.Context, input interface{}) ([]interface{}, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockChatRepository) FindByID(ctx context.Context, id int64) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func TestChatService_CreateFromMatch(t *testing.T) {
	tests := []struct {
		name      string
		matchID   int64
		userOneID int64
		userTwoID int64
		mockSetup func(*MockChatRepository)
		wantErr   bool
	}{
		{
			name:      "successful chat creation",
			matchID:   1,
			userOneID: 100,
			userTwoID: 200,
			mockSetup: func(repo *MockChatRepository) {
				repo.On("FindByMatchID", mock.Anything, int64(1)).
					Return(nil, errors.New("not found"))
				repo.On("Create", mock.Anything, mock.Anything).
					Return(struct{ ID int64 }{ID: 1}, nil)
			},
			wantErr: false,
		},
		{
			name:      "chat already exists",
			matchID:   1,
			userOneID: 100,
			userTwoID: 200,
			mockSetup: func(repo *MockChatRepository) {
				repo.On("FindByMatchID", mock.Anything, int64(1)).
					Return(struct{ ID int64 }{ID: 1}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			tt.mockSetup(mockRepo)

		})
	}
}

func TestChatService_List(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		limit     int
		offset    int
		mockSetup func(*MockChatRepository)
		wantErr   bool
	}{
		{
			name:   "successful list",
			userID: 100,
			limit:  20,
			offset: 0,
			mockSetup: func(repo *MockChatRepository) {
				repo.On("ListByUserID", mock.Anything, mock.Anything).
					Return([]interface{}{}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			tt.mockSetup(mockRepo)

		})
	}
}

func TestChatService_FindByID(t *testing.T) {
	tests := []struct {
		name      string
		chatID    int64
		mockSetup func(*MockChatRepository)
		wantErr   bool
	}{
		{
			name:   "chat found",
			chatID: 1,
			mockSetup: func(repo *MockChatRepository) {
				repo.On("FindByID", mock.Anything, int64(1)).
					Return(struct{ ID int64 }{ID: 1}, nil)
			},
			wantErr: false,
		},
		{
			name:   "chat not found",
			chatID: 999,
			mockSetup: func(repo *MockChatRepository) {
				repo.On("FindByID", mock.Anything, int64(999)).
					Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			tt.mockSetup(mockRepo)

		})
	}
}
