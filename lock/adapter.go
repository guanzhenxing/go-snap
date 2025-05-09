package lock

import (
	"context"
	"fmt"

	"github.com/guanzhenxing/go-snap/cache"
	"github.com/redis/go-redis/v9"
)

// FromRedisCache 从Redis缓存创建锁
func FromRedisCache(redisCache *cache.RedisCache, key string, opts ...Options) (Lock, error) {
	client, err := getRedisClientFromCache(redisCache)
	if err != nil {
		return nil, err
	}

	return NewRedisLock(client, key, opts...), nil
}

// WithCacheLock 使用缓存创建的锁执行操作
func WithCacheLock(ctx context.Context, c interface{}, key string, fn func() error, opts ...Options) error {
	l, err := FromCache(c, key, opts...)
	if err != nil {
		return err
	}

	return WithLock(ctx, l, fn)
}

// FromCache 从任意类型的缓存创建锁
func FromCache(c interface{}, key string, opts ...Options) (Lock, error) {
	switch cacheImpl := c.(type) {
	case *cache.RedisCache:
		return FromRedisCache(cacheImpl, key, opts...)
	default:
		return nil, fmt.Errorf("unsupported cache type for locking: %T", c)
	}
}

// getRedisClientFromCache 从Redis缓存获取Redis客户端
func getRedisClientFromCache(redisCache *cache.RedisCache) (redis.UniversalClient, error) {
	// 使用RedisCache.GetClient()方法获取Redis客户端
	return redisCache.GetClient(), nil
}
