package profile

import (
	"context"
	"errors"
	"fmt"
	"time"

	dto "winx-profile/internal/app/domain/core/dto/services/profile"
	interestrepo "winx-profile/internal/app/domain/repositories/interest"
	repository "winx-profile/internal/app/domain/repositories/profile"
	profilemeta "winx-profile/internal/app/domain/repositories/profilemeta"
	"winx-profile/internal/app/models/models"

	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("profile not found")
var ErrInvalidInterestIDs = errors.New("one or more interest_ids are invalid")

type Details struct {
	Profile   models.Profile
	Interests []models.Interest
}

type Service interface {
	GetByUserID(ctx context.Context, userID int64) (models.Profile, error)
	GetDetailsByUserID(ctx context.Context, userID int64) (Details, error)
	Store(ctx context.Context, input dto.StoreDTO) (models.Profile, bool, error)
}

type service struct {
	repository     repository.Repository
	interestRepo   interestrepo.Repository
	metaRepository profilemeta.Repository
}

func NewService(mongoDB *mongodriver.Database, pgdb *gorm.DB) (Service, error) {
	repo := repository.NewRepository(mongoDB)

	if err := repo.EnsureIndexes(context.Background()); err != nil {
		return nil, err
	}

	return &service{
		repository:     repo,
		interestRepo:   interestrepo.NewRepository(pgdb),
		metaRepository: profilemeta.NewRepository(pgdb),
	}, nil
}

func (s *service) GetByUserID(ctx context.Context, userID int64) (models.Profile, error) {
	profile, err := s.repository.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.Profile{}, ErrNotFound
		}

		return models.Profile{}, fmt.Errorf("get profile: %w", err)
	}

	return profile, nil
}

func (s *service) GetDetailsByUserID(ctx context.Context, userID int64) (Details, error) {
	profile, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return Details{}, err
	}

	interests, err := s.interestRepo.GetByIDs(ctx, profile.InterestIDs)
	if err != nil {
		return Details{}, fmt.Errorf("get profile interests: %w", err)
	}

	return Details{
		Profile:   profile,
		Interests: interests,
	}, nil
}

func (s *service) Store(ctx context.Context, input dto.StoreDTO) (models.Profile, bool, error) {
	current, err := s.repository.GetByUserID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			profile := newProfile(input)
			created, createErr := s.repository.Create(ctx, profile)
			if createErr != nil {
				return models.Profile{}, false, fmt.Errorf("create profile: %w", createErr)
			}

			if syncErr := s.metaRepository.Sync(ctx, created); syncErr != nil {
				if errors.Is(syncErr, profilemeta.ErrInvalidInterestIDs) {
					if rollbackErr := s.repository.DeleteByUserID(ctx, created.UserID); rollbackErr != nil {
						return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w; rollback profile create: %v", ErrInvalidInterestIDs, rollbackErr)
					}

					return models.Profile{}, false, ErrInvalidInterestIDs
				}

				if rollbackErr := s.repository.DeleteByUserID(ctx, created.UserID); rollbackErr != nil {
					return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w; rollback profile create: %v", syncErr, rollbackErr)
				}

				return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w", syncErr)
			}

			return created, true, nil
		}

		return models.Profile{}, false, fmt.Errorf("get profile: %w", err)
	}

	merged := mergeProfile(current, input)
	updated, err := s.repository.Update(ctx, merged)
	if err != nil {
		return models.Profile{}, false, fmt.Errorf("update profile: %w", err)
	}

	if err := s.metaRepository.Sync(ctx, updated); err != nil {
		if errors.Is(err, profilemeta.ErrInvalidInterestIDs) {
			if _, rollbackErr := s.repository.Update(ctx, current); rollbackErr != nil {
				return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w; rollback profile update: %v", ErrInvalidInterestIDs, rollbackErr)
			}

			return models.Profile{}, false, ErrInvalidInterestIDs
		}

		if _, rollbackErr := s.repository.Update(ctx, current); rollbackErr != nil {
			return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w; rollback profile update: %v", err, rollbackErr)
		}

		return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w", err)
	}

	return updated, false, nil
}

func newProfile(input dto.StoreDTO) models.Profile {
	now := time.Now().UTC()

	profile := models.Profile{
		UserID:    input.UserID,
		Email:     input.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	applyProfileInput(&profile, input)

	return profile
}

func mergeProfile(profile models.Profile, input dto.StoreDTO) models.Profile {
	profile.Email = input.Email
	profile.UpdatedAt = time.Now().UTC()
	applyProfileInput(&profile, input)

	return profile
}

func applyProfileInput(profile *models.Profile, input dto.StoreDTO) {
	if input.FirstName != nil {
		profile.FirstName = *input.FirstName
	}
	if input.BirthDate != nil {
		profile.BirthDate = *input.BirthDate
	}
	if input.Gender != nil {
		profile.Gender = *input.Gender
	}
	if input.ShowGender != nil {
		profile.ShowGender = input.ShowGender
	}
	if input.ShowAge != nil {
		profile.ShowAge = input.ShowAge
	}
	if input.AboutMe != nil {
		profile.AboutMe = *input.AboutMe
	}
	if input.JobTitle != nil {
		profile.JobTitle = *input.JobTitle
	}
	if input.Company != nil {
		profile.Company = *input.Company
	}
	if input.School != nil {
		profile.School = *input.School
	}
	if input.InterestIDs != nil {
		profile.InterestIDs = input.InterestIDs
	}
	if input.InterestedIn != nil {
		profile.InterestedIn = input.InterestedIn
	}
	if input.LookingFor != nil {
		profile.LookingFor = *input.LookingFor
	}
	if input.Lifestyle != nil {
		profile.Lifestyle = input.Lifestyle
	}
	if input.Preferences != nil {
		profile.Preferences = input.Preferences
	}
	if input.Location != nil {
		profile.Location = &models.ProfileLocation{
			City: input.Location.City,
		}
		if input.Location.Country != nil {
			profile.Location.Country = *input.Location.Country
		}
		if input.Location.CurrentLocation != nil {
			profile.Location.CurrentLocation = &models.ProfileGeoPoint{
				Latitude:  input.Location.CurrentLocation.Latitude,
				Longitude: input.Location.CurrentLocation.Longitude,
			}
		}
	}
}
