package academic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("academic profile not found")

type Repository interface {
	GetByUserID(ctx context.Context, userID int64) (models.AcademicProfile, error)
	Upsert(ctx context.Context, profile models.AcademicProfile) (models.AcademicProfile, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

type row struct {
	UserID        int64
	Skills        []byte
	PreferredRole string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (r row) toModel() (models.AcademicProfile, error) {
	skills := []string{}
	if len(r.Skills) > 0 {
		if err := json.Unmarshal(r.Skills, &skills); err != nil {
			return models.AcademicProfile{}, fmt.Errorf("decode skills: %w", err)
		}
	}

	return models.AcademicProfile{
		UserID:        r.UserID,
		Skills:        skills,
		PreferredRole: r.PreferredRole,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}, nil
}

func (r *repository) GetByUserID(ctx context.Context, userID int64) (models.AcademicProfile, error) {
	var rec row

	res := r.db.WithContext(ctx).Raw(
		`SELECT user_id, skills, preferred_role, created_at, updated_at
		   FROM academic_profiles WHERE user_id = ?`, userID,
	).Scan(&rec)
	if res.Error != nil {
		return models.AcademicProfile{}, fmt.Errorf("get academic profile: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return models.AcademicProfile{}, ErrNotFound
	}

	return rec.toModel()
}

func (r *repository) Upsert(ctx context.Context, profile models.AcademicProfile) (models.AcademicProfile, error) {
	skills, err := json.Marshal(profile.Skills)
	if err != nil {
		return models.AcademicProfile{}, fmt.Errorf("encode skills: %w", err)
	}

	var rec row

	res := r.db.WithContext(ctx).Raw(
		`INSERT INTO academic_profiles (user_id, skills, preferred_role)
		 VALUES (?, ?::jsonb, ?)
		 ON CONFLICT (user_id) DO UPDATE
		   SET skills = EXCLUDED.skills,
		       preferred_role = EXCLUDED.preferred_role,
		       updated_at = NOW()
		 RETURNING user_id, skills, preferred_role, created_at, updated_at`,
		profile.UserID, string(skills), profile.PreferredRole,
	).Scan(&rec)
	if res.Error != nil {
		return models.AcademicProfile{}, fmt.Errorf("upsert academic profile: %w", res.Error)
	}

	return rec.toModel()
}
