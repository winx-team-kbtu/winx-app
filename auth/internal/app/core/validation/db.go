package validation

import "context"

type DBValidator interface {
	Unique(ctx context.Context, table string, scope, ignore map[string]any) (bool, error)
	Exists(ctx context.Context, table string, scope, ignore map[string]any) (bool, error)
}
