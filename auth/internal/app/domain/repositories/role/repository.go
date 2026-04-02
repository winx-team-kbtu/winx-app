package role

import (
	dto "auth/internal/app/domain/core/dto/repositories/role"
	"auth/internal/app/models/models"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrNotFound  = errors.New("role not found")
	ErrDuplicate = errors.New("role slug already exists")
)

type Repository interface {
	Store(ctx context.Context, input dto.StoreDTO) (models.Role, error)
	Update(ctx context.Context, input dto.UpdateDTO) (models.Role, error)
	Delete(ctx context.Context, id int64) (bool, error)
	GetBySlug(ctx context.Context, slug string) (models.Role, error)
	GetByID(ctx context.Context, id int64) (models.Role, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Store(ctx context.Context, input dto.StoreDTO) (models.Role, error) {
	role := models.Role{
		Name: input.Name,
		Slug: input.Slug,
	}

	if err := r.db.WithContext(ctx).Create(&role).Error; err != nil {
		return models.Role{}, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

func (r *repository) Update(ctx context.Context, input dto.UpdateDTO) (models.Role, error) {
	res := r.db.WithContext(ctx).
		Model(&models.Role{}).
		Where("id = ?", input.ID).
		Updates(map[string]interface{}{
			"name": input.Name,
			"slug": input.Slug,
		})

	if res.Error != nil {
		return models.Role{}, fmt.Errorf("failed to update role: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return models.Role{}, ErrNotFound
	}

	return r.GetByID(ctx, input.ID)
}

func (r *repository) Delete(ctx context.Context, id int64) (bool, error) {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Role{})
	if res.Error != nil {
		return false, fmt.Errorf("failed to delete role: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return false, ErrNotFound
	}

	return true, nil
}

func (r *repository) GetBySlug(ctx context.Context, slug string) (models.Role, error) {
	var role models.Role

	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Role{}, ErrNotFound
		}
		return models.Role{}, fmt.Errorf("failed to get role by slug: %w", err)
	}

	return role, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (models.Role, error) {
	var role models.Role

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Role{}, ErrNotFound
		}
		return models.Role{}, fmt.Errorf("failed to get role by id: %w", err)
	}

	return role, nil
}
