//go:build integration

package deck

import (
	"context"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requires WINX_TEST_DSN pointing at a database with the profile migrations
// applied. Users 101..104 are created and cleaned up by the test.
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

func TestRepository_DeckFiltersByCourseAndExcludesSelf(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()

	cleanup := func() {
		db.Exec(`DELETE FROM users WHERE id BETWEEN 101 AND 104`)
		db.Exec(`DELETE FROM courses WHERE code IN ('TSTDECKA', 'TSTDECKB')`)
	}
	cleanup()
	t.Cleanup(cleanup)

	for id := 101; id <= 104; id++ {
		db.Exec(`INSERT INTO users (id) VALUES (?)`, id)
	}

	db.Exec(`INSERT INTO profiles (user_id, email, name) VALUES (102, 'b@x.io', 'Bob'), (103, 'c@x.io', 'Cat')`)
	db.Exec(`INSERT INTO academic_profiles (user_id, skills, preferred_role) VALUES (102, '["Go","SQL"]', 'Backend')`)

	// Dedicated courses keep the test independent of any other data in the database.
	var courseA, courseB int64
	db.Raw(`INSERT INTO courses (code, title) VALUES ('TSTDECKA', 'Deck test A') RETURNING id`).Scan(&courseA)
	db.Raw(`INSERT INTO courses (code, title) VALUES ('TSTDECKB', 'Deck test B') RETURNING id`).Scan(&courseB)

	db.Exec(`INSERT INTO user_courses (user_id, course_id) VALUES (101, ?), (102, ?), (103, ?), (104, ?)`,
		courseA, courseA, courseA, courseB)

	repo := NewRepository(db)

	has, err := repo.HasCourse(ctx, 101, courseA)
	if err != nil || !has {
		t.Fatalf("101 selected course A: has=%v err=%v", has, err)
	}

	has, _ = repo.HasCourse(ctx, 101, courseB)
	if has {
		t.Fatal("101 did not select course B")
	}

	got, err := repo.ListCandidates(ctx, 101, courseA, 20, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 || got[0].UserID != 102 || got[1].UserID != 103 {
		t.Fatalf("expected users 102,103 only (no self, no other course): %+v", got)
	}

	if got[0].Name != "Bob" || len(got[0].Skills) != 2 || got[0].PreferredRole != "Backend" {
		t.Fatalf("academic fields wrong: %+v", got[0])
	}

	if got[1].Skills == nil || len(got[1].Skills) != 0 {
		t.Fatalf("missing academic profile must yield empty skills: %+v", got[1])
	}

	page, _ := repo.ListCandidates(ctx, 101, courseA, 1, 1)
	if len(page) != 1 || page[0].UserID != 103 {
		t.Fatalf("paging broken: %+v", page)
	}
}
