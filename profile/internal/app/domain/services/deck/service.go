package deck

import (
	"context"
	"errors"

	repository "winx-profile/internal/app/domain/repositories/deck"
	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

const (
	DefaultLimit = 20
	MaxLimit     = 50
)

// ErrCourseNotSelected is returned both for a course the user has not selected
// and for a course that does not exist, so the deck never reveals the catalogue
// membership of other students.
var ErrCourseNotSelected = errors.New("course is not selected")

type Service interface {
	Candidates(ctx context.Context, userID, courseID, limit, offset int64) ([]models.DeckCandidate, error)
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

func (s *service) Candidates(ctx context.Context, userID, courseID, limit, offset int64) ([]models.DeckCandidate, error) {
	selected, err := s.repository.HasCourse(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}

	if !selected {
		return nil, ErrCourseNotSelected
	}

	if limit <= 0 {
		limit = DefaultLimit
	}

	if limit > MaxLimit {
		limit = MaxLimit
	}

	if offset < 0 {
		offset = 0
	}

	return s.repository.ListCandidates(ctx, userID, courseID, limit, offset)
}
