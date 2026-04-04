package general

import (
	"context"
	"errors"
	"fmt"
	matchService "winx-match/internal/app/domain/services/match"
	swipeService "winx-match/internal/app/domain/services/swipe"
	"winx-match/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("match not found")

type Service interface {
	HandleSwipeRight(ctx context.Context, swiperID, swipedID int64) (models.Match, bool, error)
	HandleSwipeLeft(ctx context.Context, swiperID, swipedID int64) error
}

type service struct {
	matchService matchService.Service
	swipeService swipeService.Service
}

func NewService(db *gorm.DB) Service {
	return &service{
		matchService: matchService.NewService(db),
		swipeService: swipeService.NewService(db),
	}
}

func (s *service) HandleSwipeRight(ctx context.Context, swiperID, swipedID int64) (models.Match, bool, error) {
	direction, err := s.swipeService.FindSwipe(ctx, swiperID, swipedID)
	if err != nil {
		return models.Match{}, false, fmt.Errorf("handle swipe right: %w", err)
	}

	if direction == "" {
		if _, err := s.swipeService.Create(ctx, swiperID, swipedID, direction); err != nil {
			return models.Match{}, false, fmt.Errorf("create swipe: %w", err)
		}
		return models.Match{}, false, nil
	}

	if err := s.swipeService.Delete(ctx, swiperID, swipedID); err != nil {
		return models.Match{}, false, fmt.Errorf("delete swipe: %w", err)
	}

	if direction == "right" {
		match, err := s.matchService.Create(ctx, swiperID, swipedID)
		if err != nil {
			return models.Match{}, false, fmt.Errorf("create swipe: %w", err)
		}
		return match, true, nil
	}
	return models.Match{}, false, nil
}

func (s *service) HandleSwipeLeft(ctx context.Context, swiperID, swipedID int64) error {
	direction, err := s.swipeService.FindSwipe(ctx, swiperID, swipedID)
	if err != nil {
		return fmt.Errorf("handle swipe right: %w", err)
	}

	if direction == "" {
		if _, err := s.swipeService.Create(ctx, swiperID, swipedID, direction); err != nil {
			return fmt.Errorf("create swipe: %w", err)
		}
		return nil
	}

	if err := s.swipeService.Delete(ctx, swiperID, swipedID); err != nil {
		return fmt.Errorf("delete swipe: %w", err)
	}

	return nil
}
