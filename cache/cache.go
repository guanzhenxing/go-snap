// Package cache 提供统一的缓存接口和多种实现
package cache

import (
	"context"
	"time"
)

// Item 表示缓存项
type Item struct {
	Value      interface{}
	Expiration time.Duration
	Tags       []string // 缓存标签，用于批量操作
}

// Cache 定义缓存的通用接口
type Cache interface {
	// Get 获取缓存，如果未找到返回 nil, false
	Get(ctx context.Context, key string) (interface{}, bool)

	// GetWithTTL 获取缓存和剩余TTL
	GetWithTTL(ctx context.Context, key string) (interface{}, time.Duration, bool)

	// Set 设置缓存
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// SetItem 设置带选项的缓存项
	SetItem(ctx context.Context, key string, item *Item) error

	// Delete 删除缓存
	Delete(ctx context.Context, key string) error

	// DeleteByPattern 通过模式删除缓存
	DeleteByPattern(ctx context.Context, pattern string) error

	// DeleteByTag 删除带特定标签的所有缓存
	DeleteByTag(ctx context.Context, tag string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Increment 增加数值
	Increment(ctx context.Context, key string, value int64) (int64, error)

	// Decrement 减少数值
	Decrement(ctx context.Context, key string, value int64) (int64, error)

	// Flush 清空所有缓存
	Flush(ctx context.Context) error

	// Close 关闭缓存连接
	Close() error
}

// Options 缓存选项
type Options struct {
	// 默认TTL，如果为0则永不过期
	DefaultTTL time.Duration

	// 清理间隔，用于内存缓存过期项清理
	CleanupInterval time.Duration
}

// DefaultOptions 返回默认缓存选项
func DefaultOptions() Options {
	return Options{
		DefaultTTL:      time.Hour,
		CleanupInterval: 10 * time.Minute,
	}
}
