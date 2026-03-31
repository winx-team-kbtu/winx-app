package interest

import (
	"context"

	repository "winx-profile/internal/app/domain/repositories/interest"
	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

type Service interface {
	List(ctx context.Context) ([]models.Interest, error)
}

type service struct {
	repository repository.Repository
}

func NewService(db *gorm.DB) Service {
	return &service{
		repository: repository.NewRepository(db),
	}
}

func (s *service) List(ctx context.Context) ([]models.Interest, error) {
	return s.repository.List(ctx)
}
