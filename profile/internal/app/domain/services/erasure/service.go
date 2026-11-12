package erasure

import (
	"context"
	"fmt"
)

// Narrow ports, each satisfied by an existing repository or storage client.
type (
	Rows interface {
		PhotoObjectKeys(ctx context.Context, userID int64) ([]string, error)
		DeleteUserRows(ctx context.Context, userID int64) error
	}

	Objects interface {
		Delete(ctx context.Context, key string) error
	}

	Profiles interface {
		DeleteByUserID(ctx context.Context, userID int64) error
	}
)

type Service interface {
	// DeleteUserData removes everything the profile service stores for the user.
	DeleteUserData(ctx context.Context, userID int64) error
}

type service struct {
	rows     Rows
	objects  Objects
	profiles Profiles
}

func NewService(rows Rows, objects Objects, profiles Profiles) Service {
	return &service{rows: rows, objects: objects, profiles: profiles}
}

// DeleteUserData is idempotent and safe to retry after a partial failure. External
// stores (S3, MongoDB) are cleared first and the PostgreSQL rows last, because the rows
// hold the S3 object keys: if any step fails, the next attempt can still find them.
func (s *service) DeleteUserData(ctx context.Context, userID int64) error {
	keys, err := s.rows.PhotoObjectKeys(ctx, userID)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := s.objects.Delete(ctx, key); err != nil {
			return fmt.Errorf("delete photo object %q: %w", key, err)
		}
	}

	if err := s.profiles.DeleteByUserID(ctx, userID); err != nil {
		return err
	}

	return s.rows.DeleteUserRows(ctx, userID)
}
