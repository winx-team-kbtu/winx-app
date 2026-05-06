package swipe

import (
	"context"
	"errors"
	"fmt"

	swipeRepoDto "winx-match/internal/app/domain/core/dto/repositories/swipe"
	"winx-match/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("swipe not found")

type Repository interface {
	Create(ctx context.Context, input swipeRepoDto.CreateDTO) (models.Swipe, error)
	// FindMutual returns the direction of the swipe where swiper=input.SwiperID swiped on input.SwipedID.
	FindMutual(ctx context.Context, input swipeRepoDto.FindDTO) (string, error)
	// FindOwn returns the direction of an existing swipe from swiperID → swipedID, or "" if none.
	FindOwn(ctx context.Context, swiperID, swipedID int64) (string, error)
	Delete(ctx context.Context, input swipeRepoDto.DeleteDTO) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, input swipeRepoDto.CreateDTO) (models.Swipe, error) {
	swipe := models.Swipe{
		SwiperID:  input.SwiperID,
		SwipedID:  input.SwipedID,
		Direction: input.Direction,
	}

	if err := r.db.WithContext(ctx).Create(&swipe).Error; err != nil {
		return models.Swipe{}, fmt.Errorf("failed to create swipe: %w", err)
	}

	return swipe, nil
}

func (r *repository) FindMutual(ctx context.Context, input swipeRepoDto.FindDTO) (string, error) {
	var swipe models.Swipe
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Swipe{}).
		Where("swiper_id = ? AND swiped_id = ?", input.SwiperID, input.SwipedID).
		First(&swipe).
		Count(&count).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("failed to check mutual swipe: %w", err)
	}

	return swipe.Direction, nil
}

func (r *repository) FindOwn(ctx context.Context, swiperID, swipedID int64) (string, error) {
	var swipe models.Swipe
	err := r.db.WithContext(ctx).
		Where("swiper_id = ? AND swiped_id = ?", swiperID, swipedID).
		First(&swipe).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("find own swipe: %w", err)
	}
	return swipe.Direction, nil
}

func (r *repository) Delete(ctx context.Context, input swipeRepoDto.DeleteDTO) error {
	if err := r.db.WithContext(ctx).Delete(&models.Swipe{}, "swiper_id = ? AND swiped_id = ?", input.SwiperID, input.SwipedID).Error; err != nil {
		return fmt.Errorf("failed to delete swipe: %w", err)
	}

	return nil
}
