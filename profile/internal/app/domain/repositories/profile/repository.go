package profile

import (
	dto "winx-profile/internal/app/domain/core/dto/repositories/profile"
	"winx-profile/internal/app/models/models"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("profile not found")

type Repository interface {
	GetByUserID(ctx context.Context, userID int64) (models.Profile, error)
	Create(ctx context.Context, input dto.CreateDTO) (models.Profile, error)
	Update(ctx context.Context, input dto.UpdateDTO) (models.Profile, error)
	UpdateRole(ctx context.Context, input dto.UpdateRoleDTO) (models.Profile, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByUserID(ctx context.Context, userID int64) (models.Profile, error) {
	var profile models.Profile

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Profile{}, ErrNotFound
		}
		return models.Profile{}, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

func (r *repository) Create(ctx context.Context, input dto.CreateDTO) (models.Profile, error) {
	profile := models.Profile{
		UserID: input.UserID,
		Role:   models.RoleUser,
	}

	if err := r.db.WithContext(ctx).Create(&profile).Error; err != nil {
		return models.Profile{}, fmt.Errorf("failed to create profile: %w", err)
	}

	return profile, nil
}

func (r *repository) Update(ctx context.Context, input dto.UpdateDTO) (models.Profile, error) {
	updates := map[string]interface{}{}

	if input.FirstName != nil {
		updates["first_name"] = *input.FirstName
	}
	if input.LastName != nil {
		updates["last_name"] = *input.LastName
	}
	if input.Bio != nil {
		updates["bio"] = *input.Bio
	}
	if input.AvatarURL != nil {
		updates["avatar_url"] = *input.AvatarURL
	}

	res := r.db.WithContext(ctx).
		Model(&models.Profile{}).
		Where("user_id = ?", input.UserID).
		Updates(updates)

	if res.Error != nil {
		return models.Profile{}, fmt.Errorf("failed to update profile: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return models.Profile{}, ErrNotFound
	}

	return r.GetByUserID(ctx, input.UserID)
}

func (r *repository) UpdateRole(ctx context.Context, input dto.UpdateRoleDTO) (models.Profile, error) {
	res := r.db.WithContext(ctx).
		Model(&models.Profile{}).
		Where("user_id = ?", input.UserID).
		Update("role", input.Role)

	if res.Error != nil {
		return models.Profile{}, fmt.Errorf("failed to update role: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return models.Profile{}, ErrNotFound
	}

	return r.GetByUserID(ctx, input.UserID)
}
