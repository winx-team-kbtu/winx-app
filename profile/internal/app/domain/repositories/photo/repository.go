package photo

import (
	"context"
	"errors"
	"fmt"

	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("profile photo not found")

type Repository interface {
	GetByUserID(ctx context.Context, userID int64) (models.ProfilePhoto, error)
	Create(ctx context.Context, photo models.ProfilePhoto) (models.ProfilePhoto, error)
	Update(ctx context.Context, photo models.ProfilePhoto) (models.ProfilePhoto, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByUserID(ctx context.Context, userID int64) (models.ProfilePhoto, error) {
	var photo models.ProfilePhoto

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&photo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ProfilePhoto{}, ErrNotFound
		}

		return models.ProfilePhoto{}, fmt.Errorf("get profile photo by user id: %w", err)
	}

	return photo, nil
}

func (r *repository) Create(ctx context.Context, photo models.ProfilePhoto) (models.ProfilePhoto, error) {
	if err := r.db.WithContext(ctx).Create(&photo).Error; err != nil {
		return models.ProfilePhoto{}, fmt.Errorf("create profile photo: %w", err)
	}

	return photo, nil
}

func (r *repository) Update(ctx context.Context, photo models.ProfilePhoto) (models.ProfilePhoto, error) {
	updates := map[string]any{
		"email":      photo.Email,
		"provider":   photo.Provider,
		"bucket":     photo.Bucket,
		"object_key": photo.ObjectKey,
		"url":        photo.URL,
		"mime_type":  photo.MimeType,
		"size_bytes": photo.SizeBytes,
		"width":      photo.Width,
		"height":     photo.Height,
	}

	res := r.db.WithContext(ctx).
		Model(&models.ProfilePhoto{}).
		Where("user_id = ?", photo.UserID).
		Updates(updates)
	if res.Error != nil {
		return models.ProfilePhoto{}, fmt.Errorf("update profile photo: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return models.ProfilePhoto{}, ErrNotFound
	}

	return r.GetByUserID(ctx, photo.UserID)
}
