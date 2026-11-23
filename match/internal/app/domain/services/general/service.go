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

var ErrUnavailableUser = errors.New("Этот профиль больше недоступен.")

// MatchPublisher receives new matches after the database transaction commits.
type MatchPublisher interface {
	Publish(ctx context.Context, match models.Match) error
}

type Service interface {
	HandleSwipeRight(ctx context.Context, swiperID, swipedID int64) (models.Match, bool, error)
	HandleSwipeLeft(ctx context.Context, swiperID, swipedID int64) error
}

type service struct {
	db             *gorm.DB
	cache          cache.Cache
	matchPublisher MatchPublisher
}

func NewService(db *gorm.DB, c cache.Cache, publisher MatchPublisher) Service {
	return &service{db: db, cache: c, matchPublisher: publisher}
}

func (s *service) HandleSwipeRight(ctx context.Context, swiperID, swipedID int64) (models.Match, bool, error) {
	match, created, err := s.recordSwipe(ctx, swiperID, swipedID, "right")
	if err == nil && match.ID != 0 && created && s.matchPublisher != nil {
		if err := s.matchPublisher.Publish(ctx, match); err != nil {
			fmt.Printf("warn: publish match event: %v\n", err)
		}
	}
	return match, created, err
}

func (s *service) HandleSwipeLeft(ctx context.Context, swiperID, swipedID int64) error {
	_, _, err := s.recordSwipe(ctx, swiperID, swipedID, "left")
	return err
}

func (s *service) recordSwipe(ctx context.Context, swiper, peer int64, direction string) (models.Match, bool, error) {
	if swiper == peer {
		return models.Match{}, false, ErrUnavailableUser
	}

	var match models.Match
	var created bool
	var popularity int64
	first, second := min(swiper, peer), max(swiper, peer)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Both directions take the same pair lock before reading swipes.
		key := fmt.Sprintf("winx-social:%d:%d", first, second)
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended(?,0))`, key).Error; err != nil {
			return err
		}
		var active []int64
		if err := tx.Raw(`SELECT id FROM users WHERE id IN (?,?) AND deletion_requested_at IS NULL ORDER BY id FOR SHARE`, first, second).Scan(&active).Error; err != nil {
			return err
		}
		if len(active) != 2 {
			return ErrUnavailableUser
		}
		if err := tx.Raw(`SELECT * FROM matches WHERE (user_one_id=? AND user_two_id=?) OR (user_one_id=? AND user_two_id=?) LIMIT 1 FOR KEY SHARE`,
			first, second, second, first).Scan(&match).Error; err != nil {
			return err
		}
		if match.ID != 0 {
			if direction == "right" {
				return openMatchChat(tx, &match)
			}
			return nil
		}

		swipes := swipeService.NewService(tx)
		own, err := swipes.FindOwn(ctx, swiper, peer)
		if err != nil {
			return err
		}
		if own != "" && (own != "right" || direction != "right") {
			return nil
		}
		theirs, err := swipes.FindSwipe(ctx, swiper, peer)
		if err != nil {
			return err
		}
		if own == "" {
			if _, err := swipes.Create(ctx, swiper, peer, direction); err != nil {
				return err
			}
			popularity = 1
		}
		if direction == "left" {
			popularity = -1
			if theirs != "" {
				return swipes.Delete(ctx, peer, swiper)
			}
			return nil
		}
		if theirs != "right" {
			return nil
		}
		match, err = matchService.NewService(tx).Create(ctx, first, second)
		if err != nil {
			return err
		}
		created = true
		popularity = 2
		return openMatchChat(tx, &match)
	})
	if err != nil {
		return models.Match{}, false, err
	}

	if popularity != 0 {
		_, _ = s.cache.Increment(ctx, popularityKey(peer), popularity)
		_ = s.cache.Delete(ctx, recCacheKey(swiper))
	}
	if created {
		_ = s.cache.Delete(ctx, recCacheKey(peer))
	}
	return match, created, nil
}

func openMatchChat(tx *gorm.DB, match *models.Match) error {
	if err := tx.Exec(`INSERT INTO chats(match_id,user_one_id,user_two_id) VALUES(?,?,?)
 ON CONFLICT(match_id) DO NOTHING`, match.ID, match.UserOneID, match.UserTwoID).Error; err != nil {
		return err
	}
	return tx.Raw("SELECT id FROM chats WHERE match_id=?", match.ID).Scan(&match.ChatID).Error
}

func popularityKey(userID int64) string {
	return fmt.Sprintf("pop:%d", userID)
}

// recCacheKey must match the recommendation service's cache prefix.
func recCacheKey(userID int64) string {
	return fmt.Sprintf("rec:%d", userID)
}
