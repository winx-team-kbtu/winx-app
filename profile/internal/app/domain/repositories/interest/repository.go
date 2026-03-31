package interest

import (
	"context"
	"fmt"

	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context) ([]models.Interest, error)
	GetByIDs(ctx context.Context, ids []int64) ([]models.Interest, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]models.Interest, error) {
	var items []models.Interest

	if err := r.db.WithContext(ctx).
		Order("category ASC").
		Order("name ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list interests: %w", err)
	}

	return items, nil
}

func (r *repository) GetByIDs(ctx context.Context, ids []int64) ([]models.Interest, error) {
	var items []models.Interest

	if len(ids) == 0 {
		return items, nil
	}

	if err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("get interests by ids: %w", err)
	}

	byID := make(map[int64]models.Interest, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}

	ordered := make([]models.Interest, 0, len(ids))
	for _, id := range ids {
		if item, ok := byID[id]; ok {
			ordered = append(ordered, item)
		}
	}

	return ordered, nil
}
