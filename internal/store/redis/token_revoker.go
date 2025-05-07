package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/distributedJob/internal/store"
	"github.com/go-redis/redis/v8"
)

// RedisTokenRevoker implements TokenRevoker interface using Redis
type RedisTokenRevoker struct {
	client *redis.Client
	ctx    context.Context
	prefix string
}

// NewRedisTokenRevoker creates a new Redis token revoker
func NewRedisTokenRevoker(client *redis.Client, keyPrefix string) store.TokenRevoker {
	return &RedisTokenRevoker{
		client: client,
		ctx:    context.Background(),
		prefix: keyPrefix,
	}
}

// RevokeToken adds a token to the revocation list with a specified TTL
func (r *RedisTokenRevoker) RevokeToken(jti string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", r.prefix, jti)
	return r.client.Set(r.ctx, key, "1", ttl).Err()
}

// IsRevoked checks if a token is in the revocation list
func (r *RedisTokenRevoker) IsRevoked(jti string) bool {
	key := fmt.Sprintf("%s:%s", r.prefix, jti)
	exists, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		// Log error if needed
		return false
	}
	return exists > 0
}
