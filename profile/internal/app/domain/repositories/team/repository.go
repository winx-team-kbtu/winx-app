package team

import (
	"context"
	"errors"
	"fmt"

	"winx-profile/internal/app/models/models"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const (
	uniqueViolation = "23505"

	constraintJoinCode   = "teams_join_code_key"
	constraintMembership = "team_members_user_course_key"
)

var (
	ErrNotFound      = errors.New("team not found")
	ErrAlreadyInTeam = errors.New("already in a team for this course")
	ErrCodeTaken     = errors.New("join code already taken")
)

type Repository interface {
	HasCourse(ctx context.Context, userID, courseID int64) (bool, error)
	// Create stores the team and its first member atomically.
	Create(ctx context.Context, userID, courseID int64, name, joinCode string) (models.Team, error)
	GetByUserCourse(ctx context.Context, userID, courseID int64) (models.Team, error)
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

func (r *repository) Create(ctx context.Context, userID, courseID int64, name, joinCode string) (models.Team, error) {
	var team models.Team

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(
			`INSERT INTO teams (course_id, name, join_code, created_by)
			 VALUES (?, ?, ?, ?)
			 RETURNING id, course_id, name, join_code, created_at`,
			courseID, name, joinCode, userID,
		).Scan(&team).Error; err != nil {
			return err
		}

		return tx.Exec(
			`INSERT INTO team_members (team_id, user_id, course_id) VALUES (?, ?, ?)`,
			team.ID, userID, courseID,
		).Error
	})
	if err != nil {
		return models.Team{}, mapCreateError(err)
	}

	team.MemberCount = 1

	return team, nil
}

func mapCreateError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		switch pgErr.ConstraintName {
		case constraintJoinCode:
			return ErrCodeTaken
		case constraintMembership:
			return ErrAlreadyInTeam
		}
	}

	return fmt.Errorf("create team: %w", err)
}

func (r *repository) GetByUserCourse(ctx context.Context, userID, courseID int64) (models.Team, error) {
	var team models.Team

	res := r.db.WithContext(ctx).Raw(
		`SELECT t.id, t.course_id, t.name, t.join_code, t.created_at,
		        (SELECT COUNT(*) FROM team_members m WHERE m.team_id = t.id) AS member_count
		   FROM teams t
		   JOIN team_members me ON me.team_id = t.id
		  WHERE me.user_id = ? AND t.course_id = ?`,
		userID, courseID,
	).Scan(&team)
	if res.Error != nil {
		return models.Team{}, fmt.Errorf("get team: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return models.Team{}, ErrNotFound
	}

	return team, nil
}
