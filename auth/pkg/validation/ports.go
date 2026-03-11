package validation

import "context"

type UniquenessChecker interface {
	IsUnique(ctx context.Context, table, column string, value any, scope map[string]any, ignore map[string]any) (bool, error)
}
