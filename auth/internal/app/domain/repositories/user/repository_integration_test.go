//go:build integration

package user

import (
	"context"
	"os"
	"testing"
	"time"

	dto "auth/internal/app/domain/core/dto/repositories/user"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requires WINX_TEST_DSN pointing at a database with the auth migrations applied.
func openDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("WINX_TEST_DSN")
	if dsn == "" {
		t.Skip("WINX_TEST_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	return db
}

func TestRepository_CreatePersistsConsent(t *testing.T) {
	db := openDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	db.Exec(`DELETE FROM users WHERE email IN ('consent@test.io', 'legacy@test.io')`)
	t.Cleanup(func() { db.Exec(`DELETE FROM users WHERE email IN ('consent@test.io', 'legacy@test.io')`) })

	acceptedAt := time.Now().UTC().Truncate(time.Microsecond)

	created, err := repo.Create(ctx, dto.CreateDTO{
		Email:                "consent@test.io",
		Password:             "hash",
		ConsentNoticeVersion: "pilot-2026-11-v1",
		ConsentAcceptedAt:    acceptedAt,
	})
	require.NoError(t, err)

	var version string
	var storedAt time.Time

	require.NoError(t, db.Raw(
		`SELECT consent_notice_version, consent_accepted_at FROM users WHERE id = ?`, created.ID,
	).Row().Scan(&version, &storedAt))

	require.Equal(t, "pilot-2026-11-v1", version)
	require.True(t, acceptedAt.Equal(storedAt), "stored %s, expected %s", storedAt, acceptedAt)
}

func TestRepository_CreateWithoutConsentLeavesColumnsNull(t *testing.T) {
	db := openDB(t)
	repo := NewRepository(db)

	db.Exec(`DELETE FROM users WHERE email = 'legacy@test.io'`)
	t.Cleanup(func() { db.Exec(`DELETE FROM users WHERE email = 'legacy@test.io'`) })

	created, err := repo.Create(context.Background(), dto.CreateDTO{Email: "legacy@test.io", Password: "hash"})
	require.NoError(t, err)

	var nulls bool
	require.NoError(t, db.Raw(
		`SELECT consent_notice_version IS NULL AND consent_accepted_at IS NULL FROM users WHERE id = ?`, created.ID,
	).Scan(&nulls).Error)

	require.True(t, nulls, "admin/legacy accounts must not get a fabricated consent record")
}
