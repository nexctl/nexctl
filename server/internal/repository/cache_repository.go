package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// NodeSessionCache defines cached online-session operations.
type NodeSessionCache interface {
	MarkOnline(ctx context.Context, nodeID int64, ttl time.Duration) error
}

// RedisNodeSessionCache is the Redis online-session cache implementation.
type RedisNodeSessionCache struct {
	rdb *redis.Client
}

// NewNodeSessionCache creates a Redis node session cache.
func NewNodeSessionCache(rdb *redis.Client) *RedisNodeSessionCache {
	return &RedisNodeSessionCache{rdb: rdb}
}

// MarkOnline records the node online marker in Redis.
func (c *RedisNodeSessionCache) MarkOnline(ctx context.Context, nodeID int64, ttl time.Duration) error {
	key := fmt.Sprintf("node:%d:online", nodeID)
	return c.rdb.Set(ctx, key, "1", ttl).Err()
}
