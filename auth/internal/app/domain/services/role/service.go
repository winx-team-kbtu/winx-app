package role

import (
	repoDto "auth/internal/app/domain/core/dto/repositories/role"
	dto "auth/internal/app/domain/core/dto/services/role"
	repository "auth/internal/app/domain/repositories/role"
	"auth/internal/app/models/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNotFound  = errors.New("role not found")
	ErrDuplicate = errors.New("role slug already exists")
)

type Service interface {
	Store(ctx context.Context, input dto.StoreDTO) (models.Role, error)
	Update(ctx context.Context, input dto.UpdateDTO) (models.Role, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type service struct {
	repository repository.Repository
}

func NewService(db *gorm.DB) Service {
	return &service{
		repository: repository.NewRepository(db),
	}
}

func (s *service) Store(ctx context.Context, input dto.StoreDTO) (models.Role, error) {
	role, err := s.repository.Store(ctx, repoDto.StoreDTO{
		Name: input.Name,
		Slug: input.Slug,
	})
	if err != nil {
		return models.Role{}, err
	}

	return role, nil
}

func (s *service) Update(ctx context.Context, input dto.UpdateDTO) (models.Role, error) {
	role, err := s.repository.Update(ctx, repoDto.UpdateDTO{
		ID:   input.ID,
		Name: input.Name,
		Slug: input.Slug,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.Role{}, ErrNotFound
		}
		return models.Role{}, err
	}

	return role, nil
}

func (s *service) Delete(ctx context.Context, id int64) (bool, error) {
	ok, err := s.repository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return false, ErrNotFound
		}
		return false, err
	}

	return ok, nil
}
