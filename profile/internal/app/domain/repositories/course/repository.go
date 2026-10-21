package course

import (
	"context"
	"fmt"

	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context) ([]models.Course, error)
	GetByIDs(ctx context.Context, ids []int64) ([]models.Course, error)
	ListByUserID(ctx context.Context, userID int64) ([]models.Course, error)
	ReplaceUserCourses(ctx context.Context, userID int64, courseIDs []int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]models.Course, error) {
	var items []models.Course

	if err := r.db.WithContext(ctx).Order("code ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}

	return items, nil
}

func (r *repository) GetByIDs(ctx context.Context, ids []int64) ([]models.Course, error) {
	var items []models.Course

	if len(ids) == 0 {
		return items, nil
	}

	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Order("code ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("get courses by ids: %w", err)
	}

	return items, nil
}

func (r *repository) ListByUserID(ctx context.Context, userID int64) ([]models.Course, error) {
	var items []models.Course

	if err := r.db.WithContext(ctx).
		Joins("JOIN user_courses uc ON uc.course_id = courses.id").
		Where("uc.user_id = ?", userID).
		Order("courses.code ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list user courses: %w", err)
	}

	return items, nil
}

// ReplaceUserCourses atomically replaces the user's selected courses.
func (r *repository) ReplaceUserCourses(ctx context.Context, userID int64, courseIDs []int64) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM user_courses WHERE user_id = ?`, userID).Error; err != nil {
			return fmt.Errorf("clear user courses: %w", err)
		}

		for _, courseID := range courseIDs {
			if err := tx.Exec(
				`INSERT INTO user_courses (user_id, course_id) VALUES (?, ?)`, userID, courseID,
			).Error; err != nil {
				return fmt.Errorf("insert user course: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("replace user courses: %w", err)
	}

	return nil
}
