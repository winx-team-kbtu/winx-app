package accountguard

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Guard tells event consumers whether the users an event is about still exist and are not being
// deleted, so a late event cannot recreate data of an erased account.
type Guard struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Guard {
	return &Guard{db: db}
}

// UsersActive reports whether every given user exists and has no pending deletion.
func (g *Guard) UsersActive(ctx context.Context, ids ...int64) (bool, error) {
	unique := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}

	if len(unique) == 0 {
		return false, nil
	}

	list := make([]int64, 0, len(unique))
	for id := range unique {
		list = append(list, id)
	}

	var active int64

	if err := g.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM users WHERE id IN ? AND deletion_requested_at IS NULL`, list,
	).Scan(&active).Error; err != nil {
		return false, fmt.Errorf("check users state: %w", err)
	}

	return active == int64(len(list)), nil
}
