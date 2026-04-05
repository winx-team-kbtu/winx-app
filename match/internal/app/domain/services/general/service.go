package general

import (
	"context"
	"errors"
	"fmt"

	matchService "winx-match/internal/app/domain/services/match"
	swipeService "winx-match/internal/app/domain/services/swipe"
	"winx-match/internal/app/models/models"
	"winx-match/pkg/graylog/logger"

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
	logger.LogInfo("Match service initialized")
	return &service{
		matchService: matchService.NewService(db),
		swipeService: swipeService.NewService(db),
	}
}

func (s *service) HandleSwipeRight(ctx context.Context, swiperID, swipedID int64) (models.Match, bool, error) {
	logger.LogInfo(fmt.Sprintf("Processing right swipe from %d to %d", swiperID, swipedID))

	direction, err := s.swipeService.FindSwipe(ctx, swipedID, swiperID)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to find swipe between %d and %d", swipedID, swiperID), err)
		return models.Match{}, false, fmt.Errorf("handle swipe right: %w", err)
	}

	if direction == "" {
		logger.LogDebug(fmt.Sprintf("No existing swipe from %d to %d, recording new swipe", swipedID, swiperID))
		if _, err := s.swipeService.Create(ctx, swiperID, swipedID, "right"); err != nil {
			logger.LogError("Failed to create swipe record", err)
			return models.Match{}, false, fmt.Errorf("create swipe: %w", err)
		}
		return models.Match{}, false, nil
	}

	if err := s.swipeService.Delete(ctx, swipedID, swiperID); err != nil {
		logger.LogError("Failed to delete existing swipe record", err)
		return models.Match{}, false, fmt.Errorf("delete swipe: %w", err)
	}

	if direction == "right" {
		logger.LogInfo(fmt.Sprintf("Match created between %d and %d", swiperID, swipedID))
		match, err := s.matchService.Create(ctx, swiperID, swipedID)
		if err != nil {
			logger.LogError("Failed to create match", err)
			return models.Match{}, false, fmt.Errorf("create match: %w", err)
		}
		return match, true, nil
	}

	logger.LogDebug(fmt.Sprintf("Swipe from %d to %d was %s, no match created", swipedID, swiperID, direction))
	return models.Match{}, false, nil
}

func (s *service) HandleSwipeLeft(ctx context.Context, swiperID, swipedID int64) error {
	logger.LogInfo(fmt.Sprintf("Processing left swipe from %d to %d", swiperID, swipedID))

	direction, err := s.swipeService.FindSwipe(ctx, swipedID, swiperID)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to find swipe between %d and %d", swipedID, swiperID), err)
		return fmt.Errorf("handle swipe left: %w", err)
	}

	if direction == "" {
		logger.LogDebug(fmt.Sprintf("No existing swipe from %d to %d, recording left swipe", swipedID, swiperID))
		if _, err := s.swipeService.Create(ctx, swiperID, swipedID, "left"); err != nil {
			logger.LogError("Failed to create left swipe record", err)
			return fmt.Errorf("create swipe: %w", err)
		}
		return nil
	}

	if err := s.swipeService.Delete(ctx, swipedID, swiperID); err != nil {
		logger.LogError("Failed to delete existing swipe record", err)
		return fmt.Errorf("delete swipe: %w", err)
	}

	logger.LogDebug(fmt.Sprintf("Left swipe from %d to %d, existing swipe deleted", swiperID, swipedID))
	return nil
}