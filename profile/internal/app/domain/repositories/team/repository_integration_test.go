//go:build integration

package team

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	deckrepo "winx-profile/internal/app/domain/repositories/deck"
	"winx-profile/internal/app/models/models"

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
		db.Exec(`DELETE FROM users WHERE id BETWEEN 401 AND 404 OR id BETWEEN 410 AND 429`)
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

func addEligibleUsers(db *gorm.DB, courseID int64, from, to int) {
	for id := from; id <= to; id++ {
		db.Exec(`INSERT INTO users (id) VALUES (?)`, id)
		db.Exec(`INSERT INTO user_courses (user_id, course_id) VALUES (?, ?)`, id, courseID)
	}
}

func joinConcurrently(repo Repository, code string, from, to int) (succeeded, full int) {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	start := make(chan struct{})

	for id := from; id <= to; id++ {
		wg.Add(1)

		go func(userID int64) {
			defer wg.Done()

			<-start

			_, err := repo.Join(context.Background(), userID, code)

			mu.Lock()
			defer mu.Unlock()

			switch {
			case err == nil:
				succeeded++
			case errors.Is(err, ErrTeamFull):
				full++
			}
		}(int64(id))
	}

	close(start)
	wg.Wait()

	return succeeded, full
}

func memberCount(db *gorm.DB, courseID int64) int64 {
	var n int64
	db.Raw(`SELECT count(*) FROM team_members WHERE course_id = ?`, courseID).Scan(&n)

	return n
}

func TestRepository_TeamNeverExceedsFiveUnderRace(t *testing.T) {
	db, courseID := setup(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// 401 creates the team; 402..404 join sequentially -> size 4.
	if _, err := repo.Create(ctx, 401, courseID, "Race", "RACE22"); err != nil {
		t.Fatal(err)
	}

	for _, id := range []int64{402, 403, 404} {
		if _, err := repo.Join(ctx, id, "RACE22"); err != nil {
			t.Fatalf("setup join %d: %v", id, err)
		}
	}

	// Ten students race for the single remaining seat.
	addEligibleUsers(db, courseID, 410, 419)

	succeeded, full := joinConcurrently(repo, "RACE22", 410, 419)

	if succeeded != 1 || full != 9 {
		t.Fatalf("size-four race: expected 1 success and 9 full, got success=%d full=%d", succeeded, full)
	}

	if got := memberCount(db, courseID); got != 5 {
		t.Fatalf("team must have exactly 5 members, got %d", got)
	}
}

func TestRepository_ManyJoinersFillExactlyToFive(t *testing.T) {
	db, courseID := setup(t)
	repo := NewRepository(db)

	if _, err := repo.Create(context.Background(), 401, courseID, "Fill", "FILL22"); err != nil {
		t.Fatal(err)
	}

	addEligibleUsers(db, courseID, 410, 429)

	succeeded, full := joinConcurrently(repo, "FILL22", 410, 429)

	if succeeded != 4 || full != 16 {
		t.Fatalf("expected 4 successes and 16 full, got success=%d full=%d", succeeded, full)
	}

	if got := memberCount(db, courseID); got != 5 {
		t.Fatalf("team must have exactly 5 members, got %d", got)
	}
}

func TestRepository_FullTeamRejectsWithoutMutation(t *testing.T) {
	db, courseID := setup(t)
	repo := NewRepository(db)
	ctx := context.Background()

	if _, err := repo.Create(ctx, 401, courseID, "Full", "FULL22"); err != nil {
		t.Fatal(err)
	}

	addEligibleUsers(db, courseID, 410, 413)

	for id := int64(410); id <= 413; id++ {
		if _, err := repo.Join(ctx, id, "FULL22"); err != nil {
			t.Fatalf("join %d: %v", id, err)
		}
	}

	if _, err := repo.Join(ctx, 402, "FULL22"); !errors.Is(err, ErrTeamFull) {
		t.Fatalf("sixth member: expected ErrTeamFull, got %v", err)
	}

	if _, err := repo.GetByUserCourse(ctx, 402, courseID); !errors.Is(err, ErrNotFound) {
		t.Fatal("rejected student must not be enrolled")
	}

	if got := memberCount(db, courseID); got != 5 {
		t.Fatalf("roster changed by a rejected join: %d", got)
	}
}

func TestRepository_RosterShowsAcademicFieldsToMembersOnly(t *testing.T) {
	db, courseID := setup(t)
	repo := NewRepository(db)
	ctx := context.Background()

	db.Exec(`INSERT INTO profiles (user_id, email, name) VALUES (401, 'secret-a@x.io', 'Ann'), (402, 'secret-b@x.io', 'Bob')`)
	db.Exec(`INSERT INTO academic_profiles (user_id, skills, preferred_role) VALUES (401, '["Go"]', 'Backend')`)

	if _, err := repo.Create(ctx, 401, courseID, "Alpha", "ROST22"); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.Join(ctx, 402, "ROST22"); err != nil {
		t.Fatal(err)
	}

	roster, err := repo.Roster(ctx, 402, courseID)
	if err != nil {
		t.Fatal(err)
	}

	if roster.TeamName != "Alpha" || roster.CourseCode != "TSTTEAM" || len(roster.Members) != 2 {
		t.Fatalf("unexpected roster %+v", roster)
	}

	if roster.Members[0].UserID != 401 || roster.Members[0].Name != "Ann" || roster.Members[0].PreferredRole != "Backend" || len(roster.Members[0].Skills) != 1 {
		t.Fatalf("creator entry wrong: %+v", roster.Members[0])
	}

	if roster.Members[1].UserID != 402 || roster.Members[1].Skills == nil || len(roster.Members[1].Skills) != 0 {
		t.Fatalf("member without academic profile must have empty skills: %+v", roster.Members[1])
	}

	if _, err := repo.Roster(ctx, 403, courseID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("outsider must get ErrNotFound, got %v", err)
	}
}

func TestRepository_StatusFollowsMembership(t *testing.T) {
	db, courseID := setup(t)
	repo := NewRepository(db)
	ctx := context.Background()

	team, err := repo.Create(ctx, 401, courseID, "Status", "STAT22")
	if err != nil {
		t.Fatal(err)
	}

	expect := func(want models.TeamStatus) {
		t.Helper()

		mine, err := repo.GetByUserCourse(ctx, 401, courseID)
		if err != nil {
			t.Fatal(err)
		}

		if got := mine.Status(); got != want {
			t.Fatalf("size %d: expected %s, got %s", mine.MemberCount, want, got)
		}
	}

	expect(models.TeamStatusForming) // 1

	if _, err := repo.Join(ctx, 402, "STAT22"); err != nil {
		t.Fatal(err)
	}

	expect(models.TeamStatusForming) // 2

	if _, err := repo.Join(ctx, 403, "STAT22"); err != nil {
		t.Fatal(err)
	}

	expect(models.TeamStatusReady) // 3: third join turns the team Ready

	if _, err := repo.Join(ctx, 404, "STAT22"); err != nil {
		t.Fatal(err)
	}

	addEligibleUsers(db, courseID, 410, 410)

	if _, err := repo.Join(ctx, 410, "STAT22"); err != nil {
		t.Fatal(err)
	}

	expect(models.TeamStatusFull) // 5

	// Removing a member (account deletion cascades) recalculates the status.
	db.Exec(`DELETE FROM users WHERE id = 410`)
	expect(models.TeamStatusReady) // 4

	db.Exec(`DELETE FROM users WHERE id IN (403, 404)`)
	expect(models.TeamStatusForming) // 2

	_ = team
}
