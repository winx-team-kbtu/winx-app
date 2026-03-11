package notifications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"winx-notification/internal/app/models/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	StatusPending     = "pending"
	StatusSent        = "sent"
	StatusSentMock    = "sent_mock"
	StatusFailed      = "failed"
	ChannelEmail      = "email"
	TypeWelcome       = "welcome"
	TypePasswordReset = "password_reset"
)

var ErrTypeNotFound = errors.New("notification type not found")

type Store struct {
	db *gorm.DB
}

type CreateInput struct {
	TypeCode  string
	Recipient string
	Subject   string
	Body      string
	Payload   datatypes.JSON
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, input CreateInput) (*models.Notification, error) {
	var notificationType models.NotificationType
	if err := s.db.WithContext(ctx).
		Where("code = ?", input.TypeCode).
		First(&notificationType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTypeNotFound
		}

		return nil, fmt.Errorf("find notification type: %w", err)
	}

	notification := &models.Notification{
		NotificationTypeID: notificationType.ID,
		Recipient:          input.Recipient,
		Subject:            input.Subject,
		Body:               input.Body,
		Payload:            input.Payload,
		Status:             StatusPending,
		Channel:            notificationType.Channel,
	}

	if err := s.db.WithContext(ctx).Create(notification).Error; err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	return notification, nil
}

func (s *Store) ClaimPending(ctx context.Context, limit int) ([]models.Notification, error) {
	var items []models.Notification
	if err := s.db.WithContext(ctx).
		Preload("Type").
		Where("status = ?", StatusPending).
		Order("id ASC").
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("load pending notifications: %w", err)
	}

	return items, nil
}

func (s *Store) MarkSent(ctx context.Context, id int64, mock bool) error {
	status := StatusSent
	if mock {
		status = StatusSentMock
	}

	now := time.Now()
	if err := s.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     status,
			"sent_at":    now,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("mark notification sent: %w", err)
	}

	return nil
}

func (s *Store) MarkFailed(ctx context.Context, id int64, msg string) error {
	now := time.Now()
	if err := s.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        StatusFailed,
			"error_message": msg,
			"updated_at":    now,
		}).Error; err != nil {
		return fmt.Errorf("mark notification failed: %w", err)
	}

	return nil
}
