package general

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockMatchService struct {
	mock.Mock
}

type MockSwipeService struct {
	mock.Mock
}

type MockCache struct {
	mock.Mock
}

type MockPublisher struct {
	mock.Mock
}

func TestGeneralService_HandleSwipeRight(t *testing.T) {
	tests := []struct {
		name      string
		swiperID  int64
		swipedID  int64
		mockSetup func(*MockSwipeService, *MockMatchService, *MockCache, *MockPublisher)
		wantMatch bool
		wantErr   bool
	}{
		{
			name:     "first right swipe - no match",
			swiperID: 100,
			swipedID: 200,
			mockSetup: func(swipeSvc *MockSwipeService, matchSvc *MockMatchService, cache *MockCache, pub *MockPublisher) {
				swipeSvc.On("FindSwipe", mock.Anything, int64(100), int64(200)).
					Return("", nil)
				swipeSvc.On("Create", mock.Anything, int64(100), int64(200), "right").
					Return(nil, nil)
				cache.On("Increment", mock.Anything, "pop:200", int64(1)).
					Return(int64(1), nil)
			},
			wantMatch: false,
			wantErr:   false,
		},
		{
			name:     "mutual right swipe - match created",
			swiperID: 100,
			swipedID: 200,
			mockSetup: func(swipeSvc *MockSwipeService, matchSvc *MockMatchService, cache *MockCache, pub *MockPublisher) {
				swipeSvc.On("FindSwipe", mock.Anything, int64(100), int64(200)).
					Return("right", nil)
				swipeSvc.On("Delete", mock.Anything, int64(200), int64(100)).
					Return(nil)
				matchSvc.On("Create", mock.Anything, int64(100), int64(200)).
					Return(struct{ ID int64 }{ID: 1}, nil)
				cache.On("Increment", mock.Anything, "pop:200", int64(2)).
					Return(int64(2), nil)
				pub.On("Publish", mock.Anything, mock.Anything).
					Return(nil)
			},
			wantMatch: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestGeneralService_HandleSwipeLeft(t *testing.T) {
	tests := []struct {
		name      string
		swiperID  int64
		swipedID  int64
		mockSetup func(*MockSwipeService, *MockCache)
		wantErr   bool
	}{
		{
			name:     "first left swipe",
			swiperID: 100,
			swipedID: 200,
			mockSetup: func(swipeSvc *MockSwipeService, cache *MockCache) {
				swipeSvc.On("FindSwipe", mock.Anything, int64(100), int64(200)).
					Return("", nil)
				swipeSvc.On("Create", mock.Anything, int64(100), int64(200), "left").
					Return(nil, nil)
				cache.On("Increment", mock.Anything, "pop:200", int64(-1)).
					Return(int64(-1), nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
