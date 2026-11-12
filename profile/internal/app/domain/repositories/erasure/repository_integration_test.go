//go:build integration

package erasure

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requires WINX_TEST_DSN pointing at a database with the profile migrations applied
// (the auth users table too). Users 501..504 and course TSTERASE are created and removed.
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

func count(t *testing.T, db *gorm.DB, query string, args ...any) int64 {
	t.Helper()

	var n int64
	require.NoError(t, db.Raw(query, args...).Scan(&n).Error)

	return n
}

func seed(t *testing.T) (*gorm.DB, int64) {
	t.Helper()

	db := openDB(t)

	cleanup := func() {
		db.Exec(`DELETE FROM teams WHERE course_id IN (SELECT id FROM courses WHERE code = 'TSTERASE')`)
		db.Exec(`DELETE FROM users WHERE id BETWEEN 501 AND 504`)
		db.Exec(`DELETE FROM courses WHERE code = 'TSTERASE'`)
	}
	cleanup()
	t.Cleanup(cleanup)

	for id := 501; id <= 504; id++ {
		require.NoError(t, db.Exec(`INSERT INTO users (id, email, password) VALUES (?, ?, 'x')`, id, "erase"+string(rune('0'+id-500))+"@test.io").Error)
		require.NoError(t, db.Exec(`INSERT INTO profiles (user_id, email, name) VALUES (?, 'p@test.io', 'N')`, id).Error)
	}

	var courseID int64
	require.NoError(t, db.Raw(`INSERT INTO courses (code, title) VALUES ('TSTERASE', 'Erase') RETURNING id`).Scan(&courseID).Error)

	for id := 501; id <= 504; id++ {
		require.NoError(t, db.Exec(`INSERT INTO user_courses (user_id, course_id) VALUES (?, ?)`, id, courseID).Error)
	}

	return db, courseID
}

func TestDeleteUserRows_RemovesEverythingOwnedByTheUser(t *testing.T) {
	db, courseID := seed(t)
	ctx := context.Background()

	require.NoError(t, db.Exec(`INSERT INTO academic_profiles (user_id, skills, preferred_role) VALUES (501, '["Go"]', 'BE')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO course_choices (user_id, candidate_id, course_id, choice) VALUES (501, 502, ?, 'like'), (502, 501, ?, 'like'), (502, 503, ?, 'pass')`,
		courseID, courseID, courseID).Error)
	require.NoError(t, db.Exec(`INSERT INTO profile_photos (user_id, email, provider, bucket, object_key, url, mime_type, size_bytes) VALUES (501, 'p@test.io', 's3', 'b', 'photos/501.jpg', 'u', 'image/jpeg', 1)`).Error)

	repo := NewRepository(db)

	keys, err := repo.PhotoObjectKeys(ctx, 501)
	require.NoError(t, err)
	require.Equal(t, []string{"photos/501.jpg"}, keys)

	require.NoError(t, repo.DeleteUserRows(ctx, 501))

	for _, table := range []string{"profiles", "academic_profiles", "user_courses", "profile_photos"} {
		require.Zero(t, count(t, db, `SELECT count(*) FROM `+table+` WHERE user_id = 501`), table)
	}

	require.Zero(t, count(t, db, `SELECT count(*) FROM course_choices WHERE user_id = 501 OR candidate_id = 501`),
		"choices made by the user and about the user must both go")

	require.EqualValues(t, 1, count(t, db, `SELECT count(*) FROM course_choices WHERE user_id = 502 AND candidate_id = 503`),
		"unrelated users' choices must stay")
	require.EqualValues(t, 1, count(t, db, `SELECT count(*) FROM profiles WHERE user_id = 502`), "other users untouched")
}

func TestDeleteUserRows_IsIdempotent(t *testing.T) {
	db, _ := seed(t)
	repo := NewRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.DeleteUserRows(ctx, 501))
	require.NoError(t, repo.DeleteUserRows(ctx, 501), "a retry for an already cleaned user must succeed")
	require.NoError(t, repo.DeleteUserRows(ctx, 99999), "an unknown user is a no-op")
}

func TestDeleteUserRows_TeamsShrinkAndEmptyTeamsAreRemoved(t *testing.T) {
	db, courseID := seed(t)
	repo := NewRepository(db)
	ctx := context.Background()

	var teamA, teamB int64
	require.NoError(t, db.Raw(`INSERT INTO teams (course_id, name, join_code, created_by) VALUES (?, 'A', 'ERASE1', 501) RETURNING id`, courseID).Scan(&teamA).Error)
	require.NoError(t, db.Raw(`INSERT INTO teams (course_id, name, join_code, created_by) VALUES (?, 'B', 'ERASE2', 503) RETURNING id`, courseID).Scan(&teamB).Error)

	// Team A: 501 (creator) + 502 + 503... 503 must stay out; use 501 and 502 only.
	require.NoError(t, db.Exec(`INSERT INTO team_members (team_id, user_id, course_id) VALUES (?, 501, ?), (?, 502, ?)`, teamA, courseID, teamA, courseID).Error)
	// Team B: only 503, but 503 stays, so B must be untouched.
	require.NoError(t, db.Exec(`INSERT INTO team_members (team_id, user_id, course_id) VALUES (?, 503, ?)`, teamB, courseID).Error)

	require.NoError(t, repo.DeleteUserRows(ctx, 501))

	require.EqualValues(t, 1, count(t, db, `SELECT count(*) FROM teams WHERE id = ?`, teamA), "team with a remaining member survives")
	require.EqualValues(t, 1, count(t, db, `SELECT count(*) FROM team_members WHERE team_id = ?`, teamA), "size recalculates to 1")
	require.Zero(t, count(t, db, `SELECT count(*) FROM teams WHERE id = ? AND created_by IS NOT NULL`, teamA), "creator link is cleared")
	require.EqualValues(t, 1, count(t, db, `SELECT count(*) FROM teams WHERE id = ?`, teamB), "unrelated team untouched")

	// The last member leaves: the team must disappear.
	require.NoError(t, repo.DeleteUserRows(ctx, 502))
	require.Zero(t, count(t, db, `SELECT count(*) FROM teams WHERE id = ?`, teamA), "empty team must be removed")
	require.EqualValues(t, 1, count(t, db, `SELECT count(*) FROM teams WHERE id = ?`, teamB))
}
