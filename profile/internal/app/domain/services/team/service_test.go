package team

import (
	"context"
	"errors"
	"strings"
	"testing"

	repository "winx-profile/internal/app/domain/repositories/team"
	"winx-profile/internal/app/models/models"
)

type fakeRepo struct {
	selected map[[2]int64]bool
	teams    map[[2]int64]models.Team // keyed by (user, course)
	codes    map[string]bool
	attempts []string
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		selected: map[[2]int64]bool{{1, 10}: true, {2, 10}: true},
		teams:    map[[2]int64]models.Team{},
		codes:    map[string]bool{},
	}
}

func (f *fakeRepo) HasCourse(_ context.Context, userID, courseID int64) (bool, error) {
	return f.selected[[2]int64{userID, courseID}], nil
}

func (f *fakeRepo) Create(_ context.Context, userID, courseID int64, name, code string) (models.Team, error) {
	f.attempts = append(f.attempts, code)

	if _, ok := f.teams[[2]int64{userID, courseID}]; ok {
		return models.Team{}, repository.ErrAlreadyInTeam
	}

	if f.codes[code] {
		return models.Team{}, repository.ErrCodeTaken
	}

	f.codes[code] = true
	t := models.Team{ID: int64(len(f.teams) + 1), CourseID: courseID, Name: name, JoinCode: code, MemberCount: 1}
	f.teams[[2]int64{userID, courseID}] = t

	return t, nil
}

func (f *fakeRepo) GetByUserCourse(_ context.Context, userID, courseID int64) (models.Team, error) {
	t, ok := f.teams[[2]int64{userID, courseID}]
	if !ok {
		return models.Team{}, repository.ErrNotFound
	}

	return t, nil
}

func fixedCodes(codes ...string) func() (string, error) {
	i := 0

	return func() (string, error) {
		c := codes[i%len(codes)]
		i++

		return c, nil
	}
}

func TestCreate_MakesOneMemberTeamWithCode(t *testing.T) {
	svc := NewServiceWithRepository(newFakeRepo(), fixedCodes("ABC234"))

	team, err := svc.Create(context.Background(), 1, 10, "  Dream Team ")
	if err != nil {
		t.Fatal(err)
	}

	if team.Name != "Dream Team" || team.JoinCode != "ABC234" || team.MemberCount != 1 || team.CourseID != 10 {
		t.Fatalf("unexpected team %+v", team)
	}
}

func TestCreate_RejectsDuplicateMembershipPerCourse(t *testing.T) {
	svc := NewServiceWithRepository(newFakeRepo(), fixedCodes("AAAAAA", "BBBBBB"))

	if _, err := svc.Create(context.Background(), 1, 10, "First"); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Create(context.Background(), 1, 10, "Second"); !errors.Is(err, ErrAlreadyInTeam) {
		t.Fatalf("expected ErrAlreadyInTeam, got %v", err)
	}
}

func TestCreate_RejectsUnselectedCourse(t *testing.T) {
	repo := newFakeRepo()

	if _, err := NewServiceWithRepository(repo, fixedCodes("AAAAAA")).Create(context.Background(), 1, 99, "Team"); !errors.Is(err, ErrCourseNotSelected) {
		t.Fatalf("expected ErrCourseNotSelected, got %v", err)
	}

	if len(repo.attempts) != 0 {
		t.Fatal("nothing must be written for an unselected course")
	}
}

func TestCreate_ValidatesName(t *testing.T) {
	for _, name := range []string{"", " ", "a", strings.Repeat("x", MaxNameLength+1)} {
		repo := newFakeRepo()

		if _, err := NewServiceWithRepository(repo, fixedCodes("AAAAAA")).Create(context.Background(), 1, 10, name); !errors.Is(err, ErrInvalidName) {
			t.Fatalf("%q: expected ErrInvalidName, got %v", name, err)
		}

		if len(repo.attempts) != 0 {
			t.Fatalf("%q: invalid name must not reach the repository", name)
		}
	}
}

func TestCreate_RetriesOnJoinCodeCollision(t *testing.T) {
	repo := newFakeRepo()
	repo.codes["TAKEN2"] = true

	team, err := NewServiceWithRepository(repo, fixedCodes("TAKEN2", "TAKEN2", "FRESH3")).Create(context.Background(), 1, 10, "Team")
	if err != nil {
		t.Fatal(err)
	}

	if team.JoinCode != "FRESH3" || len(repo.attempts) != 3 {
		t.Fatalf("expected success on third code, got %+v attempts=%v", team, repo.attempts)
	}
}

func TestCreate_GivesUpAfterRepeatedCollisions(t *testing.T) {
	repo := newFakeRepo()
	repo.codes["SAME22"] = true

	if _, err := NewServiceWithRepository(repo, fixedCodes("SAME22")).Create(context.Background(), 1, 10, "Team"); !errors.Is(err, repository.ErrCodeTaken) {
		t.Fatalf("expected ErrCodeTaken, got %v", err)
	}

	if len(repo.attempts) != codeAttempts {
		t.Fatalf("expected %d attempts, got %d", codeAttempts, len(repo.attempts))
	}
}

func TestGenerateCode_FormatAndAlphabet(t *testing.T) {
	seen := map[string]bool{}

	for i := 0; i < 200; i++ {
		code, err := generateCode()
		if err != nil {
			t.Fatal(err)
		}

		if len(code) != codeLength {
			t.Fatalf("bad length: %q", code)
		}

		for _, r := range code {
			if !strings.ContainsRune(codeAlphabet, r) {
				t.Fatalf("char %q outside alphabet in %q", r, code)
			}
		}

		seen[code] = true
	}

	if len(seen) < 190 {
		t.Fatalf("codes are not random enough: %d unique of 200", len(seen))
	}
}

func TestGetMine_NotFound(t *testing.T) {
	if _, err := NewServiceWithRepository(newFakeRepo(), fixedCodes("AAAAAA")).GetMine(context.Background(), 1, 10); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
