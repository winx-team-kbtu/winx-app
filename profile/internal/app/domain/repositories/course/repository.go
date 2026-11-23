package course

import (
	"context"
	"errors"
	"fmt"

	"winx-profile/internal/app/models/models"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var ErrTeamCourse = errors.New("Сначала выйдите из команды по предмету, который хотите убрать.")

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
		// Serialize edits for the same account and leave retained enrollment rows in place.
		var locked int64
		if err := tx.Raw(`SELECT id FROM users WHERE id=? FOR NO KEY UPDATE`, userID).Scan(&locked).Error; err != nil {
			return err
		}
		remove := tx.Table("user_courses").Where("user_id = ?", userID)
		if len(courseIDs) > 0 {
			remove = remove.Where("course_id NOT IN ?", courseIDs)
		}
		if err := remove.Delete(&models.Course{}).Error; err != nil {
			return fmt.Errorf("remove unselected courses: %w", err)
		}

		for _, courseID := range courseIDs {
			if err := tx.Exec(
				`INSERT INTO user_courses (user_id, course_id) VALUES (?, ?) ON CONFLICT DO NOTHING`, userID, courseID,
			).Error; err != nil {
				return fmt.Errorf("insert user course: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "team_members_enrollment_fkey" {
			return ErrTeamCourse
		}
		return fmt.Errorf("replace user courses: %w", err)
	}

	return nil
}
