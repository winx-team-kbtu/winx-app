package course

import (
	"context"
	"errors"

	repository "winx-profile/internal/app/domain/repositories/course"
	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

const MaxUserCourses = 10

var ErrInvalidCourseIDs = errors.New("invalid course ids")

type Service interface {
	List(ctx context.Context) ([]models.Course, error)
	GetUserCourses(ctx context.Context, userID int64) ([]models.Course, error)
	SaveUserCourses(ctx context.Context, userID int64, courseIDs []int64) ([]models.Course, error)
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

func (s *service) List(ctx context.Context) ([]models.Course, error) {
	return s.repository.List(ctx)
}

func (s *service) GetUserCourses(ctx context.Context, userID int64) ([]models.Course, error) {
	return s.repository.ListByUserID(ctx, userID)
}

// SaveUserCourses replaces the user's selection. Every id must exist in the
// seeded catalogue, duplicates are collapsed and an empty list clears the selection.
func (s *service) SaveUserCourses(ctx context.Context, userID int64, courseIDs []int64) ([]models.Course, error) {
	unique := uniqueIDs(courseIDs)
	if len(unique) > MaxUserCourses {
		return nil, ErrInvalidCourseIDs
	}

	found, err := s.repository.GetByIDs(ctx, unique)
	if err != nil {
		return nil, err
	}

	if len(found) != len(unique) {
		return nil, ErrInvalidCourseIDs
	}

	if err := s.repository.ReplaceUserCourses(ctx, userID, unique); err != nil {
		return nil, err
	}

	return s.repository.ListByUserID(ctx, userID)
}

func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))

	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		out = append(out, id)
	}

	return out
}
