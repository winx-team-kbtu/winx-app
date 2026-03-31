package profilemeta

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"winx-profile/internal/app/models/models"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInvalidInterestIDs = errors.New("one or more interest ids are invalid")

type Repository interface {
	Sync(ctx context.Context, profile models.Profile) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Sync(ctx context.Context, profile models.Profile) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.upsertProfileRecord(tx, profile); err != nil {
			return err
		}

		if err := r.syncUserInterests(tx, profile); err != nil {
			return err
		}

		if err := r.upsertMatchingPreferences(tx, profile); err != nil {
			return err
		}

		return nil
	})
}

func (r *repository) upsertProfileRecord(tx *gorm.DB, profile models.Profile) error {
	name := strings.TrimSpace(profile.FirstName)
	if name == "" {
		name = profile.Email
	}
	city, country, latitude, longitude := profileGeoFields(profile)
	updatedAt := time.Now().UTC()

	query := `
		INSERT INTO profiles (user_id, email, name, city, country, current_location, updated_at)
		VALUES (?, ?, ?, ?, ?, %s, ?)
		ON CONFLICT (user_id) DO UPDATE
		SET email = EXCLUDED.email,
		    name = EXCLUDED.name,
		    city = EXCLUDED.city,
		    country = EXCLUDED.country,
		    current_location = EXCLUDED.current_location,
		    updated_at = EXCLUDED.updated_at
	`

	args := []any{
		profile.UserID,
		profile.Email,
		name,
		city,
		country,
	}

	locationExpr := "NULL"
	if latitude != nil && longitude != nil {
		locationExpr = "ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography"
		args = append(args, longitude, latitude)
	}

	args = append(args, updatedAt)

	if err := tx.Exec(fmt.Sprintf(query, locationExpr), args...).Error; err != nil {
		return fmt.Errorf("upsert profiles record: %w", err)
	}

	return nil
}

func (r *repository) syncUserInterests(tx *gorm.DB, profile models.Profile) error {
	if profile.InterestIDs == nil {
		return nil
	}

	if err := tx.Where("user_id = ?", profile.UserID).Delete(&models.UserInterest{}).Error; err != nil {
		return fmt.Errorf("delete user interests: %w", err)
	}

	interestIDs := normalizeInterestIDs(profile.InterestIDs)
	if len(interestIDs) == 0 {
		return nil
	}

	exists, err := r.countInterestsByIDs(tx, interestIDs)
	if err != nil {
		return err
	}
	if exists != int64(len(interestIDs)) {
		return ErrInvalidInterestIDs
	}

	userInterests := make([]models.UserInterest, 0, len(interestIDs))
	for _, interestID := range interestIDs {
		userInterests = append(userInterests, models.UserInterest{
			UserID:     profile.UserID,
			InterestID: interestID,
			AddedAt:    time.Now().UTC(),
		})
	}

	if err := tx.Create(&userInterests).Error; err != nil {
		return fmt.Errorf("create user interests: %w", err)
	}

	return nil
}

func (r *repository) countInterestsByIDs(tx *gorm.DB, ids []int64) (int64, error) {
	var count int64

	if err := tx.Model(&models.Interest{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count interests by ids: %w", err)
	}

	return count, nil
}

func (r *repository) upsertMatchingPreferences(tx *gorm.DB, profile models.Profile) error {
	preferences := models.MatchingPreferences{
		UserID:           profile.UserID,
		MinAge:           18,
		MaxAge:           99,
		MaxDistanceKM:    50,
		InterestedIn:     pq.StringArray(normalizeStrings(profile.InterestedIn)),
		ShowMeGlobal:     false,
		OnlyShowVerified: false,
		UpdatedAt:        time.Now().UTC(),
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"interested_in", "updated_at"}),
	}).Create(&preferences).Error; err != nil {
		return fmt.Errorf("upsert matching preferences: %w", err)
	}

	return nil
}

func normalizeInterestIDs(items []int64) []int64 {
	seen := map[int64]struct{}{}
	out := make([]int64, 0, len(items))

	for _, item := range items {
		if item <= 0 {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		out = append(out, item)
	}

	return out
}

func normalizeStrings(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		out = append(out, item)
	}

	return out
}

func profileGeoFields(profile models.Profile) (city any, country any, latitude any, longitude any) {
	if profile.Location == nil {
		return nil, nil, nil, nil
	}

	if strings.TrimSpace(profile.Location.City) != "" {
		city = profile.Location.City
	}

	if strings.TrimSpace(profile.Location.Country) != "" {
		country = profile.Location.Country
	}

	if profile.Location.CurrentLocation != nil {
		latitude = profile.Location.CurrentLocation.Latitude
		longitude = profile.Location.CurrentLocation.Longitude
	}

	return city, country, latitude, longitude
}
