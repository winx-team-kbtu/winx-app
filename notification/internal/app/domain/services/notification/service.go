package notification

import (
	"context"
	"errors"

	repository "winx-notification/internal/app/domain/repositories/notification"
	"winx-notification/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("notification not found")

type Service interface {
	ListByRecipient(ctx context.Context, recipient string) ([]models.Notification, error)
	DeleteByIDAndRecipient(ctx context.Context, id int64, recipient string) (bool, error)
}

type service struct {
	db         *gorm.DB
	repository repository.Repository
}

func NewService(db *gorm.DB) Service {
	return &service{
		db:         db,
		repository: repository.NewRepository(db),
	}
}

func (s *service) ListByRecipient(ctx context.Context, recipient string) ([]models.Notification, error) {
	return s.repository.ListByRecipient(ctx, recipient)
}

func (s *service) DeleteByIDAndRecipient(ctx context.Context, id int64, recipient string) (bool, error) {
	ok, err := s.repository.DeleteByIDAndRecipient(ctx, id, recipient)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return false, ErrNotFound
	}

	return ok, err
}
