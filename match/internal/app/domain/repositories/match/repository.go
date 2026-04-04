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
