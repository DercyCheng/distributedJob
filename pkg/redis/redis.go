package redis

import (
	"context"
	"fmt"
	"go-job/pkg/config"
	"time"

	"github.com/go-redis/redis/v8"
)

var client *redis.Client

// Init 初始化 Redis 连接
func Init(cfg *config.Config) error {
	client = redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}

	return nil
}

// GetClient 获取 Redis 客户端
func GetClient() *redis.Client {
	return client
}

// Close 关闭 Redis 连接
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// Set 设置键值对
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func Get(ctx context.Context, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

// Del 删除键
func Del(ctx context.Context, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(ctx context.Context, keys ...string) (int64, error) {
	return client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return client.Expire(ctx, key, expiration).Err()
}

// HSet 设置哈希字段
func HSet(ctx context.Context, key string, values ...interface{}) error {
	return client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段
func HGet(ctx context.Context, key, field string) (string, error) {
	return client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func HDel(ctx context.Context, key string, fields ...string) error {
	return client.HDel(ctx, key, fields...).Err()
}

// LPush 从左边推入列表
func LPush(ctx context.Context, key string, values ...interface{}) error {
	return client.LPush(ctx, key, values...).Err()
}

// RPop 从右边弹出列表
func RPop(ctx context.Context, key string) (string, error) {
	return client.RPop(ctx, key).Result()
}

// BRPop 阻塞式从右边弹出列表
func BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return client.BRPop(ctx, timeout, keys...).Result()
}

// LLen 获取列表长度
func LLen(ctx context.Context, key string) (int64, error) {
	return client.LLen(ctx, key).Result()
}

// SAdd 添加到集合
func SAdd(ctx context.Context, key string, members ...interface{}) error {
	return client.SAdd(ctx, key, members...).Err()
}

// SRem 从集合移除
func SRem(ctx context.Context, key string, members ...interface{}) error {
	return client.SRem(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func SMembers(ctx context.Context, key string) ([]string, error) {
	return client.SMembers(ctx, key).Result()
}

// ZAdd 添加到有序集合
func ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	return client.ZAdd(ctx, key, members...).Err()
}

// ZRangeByScore 按分数范围获取有序集合成员
func ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return client.ZRangeByScore(ctx, key, opt).Result()
}

// ZRem 从有序集合移除成员
func ZRem(ctx context.Context, key string, members ...interface{}) error {
	return client.ZRem(ctx, key, members...).Err()
}

// IsConnected 检查 Redis 连接状态
func IsConnected() bool {
	if client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	return err == nil
}
