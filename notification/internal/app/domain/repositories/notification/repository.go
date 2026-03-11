package notification

import (
	"context"
	"errors"
	"fmt"

	"winx-notification/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("notification not found")

type Repository interface {
	ListByRecipient(ctx context.Context, recipient string) ([]models.Notification, error)
	DeleteByIDAndRecipient(ctx context.Context, id int64, recipient string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListByRecipient(ctx context.Context, recipient string) ([]models.Notification, error) {
	var items []models.Notification

	if err := r.db.WithContext(ctx).
		Preload("Type").
		Where("recipient = ?", recipient).
		Order("id DESC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list notifications by recipient: %w", err)
	}

	return items, nil
}

func (r *repository) DeleteByIDAndRecipient(ctx context.Context, id int64, recipient string) (bool, error) {
	res := r.db.WithContext(ctx).
		Where("id = ? AND recipient = ?", id, recipient).
		Delete(&models.Notification{})
	if res.Error != nil {
		return false, fmt.Errorf("delete notification: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return false, ErrNotFound
	}

	return true, nil
}
