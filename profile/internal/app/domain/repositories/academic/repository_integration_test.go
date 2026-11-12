//go:build integration

package academic

import (
	"context"
	"errors"
	"os"
	"testing"

	"winx-profile/internal/app/models/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requires WINX_TEST_DSN pointing at a database with the profile migrations
// applied and a users row with id 1.
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

func TestRepository_UpsertAndGet(t *testing.T) {
	db := openDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	db.Exec(`DELETE FROM academic_profiles WHERE user_id = 1`)

	if _, err := repo.GetByUserID(ctx, 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if _, err := repo.Upsert(ctx, models.AcademicProfile{UserID: 1, Skills: []string{"Go", "SQL"}, PreferredRole: "Backend"}); err != nil {
		t.Fatal(err)
	}

	updated, err := repo.Upsert(ctx, models.AcademicProfile{UserID: 1, Skills: []string{"Figma"}, PreferredRole: "Designer"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByUserID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Skills) != 1 || got.Skills[0] != "Figma" || got.PreferredRole != "Designer" {
		t.Fatalf("second upsert must replace first: %+v", got)
	}

	if !updated.UpdatedAt.Equal(got.UpdatedAt) {
		t.Fatalf("returning row differs from stored row")
	}
}
