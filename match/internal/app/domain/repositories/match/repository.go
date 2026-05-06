package match

import (
	"context"
	"errors"
	"fmt"

	repodto "winx-match/internal/app/domain/core/dto/repositories/match"
	"winx-match/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("match not found")

type Repository interface {
	Create(ctx context.Context, input repodto.CreateDTO) (models.Match, error)
	ListByUserID(ctx context.Context, input repodto.ListDTO) ([]models.Match, error)
	ListWithProfiles(ctx context.Context, input repodto.ListDTO) ([]repodto.MatchRow, error)
	DeleteByIDAndUserID(ctx context.Context, id, userID int64) (bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, input repodto.CreateDTO) (models.Match, error) {
	match := models.Match{
		UserOneID: input.UserOneID,
		UserTwoID: input.UserTwoID,
	}

	if err := r.db.WithContext(ctx).Create(&match).Error; err != nil {
		return models.Match{}, fmt.Errorf("failed to create match: %w", err)
	}

	return match, nil
}

func (r *repository) ListByUserID(ctx context.Context, input repodto.ListDTO) ([]models.Match, error) {
	var matches []models.Match

	q := r.db.WithContext(ctx).
		Where("user_one_id = ? OR user_two_id = ?", input.UserID, input.UserID).
		Order("created_at DESC")

	if input.Limit > 0 {
		q = q.Limit(input.Limit)
	}
	if input.Offset > 0 {
		q = q.Offset(input.Offset)
	}

	if err := q.Find(&matches).Error; err != nil {
		return nil, fmt.Errorf("failed to list matches: %w", err)
	}

	return matches, nil
}

func (r *repository) ListWithProfiles(ctx context.Context, input repodto.ListDTO) ([]repodto.MatchRow, error) {
	limit := input.Limit
	if limit == 0 {
		limit = 20
	}

	query := `
		WITH ranked AS (
			SELECT
				m.id,
				m.user_one_id,
				m.user_two_id,
				m.created_at,
				CASE WHEN m.user_one_id = ? THEN m.user_two_id ELSE m.user_one_id END AS matched_user_id
			FROM matches m
			WHERE m.user_one_id = ? OR m.user_two_id = ?
			ORDER BY m.created_at DESC
			LIMIT ? OFFSET ?
		)
		SELECT
			r.id,
			r.user_one_id,
			r.user_two_id,
			r.matched_user_id,
			r.created_at,
			COALESCE(p.name, '')    AS name,
			COALESCE(p.city, '')    AS city,
			COALESCE(p.country, '') AS country,
			COALESCE(ph.url, '')    AS photo_url
		FROM ranked r
		LEFT JOIN profiles p ON p.user_id = r.matched_user_id
		LEFT JOIN (
			SELECT DISTINCT ON (user_id) user_id, url
			FROM profile_photos
			ORDER BY user_id, id ASC
		) ph ON ph.user_id = r.matched_user_id
	`

	var rows []repodto.MatchRow
	err := r.db.WithContext(ctx).Raw(
		query,
		input.UserID, // CASE WHEN
		input.UserID, // WHERE user_one_id
		input.UserID, // WHERE user_two_id
		limit,
		input.Offset,
	).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list matches with profiles: %w", err)
	}
	return rows, nil
}

func (r *repository) DeleteByIDAndUserID(ctx context.Context, id, userID int64) (bool, error) {
	res := r.db.WithContext(ctx).
		Where("id = ? AND (user_one_id = ? OR user_two_id = ?)", id, userID, userID).
		Delete(&models.Match{})

	if res.Error != nil {
		return false, fmt.Errorf("failed to delete match: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return false, ErrNotFound
	}

	return true, nil
}
