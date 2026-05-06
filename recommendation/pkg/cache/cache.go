package cache

import (
	"context"
	"errors"
	"time"
)

var ErrCacheMiss = errors.New("cache: miss")

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	// Increment atomically adds delta to the integer stored at key and returns
	// the new value. The key is created with value 0 if it does not exist.
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	// MGet fetches multiple keys in a single round trip. Missing keys are
	// omitted from the returned map (not an error).
	MGet(ctx context.Context, keys ...string) (map[string][]byte, error)
}
