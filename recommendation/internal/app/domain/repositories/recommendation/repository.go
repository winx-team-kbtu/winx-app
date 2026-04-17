package recommendation

import (
	"context"
	"errors"
	"fmt"

	repoDto "winx-recommendation/internal/app/domain/core/dto/repositories/recommendation"
	"winx-recommendation/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("recommendation not found")

type Repository interface {
	List(ctx context.Context, input repoDto.ListDTO) ([]models.Recommendation, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context, input repoDto.ListDTO) ([]models.Recommendation, error) {
	var recs []models.Recommendation

	q := r.db.WithContext(ctx).
		Where("user_id = ?", input.UserID).
		Order("score DESC").
		Limit(input.Limit).
		Offset(input.Offset)

	if err := q.Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("list recommendations: %w", err)
	}

	return recs, nil
}
