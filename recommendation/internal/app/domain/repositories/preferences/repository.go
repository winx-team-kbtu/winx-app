package preferences

import (
	"context"
	"errors"
	"fmt"

	"winx-recommendation/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("preferences not found")

// UserContext holds everything the algorithm needs about the requesting user
// that lives in Postgres.
type UserContext struct {
	Preferences models.MatchingPreferences
	InterestIDs []int64
	// Location taken from profiles.current_location (if set)
	Latitude  *float64
	Longitude *float64
}

type Repository interface {
	GetUserContext(ctx context.Context, userID int64) (UserContext, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// GetUserContext fetches matching_preferences, interest IDs and location
// for the given user in three lightweight queries.
func (r *repository) GetUserContext(ctx context.Context, userID int64) (UserContext, error) {
	var uc UserContext

	// 1. Matching preferences (may not exist yet — use defaults in that case)
	var prefs models.MatchingPreferences
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&prefs).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return UserContext{}, fmt.Errorf("get matching preferences: %w", err)
		}
		// sensible defaults
		prefs = models.MatchingPreferences{
			UserID:        userID,
			MinAge:        18,
			MaxAge:        99,
			MaxDistanceKM: 50,
			ShowMeGlobal:  false,
		}
	}
	uc.Preferences = prefs

	// 2. Interest IDs
	var interests []models.UserInterest
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&interests).Error; err != nil {
		return UserContext{}, fmt.Errorf("get user interests: %w", err)
	}
	uc.InterestIDs = make([]int64, 0, len(interests))
	for _, i := range interests {
		uc.InterestIDs = append(uc.InterestIDs, i.InterestID)
	}

	// 3. Current location from profiles table
	type locationRow struct {
		Longitude *float64
		Latitude  *float64
	}
	var loc locationRow
	err = r.db.WithContext(ctx).Raw(`
		SELECT
			ST_X(current_location::geometry) AS longitude,
			ST_Y(current_location::geometry) AS latitude
		FROM profiles
		WHERE user_id = ? AND current_location IS NOT NULL
		LIMIT 1
	`, userID).Scan(&loc).Error
	if err != nil {
		return UserContext{}, fmt.Errorf("get user location: %w", err)
	}
	uc.Latitude = loc.Latitude
	uc.Longitude = loc.Longitude

	return uc, nil
}
