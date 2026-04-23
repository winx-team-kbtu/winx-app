package geoip

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"winx-profile/pkg/cache"
)

const cacheTTL = 24 * time.Hour

type cachedClient struct {
	inner Client
	cache cache.Cache
}

// NewCachedClient wraps any Client with a Redis-backed cache.
// Each unique IP is resolved at most once per 24 hours.
func NewCachedClient(inner Client, c cache.Cache) Client {
	return &cachedClient{inner: inner, cache: c}
}

func (c *cachedClient) Lookup(ctx context.Context, ip string) (Result, error) {
	key := "lookup:" + ip

	if b, err := c.cache.Get(ctx, key); err == nil {
		var cached Result
		if json.Unmarshal(b, &cached) == nil {
			return cached, nil
		}
	}

	result, err := c.inner.Lookup(ctx, ip)
	if err != nil {
		if errors.Is(err, ErrLookupFailed) {
			return Result{}, err
		}
		return Result{}, err
	}

	if b, err := json.Marshal(result); err == nil {
		_ = c.cache.Set(ctx, key, b, cacheTTL)
	}

	return result, nil
}
