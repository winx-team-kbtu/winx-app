package erasure

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	PhotoObjectKeys(ctx context.Context, userID int64) ([]string, error)
	// DeleteUserRows removes every PostgreSQL row owned by the profile service for the user,
	// leaves teams without members out of the database and keeps remaining teams consistent.
	// It is idempotent: running it for a user with no data is a no-op.
	DeleteUserRows(ctx context.Context, userID int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) PhotoObjectKeys(ctx context.Context, userID int64) ([]string, error) {
	var keys []string

	if err := r.db.WithContext(ctx).Raw(
		`SELECT object_key FROM profile_photos WHERE user_id = ? AND object_key <> ''`, userID,
	).Scan(&keys).Error; err != nil {
		return nil, fmt.Errorf("list photo object keys: %w", err)
	}

	return keys, nil
}

// userTables are deleted by user_id, in an order that never violates a foreign key.
var userTables = []string{
	"user_courses",
	"academic_profiles",
	"user_interests",
	"matching_preferences",
	"profile_photos",
	"profiles",
}

func (r *repository) DeleteUserRows(ctx context.Context, userID int64) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var teamIDs []int64

		// Lock the user's teams first. Joins lock the same row, so a join can never slip
		// into a team that is being emptied and removed.
		if err := tx.Raw(
			`SELECT id FROM teams
			  WHERE id IN (SELECT team_id FROM team_members WHERE user_id = ?)
			  ORDER BY id FOR UPDATE`, userID,
		).Scan(&teamIDs).Error; err != nil {
			return fmt.Errorf("lock user teams: %w", err)
		}

		if err := tx.Exec(`DELETE FROM team_members WHERE user_id = ?`, userID).Error; err != nil {
			return fmt.Errorf("delete team membership: %w", err)
		}

		if len(teamIDs) > 0 {
			if err := tx.Exec(
				`DELETE FROM teams t
				  WHERE t.id IN ?
				    AND NOT EXISTS (SELECT 1 FROM team_members m WHERE m.team_id = t.id)`, teamIDs,
			).Error; err != nil {
				return fmt.Errorf("delete empty teams: %w", err)
			}
		}

		if err := tx.Exec(`UPDATE teams SET created_by = NULL WHERE created_by = ?`, userID).Error; err != nil {
			return fmt.Errorf("detach team creator: %w", err)
		}

		if err := tx.Exec(
			`DELETE FROM course_choices WHERE user_id = ? OR candidate_id = ?`, userID, userID,
		).Error; err != nil {
			return fmt.Errorf("delete course choices: %w", err)
		}

		for _, table := range userTables {
			if err := tx.Exec(`DELETE FROM `+table+` WHERE user_id = ?`, userID).Error; err != nil {
				return fmt.Errorf("delete from %s: %w", table, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("delete user rows: %w", err)
	}

	return nil
}
