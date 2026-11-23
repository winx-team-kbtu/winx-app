package team

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"winx-profile/internal/app/models/models"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const (
	uniqueViolation = "23505"

	constraintJoinCode   = "teams_join_code_key"
	constraintMembership = "team_members_user_course_key"
	constraintMemberPK   = "team_members_pkey"
)

var (
	ErrNotFound      = errors.New("team not found")
	ErrAlreadyInTeam = errors.New("already in a team for this course")
	ErrCodeTaken     = errors.New("join code already taken")

	ErrCourseNotSelected = errors.New("course is not selected")
	ErrTeamFull          = errors.New("team is full")
)

// MaxMembers is the hard cap on team size.
const MaxMembers = 5

type Repository interface {
	HasCourse(ctx context.Context, userID, courseID int64) (bool, error)
	// Create stores the team and its first member atomically.
	Create(ctx context.Context, userID, courseID int64, name, joinCode string) (models.Team, error)
	GetByUserCourse(ctx context.Context, userID, courseID int64) (models.Team, error)
	// Roster returns the roster of the team the user belongs to for the course.
	Roster(ctx context.Context, userID, courseID int64) (models.TeamRoster, error)
	// Join adds the user to the team that owns joinCode in one transaction.
	Join(ctx context.Context, userID int64, joinCode string) (models.Team, error)
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
	if errors.As(err, &pgErr) && pgErr.ConstraintName == "team_members_enrollment_fkey" {
		return ErrCourseNotSelected
	}
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		switch pgErr.ConstraintName {
		case constraintJoinCode:
			return ErrCodeTaken
		case constraintMembership, constraintMemberPK:
			return ErrAlreadyInTeam
		}
	}

	return fmt.Errorf("team write: %w", err)
}

func (r *repository) GetByUserCourse(ctx context.Context, userID, courseID int64) (models.Team, error) {
	var team models.Team

	res := r.db.WithContext(ctx).Raw(
		`SELECT t.id, t.created_by,t.course_id, t.name, t.join_code, t.created_at,
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

func (r *repository) Join(ctx context.Context, userID int64, joinCode string) (models.Team, error) {
	var team models.Team

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Raw(
			// FOR UPDATE serialises concurrent joins to the same team, so the size
			// check below and the insert are atomic with respect to other joiners.
			`SELECT id, course_id, name, join_code, created_by, created_at FROM teams WHERE join_code = ? FOR UPDATE`, joinCode,
		).Scan(&team)
		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected == 0 {
			return ErrNotFound
		}

		var eligible bool
		if err := tx.Raw(
			`SELECT EXISTS (SELECT 1 FROM user_courses WHERE user_id = ? AND course_id = ?)`,
			userID, team.CourseID,
		).Scan(&eligible).Error; err != nil {
			return err
		}

		if !eligible {
			return ErrCourseNotSelected
		}
		var existingTeam int64
		if err := tx.Raw(`SELECT team_id FROM team_members WHERE user_id=? AND course_id=?`, userID, team.CourseID).Scan(&existingTeam).Error; err != nil {
			return err
		}
		if existingTeam != 0 && existingTeam != team.ID {
			return ErrAlreadyInTeam
		}

		var size int64
		if err := tx.Raw(`SELECT COUNT(*) FROM team_members WHERE team_id = ?`, team.ID).Scan(&size).Error; err != nil {
			return err
		}
		if existingTeam == team.ID {
			team.MemberCount = size
			return nil
		}

		if size >= MaxMembers {
			return ErrTeamFull
		}

		if err := tx.Exec(
			`INSERT INTO team_members (team_id, user_id, course_id) VALUES (?, ?, ?)`,
			team.ID, userID, team.CourseID,
		).Error; err != nil {
			return err
		}

		return tx.Raw(`SELECT COUNT(*) FROM team_members WHERE team_id = ?`, team.ID).Scan(&team.MemberCount).Error
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrCourseNotSelected) || errors.Is(err, ErrTeamFull) || errors.Is(err, ErrAlreadyInTeam) {
			return models.Team{}, err
		}

		return models.Team{}, mapCreateError(err)
	}

	return team, nil
}

func (r *repository) Roster(ctx context.Context, userID, courseID int64) (models.TeamRoster, error) {
	type row struct {
		TeamID, CourseID, UserID                    int64
		TeamName, CourseCode, CourseTitle, JoinCode string
		CreatedBy                                   *int64
		CreatedAt                                   time.Time
		Name, PreferredRole                         string
		Skills                                      []byte
	}

	var rows []row
	if err := r.db.WithContext(ctx).Raw(
		`SELECT t.id AS team_id,t.name AS team_name,t.join_code,t.created_by,t.created_at,
		        c.id AS course_id,c.code AS course_code,c.title AS course_title,m.user_id,
		        COALESCE(p.name,'') AS name,COALESCE(ap.skills,'[]'::jsonb) AS skills,
		        COALESCE(ap.preferred_role,'') AS preferred_role
		   FROM team_members me
		   JOIN teams t ON t.id=me.team_id
		   JOIN courses c ON c.id=t.course_id
		   JOIN team_members m ON m.team_id=t.id
		   LEFT JOIN users active ON active.id=m.user_id AND active.deletion_requested_at IS NULL
		   LEFT JOIN profiles p ON p.user_id=active.id
		   LEFT JOIN academic_profiles ap ON ap.user_id=active.id
		  WHERE me.user_id=? AND me.course_id=?
		  ORDER BY m.joined_at,m.user_id`, userID, courseID,
	).Scan(&rows).Error; err != nil {
		return models.TeamRoster{}, fmt.Errorf("get roster: %w", err)
	}
	if len(rows) == 0 {
		return models.TeamRoster{}, ErrNotFound
	}
	head := rows[0]
	roster := models.TeamRoster{
		TeamID: head.TeamID, TeamName: head.TeamName, CourseID: head.CourseID,
		CourseCode: head.CourseCode, CourseTitle: head.CourseTitle,
		CreatedBy: head.CreatedBy, JoinCode: head.JoinCode, CreatedAt: head.CreatedAt,
		Members: make([]models.DeckCandidate, 0, len(rows)),
	}
	for _, rec := range rows {
		skills := []string{}
		if len(rec.Skills) > 0 {
			if err := json.Unmarshal(rec.Skills, &skills); err != nil {
				return models.TeamRoster{}, fmt.Errorf("decode member skills: %w", err)
			}
		}
		roster.Members = append(roster.Members, models.DeckCandidate{
			UserID: rec.UserID, Name: rec.Name, Skills: skills, PreferredRole: rec.PreferredRole,
		})
	}
	return roster, nil
}
