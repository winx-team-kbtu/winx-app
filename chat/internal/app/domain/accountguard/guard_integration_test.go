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

	cleanup := func() { db.Exec(`DELETE FROM users WHERE id BETWEEN 801 AND 803`) }
	cleanup()
	t.Cleanup(cleanup)

	for id := 801; id <= 803; id++ {
		require.NoError(t, db.Exec(`INSERT INTO users (id, email, password) VALUES (?, ?, 'x')`, id, "chatguard"+string(rune('0'+id-800))+"@test.io").Error)
	}

	return db
}

func TestUsersActive(t *testing.T) {
	db := setup(t)
	guard := New(db)
	ctx := context.Background()

	ok, err := guard.UsersActive(ctx, 801, 802)
	require.NoError(t, err)
	assert.True(t, ok, "two existing, ordinary users")

	require.NoError(t, db.Exec(`UPDATE users SET deletion_requested_at = NOW() WHERE id = 802`).Error)

	ok, err = guard.UsersActive(ctx, 801, 802)
	require.NoError(t, err)
	assert.False(t, ok, "one user is being deleted")

	ok, err = guard.UsersActive(ctx, 801, 999999)
	require.NoError(t, err)
	assert.False(t, ok, "one user no longer exists")

	ok, err = guard.UsersActive(ctx, 801, 801)
	require.NoError(t, err)
	assert.True(t, ok, "duplicate ids are counted once")

	ok, err = guard.UsersActive(ctx)
	require.NoError(t, err)
	assert.False(t, ok, "no users is never active")
}
