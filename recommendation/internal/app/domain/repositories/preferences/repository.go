package preferences

import (
	"context"
	"errors"
	"fmt"

	"winx-recommendation/internal/app/models/models"

	"gorm.io/gorm"
)

type prefsResult struct {
	prefs models.MatchingPreferences
	err   error
}

type interestsResult struct {
	ids []int64
	err error
}

type locationResult struct {
	lat, lon *float64
	err      error
}

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
// for the given user. All three queries run in parallel.
func (r *repository) GetUserContext(ctx context.Context, userID int64) (UserContext, error) {
	prefsCh := make(chan prefsResult, 1)
	interestsCh := make(chan interestsResult, 1)
	locationCh := make(chan locationResult, 1)

	go func() {
		var prefs models.MatchingPreferences
		err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&prefs).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			prefs = models.MatchingPreferences{
				UserID:        userID,
				MinAge:        18,
				MaxAge:        99,
				MaxDistanceKM: 50,
				ShowMeGlobal:  false,
			}
			err = nil
		}
		prefsCh <- prefsResult{prefs, err}
	}()

	go func() {
		var interests []models.UserInterest
		err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&interests).Error
		ids := make([]int64, 0, len(interests))
		for _, i := range interests {
			ids = append(ids, i.InterestID)
		}
		interestsCh <- interestsResult{ids, err}
	}()

	go func() {
		type locationRow struct {
			Longitude *float64
			Latitude  *float64
		}
		var loc locationRow
		err := r.db.WithContext(ctx).Raw(`
			SELECT
				ST_X(current_location::geometry) AS longitude,
				ST_Y(current_location::geometry) AS latitude
			FROM profiles
			WHERE user_id = ? AND current_location IS NOT NULL
			LIMIT 1
		`, userID).Scan(&loc).Error
		locationCh <- locationResult{loc.Latitude, loc.Longitude, err}
	}()

	pr := <-prefsCh
	if pr.err != nil {
		return UserContext{}, fmt.Errorf("get matching preferences: %w", pr.err)
	}

	ir := <-interestsCh
	if ir.err != nil {
		return UserContext{}, fmt.Errorf("get user interests: %w", ir.err)
	}

	lr := <-locationCh
	if lr.err != nil {
		return UserContext{}, fmt.Errorf("get user location: %w", lr.err)
	}

	return UserContext{
		Preferences: pr.prefs,
		InterestIDs: ir.ids,
		Latitude:    lr.lat,
		Longitude:   lr.lon,
	}, nil
}
