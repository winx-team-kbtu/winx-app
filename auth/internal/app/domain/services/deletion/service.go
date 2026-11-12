package deletion

import (
	"context"
	"errors"
	"fmt"

	repository "auth/internal/app/domain/repositories/deletion"
	"auth/internal/app/models/models"

	"golang.org/x/crypto/bcrypt"
)

// errNotFoundSentinel is what the user repository returns for a missing user.
var errNotFoundSentinel = repository.ErrNotFound

var (
	ErrNotFound        = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	// ErrPending means the deletion started but a step failed. The account stays blocked and
	// its cleanup remains pending; it is never reported as done.
	ErrPending = errors.New("account deletion is pending")
)

// Narrow ports, each satisfied by a repository or client wired in bootstrap.
type (
	Users interface {
		GetByID(ctx context.Context, id int64) (models.User, error)
		MarkDeletionRequested(ctx context.Context, id int64) error
		DeleteAccount(ctx context.Context, user models.User) error
	}

	Tokens interface {
		AccessTokens(ctx context.Context, userID int64) ([]string, error)
		DeleteTokens(ctx context.Context, userID int64) error
	}

	Cache interface {
		Delete(ctx context.Context, keys ...string) error
	}

	ProfileData interface {
		DeleteUserData(ctx context.Context, userID int64) error
	}
)

type Service interface {
	// Delete removes the account after the user confirmed with their password.
	Delete(ctx context.Context, userID int64, password string) error
}

type service struct {
	users   Users
	tokens  Tokens
	cache   Cache
	profile ProfileData
}

func NewService(users Users, tokens Tokens, cache Cache, profile ProfileData) Service {
	return &service{users: users, tokens: tokens, cache: cache, profile: profile}
}

func (s *service) Delete(ctx context.Context, userID int64, password string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errNotFoundSentinel) {
			return ErrNotFound
		}

		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return ErrInvalidPassword
	}

	return s.run(ctx, user)
}

// run executes the deletion steps in order. Every step is idempotent, so the whole sequence
// can be repeated after a failure. The account is blocked first, then sessions are revoked,
// then data is removed from the other stores, and the user row goes last.
func (s *service) run(ctx context.Context, user models.User) error {
	steps := []struct {
		name string
		do   func(context.Context) error
	}{
		{"block account", func(ctx context.Context) error { return s.users.MarkDeletionRequested(ctx, user.ID) }},
		{"revoke sessions", func(ctx context.Context) error { return s.revokeSessions(ctx, user.ID) }},
		{"delete profile data", func(ctx context.Context) error { return s.profile.DeleteUserData(ctx, user.ID) }},
		{"delete account", func(ctx context.Context) error { return s.users.DeleteAccount(ctx, user) }},
	}

	for _, step := range steps {
		if err := step.do(ctx); err != nil {
			return fmt.Errorf("%w: %s: %w", ErrPending, step.name, err)
		}
	}

	return nil
}

// revokeSessions invalidates every session of the user in all services. Services authenticate a
// request through the user-level cache keys, so removing those cuts access immediately, even for
// tokens that are not known any more. The per-token keys and the stored tokens go afterwards.
func (s *service) revokeSessions(ctx context.Context, userID int64) error {
	if err := s.cache.Delete(ctx, fmt.Sprintf("user_id:%d", userID), fmt.Sprintf("user_email:%d", userID)); err != nil {
		return fmt.Errorf("delete user session keys: %w", err)
	}

	tokens, err := s.tokens.AccessTokens(ctx, userID)
	if err != nil {
		return err
	}

	if len(tokens) > 0 {
		keys := make([]string, 0, len(tokens))
		for _, token := range tokens {
			keys = append(keys, "access_token:"+token)
		}

		if err := s.cache.Delete(ctx, keys...); err != nil {
			return fmt.Errorf("delete access token keys: %w", err)
		}
	}

	return s.tokens.DeleteTokens(ctx, userID)
}
