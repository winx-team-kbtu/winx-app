package message

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(ctx context.Context, input interface{}) (interface{}, error) {
	args := m.Called(ctx, input)
	return args.Get(0), args.Error(1)
}

func (m *MockMessageRepository) ListByChatID(ctx context.Context, input interface{}) ([]interface{}, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]interface{}), args.Error(1)
}

type MockChatRepositoryForMsg struct {
	mock.Mock
}

func (m *MockChatRepositoryForMsg) FindByID(ctx context.Context, id int64) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func TestMessageService_Create(t *testing.T) {
	tests := []struct {
		name      string
		chatID    int64
		senderID  int64
		content   string
		mockSetup func(*MockChatRepositoryForMsg, *MockMessageRepository)
		wantErr   bool
	}{
		{
			name:     "successful message creation",
			chatID:   1,
			senderID: 100,
			content:  "Hello, world!",
			mockSetup: func(chatRepo *MockChatRepositoryForMsg, msgRepo *MockMessageRepository) {
				chatRepo.On("FindByID", mock.Anything, int64(1)).
					Return(struct {
						UserOneID int64
						UserTwoID int64
					}{UserOneID: 100, UserTwoID: 200}, nil)
				msgRepo.On("Create", mock.Anything, mock.Anything).
					Return(struct{ ID int64 }{ID: 1}, nil)
			},
			wantErr: false,
		},
		{
			name:     "user not a member",
			chatID:   1,
			senderID: 300,
			content:  "Hello",
			mockSetup: func(chatRepo *MockChatRepositoryForMsg, msgRepo *MockMessageRepository) {
				chatRepo.On("FindByID", mock.Anything, int64(1)).
					Return(struct {
						UserOneID int64
						UserTwoID int64
					}{UserOneID: 100, UserTwoID: 200}, nil)
			},
			wantErr: true,
		},
		{
			name:     "chat not found",
			chatID:   999,
			senderID: 100,
			content:  "Hello",
			mockSetup: func(chatRepo *MockChatRepositoryForMsg, msgRepo *MockMessageRepository) {
				chatRepo.On("FindByID", mock.Anything, int64(999)).
					Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockChatRepo := new(MockChatRepositoryForMsg)
			mockMsgRepo := new(MockMessageRepository)
			tt.mockSetup(mockChatRepo, mockMsgRepo)

		})
	}
}

func TestMessageService_List(t *testing.T) {
	tests := []struct {
		name      string
		chatID    int64
		userID    int64
		limit     int
		offset    int
		mockSetup func(*MockChatRepositoryForMsg, *MockMessageRepository)
		wantErr   bool
	}{
		{
			name:   "successful list messages",
			chatID: 1,
			userID: 100,
			limit:  50,
			offset: 0,
			mockSetup: func(chatRepo *MockChatRepositoryForMsg, msgRepo *MockMessageRepository) {
				chatRepo.On("FindByID", mock.Anything, int64(1)).
					Return(struct {
						UserOneID int64
						UserTwoID int64
					}{UserOneID: 100, UserTwoID: 200}, nil)
				msgRepo.On("ListByChatID", mock.Anything, mock.Anything).
					Return([]interface{}{}, nil)
			},
			wantErr: false,
		},
		{
			name:   "user not a member",
			chatID: 1,
			userID: 300,
			limit:  50,
			offset: 0,
			mockSetup: func(chatRepo *MockChatRepositoryForMsg, msgRepo *MockMessageRepository) {
				chatRepo.On("FindByID", mock.Anything, int64(1)).
					Return(struct {
						UserOneID int64
						UserTwoID int64
					}{UserOneID: 100, UserTwoID: 200}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockChatRepo := new(MockChatRepositoryForMsg)
			mockMsgRepo := new(MockMessageRepository)
			tt.mockSetup(mockChatRepo, mockMsgRepo)

		})
	}
}
