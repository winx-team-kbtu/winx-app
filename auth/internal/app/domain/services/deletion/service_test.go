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

	assert.Equal(t, []string{"block", "cache", "tokens.list", "cache", "tokens.delete", "profile", "account"}, r.calls,
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
		"block":         {"block", []string{"block"}},
		"cache":         {"cache", []string{"block", "cache"}},
		"tokens.list":   {"tokens.list", []string{"block", "cache", "tokens.list"}},
		"tokens.delete": {"tokens.delete", []string{"block", "cache", "tokens.list", "tokens.delete"}},
		"profile":       {"profile", []string{"block", "cache", "tokens.list", "tokens.delete", "profile"}},
		"account":       {"account", []string{"block", "cache", "tokens.list", "tokens.delete", "profile", "account"}},
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
