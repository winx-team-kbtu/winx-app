package deletion

import (
	"context"
	"errors"
	"testing"

	repository "auth/internal/app/domain/repositories/deletion"
	"auth/internal/app/models/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type recorder struct {
	calls []string

	user      models.User
	getErr    error
	failStep  map[string]error
	tokens    []string
	cacheKeys [][]string

	claimed  []models.User
	claimErr error
	failures []string
	attempts int
}

func (r *recorder) fail(step string) error { return r.failStep[step] }

func (r *recorder) GetByID(context.Context, int64) (models.User, error) {
	return r.user, r.getErr
}

func (r *recorder) MarkDeletionRequested(context.Context, int64) error {
	r.calls = append(r.calls, "block")

	return r.fail("block")
}

func (r *recorder) DeleteAccount(context.Context, models.User) error {
	r.calls = append(r.calls, "account")

	return r.fail("account")
}

func (r *recorder) TouchAttempt(context.Context, int64) error {
	r.calls = append(r.calls, "touch")
	r.attempts++

	return r.fail("touch")
}

func (r *recorder) RecordFailure(_ context.Context, _ int64, message string) error {
	r.failures = append(r.failures, message)

	return nil
}

func (r *recorder) ClaimPending(context.Context, int) ([]models.User, error) {
	return r.claimed, r.claimErr
}

func (r *recorder) AccessTokens(context.Context, int64) ([]string, error) {
	r.calls = append(r.calls, "tokens.list")

	return r.tokens, r.fail("tokens.list")
}

func (r *recorder) DeleteTokens(context.Context, int64) error {
	r.calls = append(r.calls, "tokens.delete")

	return r.fail("tokens.delete")
}

func (r *recorder) Delete(_ context.Context, keys ...string) error {
	r.calls = append(r.calls, "cache")
	r.cacheKeys = append(r.cacheKeys, keys)

	return r.fail("cache")
}

func (r *recorder) DeleteUserData(context.Context, int64) error {
	r.calls = append(r.calls, "profile")

	return r.fail("profile")
}

func newRecorder(t *testing.T) *recorder {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte("correct-horse"), bcrypt.MinCost)
	require.NoError(t, err)

	return &recorder{user: models.User{ID: 42, Email: "u@test.io", Password: string(hash)}, failStep: map[string]error{}}
}

func (r *recorder) service() Service { return NewService(r, r, r, r) }

func TestDelete_RunsEveryStepInOrder(t *testing.T) {
	r := newRecorder(t)
	r.tokens = []string{"tok-a", "tok-b"}

	require.NoError(t, r.service().Delete(context.Background(), 42, "correct-horse"))

	assert.Equal(t, []string{"touch", "block", "cache", "tokens.list", "cache", "tokens.delete", "profile", "account"}, r.calls,
		"block first so sign-in stops, user row last")
}

func TestDelete_RevokesUserAndTokenSessionKeys(t *testing.T) {
	r := newRecorder(t)
	r.tokens = []string{"tok-a", "tok-b"}

	require.NoError(t, r.service().Delete(context.Background(), 42, "correct-horse"))

	require.Len(t, r.cacheKeys, 2)
	assert.ElementsMatch(t, []string{"user_id:42", "user_email:42"}, r.cacheKeys[0],
		"user-level keys cut access in every service even if tokens are unknown")
	assert.ElementsMatch(t, []string{"access_token:tok-a", "access_token:tok-b"}, r.cacheKeys[1])
}

func TestDelete_NoTokensSkipsTokenKeyDeletion(t *testing.T) {
	r := newRecorder(t)

	require.NoError(t, r.service().Delete(context.Background(), 42, "correct-horse"))

	assert.Len(t, r.cacheKeys, 1, "only the user-level keys")
}

func TestDelete_WrongPasswordChangesNothing(t *testing.T) {
	r := newRecorder(t)

	err := r.service().Delete(context.Background(), 42, "wrong")

	require.ErrorIs(t, err, ErrInvalidPassword)
	assert.Empty(t, r.calls, "a wrong confirmation must not block the account or touch any store")
}

func TestDelete_UnknownUser(t *testing.T) {
	r := newRecorder(t)
	r.getErr = repository.ErrNotFound

	require.ErrorIs(t, r.service().Delete(context.Background(), 42, "x"), ErrNotFound)
	assert.Empty(t, r.calls)
}

func TestDelete_FailureLeavesAccountPendingAndStopsLaterSteps(t *testing.T) {
	boom := errors.New("boom")

	cases := map[string]struct {
		failAt    string
		wantCalls []string
	}{
		"block":         {"block", []string{"touch", "block"}},
		"cache":         {"cache", []string{"touch", "block", "cache"}},
		"tokens.list":   {"tokens.list", []string{"touch", "block", "cache", "tokens.list"}},
		"tokens.delete": {"tokens.delete", []string{"touch", "block", "cache", "tokens.list", "tokens.delete"}},
		"profile":       {"profile", []string{"touch", "block", "cache", "tokens.list", "tokens.delete", "profile"}},
		"account":       {"account", []string{"touch", "block", "cache", "tokens.list", "tokens.delete", "profile", "account"}},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := newRecorder(t)
			r.failStep[tc.failAt] = boom

			err := r.service().Delete(context.Background(), 42, "correct-horse")

			require.ErrorIs(t, err, ErrPending, "a failure must never look like success")
			require.ErrorIs(t, err, boom, "the cause stays available for logs")
			assert.Equal(t, tc.wantCalls, r.calls)
		})
	}
}

func TestDelete_RepeatingAfterFailureCompletes(t *testing.T) {
	r := newRecorder(t)
	r.failStep["profile"] = errors.New("profile service down")
	svc := r.service()

	require.ErrorIs(t, svc.Delete(context.Background(), 42, "correct-horse"), ErrPending)
	assert.NotContains(t, r.calls, "account", "the user row must survive a failed cleanup")

	delete(r.failStep, "profile")
	r.calls = nil

	require.NoError(t, svc.Delete(context.Background(), 42, "correct-horse"))
	assert.Equal(t, "account", r.calls[len(r.calls)-1])
}

func TestDelete_RecordsWhyTheAttemptFailed(t *testing.T) {
	r := newRecorder(t)
	r.failStep["profile"] = errors.New("profile cleanup returned 500")

	require.ErrorIs(t, r.service().Delete(context.Background(), 42, "correct-horse"), ErrPending)

	require.Len(t, r.failures, 1)
	assert.Contains(t, r.failures[0], "delete profile data")
	assert.Contains(t, r.failures[0], "profile cleanup returned 500")
}

func TestDelete_CountsTheAttemptBeforeRunning(t *testing.T) {
	r := newRecorder(t)
	r.failStep["touch"] = errors.New("db down")

	err := r.service().Delete(context.Background(), 42, "correct-horse")

	require.ErrorIs(t, err, ErrPending)
	assert.Equal(t, []string{"touch"}, r.calls, "nothing runs if the attempt cannot even be recorded")
}

func TestRetryPending_CompletesClaimedAccountsAndReportsFailures(t *testing.T) {
	r := newRecorder(t)
	r.claimed = []models.User{{ID: 1, Email: "a@test.io"}, {ID: 2, Email: "b@test.io"}}

	report, err := r.service().RetryPending(context.Background())

	require.NoError(t, err)
	assert.Equal(t, Report{Attempted: 2, Completed: 2}, report)
	assert.Equal(t, 2, countOf(r.calls, "account"), "both user rows are removed")
	assert.Zero(t, r.attempts, "claiming already counted the attempts, retries must not count twice")
}

func TestRetryPending_FailedAccountStaysPendingAndOthersContinue(t *testing.T) {
	r := newRecorder(t)
	r.claimed = []models.User{{ID: 1}, {ID: 2}}

	failFirst := &failOnce{recorder: r}
	svc := NewService(r, r, r, failFirst)

	report, err := svc.RetryPending(context.Background())

	require.NoError(t, err)
	assert.Equal(t, Report{Attempted: 2, Completed: 1, Failed: 1}, report, "one failure must not stop the batch")
	assert.Len(t, r.failures, 1, "the reason is recorded for the failed account")
	assert.Equal(t, 1, countOf(r.calls, "account"), "only the successful account row is removed")
}

func TestRetryPending_NothingDueIsANoOp(t *testing.T) {
	r := newRecorder(t)

	report, err := r.service().RetryPending(context.Background())

	require.NoError(t, err)
	assert.Equal(t, Report{}, report)
	assert.Empty(t, r.calls)
}

func TestRetryPending_ClaimErrorIsReturned(t *testing.T) {
	r := newRecorder(t)
	r.claimErr = errors.New("db down")

	_, err := r.service().RetryPending(context.Background())

	require.ErrorIs(t, err, r.claimErr)
}

func countOf(calls []string, want string) int {
	n := 0

	for _, c := range calls {
		if c == want {
			n++
		}
	}

	return n
}

// failOnce makes the profile step fail for the first account only.
type failOnce struct {
	*recorder
	used bool
}

func (f *failOnce) DeleteUserData(ctx context.Context, id int64) error {
	if !f.used {
		f.used = true

		return errors.New("profile service down")
	}

	return f.recorder.DeleteUserData(ctx, id)
}
