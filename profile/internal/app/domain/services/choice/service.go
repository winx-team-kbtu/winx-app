package choice

import (
	"context"
	"errors"

	repository "winx-profile/internal/app/domain/repositories/choice"
	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

const (
	Like = "like"
	Pass = "pass"

	DefaultLimit = 20
	MaxLimit     = 50
)

var (
	ErrCourseNotSelected = errors.New("course is not selected")
	ErrInvalidCandidate  = errors.New("invalid candidate")
	ErrInvalidChoice     = errors.New("invalid choice")
)

type Service interface {
	Choose(ctx context.Context, userID, candidateID, courseID int64, choice string) error
	Shortlist(ctx context.Context, userID, courseID, limit, offset int64) ([]models.DeckCandidate, error)
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

// Choose records a Like or Pass for a candidate within one course. The call is
// idempotent, so a client can safely retry after a failure; repeating or
// changing a choice never creates a second record. Nothing is shared with the
// candidate, a Like is not a match and a Pass stays private.
func (s *service) Choose(ctx context.Context, userID, candidateID, courseID int64, choice string) error {
	if choice != Like && choice != Pass {
		return ErrInvalidChoice
	}

	if userID == candidateID {
		return ErrInvalidCandidate
	}

	selected, err := s.repository.HasCourse(ctx, userID, courseID)
	if err != nil {
		return err
	}

	if !selected {
		return ErrCourseNotSelected
	}

	stored, err := s.repository.Upsert(ctx, userID, candidateID, courseID, choice)
	if err != nil {
		return err
	}

	if !stored {
		return ErrInvalidCandidate
	}

	return nil
}

func (s *service) Shortlist(ctx context.Context, userID, courseID, limit, offset int64) ([]models.DeckCandidate, error) {
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

	return s.repository.ListShortlist(ctx, userID, courseID, limit, offset)
}
