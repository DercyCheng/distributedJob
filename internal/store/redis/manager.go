package redis

import (
	"context"
	"time"

	"distributedJob/internal/config"
	"github.com/go-redis/redis/v8"
)

// Manager manages Redis connections and operations
type Manager struct {
	client *redis.Client
	ctx    context.Context
}

// NewManager creates a new Redis manager from configuration
func NewManager(cfg *config.Config) (*Manager, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.URL,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.MaxActive,
		MinIdleConns: cfg.Redis.MaxIdle,
		DialTimeout:  time.Duration(cfg.Redis.ConnectTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Redis.IdleTimeout) * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	ctx := context.Background()

	// Test the connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &Manager{
		client: client,
		ctx:    ctx,
	}, nil
}

// Client returns the Redis client
func (m *Manager) Client() *redis.Client {
	return m.client
}

// Close closes the Redis client
func (m *Manager) Close() error {
	return m.client.Close()
}

// CreateTokenRevoker creates a token revoker using this Redis manager
func (m *Manager) CreateTokenRevoker() *RedisTokenRevoker {
	return &RedisTokenRevoker{
		client: m.client,
		ctx:    m.ctx,
		prefix: "token:revoked:",
	}
}
