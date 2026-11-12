//go:build integration

package choice

import (
	"context"
	"os"
	"testing"

	deckrepo "winx-profile/internal/app/domain/repositories/deck"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requires WINX_TEST_DSN pointing at a database with the profile migrations
// applied. Users 201..204 and two test courses are created and cleaned up.
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

func TestRepository_ChoicesShortlistAndDeck(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()

	cleanup := func() {
		db.Exec(`DELETE FROM users WHERE id BETWEEN 201 AND 204`)
		db.Exec(`DELETE FROM courses WHERE code IN ('TSTCHOICEA', 'TSTCHOICEB')`)
	}
	cleanup()
	t.Cleanup(cleanup)

	for id := 201; id <= 204; id++ {
		db.Exec(`INSERT INTO users (id, email, password) VALUES (?, 'it-' || (?::bigint)::text || '@test.io', 'x')`, id, id)
	}

	db.Exec(`INSERT INTO profiles (user_id, email, name) VALUES (202, 'b@x.io', 'Bob'), (203, 'c@x.io', 'Cat')`)

	var courseA, courseB int64
	db.Raw(`INSERT INTO courses (code, title) VALUES ('TSTCHOICEA', 'A') RETURNING id`).Scan(&courseA)
	db.Raw(`INSERT INTO courses (code, title) VALUES ('TSTCHOICEB', 'B') RETURNING id`).Scan(&courseB)

	// 201,202,203 in A; 204 only in B.
	db.Exec(`INSERT INTO user_courses (user_id, course_id) VALUES (201, ?), (202, ?), (203, ?), (204, ?), (201, ?)`,
		courseA, courseA, courseA, courseB, courseB)

	repo := NewRepository(db)

	stored, err := repo.Upsert(ctx, 201, 202, courseA, "like")
	if err != nil || !stored {
		t.Fatalf("like 202: stored=%v err=%v", stored, err)
	}

	// Retry and change must keep exactly one row.
	for _, c := range []string{"like", "like", "pass", "like"} {
		if ok, err := repo.Upsert(ctx, 201, 202, courseA, c); err != nil || !ok {
			t.Fatalf("upsert %s: ok=%v err=%v", c, ok, err)
		}
	}

	var rowsFor202 int64
	db.Raw(`SELECT count(*) FROM course_choices WHERE user_id = 201 AND candidate_id = 202 AND course_id = ?`, courseA).Scan(&rowsFor202)
	if rowsFor202 != 1 {
		t.Fatalf("exactly one persisted choice expected, got %d", rowsFor202)
	}

	// Candidate that is not in the course is refused and nothing is written.
	if ok, err := repo.Upsert(ctx, 201, 204, courseA, "like"); err != nil || ok {
		t.Fatalf("candidate outside course: ok=%v err=%v", ok, err)
	}

	if ok, _ := repo.Upsert(ctx, 201, 203, courseA, "pass"); !ok {
		t.Fatal("pass 203 should be stored")
	}

	shortlist, err := repo.ListShortlist(ctx, 201, courseA, 20, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(shortlist) != 1 || shortlist[0].UserID != 202 || shortlist[0].Name != "Bob" {
		t.Fatalf("shortlist must hold only likes: %+v", shortlist)
	}

	// Course isolation: a choice in A must not show up in B's shortlist.
	other, _ := repo.ListShortlist(ctx, 201, courseB, 20, 0)
	if len(other) != 0 {
		t.Fatalf("course B shortlist must be empty: %+v", other)
	}

	// Pass is private: the candidate sees nothing about it.
	theirs, _ := repo.ListShortlist(ctx, 203, courseA, 20, 0)
	if len(theirs) != 0 {
		t.Fatalf("candidate must not see decisions made about them: %+v", theirs)
	}

	// Decided candidates (like and pass) disappear from the deck.
	deck, err := deckrepo.NewRepository(db).ListCandidates(ctx, 201, courseA, 20, 0)
	if err != nil || len(deck) != 0 {
		t.Fatalf("deck must hide already decided candidates: %+v err=%v", deck, err)
	}

}
