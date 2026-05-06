package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	rdb    *redis.Client
	prefix string
}

func NewRedisCache(rdb *redis.Client, prefix string) *RedisCache {
	p := strings.Trim(prefix, ":")
	if p != "" {
		p += ":"
	}
	return &RedisCache{rdb: rdb, prefix: p}
}

func (c *RedisCache) k(key string) string {
	if c.prefix == "" {
		return key
	}
	return c.prefix + key
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := c.rdb.Get(ctx, c.k(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	return b, err
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.rdb.Set(ctx, c.k(key), value, ttl).Err()
}

func (c *RedisCache) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, c.k(key), value, ttl).Result()
}

func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	ks := make([]string, 0, len(keys))
	for _, k := range keys {
		ks = append(ks, c.k(k))
	}
	return c.rdb.Del(ctx, ks...).Err()
}

func (c *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	ks := make([]string, 0, len(keys))
	for _, k := range keys {
		ks = append(ks, c.k(k))
	}
	return c.rdb.Exists(ctx, ks...).Result()
}

func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, c.k(key)).Result()
}

func (c *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return c.rdb.IncrBy(ctx, c.k(key), delta).Result()
}

func (c *RedisCache) MGet(ctx context.Context, keys ...string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return map[string][]byte{}, nil
	}
	ks := make([]string, len(keys))
	for i, k := range keys {
		ks[i] = c.k(k)
	}
	vals, err := c.rdb.MGet(ctx, ks...).Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string][]byte, len(keys))
	for i, v := range vals {
		if v != nil {
			result[keys[i]] = []byte(v.(string))
		}
	}
	return result, nil
}
