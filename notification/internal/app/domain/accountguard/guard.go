package accountguard

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Guard tells event consumers whether the account an event is about still exists and is not
// being deleted. Events can arrive long after they were published, so without this check a
// late event would recreate personal data of a user that has already been erased.
type Guard struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Guard {
	return &Guard{db: db}
}

// AccountActive checks the exact account the event was published for: same id and same
// email. A different account that later reused the email address does not count.
func (g *Guard) AccountActive(ctx context.Context, id int64, email string) (bool, error) {
	return g.exists(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = ? AND email = ? AND deletion_requested_at IS NULL)`, id, email)
}

// EmailActive is for events that only carry an email address.
func (g *Guard) EmailActive(ctx context.Context, email string) (bool, error) {
	return g.exists(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE email = ? AND deletion_requested_at IS NULL)`, email)
}

func (g *Guard) UserActive(ctx context.Context, id int64) (bool, error) {
	return g.exists(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = ? AND deletion_requested_at IS NULL)`, id)
}

func (g *Guard) exists(ctx context.Context, query string, args ...any) (bool, error) {
	var ok bool

	if err := g.db.WithContext(ctx).Raw(query, args...).Scan(&ok).Error; err != nil {
		return false, fmt.Errorf("check account state: %w", err)
	}

	return ok, nil
}
