package swipe

import (
	"context"
	"errors"
	swipeRepoDto "winx-match/internal/app/domain/core/dto/repositories/swipe"
	swipeRepo "winx-match/internal/app/domain/repositories/swipe"
	"winx-match/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("swipe not found")

// Service — интерфейс сервиса свайпов.
// Добавляй новые методы сюда и реализуй в service struct.
type Service interface {
	FindSwipe(ctx context.Context, swiperID, swipedID int64) (string, error)
	// FindOwn returns the direction of swiperID's own swipe on swipedID, or "" if none exists.
	FindOwn(ctx context.Context, swiperID, swipedID int64) (string, error)
	Create(ctx context.Context, swiperID, swipedID int64, direction string) (models.Swipe, error)
	Delete(ctx context.Context, swiperID, swipedID int64) error
}

type service struct {
	swipeRepo swipeRepo.Repository
}

func NewService(db *gorm.DB) Service {
	return &service{
		swipeRepo: swipeRepo.NewRepository(db),
	}
}

func (s *service) FindOwn(ctx context.Context, swiperID, swipedID int64) (string, error) {
	return s.swipeRepo.FindOwn(ctx, swiperID, swipedID)
}

func (s *service) FindSwipe(ctx context.Context, swiperID, swipedID int64) (string, error) {
	// Проверяем, свайпнул ли swipedID вправо на swiperID ранее
	direction, err := s.swipeRepo.FindMutual(ctx, swipeRepoDto.FindDTO{
		SwiperID: swipedID,
		SwipedID: swiperID,
	})

	return direction, err
}

func (s *service) Create(ctx context.Context, swiperID, swipedID int64, direction string) (models.Swipe, error) {
	return s.swipeRepo.Create(ctx, swipeRepoDto.CreateDTO{
		SwiperID:  swiperID,
		SwipedID:  swipedID,
		Direction: direction,
	})
}

func (s *service) Delete(ctx context.Context, swiperID, swipedID int64) error {
	return s.swipeRepo.Delete(ctx, swipeRepoDto.DeleteDTO{
		SwiperID: swiperID,
		SwipedID: swipedID,
	})
}
