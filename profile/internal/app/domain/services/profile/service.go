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
	"winx-profile/pkg/graylog/logger"

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
	logger.LogInfo("Profile service initializing")

	repo := repository.NewRepository(mongoDB)

	if err := repo.EnsureIndexes(context.Background()); err != nil {
		logger.LogError("Failed to ensure MongoDB indexes", err)
		return nil, err
	}

	logger.LogInfo("Profile service initialized successfully")
	return &service{
		repository:     repo,
		interestRepo:   interestrepo.NewRepository(pgdb),
		metaRepository: profilemeta.NewRepository(pgdb),
	}, nil
}

func (s *service) GetByUserID(ctx context.Context, userID int64) (models.Profile, error) {
	logger.LogDebug(fmt.Sprintf("Fetching profile for user %d", userID))

	profile, err := s.repository.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			logger.LogDebug(fmt.Sprintf("Profile not found for user %d", userID))
			return models.Profile{}, ErrNotFound
		}
		logger.LogError(fmt.Sprintf("Failed to get profile for user %d", userID), err)
		return models.Profile{}, fmt.Errorf("get profile: %w", err)
	}

	return profile, nil
}

func (s *service) GetDetailsByUserID(ctx context.Context, userID int64) (Details, error) {
	logger.LogDebug(fmt.Sprintf("Fetching profile details for user %d", userID))

	profile, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return Details{}, err
	}

	interests, err := s.interestRepo.GetByIDs(ctx, profile.InterestIDs)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to get interests for user %d", userID), err)
		return Details{}, fmt.Errorf("get profile interests: %w", err)
	}

	logger.LogDebug(fmt.Sprintf("Profile details fetched for user %d with %d interests", userID, len(interests)))
	return Details{
		Profile:   profile,
		Interests: interests,
	}, nil
}

func (s *service) Store(ctx context.Context, input dto.StoreDTO) (models.Profile, bool, error) {
	logger.LogInfo(fmt.Sprintf("Storing profile for user %d", input.UserID))

	current, err := s.repository.GetByUserID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			logger.LogDebug(fmt.Sprintf("Creating new profile for user %d", input.UserID))
			profile := newProfile(input)
			created, createErr := s.repository.Create(ctx, profile)
			if createErr != nil {
				logger.LogError(fmt.Sprintf("Failed to create profile for user %d", input.UserID), createErr)
				return models.Profile{}, false, fmt.Errorf("create profile: %w", createErr)
			}

			if syncErr := s.metaRepository.Sync(ctx, created); syncErr != nil {
				logger.LogError(fmt.Sprintf("Failed to sync profile metadata for user %d", input.UserID), syncErr)
				if errors.Is(syncErr, profilemeta.ErrInvalidInterestIDs) {
					if rollbackErr := s.repository.DeleteByUserID(ctx, created.UserID); rollbackErr != nil {
						logger.LogError("Failed to rollback profile creation", rollbackErr)
						return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w; rollback profile create: %v", ErrInvalidInterestIDs, rollbackErr)
					}
					return models.Profile{}, false, ErrInvalidInterestIDs
				}
				if rollbackErr := s.repository.DeleteByUserID(ctx, created.UserID); rollbackErr != nil {
					logger.LogError("Failed to rollback profile creation", rollbackErr)
					return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w; rollback profile create: %v", syncErr, rollbackErr)
				}
				return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w", syncErr)
			}

			logger.LogInfo(fmt.Sprintf("Profile created successfully for user %d", input.UserID))
			return created, true, nil
		}
		logger.LogError(fmt.Sprintf("Failed to get existing profile for user %d", input.UserID), err)
		return models.Profile{}, false, fmt.Errorf("get profile: %w", err)
	}

	merged := mergeProfile(current, input)
	updated, err := s.repository.Update(ctx, merged)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to update profile for user %d", input.UserID), err)
		return models.Profile{}, false, fmt.Errorf("update profile: %w", err)
	}

	if err := s.metaRepository.Sync(ctx, updated); err != nil {
		logger.LogError(fmt.Sprintf("Failed to sync profile metadata for user %d", input.UserID), err)
		if errors.Is(err, profilemeta.ErrInvalidInterestIDs) {
			if _, rollbackErr := s.repository.Update(ctx, current); rollbackErr != nil {
				logger.LogError("Failed to rollback profile update", rollbackErr)
				return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w; rollback profile update: %v", ErrInvalidInterestIDs, rollbackErr)
			}
			return models.Profile{}, false, ErrInvalidInterestIDs
		}
		if _, rollbackErr := s.repository.Update(ctx, current); rollbackErr != nil {
			logger.LogError("Failed to rollback profile update", rollbackErr)
			return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w; rollback profile update: %v", err, rollbackErr)
		}
		return models.Profile{}, false, fmt.Errorf("sync profile metadata: %w", err)
	}

	logger.LogInfo(fmt.Sprintf("Profile updated successfully for user %d", input.UserID))
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
		if profile.Location == nil {
			profile.Location = &models.ProfileLocation{}
		}
		profile.Location.City = input.Location.City
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