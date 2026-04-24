package swipe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockSwipeRepository struct {
	mock.Mock
}

func (m *MockSwipeRepository) Create(ctx context.Context, input interface{}) (interface{}, error) {
	args := m.Called(ctx, input)
	return args.Get(0), args.Error(1)
}

func (m *MockSwipeRepository) FindMutual(ctx context.Context, input interface{}) (string, error) {
	args := m.Called(ctx, input)
	return args.String(0), args.Error(1)
}

func (m *MockSwipeRepository) Delete(ctx context.Context, input interface{}) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func TestSwipeService_Create(t *testing.T) {
	tests := []struct {
		name      string
		swiperID  int64
		swipedID  int64
		direction string
		mockSetup func(*MockSwipeRepository)
		wantErr   bool
	}{
		{
			name:      "successful left swipe",
			swiperID:  100,
			swipedID:  200,
			direction: "left",
			mockSetup: func(repo *MockSwipeRepository) {
				repo.On("Create", mock.Anything, mock.Anything).
					Return(struct{ ID int64 }{ID: 1}, nil)
			},
			wantErr: false,
		},
		{
			name:      "successful right swipe",
			swiperID:  100,
			swipedID:  200,
			direction: "right",
			mockSetup: func(repo *MockSwipeRepository) {
				repo.On("Create", mock.Anything, mock.Anything).
					Return(struct{ ID int64 }{ID: 1}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockSwipeRepository)
			tt.mockSetup(mockRepo)

		})
	}
}

func TestSwipeService_FindSwipe(t *testing.T) {
	tests := []struct {
		name          string
		swiperID      int64
		swipedID      int64
		mockDirection string
		mockError     error
		wantDirection string
		wantErr       bool
	}{
		{
			name:          "existing right swipe",
			swiperID:      100,
			swipedID:      200,
			mockDirection: "right",
			mockError:     nil,
			wantDirection: "right",
			wantErr:       false,
		},
		{
			name:          "no swipe exists",
			swiperID:      100,
			swipedID:      200,
			mockDirection: "",
			mockError:     nil,
			wantDirection: "",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockSwipeRepository)
			mockRepo.On("FindMutual", mock.Anything, mock.Anything).
				Return(tt.mockDirection, tt.mockError)

		})
	}
}

func TestSwipeService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		swiperID  int64
		swipedID  int64
		mockSetup func(*MockSwipeRepository)
		wantErr   bool
	}{
		{
			name:     "successful delete",
			swiperID: 100,
			swipedID: 200,
			mockSetup: func(repo *MockSwipeRepository) {
				repo.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockSwipeRepository)
			tt.mockSetup(mockRepo)

		})
	}
}
