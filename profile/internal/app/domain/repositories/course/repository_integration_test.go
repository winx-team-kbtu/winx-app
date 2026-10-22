//go:build integration

package course

import (
	"context"
	"os"
	"testing"

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

	return db
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
