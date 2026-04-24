package match

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockMatchRepository struct {
	mock.Mock
}

func (m *MockMatchRepository) Create(ctx context.Context, input interface{}) (interface{}, error) {
	args := m.Called(ctx, input)
	return args.Get(0), args.Error(1)
}

func (m *MockMatchRepository) ListByUserID(ctx context.Context, input interface{}) ([]interface{}, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockMatchRepository) DeleteByIDAndUserID(ctx context.Context, id, userID int64) (bool, error) {
	args := m.Called(ctx, id, userID)
	return args.Bool(0), args.Error(1)
}

func TestMatchService_Create(t *testing.T) {
	tests := []struct {
		name      string
		userOneID int64
		userTwoID int64
		mockSetup func(*MockMatchRepository)
		wantErr   bool
	}{
		{
			name:      "successful match creation",
			userOneID: 100,
			userTwoID: 200,
			mockSetup: func(repo *MockMatchRepository) {
				repo.On("Create", mock.Anything, mock.Anything).
					Return(struct{ ID int64 }{ID: 1}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMatchRepository)
			tt.mockSetup(mockRepo)

		})
	}
}

func TestMatchService_List(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		limit     int
		offset    int
		mockSetup func(*MockMatchRepository)
		wantErr   bool
	}{
		{
			name:   "successful list",
			userID: 100,
			limit:  20,
			offset: 0,
			mockSetup: func(repo *MockMatchRepository) {
				repo.On("ListByUserID", mock.Anything, mock.Anything).
					Return([]interface{}{}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMatchRepository)
			tt.mockSetup(mockRepo)

		})
	}
}

func TestMatchService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		matchID   int64
		userID    int64
		mockSetup func(*MockMatchRepository)
		wantErr   bool
	}{
		{
			name:    "successful delete",
			matchID: 1,
			userID:  100,
			mockSetup: func(repo *MockMatchRepository) {
				repo.On("DeleteByIDAndUserID", mock.Anything, int64(1), int64(100)).
					Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:    "match not found",
			matchID: 999,
			userID:  100,
			mockSetup: func(repo *MockMatchRepository) {
				repo.On("DeleteByIDAndUserID", mock.Anything, int64(999), int64(100)).
					Return(false, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMatchRepository)
			tt.mockSetup(mockRepo)

		})
	}
}
