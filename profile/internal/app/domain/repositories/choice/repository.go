package choice

import (
	"context"
	"encoding/json"
	"fmt"

	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

type Repository interface {
	HasCourse(ctx context.Context, userID, courseID int64) (bool, error)
	// Upsert stores the choice only if the candidate also selected the course.
	// It reports false (and writes nothing) when the candidate is not in the course.
	Upsert(ctx context.Context, userID, candidateID, courseID int64, choice string) (bool, error)
	ListShortlist(ctx context.Context, userID, courseID, limit, offset int64) ([]models.DeckCandidate, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) HasCourse(ctx context.Context, userID, courseID int64) (bool, error) {
	var exists bool

	if err := r.db.WithContext(ctx).Raw(
		`SELECT EXISTS (SELECT 1 FROM user_courses WHERE user_id = ? AND course_id = ?)`,
		userID, courseID,
	).Scan(&exists).Error; err != nil {
		return false, fmt.Errorf("check user course: %w", err)
	}

	return exists, nil
}

func (r *repository) Upsert(ctx context.Context, userID, candidateID, courseID int64, choice string) (bool, error) {
	res := r.db.WithContext(ctx).Exec(
		`INSERT INTO course_choices (user_id, candidate_id, course_id, choice)
		 SELECT ?, ?, ?, ?
		  WHERE EXISTS (SELECT 1 FROM user_courses WHERE user_id = ? AND course_id = ?)
		 ON CONFLICT (user_id, candidate_id, course_id) DO UPDATE
		   SET choice = EXCLUDED.choice, updated_at = NOW()`,
		userID, candidateID, courseID, choice, candidateID, courseID,
	)
	if res.Error != nil {
		return false, fmt.Errorf("upsert course choice: %w", res.Error)
	}

	return res.RowsAffected > 0, nil
}

type row struct {
	UserID        int64
	Name          string
	Skills        []byte
	PreferredRole string
}

func (r *repository) ListShortlist(ctx context.Context, userID, courseID, limit, offset int64) ([]models.DeckCandidate, error) {
	var rows []row

	if err := r.db.WithContext(ctx).Raw(
		`SELECT cc.candidate_id                   AS user_id,
		        COALESCE(p.name, '')              AS name,
		        COALESCE(ap.skills, '[]'::jsonb)  AS skills,
		        COALESCE(ap.preferred_role, '')   AS preferred_role
		   FROM course_choices cc
		   LEFT JOIN profiles p            ON p.user_id = cc.candidate_id
		   LEFT JOIN academic_profiles ap  ON ap.user_id = cc.candidate_id
		  WHERE cc.user_id = ? AND cc.course_id = ? AND cc.choice = 'like'
		  ORDER BY cc.updated_at DESC, cc.candidate_id
		  LIMIT ? OFFSET ?`,
		userID, courseID, limit, offset,
	).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list shortlist: %w", err)
	}

	out := make([]models.DeckCandidate, 0, len(rows))

	for _, rec := range rows {
		skills := []string{}
		if len(rec.Skills) > 0 {
			if err := json.Unmarshal(rec.Skills, &skills); err != nil {
				return nil, fmt.Errorf("decode shortlist skills: %w", err)
			}
		}

		out = append(out, models.DeckCandidate{
			UserID:        rec.UserID,
			Name:          rec.Name,
			Skills:        skills,
			PreferredRole: rec.PreferredRole,
		})
	}

	return out, nil
}
