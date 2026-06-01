package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RateLimiter 定义了限流器的接口
type RateLimiter interface {
	// Allow 检查指定的 key 是否允许通过请求
	// capacity 是桶的最大容量，rate 是每秒补充的令牌数
	Allow(ctx context.Context, key string, capacity int, rate float64) (bool, error)
}

type redisRateLimiter struct {
	client *redis.Client
	script *redis.Script
}

// 令牌桶算法 Lua 脚本
// 接受参数：
// KEYS[1]: Redis 中的限流 Key
// ARGV[1]: 桶的容量 (capacity)
// ARGV[2]: 每秒补充令牌数 (rate)
// ARGV[3]: 本次请求消耗的令牌数 (默认 1)
const tokenBucketScript = `
local rate_limit_key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local requested = tonumber(ARGV[3])

-- 获取 Redis 服务器当前时间，保证不同客户端时间一致性
local now = redis.call('TIME')
-- now[1] 是秒，now[2] 是微秒
local current_time = tonumber(now[1]) + tonumber(now[2]) / 1000000

local bucket = redis.call('HMGET', rate_limit_key, 'tokens', 'last_refreshed')
local tokens = tonumber(bucket[1])
local last_refreshed = tonumber(bucket[2])

-- 第一次访问初始化
if tokens == nil then
	tokens = capacity
	last_refreshed = current_time
else
	-- 根据时间差补充令牌
	local delta = math.max(0, current_time - last_refreshed)
	tokens = math.min(capacity, tokens + delta * rate)
end

local allowed = 0
if tokens >= requested then
	tokens = tokens - requested
	allowed = 1
end

-- 写回 Redis
redis.call('HMSET', rate_limit_key, 'tokens', tokens, 'last_refreshed', current_time)
-- 设置过期时间，避免废弃的数据一直占用内存
local expire_time = math.ceil(capacity / rate) + 1
redis.call('EXPIRE', rate_limit_key, expire_time)

return allowed
`

// NewRedisRateLimiter 创建一个新的 Redis 限流器
func NewRedisRateLimiter(client *redis.Client) RateLimiter {
	return &redisRateLimiter{
		client: client,
		// 使用 Script 加载，底层会自动使用 EVALSHA，提高性能
		script: redis.NewScript(tokenBucketScript),
	}
}

// Allow 检查指定的 key 是否允许通过请求
func (l *redisRateLimiter) Allow(ctx context.Context, key string, capacity int, rate float64) (bool, error) {
	// 执行 Lua 脚本，默认每次消耗 1 个令牌
	result, err := l.script.Run(ctx, l.client, []string{key}, capacity, rate, 1).Result()
	if err != nil {
		return false, fmt.Errorf("failed to run rate limiter script: %w", err)
	}

	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected script result type: %T", result)
	}

	return allowed == 1, nil
}
