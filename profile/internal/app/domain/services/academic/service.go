package academic

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	repository "winx-profile/internal/app/domain/repositories/academic"
	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

const (
	MaxSkills        = 20
	MaxSkillLength   = 30
	MaxRoleLength    = 50
	skillRoleMessage = "invalid skills or preferred role"
)

var ErrInvalidInput = errors.New(skillRoleMessage)

type SaveInput struct {
	Skills        []string
	PreferredRole string
}

type Service interface {
	Get(ctx context.Context, userID int64) (models.AcademicProfile, error)
	Save(ctx context.Context, userID int64, input SaveInput) (models.AcademicProfile, error)
}

type service struct {
	repository repository.Repository
}

func NewService(db *gorm.DB) Service {
	return NewServiceWithRepository(repository.NewRepository(db))
}

func NewServiceWithRepository(repo repository.Repository) Service {
	return &service{repository: repo}
}

// Get returns an empty profile when the user has not saved anything yet, so
// clients can treat "not set" and "empty" the same way.
func (s *service) Get(ctx context.Context, userID int64) (models.AcademicProfile, error) {
	profile, err := s.repository.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return models.AcademicProfile{UserID: userID, Skills: []string{}}, nil
	}

	return profile, err
}

func (s *service) Save(ctx context.Context, userID int64, input SaveInput) (models.AcademicProfile, error) {
	skills, role, err := normalize(input)
	if err != nil {
		return models.AcademicProfile{}, err
	}

	return s.repository.Upsert(ctx, models.AcademicProfile{
		UserID:        userID,
		Skills:        skills,
		PreferredRole: role,
	})
}

// normalize trims values, drops empty ones, removes case-insensitive duplicates
// and enforces the length limits. Skills and role are both optional.
func normalize(input SaveInput) ([]string, string, error) {
	role := strings.TrimSpace(input.PreferredRole)
	if utf8.RuneCountInString(role) > MaxRoleLength {
		return nil, "", ErrInvalidInput
	}

	skills := make([]string, 0, len(input.Skills))
	seen := make(map[string]struct{}, len(input.Skills))

	for _, raw := range input.Skills {
		skill := strings.TrimSpace(raw)
		if skill == "" {
			continue
		}

		if utf8.RuneCountInString(skill) > MaxSkillLength {
			return nil, "", ErrInvalidInput
		}

		key := strings.ToLower(skill)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		skills = append(skills, skill)
	}

	if len(skills) > MaxSkills {
		return nil, "", ErrInvalidInput
	}

	return skills, role, nil
}
