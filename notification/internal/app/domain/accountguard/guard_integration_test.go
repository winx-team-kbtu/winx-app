//go:build integration

package accountguard

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requires WINX_TEST_DSN pointing at a database with the auth migrations applied.
func setup(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("WINX_TEST_DSN")
	if dsn == "" {
		t.Skip("WINX_TEST_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	cleanup := func() { db.Exec(`DELETE FROM users WHERE id BETWEEN 811 AND 813`) }
	cleanup()
	t.Cleanup(cleanup)

	require.NoError(t, db.Exec(`INSERT INTO users (id, email, password) VALUES (811, 'guard811@test.io', 'x'), (812, 'guard812@test.io', 'x')`).Error)

	return db
}

func TestAccountActive_RequiresTheSameAccountNotJustTheEmail(t *testing.T) {
	db := setup(t)
	guard := New(db)
	ctx := context.Background()

	ok, err := guard.AccountActive(ctx, 811, "guard811@test.io")
	require.NoError(t, err)
	assert.True(t, ok)

	// The old account (id 811) was deleted and someone registered the same email again.
	require.NoError(t, db.Exec(`DELETE FROM users WHERE id = 811`).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, email, password) VALUES (813, 'guard811@test.io', 'x')`).Error)

	ok, err = guard.AccountActive(ctx, 811, "guard811@test.io")
	require.NoError(t, err)
	assert.False(t, ok, "a stale event about the old account must not count for the new owner of the email")

	ok, err = guard.AccountActive(ctx, 813, "guard811@test.io")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestGuard_PendingDeletionCountsAsInactive(t *testing.T) {
	db := setup(t)
	guard := New(db)
	ctx := context.Background()

	require.NoError(t, db.Exec(`UPDATE users SET deletion_requested_at = NOW() WHERE id = 812`).Error)

	byAccount, err := guard.AccountActive(ctx, 812, "guard812@test.io")
	require.NoError(t, err)
	assert.False(t, byAccount)

	byEmail, err := guard.EmailActive(ctx, "guard812@test.io")
	require.NoError(t, err)
	assert.False(t, byEmail)

	byID, err := guard.UserActive(ctx, 812)
	require.NoError(t, err)
	assert.False(t, byID)
}

func TestGuard_UnknownAccountIsInactive(t *testing.T) {
	guard := New(setup(t))
	ctx := context.Background()

	ok, err := guard.AccountActive(ctx, 999999, "nobody@test.io")
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = guard.EmailActive(ctx, "nobody@test.io")
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = guard.UserActive(ctx, 999999)
	require.NoError(t, err)
	assert.False(t, ok)
}
