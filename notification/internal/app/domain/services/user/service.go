package user

import (
	"winx-notification/internal/app/core/helpers/password"
	repoDto "winx-notification/internal/app/domain/core/dto/repositories/user"
	dto "winx-notification/internal/app/domain/core/dto/services/user"
	repository "winx-notification/internal/app/domain/repositories/user"
	"winx-notification/internal/app/models/models"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("user not found")
)

type Service interface {
	Create(ctx context.Context, input dto.CreateDTO) (models.User, error)
	Delete(ctx context.Context, input dto.DeleteDTO) (bool, error)
	Update(ctx context.Context, input dto.UpdateDTO) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
}

type service struct {
	db         *gorm.DB
	repository repository.Repository
}

func NewService(
	db *gorm.DB,
) Service {
	return &service{
		db:         db,
		repository: repository.NewRepository(db),
	}
}

func (s *service) Create(ctx context.Context, input dto.CreateDTO) (models.User, error) {
	passwd, err := password.Hash(input.Password)
	if err != nil {
		return models.User{}, fmt.Errorf("failed hashing password: %w", err)
	}

	return s.repository.Create(ctx, repoDto.CreateDTO{
		Email:    input.Email,
		Password: passwd,
	})
}

func (s *service) Delete(ctx context.Context, input dto.DeleteDTO) (bool, error) {
	ok, err := s.repository.Delete(ctx, repoDto.DeleteDTO{
		Email: input.Email,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ok, ErrNotFound
		}
	}

	return ok, err
}

func (s *service) Update(ctx context.Context, input dto.UpdateDTO) (models.User, error) {
	var passwd *string
	if input.Password != nil {
		hashed, err := password.Hash(*input.Password)
		if err != nil {
			return models.User{}, fmt.Errorf("failed hashing password: %w", err)
		}

		passwd = &hashed
	}

	user, err := s.repository.Update(ctx, repoDto.UpdateDTO{
		Email:    input.Email,
		NewEmail: input.NewEmail,
		Password: passwd,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.User{}, ErrNotFound
		}
	}

	return user, err
}

func (s *service) GetByEmail(ctx context.Context, email string) (models.User, error) {
	user, err := s.repository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.User{}, ErrNotFound
		}
	}

	return user, err
}
