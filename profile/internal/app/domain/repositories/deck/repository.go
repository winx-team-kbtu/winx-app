package deck

import (
	"context"
	"encoding/json"
	"fmt"

	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

type Repository interface {
	HasCourse(ctx context.Context, userID, courseID int64) (bool, error)
	ListCandidates(ctx context.Context, userID, courseID, limit, offset int64) ([]models.DeckCandidate, error)
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

type row struct {
	UserID        int64
	Name          string
	Skills        []byte
	PreferredRole string
}

func (r *repository) ListCandidates(ctx context.Context, userID, courseID, limit, offset int64) ([]models.DeckCandidate, error) {
	var rows []row

	if err := r.db.WithContext(ctx).Raw(
		`SELECT uc.user_id,
		        COALESCE(p.name, '')              AS name,
		        COALESCE(ap.skills, '[]'::jsonb)  AS skills,
		        COALESCE(ap.preferred_role, '')   AS preferred_role
		   FROM user_courses uc
		   LEFT JOIN profiles p            ON p.user_id = uc.user_id
		   LEFT JOIN academic_profiles ap  ON ap.user_id = uc.user_id
		  WHERE uc.course_id = ? AND uc.user_id <> ?
		    AND NOT EXISTS (SELECT 1 FROM course_choices cc
		                     WHERE cc.user_id = ? AND cc.candidate_id = uc.user_id AND cc.course_id = uc.course_id)
		    AND NOT EXISTS (SELECT 1 FROM team_members tm
		                     WHERE tm.user_id = uc.user_id AND tm.course_id = uc.course_id)
		  ORDER BY uc.user_id
		  LIMIT ? OFFSET ?`,
		courseID, userID, userID, limit, offset,
	).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list deck candidates: %w", err)
	}

	out := make([]models.DeckCandidate, 0, len(rows))

	for _, rec := range rows {
		skills := []string{}
		if len(rec.Skills) > 0 {
			if err := json.Unmarshal(rec.Skills, &skills); err != nil {
				return nil, fmt.Errorf("decode candidate skills: %w", err)
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
