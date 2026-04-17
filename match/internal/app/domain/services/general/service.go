package general

import (
	"context"
	"errors"
	"fmt"
	matchService "winx-match/internal/app/domain/services/match"
	swipeService "winx-match/internal/app/domain/services/swipe"
	"winx-match/internal/app/models/models"
	"winx-match/pkg/cache"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("match not found")

// MatchPublisher is notified whenever a new match is created.
// Implementations should be non-blocking and treat errors as best-effort.
type MatchPublisher interface {
	Publish(ctx context.Context, match models.Match) error
}

type Service interface {
	HandleSwipeRight(ctx context.Context, swiperID, swipedID int64) (models.Match, bool, error)
	HandleSwipeLeft(ctx context.Context, swiperID, swipedID int64) error
}

type service struct {
	matchService   matchService.Service
	swipeService   swipeService.Service
	cache          cache.Cache
	matchPublisher MatchPublisher
}

func NewService(db *gorm.DB, c cache.Cache, publisher MatchPublisher) Service {
	return &service{
		matchService:   matchService.NewService(db),
		swipeService:   swipeService.NewService(db),
		cache:          c,
		matchPublisher: publisher,
	}
}

func (s *service) HandleSwipeRight(ctx context.Context, swiperID, swipedID int64) (models.Match, bool, error) {
	// Check whether the other user already swiped right on us.
	direction, err := s.swipeService.FindSwipe(ctx, swiperID, swipedID)
	if err != nil {
		return models.Match{}, false, fmt.Errorf("handle swipe right: %w", err)
	}

	if direction != "right" {
		// No mutual right-swipe yet — record ours, boost the other user's popularity.
		if _, err := s.swipeService.Create(ctx, swiperID, swipedID, "right"); err != nil {
			return models.Match{}, false, fmt.Errorf("create swipe: %w", err)
		}
		_, _ = s.cache.Increment(ctx, popularityKey(swipedID), 1)
		return models.Match{}, false, nil
	}

	// Mutual right-swipe — clean up the other user's pending swipe then create the match.
	if err := s.swipeService.Delete(ctx, swipedID, swiperID); err != nil {
		return models.Match{}, false, fmt.Errorf("delete swipe: %w", err)
	}

	match, err := s.matchService.Create(ctx, swiperID, swipedID)
	if err != nil {
		return models.Match{}, false, fmt.Errorf("create match: %w", err)
	}
	// Both users swiped right — double boost for the matched user.
	_, _ = s.cache.Increment(ctx, popularityKey(swipedID), 2)
	// Publish event best-effort — a publish failure must not roll back the match.
	if err := s.matchPublisher.Publish(ctx, match); err != nil {
		fmt.Printf("warn: publish match event: %v\n", err)
	}
	return match, true, nil
}

func (s *service) HandleSwipeLeft(ctx context.Context, swiperID, swipedID int64) error {
	direction, err := s.swipeService.FindSwipe(ctx, swiperID, swipedID)
	if err != nil {
		return fmt.Errorf("handle swipe left: %w", err)
	}

	if direction == "" {
		if _, err := s.swipeService.Create(ctx, swiperID, swipedID, "left"); err != nil {
			return fmt.Errorf("create swipe: %w", err)
		}
		_, _ = s.cache.Increment(ctx, popularityKey(swipedID), -1)
		return nil
	}

	if err := s.swipeService.Delete(ctx, swiperID, swipedID); err != nil {
		return fmt.Errorf("delete swipe: %w", err)
	}

	return nil
}

func popularityKey(userID int64) string {
	return fmt.Sprintf("pop:%d", userID)
}
