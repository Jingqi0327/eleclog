package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

// NewRedisCache 创建一个基于 go-redis 的缓存实例
func NewRedisCache(client *redis.Client) Cache {
	return &redisCache{
		client: client,
	}
}

// Set 序列化数据并存入 Redis
func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache to redis: %w", err)
	}

	return nil
}

// Get 从 Redis 获取数据并反序列化到 dest
func (c *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 如果是键不存在，可以选择返回自定义错误或者原始错误，方便调用层判断
			return err
		}
		return fmt.Errorf("failed to get cache from redis: %w", err)
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return nil
}

// Delete 从 Redis 删除指定键
func (c *redisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache from redis: %w", err)
	}
	return nil
}
