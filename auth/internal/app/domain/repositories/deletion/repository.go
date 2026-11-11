package deletion

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"auth/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("user not found")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByID(ctx context.Context, id int64) (models.User, error) {
	var user models.User

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, ErrNotFound
		}

		return models.User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

// MarkDeletionRequested is idempotent: the first request time is kept.
func (r *Repository) MarkDeletionRequested(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Exec(
		`UPDATE users SET deletion_requested_at = COALESCE(deletion_requested_at, NOW()) WHERE id = ?`, id,
	).Error; err != nil {
		return fmt.Errorf("mark deletion requested: %w", err)
	}

	return nil
}

func (r *Repository) AccessTokens(ctx context.Context, userID int64) ([]string, error) {
	var tokens []string

	if err := r.db.WithContext(ctx).Raw(
		`SELECT access FROM oauth_tokens WHERE user_id = ? AND access IS NOT NULL`, strconv.FormatInt(userID, 10),
	).Scan(&tokens).Error; err != nil {
		return nil, fmt.Errorf("list access tokens: %w", err)
	}

	return tokens, nil
}

func (r *Repository) DeleteTokens(ctx context.Context, userID int64) error {
	if err := r.db.WithContext(ctx).Exec(
		`DELETE FROM oauth_tokens WHERE user_id = ?`, strconv.FormatInt(userID, 10),
	).Error; err != nil {
		return fmt.Errorf("delete tokens: %w", err)
	}

	return nil
}

// cleanup is one statement removing the user's rows from a table. Tables that belong to the
// other services share this database; the statement is skipped when a deployment does not
// have the table, so a missing optional service cannot block a deletion forever.
type cleanup struct {
	table string
	sql   string
	args  func(user models.User) []any
}

var userKey = func(user models.User) []any { return []any{user.ID} }

var cleanups = []cleanup{
	{"oauth_tokens", `DELETE FROM oauth_tokens WHERE user_id = ?`, func(u models.User) []any { return []any{strconv.FormatInt(u.ID, 10)} }},
	{"password_resets", `DELETE FROM password_resets WHERE email = ?`, func(u models.User) []any { return []any{u.Email} }},
	{"notifications", `DELETE FROM notifications WHERE recipient = ?`, func(u models.User) []any { return []any{u.Email} }},
	// Legacy Social data. It must not exist for pilot users, but it is removed if it does.
	{"messages", `DELETE FROM messages WHERE sender_id = ?`, userKey},
	{"chats", `DELETE FROM chats WHERE user_one_id = ? OR user_two_id = ?`, func(u models.User) []any { return []any{u.ID, u.ID} }},
	{"matches", `DELETE FROM matches WHERE user_one_id = ? OR user_two_id = ?`, func(u models.User) []any { return []any{u.ID, u.ID} }},
	{"swipes", `DELETE FROM swipes WHERE swiper_id = ? OR swiped_id = ?`, func(u models.User) []any { return []any{u.ID, u.ID} }},
	{"recommendations", `DELETE FROM recommendations WHERE user_id = ? OR recommended_id = ?`, func(u models.User) []any { return []any{u.ID, u.ID} }},
}

// DeleteAccount removes the auth-owned and shared-database rows of the user and finally the
// user row itself, all in one transaction, so the account either disappears completely or not
// at all. Rows still referencing the user through ON DELETE CASCADE are removed with it.
func (r *Repository) DeleteAccount(ctx context.Context, user models.User) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, c := range cleanups {
			var exists bool
			if err := tx.Raw(`SELECT to_regclass(?) IS NOT NULL`, c.table).Scan(&exists).Error; err != nil {
				return fmt.Errorf("check table %s: %w", c.table, err)
			}

			if !exists {
				continue
			}

			if err := tx.Exec(c.sql, c.args(user)...).Error; err != nil {
				return fmt.Errorf("clean %s: %w", c.table, err)
			}
		}

		if err := tx.Exec(`DELETE FROM users WHERE id = ?`, user.ID).Error; err != nil {
			return fmt.Errorf("delete user: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}

	return nil
}
