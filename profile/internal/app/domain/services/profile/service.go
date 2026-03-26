package profile

import (
	repoDto "winx-profile/internal/app/domain/core/dto/repositories/profile"
	dto "winx-profile/internal/app/domain/core/dto/services/profile"
	repository "winx-profile/internal/app/domain/repositories/profile"
	"winx-profile/internal/app/models/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("profile not found")

type Service interface {
	Get(ctx context.Context, userID int64) (models.Profile, error)
	Create(ctx context.Context, input dto.CreateDTO) (models.Profile, error)
	Update(ctx context.Context, input dto.UpdateDTO) (models.Profile, error)
	UpdateRole(ctx context.Context, input dto.UpdateRoleDTO) (models.Profile, error)
}

type service struct {
	repository repository.Repository
}

func NewService(db *gorm.DB) Service {
	return &service{
		repository: repository.NewRepository(db),
	}
}

func (s *service) Get(ctx context.Context, userID int64) (models.Profile, error) {
	profile, err := s.repository.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.Profile{}, ErrNotFound
		}
		return models.Profile{}, err
	}

	return profile, nil
}

func (s *service) Create(ctx context.Context, input dto.CreateDTO) (models.Profile, error) {
	return s.repository.Create(ctx, repoDto.CreateDTO{
		UserID: input.UserID,
	})
}

func (s *service) Update(ctx context.Context, input dto.UpdateDTO) (models.Profile, error) {
	profile, err := s.repository.Update(ctx, repoDto.UpdateDTO{
		UserID:    input.UserID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Bio:       input.Bio,
		AvatarURL: input.AvatarURL,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.Profile{}, ErrNotFound
		}
		return models.Profile{}, err
	}

	return profile, nil
}

func (s *service) UpdateRole(ctx context.Context, input dto.UpdateRoleDTO) (models.Profile, error) {
	profile, err := s.repository.UpdateRole(ctx, repoDto.UpdateRoleDTO{
		UserID: input.UserID,
		Role:   input.Role,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.Profile{}, ErrNotFound
		}
		return models.Profile{}, err
	}

	return profile, nil
}
