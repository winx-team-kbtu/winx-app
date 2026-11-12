package team

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"
	"unicode/utf8"

	repository "winx-profile/internal/app/domain/repositories/team"
	"winx-profile/internal/app/models/models"

	"gorm.io/gorm"
)

const (
	MinNameLength = 2
	MaxNameLength = 50

	codeLength   = 6
	codeAttempts = 5
	// No 0/O/1/I so a code read aloud or typed from a screenshot is unambiguous.
	codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

var (
	ErrCourseNotSelected = errors.New("course is not selected")
	ErrInvalidName       = errors.New("invalid team name")
	ErrAlreadyInTeam     = repository.ErrAlreadyInTeam
	ErrNotFound          = repository.ErrNotFound
)

type Service interface {
	Create(ctx context.Context, userID, courseID int64, name string) (models.Team, error)
	GetMine(ctx context.Context, userID, courseID int64) (models.Team, error)
}

type service struct {
	repository repository.Repository
	newCode    func() (string, error)
}

func NewService(db *gorm.DB) Service {
	return NewServiceWithRepository(repository.NewRepository(db), generateCode)
}

func NewServiceWithRepository(repo repository.Repository, newCode func() (string, error)) Service {
	return &service{repository: repo, newCode: newCode}
}

// Create makes a team for a course the user selected and enrols the creator as
// its first, ordinary member. A user can belong to only one team per course.
func (s *service) Create(ctx context.Context, userID, courseID int64, name string) (models.Team, error) {
	name = strings.TrimSpace(name)
	if n := utf8.RuneCountInString(name); n < MinNameLength || n > MaxNameLength {
		return models.Team{}, ErrInvalidName
	}

	selected, err := s.repository.HasCourse(ctx, userID, courseID)
	if err != nil {
		return models.Team{}, err
	}

	if !selected {
		return models.Team{}, ErrCourseNotSelected
	}

	// A join-code collision is rare but possible, so retry with a fresh code.
	for attempt := 0; attempt < codeAttempts; attempt++ {
		code, err := s.newCode()
		if err != nil {
			return models.Team{}, err
		}

		team, err := s.repository.Create(ctx, userID, courseID, name, code)
		if errors.Is(err, repository.ErrCodeTaken) {
			continue
		}

		return team, err
	}

	return models.Team{}, repository.ErrCodeTaken
}

func (s *service) GetMine(ctx context.Context, userID, courseID int64) (models.Team, error) {
	return s.repository.GetByUserCourse(ctx, userID, courseID)
}

func generateCode() (string, error) {
	buf := make([]byte, codeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	var b strings.Builder
	for _, v := range buf {
		b.WriteByte(codeAlphabet[int(v)%len(codeAlphabet)])
	}

	return b.String(), nil
}
