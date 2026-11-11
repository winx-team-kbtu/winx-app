//go:build integration

package deletion

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	profileclient "auth/adapters/profile-client"
	oauthHelper "auth/internal/app/core/helpers/oauth"
	repository "auth/internal/app/domain/repositories/deletion"
	"auth/pkg/cache"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requires WINX_TEST_DSN (database with all service migrations applied) and WINX_TEST_REDIS
// (host:port). Users 601.. and their related rows are created and removed by the tests.
type env struct {
	db    *gorm.DB
	rdb   *redis.Client
	cache cache.Cache
}

func setup(t *testing.T) env {
	t.Helper()

	dsn, redisAddr := os.Getenv("WINX_TEST_DSN"), os.Getenv("WINX_TEST_REDIS")
	if dsn == "" || redisAddr == "" {
		t.Skip("WINX_TEST_DSN and WINX_TEST_REDIS are required")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	require.NoError(t, rdb.Ping(context.Background()).Err())

	t.Cleanup(func() {
		db.Exec(`DELETE FROM users WHERE id BETWEEN 601 AND 610`)
		db.Exec(`DELETE FROM notifications WHERE recipient LIKE 'del%@test.io'`)
		db.Exec(`DELETE FROM password_resets WHERE email LIKE 'del%@test.io'`)
		db.Exec(`DELETE FROM oauth_tokens WHERE user_id IN ('601','602')`)
		db.Exec(`DELETE FROM swipes WHERE swiper_id BETWEEN 601 AND 610 OR swiped_id BETWEEN 601 AND 610`)
		db.Exec(`DELETE FROM chats WHERE user_one_id BETWEEN 601 AND 610 OR user_two_id BETWEEN 601 AND 610`)
		db.Exec(`DELETE FROM matches WHERE user_one_id BETWEEN 601 AND 610 OR user_two_id BETWEEN 601 AND 610`)
		db.Exec(`DELETE FROM recommendations WHERE user_id BETWEEN 601 AND 610`)
		rdb.Del(context.Background(), "users:user_id:601", "users:user_email:601", "users:access_token:tokA", "users:access_token:tokB")
	})

	return env{db: db, rdb: rdb, cache: cache.NewRedisCache(rdb, "users")}
}

func (e env) createUser(t *testing.T, id int64, email, password string) {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)

	e.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	require.NoError(t, e.db.Exec(`INSERT INTO users (id, email, password) VALUES (?, ?, ?)`, id, email, string(hash)).Error)
}

func (e env) count(t *testing.T, query string, args ...any) int64 {
	t.Helper()

	var n int64
	require.NoError(t, e.db.Raw(query, args...).Scan(&n).Error)

	return n
}

func (e env) profileServer(status *atomic.Int32, calls *atomic.Int32) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(int(status.Load()))
	}))
}

func (e env) service(profileURL string) Service {
	repo := repository.NewRepository(e.db)

	return NewService(repo, repo, e.cache, profileclient.New(profileURL, "k", time.Second))
}

func TestIntegration_DeleteRemovesAccountSessionsAndLinkedRows(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	e.createUser(t, 601, "del601@test.io", "pw-601")
	e.createUser(t, 602, "del602@test.io", "pw-602")

	e.db.Exec(`INSERT INTO password_resets (email, pin_code, token) VALUES ('del601@test.io', '123456', 'tok-del-601')`)
	e.db.Exec(`INSERT INTO oauth_tokens (client_id, user_id, access, refresh) VALUES ('c', '601', 'tokA', 'refA'), ('c', '601', 'tokB', 'refB'), ('c', '602', 'tokOther', 'refOther')`)
	e.db.Exec(`INSERT INTO swipes (swiper_id, swiped_id, direction) VALUES (601, 602, 'right'), (602, 601, 'right'), (602, 999, 'left')`)
	e.db.Exec(`INSERT INTO matches (user_one_id, user_two_id) VALUES (601, 602)`)

	require.NoError(t, e.cache.Set(ctx, "user_id:601", []byte("tokB"), time.Minute))
	require.NoError(t, e.cache.Set(ctx, "user_email:601", []byte("del601@test.io"), time.Minute))
	require.NoError(t, e.cache.Set(ctx, "access_token:tokA", []byte("601"), time.Minute))
	require.NoError(t, e.cache.Set(ctx, "access_token:tokB", []byte("601"), time.Minute))

	// The data must really exist before deletion, otherwise "zero after" proves nothing.
	require.EqualValues(t, 2, e.count(t, `SELECT count(*) FROM oauth_tokens WHERE user_id = '601'`))
	require.EqualValues(t, 1, e.count(t, `SELECT count(*) FROM password_resets WHERE email = 'del601@test.io'`))
	require.EqualValues(t, 2, e.count(t, `SELECT count(*) FROM swipes WHERE swiper_id = 601 OR swiped_id = 601`))
	require.EqualValues(t, 1, e.count(t, `SELECT count(*) FROM matches WHERE user_one_id = 601 OR user_two_id = 601`))

	before, err := e.rdb.Exists(ctx, "users:user_id:601", "users:user_email:601", "users:access_token:tokA", "users:access_token:tokB").Result()
	require.NoError(t, err)
	require.EqualValues(t, 4, before)

	var status, calls atomic.Int32
	status.Store(http.StatusOK)

	srv := e.profileServer(&status, &calls)
	defer srv.Close()

	require.NoError(t, e.service(srv.URL).Delete(ctx, 601, "pw-601"))

	assert.EqualValues(t, 1, calls.Load(), "the profile service was asked to clean up")
	assert.Zero(t, e.count(t, `SELECT count(*) FROM users WHERE id = 601`), "account removed")
	assert.Zero(t, e.count(t, `SELECT count(*) FROM oauth_tokens WHERE user_id = '601'`), "refresh and access tokens removed")
	assert.Zero(t, e.count(t, `SELECT count(*) FROM password_resets WHERE email = 'del601@test.io'`))
	assert.Zero(t, e.count(t, `SELECT count(*) FROM swipes WHERE swiper_id = 601 OR swiped_id = 601`))
	assert.Zero(t, e.count(t, `SELECT count(*) FROM matches WHERE user_one_id = 601 OR user_two_id = 601`))

	for _, key := range []string{"users:user_id:601", "users:user_email:601", "users:access_token:tokA", "users:access_token:tokB"} {
		exists, err := e.rdb.Exists(ctx, key).Result()
		require.NoError(t, err)
		assert.Zero(t, exists, "%s must be gone so every service rejects the old session", key)
	}

	// Other users are not affected.
	assert.EqualValues(t, 1, e.count(t, `SELECT count(*) FROM users WHERE id = 602`))
	assert.EqualValues(t, 1, e.count(t, `SELECT count(*) FROM oauth_tokens WHERE user_id = '602'`))
	assert.EqualValues(t, 1, e.count(t, `SELECT count(*) FROM swipes WHERE swiper_id = 602 AND swiped_id = 999`), "unrelated swipe stays")
}

func TestIntegration_ProfileFailureKeepsAccountBlockedAndNotDeleted(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	e.createUser(t, 601, "del601@test.io", "pw-601")
	e.db.Exec(`INSERT INTO oauth_tokens (client_id, user_id, access, refresh) VALUES ('c', '601', 'tokA', 'refA')`)
	require.NoError(t, e.cache.Set(ctx, "user_id:601", []byte("tokA"), time.Minute))
	require.NoError(t, e.cache.Set(ctx, "user_email:601", []byte("del601@test.io"), time.Minute))

	require.EqualValues(t, 1, e.count(t, `SELECT count(*) FROM oauth_tokens WHERE user_id = '601'`), "token exists before deletion")

	var status, calls atomic.Int32
	status.Store(http.StatusInternalServerError)

	srv := e.profileServer(&status, &calls)
	defer srv.Close()

	svc := e.service(srv.URL)

	err := svc.Delete(ctx, 601, "pw-601")
	require.ErrorIs(t, err, ErrPending, "must not report success")

	assert.EqualValues(t, 1, e.count(t, `SELECT count(*) FROM users WHERE id = 601`), "user row kept so the cleanup can be retried")
	assert.EqualValues(t, 1, e.count(t, `SELECT count(*) FROM users WHERE id = 601 AND deletion_requested_at IS NOT NULL`), "marked pending")
	assert.Zero(t, e.count(t, `SELECT count(*) FROM oauth_tokens WHERE user_id = '601'`), "sessions are revoked even though cleanup failed")

	exists, err := e.rdb.Exists(ctx, "users:user_id:601", "users:user_email:601").Result()
	require.NoError(t, err)
	assert.Zero(t, exists, "old session is rejected while the deletion is pending")

	// The profile service recovers; repeating completes the deletion.
	status.Store(http.StatusOK)

	require.NoError(t, svc.Delete(ctx, 601, "pw-601"))
	assert.Zero(t, e.count(t, `SELECT count(*) FROM users WHERE id = 601`))
}

func TestIntegration_PendingAccountCannotSignIn(t *testing.T) {
	e := setup(t)

	e.createUser(t, 601, "del601@test.io", "pw-601")

	handler := oauthHelper.NewPasswordHandler(e.db)

	id, err := handler.Password(context.Background(), "", "del601@test.io", "pw-601")
	require.NoError(t, err)
	assert.Equal(t, strconv.Itoa(601), id, "sign-in works before deletion is requested")

	require.NoError(t, repository.NewRepository(e.db).MarkDeletionRequested(context.Background(), 601))

	_, err = handler.Password(context.Background(), "", "del601@test.io", "pw-601")
	require.Error(t, err, "a pending deletion must block sign-in with the correct password")
	assert.EqualError(t, err, "invalid credentials", fmt.Sprintf("must not reveal the reason, got %v", err))
}

func TestIntegration_MarkDeletionRequestedKeepsFirstTimestamp(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	e.createUser(t, 601, "del601@test.io", "pw-601")

	repo := repository.NewRepository(e.db)
	require.NoError(t, repo.MarkDeletionRequested(ctx, 601))

	var first time.Time
	require.NoError(t, e.db.Raw(`SELECT deletion_requested_at FROM users WHERE id = 601`).Row().Scan(&first))

	time.Sleep(20 * time.Millisecond)
	require.NoError(t, repo.MarkDeletionRequested(ctx, 601))

	var second time.Time
	require.NoError(t, e.db.Raw(`SELECT deletion_requested_at FROM users WHERE id = 601`).Row().Scan(&second))

	assert.True(t, first.Equal(second), "repeating the request must keep the original time")
}
