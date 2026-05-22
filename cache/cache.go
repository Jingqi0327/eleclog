package cache

import (
	"context"
	"time"
)

// Cache 定义了缓存的基本接口
type Cache interface {
	// Set 存入缓存数据
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	
	// Get 获取缓存数据，并将结果反序列化到 dest 指针指向的对象中
	Get(ctx context.Context, key string, dest interface{}) error
	
	// Delete 删除缓存数据
	Delete(ctx context.Context, key string) error
}
