//go:build integration

package course

import (
	"context"
	"errors"
	"os"
	"testing"

	"winx-profile/internal/app/domain/repositories/team"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requires WINX_TEST_DSN pointing at a database with the profile migrations
// applied and users rows with id 1 and 2.
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

	for _, id := range []int{1, 2} {
		db.Exec(`INSERT INTO users (id, email, password) VALUES (?, 'it-' || (?::bigint)::text || '@test.io', 'x') ON CONFLICT DO NOTHING`, id, id)
	}

	return db
}

func TestRepository_EnrollmentCannotBeRemovedWhileInTeam(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	repo := NewRepository(db)
	catalogue, err := repo.List(ctx)
	if err != nil || len(catalogue) < 2 {
		t.Fatalf("catalogue: %v %v", catalogue, err)
	}
	first, second := catalogue[0].ID, catalogue[1].ID
	if err := repo.ReplaceUserCourses(ctx, 1, []int64{first, second}); err != nil {
		t.Fatal(err)
	}
	created, err := team.NewRepository(db).Create(ctx, 1, first, "Project", "LOCK22")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM teams WHERE id=?`, created.ID) })
	if err := repo.ReplaceUserCourses(ctx, 1, nil); !errors.Is(err, ErrTeamCourse) {
		t.Fatalf("removing a team course: %v", err)
	}
	if got, err := repo.ListByUserID(ctx, 1); err != nil || len(got) != 2 {
		t.Fatalf("failed edit changed another course: %v %v", got, err)
	}
	if err := repo.ReplaceUserCourses(ctx, 1, []int64{first}); err != nil {
		t.Fatalf("removing an unrelated course: %v", err)
	}
}

func TestRepository_ReplaceAndListUserCourses(t *testing.T) {
	db := openDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	db.Exec(`DELETE FROM user_courses WHERE user_id IN (1, 2)`)

	catalogue, err := repo.List(ctx)
	if err != nil || len(catalogue) < 3 {
		t.Fatalf("seeded catalogue expected, got %d err=%v", len(catalogue), err)
	}

	if err := repo.ReplaceUserCourses(ctx, 1, []int64{catalogue[0].ID, catalogue[1].ID}); err != nil {
		t.Fatal(err)
	}

	if err := repo.ReplaceUserCourses(ctx, 1, []int64{catalogue[2].ID}); err != nil {
		t.Fatal(err)
	}

	mine, err := repo.ListByUserID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(mine) != 1 || mine[0].ID != catalogue[2].ID {
		t.Fatalf("selection must be replaced: %+v", mine)
	}

	other, err := repo.ListByUserID(ctx, 2)
	if err != nil || len(other) != 0 {
		t.Fatalf("other user must have no courses: %+v err=%v", other, err)
	}
}

func TestRepository_ReplaceIsAtomicOnFailure(t *testing.T) {
	db := openDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	db.Exec(`DELETE FROM user_courses WHERE user_id = 1`)

	catalogue, _ := repo.List(ctx)
	if err := repo.ReplaceUserCourses(ctx, 1, []int64{catalogue[0].ID}); err != nil {
		t.Fatal(err)
	}

	// Unknown course id violates the FK mid-transaction: previous selection must survive.
	if err := repo.ReplaceUserCourses(ctx, 1, []int64{catalogue[1].ID, 999999}); err == nil {
		t.Fatal("expected FK violation")
	}

	mine, err := repo.ListByUserID(ctx, 1)
	if err != nil || len(mine) != 1 || mine[0].ID != catalogue[0].ID {
		t.Fatalf("failed replace must roll back: %+v err=%v", mine, err)
	}
}
