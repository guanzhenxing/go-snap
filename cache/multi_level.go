package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// CacheLevel 表示缓存层级
type CacheLevel int

const (
	// CacheLevelLocal 本地缓存层级
	CacheLevelLocal CacheLevel = iota
	// CacheLevelRedis Redis缓存层级
	CacheLevelRedis
	// CacheLevelAll 所有缓存层级
	CacheLevelAll
)

// WriteMode 表示多级缓存的写入模式
type WriteMode int

const (
	// WriteModeWriteThrough 直写模式：同时写入所有缓存层级
	WriteModeWriteThrough WriteMode = iota
	// WriteModeWriteBack 回写模式：只写入最高层级，其他层级按需加载
	WriteModeWriteBack
)

// MultiLevelCache 实现多级缓存
type MultiLevelCache struct {
	local     Cache // 本地缓存，通常为内存缓存
	remote    Cache // 远程缓存，通常为Redis缓存
	writeMode WriteMode
	localTTL  time.Duration // 本地缓存的TTL，通常比远程缓存的TTL短
}

// MultiLevelOptions 多级缓存的选项
type MultiLevelOptions struct {
	WriteMode WriteMode
	LocalTTL  time.Duration
}

// DefaultMultiLevelOptions 返回默认的多级缓存选项
func DefaultMultiLevelOptions() MultiLevelOptions {
	return MultiLevelOptions{
		WriteMode: WriteModeWriteThrough,
		LocalTTL:  time.Minute * 5,
	}
}

// NewMultiLevelCache 创建新的多级缓存
func NewMultiLevelCache(local, remote Cache, opts ...MultiLevelOptions) (*MultiLevelCache, error) {
	if local == nil || remote == nil {
		return nil, errors.New("both local and remote cache must be provided")
	}

	var options MultiLevelOptions
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = DefaultMultiLevelOptions()
	}

	return &MultiLevelCache{
		local:     local,
		remote:    remote,
		writeMode: options.WriteMode,
		localTTL:  options.LocalTTL,
	}, nil
}

// Get 从缓存中获取值，优先从本地缓存获取，如果本地缓存没有，再从远程缓存获取
func (c *MultiLevelCache) Get(ctx context.Context, key string) (interface{}, bool) {
	// 先尝试从本地缓存获取
	value, found := c.local.Get(ctx, key)
	if found {
		return value, true
	}

	// 如果本地缓存没有，尝试从远程缓存获取
	value, found = c.remote.Get(ctx, key)
	if !found {
		return nil, false
	}

	// 如果远程缓存有，将值写入本地缓存（使用较短的TTL）
	_ = c.local.Set(ctx, key, value, c.localTTL)

	return value, true
}

// GetWithTTL 获取值和剩余TTL，优先从本地缓存获取
func (c *MultiLevelCache) GetWithTTL(ctx context.Context, key string) (interface{}, time.Duration, bool) {
	// 先尝试从本地缓存获取
	value, ttl, found := c.local.GetWithTTL(ctx, key)
	if found {
		return value, ttl, true
	}

	// 如果本地缓存没有，尝试从远程缓存获取
	value, ttl, found = c.remote.GetWithTTL(ctx, key)
	if !found {
		return nil, 0, false
	}

	// 如果远程缓存有，将值写入本地缓存（使用较短的TTL）
	localTTL := c.localTTL
	if ttl > 0 && ttl < localTTL {
		localTTL = ttl
	}
	_ = c.local.Set(ctx, key, value, localTTL)

	return value, ttl, true
}

// Set 设置缓存
func (c *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	var err error

	// 根据写入模式决定如何写入缓存
	if c.writeMode == WriteModeWriteThrough {
		// 直写模式：同时写入本地和远程缓存
		localTTL := ttl
		if ttl == 0 || ttl > c.localTTL {
			localTTL = c.localTTL
		}

		_ = c.local.Set(ctx, key, value, localTTL)
		err2 := c.remote.Set(ctx, key, value, ttl)

		// 只要远程缓存写入成功就认为成功
		if err2 != nil {
			return err2
		}
		return nil
	} else {
		// 回写模式：只写入远程缓存
		err = c.remote.Set(ctx, key, value, ttl)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetItem 设置带完整选项的缓存项
func (c *MultiLevelCache) SetItem(ctx context.Context, key string, item *Item) error {
	if item == nil {
		return errors.New("cache item cannot be nil")
	}

	var err error

	// 根据写入模式决定如何写入缓存
	if c.writeMode == WriteModeWriteThrough {
		// 直写模式：同时写入本地和远程缓存
		localItem := &Item{
			Value:      item.Value,
			Expiration: item.Expiration,
			Tags:       item.Tags,
		}

		// 为本地缓存设置较短的TTL
		if localItem.Expiration == 0 || localItem.Expiration > c.localTTL {
			localItem.Expiration = c.localTTL
		}

		_ = c.local.SetItem(ctx, key, localItem)
		err2 := c.remote.SetItem(ctx, key, item)

		// 只要远程缓存写入成功就认为成功
		if err2 != nil {
			return err2
		}
		return nil
	} else {
		// 回写模式：只写入远程缓存
		err = c.remote.SetItem(ctx, key, item)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete 从所有缓存层级删除值
func (c *MultiLevelCache) Delete(ctx context.Context, key string) error {
	// 始终从所有层级删除
	_ = c.local.Delete(ctx, key)
	return c.remote.Delete(ctx, key)
}

// DeleteByPattern 根据模式从所有缓存层级删除值
func (c *MultiLevelCache) DeleteByPattern(ctx context.Context, pattern string) error {
	// 始终从所有层级删除
	_ = c.local.DeleteByPattern(ctx, pattern)
	return c.remote.DeleteByPattern(ctx, pattern)
}

// DeleteByTag 删除带特定标签的所有缓存
func (c *MultiLevelCache) DeleteByTag(ctx context.Context, tag string) error {
	// 始终从所有层级删除
	_ = c.local.DeleteByTag(ctx, tag)
	return c.remote.DeleteByTag(ctx, tag)
}

// Exists 检查键是否存在于任意缓存层级
func (c *MultiLevelCache) Exists(ctx context.Context, key string) (bool, error) {
	// 先检查本地缓存
	exists, err := c.local.Exists(ctx, key)
	if err == nil && exists {
		return true, nil
	}

	// 再检查远程缓存
	return c.remote.Exists(ctx, key)
}

// Increment 增加数值，操作会传递到所有缓存层级
func (c *MultiLevelCache) Increment(ctx context.Context, key string, value int64) (int64, error) {
	// 在远程缓存中执行操作
	result, err := c.remote.Increment(ctx, key, value)
	if err != nil {
		return 0, err
	}

	// 如果是直写模式，同步更新本地缓存
	if c.writeMode == WriteModeWriteThrough {
		_ = c.local.Set(ctx, key, result, c.localTTL)
	} else {
		// 如果是回写模式，删除本地缓存中的旧值
		_ = c.local.Delete(ctx, key)
	}

	return result, nil
}

// Decrement 减少数值，操作会传递到所有缓存层级
func (c *MultiLevelCache) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return c.Increment(ctx, key, -value)
}

// Flush 清空所有缓存层级
func (c *MultiLevelCache) Flush(ctx context.Context) error {
	// 清空本地缓存
	_ = c.local.Flush(ctx)

	// 清空远程缓存
	return c.remote.Flush(ctx)
}

// Close 关闭所有缓存连接
func (c *MultiLevelCache) Close() error {
	// 关闭本地缓存
	_ = c.local.Close()

	// 关闭远程缓存
	return c.remote.Close()
}

// FlushLevel 清空指定缓存层级
func (c *MultiLevelCache) FlushLevel(ctx context.Context, level CacheLevel) error {
	var err error

	switch level {
	case CacheLevelLocal:
		err = c.local.Flush(ctx)
	case CacheLevelRedis:
		err = c.remote.Flush(ctx)
	case CacheLevelAll:
		_ = c.local.Flush(ctx)
		err = c.remote.Flush(ctx)
	default:
		err = fmt.Errorf("unknown cache level: %d", level)
	}

	return err
}
