//go:build integration

package team

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	deckrepo "winx-profile/internal/app/domain/repositories/deck"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requires WINX_TEST_DSN pointing at a database with the profile migrations
// applied. Users 401..404 and a test course are created and cleaned up.
func openDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("WINX_TEST_DSN")
	if dsn == "" {
		t.Skip("WINX_TEST_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func setup(t *testing.T) (*gorm.DB, int64) {
	t.Helper()

	db := openDB(t)

	cleanup := func() {
		db.Exec(`DELETE FROM teams WHERE course_id IN (SELECT id FROM courses WHERE code = 'TSTTEAM')`)
		db.Exec(`DELETE FROM users WHERE id BETWEEN 401 AND 404`)
		db.Exec(`DELETE FROM courses WHERE code = 'TSTTEAM'`)
	}
	cleanup()
	t.Cleanup(cleanup)

	for id := 401; id <= 404; id++ {
		db.Exec(`INSERT INTO users (id) VALUES (?)`, id)
	}

	var courseID int64
	db.Raw(`INSERT INTO courses (code, title) VALUES ('TSTTEAM', 'Team test') RETURNING id`).Scan(&courseID)

	for id := 401; id <= 404; id++ {
		db.Exec(`INSERT INTO user_courses (user_id, course_id) VALUES (?, ?)`, id, courseID)
	}

	return db, courseID
}

func TestRepository_CreateTeamWithFirstMember(t *testing.T) {
	db, courseID := setup(t)
	repo := NewRepository(db)
	ctx := context.Background()

	team, err := repo.Create(ctx, 401, courseID, "Alpha", "AAAAAA")
	if err != nil {
		t.Fatal(err)
	}

	if team.ID == 0 || team.MemberCount != 1 || team.JoinCode != "AAAAAA" {
		t.Fatalf("unexpected team %+v", team)
	}

	mine, err := repo.GetByUserCourse(ctx, 401, courseID)
	if err != nil || mine.ID != team.ID || mine.MemberCount != 1 {
		t.Fatalf("creator must be the single member: %+v err=%v", mine, err)
	}

	if _, err := repo.GetByUserCourse(ctx, 402, courseID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("non-member must get ErrNotFound, got %v", err)
	}
}

func TestRepository_DuplicateMembershipRollsBackTeam(t *testing.T) {
	db, courseID := setup(t)
	repo := NewRepository(db)
	ctx := context.Background()

	if _, err := repo.Create(ctx, 401, courseID, "Alpha", "AAAAAA"); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.Create(ctx, 401, courseID, "Beta", "BBBBBB"); !errors.Is(err, ErrAlreadyInTeam) {
		t.Fatalf("expected ErrAlreadyInTeam, got %v", err)
	}

	var teams int64
	db.Raw(`SELECT count(*) FROM teams WHERE course_id = ?`, courseID).Scan(&teams)

	if teams != 1 {
		t.Fatalf("failed create must not leave an orphan team, found %d", teams)
	}
}

func TestRepository_JoinCodeCollisionIsReported(t *testing.T) {
	db, courseID := setup(t)
	repo := NewRepository(db)
	ctx := context.Background()

	if _, err := repo.Create(ctx, 401, courseID, "Alpha", "AAAAAA"); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.Create(ctx, 402, courseID, "Beta", "AAAAAA"); !errors.Is(err, ErrCodeTaken) {
		t.Fatalf("expected ErrCodeTaken, got %v", err)
	}

	if _, err := repo.GetByUserCourse(ctx, 402, courseID); !errors.Is(err, ErrNotFound) {
		t.Fatal("collision must not enrol the user")
	}
}

func TestRepository_ConcurrentCreateBySameUserYieldsOneTeam(t *testing.T) {
	db, courseID := setup(t)
	repo := NewRepository(db)
	ctx := context.Background()

	codes := []string{"CCCCC1", "CCCCC2", "CCCCC3", "CCCCC4", "CCCCC5", "CCCCC6"}

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		won int
	)

	for _, code := range codes {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if _, err := repo.Create(ctx, 403, courseID, "Racer", code); err == nil {
				mu.Lock()
				won++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	var teams int64
	db.Raw(`SELECT count(*) FROM teams WHERE course_id = ?`, courseID).Scan(&teams)

	if won != 1 || teams != 1 {
		t.Fatalf("exactly one concurrent create must win: won=%d teams=%d", won, teams)
	}
}

func TestRepository_TeamedStudentsLeaveTheDeck(t *testing.T) {
	db, courseID := setup(t)
	ctx := context.Background()

	deck := deckrepo.NewRepository(db)

	before, err := deck.ListCandidates(ctx, 401, courseID, 20, 0)
	if err != nil || len(before) != 3 {
		t.Fatalf("expected 402,403,404 before teaming: %+v err=%v", before, err)
	}

	if _, err := NewRepository(db).Create(ctx, 402, courseID, "Gamma", "GGGGGG"); err != nil {
		t.Fatal(err)
	}

	after, _ := deck.ListCandidates(ctx, 401, courseID, 20, 0)
	if len(after) != 2 {
		t.Fatalf("teamed student must leave the deck: %+v", after)
	}

	for _, c := range after {
		if c.UserID == 402 {
			t.Fatal("402 is teamed and must not be shown")
		}
	}
}

func TestRepository_JoinByCode(t *testing.T) {
	db, courseID := setup(t)
	repo := NewRepository(db)
	ctx := context.Background()

	if _, err := repo.Create(ctx, 401, courseID, "Alpha", "JOIN22"); err != nil {
		t.Fatal(err)
	}

	team, err := repo.Join(ctx, 402, "JOIN22")
	if err != nil || team.MemberCount != 2 || team.CourseID != courseID {
		t.Fatalf("join failed: %+v err=%v", team, err)
	}

	mine, err := repo.GetByUserCourse(ctx, 402, courseID)
	if err != nil || mine.ID != team.ID {
		t.Fatalf("joined user must see the team: %+v err=%v", mine, err)
	}

	if _, err := repo.Join(ctx, 402, "JOIN22"); !errors.Is(err, ErrAlreadyInTeam) {
		t.Fatalf("duplicate join: expected ErrAlreadyInTeam, got %v", err)
	}

	if _, err := repo.Join(ctx, 403, "NOPE22"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown code: expected ErrNotFound, got %v", err)
	}

	// Student who did not select the course is not eligible.
	db.Exec(`INSERT INTO users (id) VALUES (499) ON CONFLICT DO NOTHING`)
	t.Cleanup(func() { db.Exec(`DELETE FROM users WHERE id = 499`) })

	if _, err := repo.Join(ctx, 499, "JOIN22"); !errors.Is(err, ErrCourseNotSelected) {
		t.Fatalf("ineligible: expected ErrCourseNotSelected, got %v", err)
	}

	var members int64
	db.Raw(`SELECT count(*) FROM team_members WHERE team_id = ?`, team.ID).Scan(&members)

	if members != 2 {
		t.Fatalf("rejected joins must not mutate the roster, members=%d", members)
	}
}
