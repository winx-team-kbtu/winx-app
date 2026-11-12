//go:build integration

package deletion

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	profileclient "auth/adapters/profile-client"
	oauthHelper "auth/internal/app/core/helpers/oauth"
	repository "auth/internal/app/domain/repositories/deletion"
	"auth/internal/app/models/models"
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
		db.Exec(`DELETE FROM users WHERE id BETWEEN 601 AND 630`)
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

type userRow = models.User

func intersect(a, b []int64) []int64 {
	set := map[int64]bool{}
	for _, v := range a {
		set[v] = true
	}

	var out []int64

	for _, v := range b {
		if set[v] {
			out = append(out, v)
		}
	}

	return out
}

func (e env) markPending(t *testing.T, id int64, attempts int, lastAttemptAgo time.Duration) {
	t.Helper()

	if lastAttemptAgo == 0 {
		require.NoError(t, e.db.Exec(`UPDATE users SET deletion_requested_at = NOW(), deletion_attempts = ?, deletion_last_attempt_at = NULL WHERE id = ?`, attempts, id).Error)

		return
	}

	require.NoError(t, e.db.Exec(
		`UPDATE users SET deletion_requested_at = NOW(), deletion_attempts = ?,
		        deletion_last_attempt_at = NOW() - make_interval(secs => ?) WHERE id = ?`,
		attempts, lastAttemptAgo.Seconds(), id).Error)
}

func claimedIDs(users []userRow) []int64 {
	ids := make([]int64, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}

	return ids
}

func TestIntegration_ClaimPendingHonoursBackoff(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	repo := repository.NewRepository(e.db)

	for id := int64(601); id <= 606; id++ {
		e.createUser(t, id, fmt.Sprintf("del%d@test.io", id), "pw")
	}

	e.markPending(t, 601, 0, 0)               // never tried: due now
	e.markPending(t, 602, 1, 2*time.Minute)   // wait 60s, 120s passed: due
	e.markPending(t, 603, 1, 10*time.Second)  // wait 60s, 10s passed: not due
	e.markPending(t, 604, 10, 59*time.Minute) // wait capped at 1h, 59m passed: not due
	e.markPending(t, 605, 10, 61*time.Minute) // wait capped at 1h, 61m passed: due
	// 606 is a normal account that never asked for deletion.

	claimed, err := repo.ClaimPending(ctx, 50)
	require.NoError(t, err)

	assert.ElementsMatch(t, []int64{601, 602, 605}, claimedIDs(claimed))

	for _, u := range claimed {
		assert.Positive(t, u.DeletionAttempts, "claiming counts the attempt")
		require.NotNil(t, u.DeletionLastAttemptAt)
	}

	again, err := repo.ClaimPending(ctx, 50)
	require.NoError(t, err)
	assert.Empty(t, again, "a just-claimed account is not handed out again until its backoff passes")

	assert.EqualValues(t, 0, e.count(t, `SELECT deletion_attempts FROM users WHERE id = 606`), "normal accounts are never touched")
}

func TestIntegration_ClaimPendingRespectsLimitAndOrder(t *testing.T) {
	e := setup(t)
	repo := repository.NewRepository(e.db)

	for id := int64(601); id <= 605; id++ {
		e.createUser(t, id, fmt.Sprintf("del%d@test.io", id), "pw")
		e.markPending(t, id, 0, 0)
	}

	first, err := repo.ClaimPending(context.Background(), 2)
	require.NoError(t, err)
	assert.Len(t, first, 2)

	second, err := repo.ClaimPending(context.Background(), 10)
	require.NoError(t, err)
	assert.Len(t, second, 3, "the rest, none of the first two again")

	assert.Empty(t, intersect(claimedIDs(first), claimedIDs(second)))
}

func TestIntegration_ConcurrentWorkersNeverClaimTheSameAccount(t *testing.T) {
	e := setup(t)
	repo := repository.NewRepository(e.db)

	const accounts = 20

	for id := int64(601); id < 601+accounts; id++ {
		e.createUser(t, id, fmt.Sprintf("del%d@test.io", id), "pw")
		e.markPending(t, id, 0, 0)
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results [][]int64
	)

	start := make(chan struct{})

	for w := 0; w < 5; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			<-start

			users, err := repo.ClaimPending(context.Background(), 7)
			if err != nil {
				t.Errorf("claim: %v", err)

				return
			}

			mu.Lock()
			results = append(results, claimedIDs(users))
			mu.Unlock()
		}()
	}

	close(start)
	wg.Wait()

	seen := map[int64]int{}
	for _, ids := range results {
		for _, id := range ids {
			seen[id]++
		}
	}

	for id, n := range seen {
		assert.Equal(t, 1, n, "account %d was claimed by %d workers", id, n)
	}

	assert.Len(t, seen, accounts, "every pending account is claimed exactly once across the workers")
}

func TestIntegration_FailedDeletionIsRetriedByTheWorkerUntilComplete(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	e.createUser(t, 601, "del601@test.io", "pw-601")
	e.db.Exec(`INSERT INTO oauth_tokens (client_id, user_id, access, refresh) VALUES ('c', '601', 'tokA', 'refA')`)

	var status, calls atomic.Int32
	status.Store(http.StatusBadGateway)

	srv := e.profileServer(&status, &calls)
	defer srv.Close()

	svc := e.service(srv.URL)

	require.ErrorIs(t, svc.Delete(ctx, 601, "pw-601"), ErrPending)

	var attempts int
	var lastError string

	require.NoError(t, e.db.Raw(`SELECT deletion_attempts, COALESCE(deletion_last_error, '') FROM users WHERE id = 601`).Row().Scan(&attempts, &lastError))
	assert.Equal(t, 1, attempts)
	assert.Contains(t, lastError, "delete profile data")
	assert.Contains(t, lastError, "502")

	// The backoff has not passed, so the worker leaves it alone.
	report, err := svc.RetryPending(ctx)
	require.NoError(t, err)
	assert.Zero(t, report.Attempted)

	// Time passes, the worker retries, the profile service is still down.
	e.db.Exec(`UPDATE users SET deletion_last_attempt_at = NOW() - INTERVAL '10 minutes' WHERE id = 601`)

	report, err = svc.RetryPending(ctx)
	require.NoError(t, err)
	assert.Equal(t, Report{Attempted: 1, Failed: 1}, report)
	assert.EqualValues(t, 1, e.count(t, `SELECT count(*) FROM users WHERE id = 601`), "still blocked, still retryable")
	assert.EqualValues(t, 2, e.count(t, `SELECT deletion_attempts FROM users WHERE id = 601`))

	// The profile service recovers and the next pass completes the deletion.
	status.Store(http.StatusOK)
	e.db.Exec(`UPDATE users SET deletion_last_attempt_at = NOW() - INTERVAL '10 minutes' WHERE id = 601`)

	report, err = svc.RetryPending(ctx)
	require.NoError(t, err)
	assert.Equal(t, Report{Attempted: 1, Completed: 1}, report)
	assert.Zero(t, e.count(t, `SELECT count(*) FROM users WHERE id = 601`), "deletion completed by the worker")
	assert.Zero(t, e.count(t, `SELECT count(*) FROM oauth_tokens WHERE user_id = '601'`))

	report, err = svc.RetryPending(ctx)
	require.NoError(t, err)
	assert.Zero(t, report.Attempted, "nothing left to retry")
}
