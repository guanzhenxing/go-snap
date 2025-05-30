// Package cache 提供统一的缓存接口和多种实现
// 内存缓存实现提供了高性能的本地内存缓存，支持TTL过期、标签索引和模式匹配删除等功能
// 适用场景：
// - 单机应用中需要临时存储的数据
// - 需要高性能且不关心持久化的缓存
// - 作为分布式缓存的本地一级缓存
//
// 性能特性：
// - 读取操作使用读写锁的读锁，多个并发读取不会阻塞
// - 写入操作使用读写锁的写锁，会阻塞所有读写操作
// - 标签索引提供了O(1)时间复杂度的标签查找
// - 过期清理由后台goroutine定期执行，不影响正常操作
//
// 限制：
// - 所有数据存储在内存中，重启后数据会丢失
// - 不适合存储大量数据，可能导致内存压力
// - 不支持跨实例的数据共享
package cache

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"
)

// MemoryCache 实现基于内存的缓存
// 提供线程安全的内存键值存储，支持TTL、标签和模式匹配等功能
// 使用读写锁保护并发访问，适合读多写少的场景
type MemoryCache struct {
	// items 存储缓存项的映射，键为缓存键，值为缓存项
	items map[string]*memoryItem
	// mu 保护items的读写锁
	mu sync.RWMutex
	// janitor 后台清理过期项的清理器
	janitor *janitor
	// options 缓存选项
	options Options
	// tagIndex 标签索引，将标签映射到使用该标签的所有键
	// 提供O(1)时间复杂度的标签查找
	tagIndex map[string]map[string]struct{}
	// tagMu 保护标签索引的读写锁
	tagMu sync.RWMutex
}

// memoryItem 内存缓存项，存储值和元数据
type memoryItem struct {
	// Value 缓存的实际值
	Value interface{}
	// Expiration 过期时间（Unix纳秒时间戳），0表示永不过期
	Expiration int64
	// Tags 该缓存项关联的标签列表
	Tags []string
}

// NewMemoryCache 创建新的内存缓存实例
// 参数：
//
//	opts: 缓存选项，如果提供多个，只使用第一个；如果不提供，使用默认选项
//
// 返回：
//
//	*MemoryCache: 配置好的内存缓存实例
//
// 注意：
//   - 如果CleanupInterval大于0，会启动后台清理goroutine
//   - 应在不再使用缓存时调用Close方法停止清理goroutine
//
// 示例：
//
//	// 创建使用默认选项的缓存
//	cache := cache.NewMemoryCache()
//
//	// 创建自定义选项的缓存
//	cache := cache.NewMemoryCache(cache.Options{
//	    DefaultTTL: time.Minute * 5,
//	    CleanupInterval: time.Minute,
//	})
func NewMemoryCache(opts ...Options) *MemoryCache {
	var options Options
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = DefaultOptions()
	}

	c := &MemoryCache{
		items:    make(map[string]*memoryItem),
		options:  options,
		tagIndex: make(map[string]map[string]struct{}),
	}

	// 启动过期项清理器
	if options.CleanupInterval > 0 {
		c.janitor = newJanitor(options.CleanupInterval)
		c.janitor.run(c)
	}

	return c
}

// Get 从缓存中检索值
// 参数：
//
//	ctx: 上下文（本实现中未使用，但保留以兼容接口）
//	key: 要检索的缓存键
//
// 返回：
//
//	interface{}: 缓存值，如果未找到则为nil
//	bool: 是否找到缓存，true表示找到，false表示未找到或已过期
//
// 性能：
//   - 使用读锁，多个并发读取不会相互阻塞
//   - 时间复杂度：O(1)
//
// 示例：
//
//	value, found := cache.Get(context.Background(), "user:123")
//	if found {
//	    user := value.(*User)
//	    // 使用用户数据
//	}
func (c *MemoryCache) Get(_ context.Context, key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// 检查是否过期
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

// GetWithTTL 获取值和剩余TTL
// 参数：
//
//	ctx: 上下文（本实现中未使用，但保留以兼容接口）
//	key: 要检索的缓存键
//
// 返回：
//
//	interface{}: 缓存值，如果未找到则为nil
//	time.Duration: 剩余生存时间，如果永不过期则为-1，如果未找到则为0
//	bool: 是否找到缓存，true表示找到，false表示未找到或已过期
//
// 示例：
//
//	value, ttl, found := cache.GetWithTTL(context.Background(), "session:abc")
//	if found {
//	    fmt.Printf("会话将在 %v 后过期\n", ttl)
//	}
func (c *MemoryCache) GetWithTTL(_ context.Context, key string) (interface{}, time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, 0, false
	}

	// 如果没有过期时间，则返回-1表示永不过期
	if item.Expiration == 0 {
		return item.Value, -1, true
	}

	// 检查是否过期
	now := time.Now().UnixNano()
	if now > item.Expiration {
		return nil, 0, false
	}

	// 计算剩余TTL
	ttl := time.Duration(item.Expiration - now)
	return item.Value, ttl, true
}

// Set 设置缓存值
// 参数：
//
//	ctx: 上下文（本实现中未使用，但保留以兼容接口）
//	key: 缓存键
//	value: 要存储的值
//	ttl: 生存时间，0表示使用默认TTL，负数表示永不过期
//
// 返回：
//
//	error: 操作过程中遇到的错误，本实现始终返回nil
//
// 注意：
//   - 如果键已存在，会更新值和过期时间，并清除旧的标签关联
//   - 使用此方法设置的项没有标签，如需标签请使用SetItem
//
// 性能：
//   - 使用写锁，会阻塞所有读写操作
//   - 如果键已存在且有标签，会有额外的标签索引更新开销
//
// 示例：
//
//	// 设置带1分钟过期时间的缓存
//	cache.Set(context.Background(), "counter", 42, time.Minute)
//
//	// 设置永不过期的缓存
//	cache.Set(context.Background(), "app:config", config, -1)
func (c *MemoryCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否已存在，如果存在，需要更新标签索引
	oldItem, found := c.items[key]
	if found {
		c.removeItemFromTagIndex(key, oldItem.Tags)
	}

	c.items[key] = &memoryItem{
		Value:      value,
		Expiration: exp,
		Tags:       []string{},
	}

	return nil
}

// SetItem 设置带完整选项的缓存项
// 参数：
//
//	ctx: 上下文（本实现中未使用，但保留以兼容接口）
//	key: 缓存键
//	item: 包含值、过期时间和标签的缓存项
//
// 返回：
//
//	error: 操作过程中遇到的错误，本实现始终返回nil
//
// 注意：
//   - 相比Set方法，此方法支持设置标签
//   - 如果键已存在，会完全替换旧项，包括更新标签索引
//
// 边界条件：
//   - 标签数量不宜过多，每个标签都会在标签索引中占用内存
//   - 大量使用相同标签的项可能导致删除标签时性能下降
//
// 示例：
//
//	cache.SetItem(context.Background(), "user:123", &cache.Item{
//	    Value: user,
//	    Expiration: time.Hour,
//	    Tags: []string{"user", "active"},
//	})
func (c *MemoryCache) SetItem(_ context.Context, key string, item *Item) error {
	var exp int64
	if item.Expiration > 0 {
		exp = time.Now().Add(item.Expiration).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否已存在，如果存在，需要更新标签索引
	oldItem, found := c.items[key]
	if found {
		c.removeItemFromTagIndex(key, oldItem.Tags)
	}

	mItem := &memoryItem{
		Value:      item.Value,
		Expiration: exp,
		Tags:       item.Tags,
	}

	c.items[key] = mItem

	// 更新标签索引
	c.updateTagIndex(key, item.Tags)

	return nil
}

// Delete 从缓存中删除项
// 参数：
//
//	ctx: 上下文（本实现中未使用，但保留以兼容接口）
//	key: 要删除的缓存键
//
// 返回：
//
//	error: 操作过程中遇到的错误，本实现始终返回nil
//
// 注意：
//   - 如果键不存在，操作无效但不会返回错误
//   - 删除操作会同时更新标签索引
//
// 示例：
//
//	cache.Delete(context.Background(), "user:123")
func (c *MemoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否存在，并更新标签索引
	if item, found := c.items[key]; found {
		c.removeItemFromTagIndex(key, item.Tags)
		delete(c.items, key)
	}

	return nil
}

// DeleteByPattern 根据正则表达式模式删除缓存
// 参数：
//
//	ctx: 上下文（本实现中未使用，但保留以兼容接口）
//	pattern: 正则表达式模式字符串
//
// 返回：
//
//	error: 如果正则表达式无效则返回错误，否则返回nil
//
// 性能：
//   - 时间复杂度：O(n)，其中n为缓存项数量
//   - 对于大量缓存项，此操作可能较慢且占用写锁时间较长
//
// 安全：
//   - 复杂的正则表达式可能导致性能问题，请谨慎使用
//
// 示例：
//
//	// 删除所有用户缓存
//	cache.DeleteByPattern(context.Background(), "^user:")
//
//	// 删除特定ID范围的项
//	cache.DeleteByPattern(context.Background(), "item:[1-9][0-9]$")
func (c *MemoryCache) DeleteByPattern(_ context.Context, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if re.MatchString(key) {
			c.removeItemFromTagIndex(key, item.Tags)
			delete(c.items, key)
		}
	}

	return nil
}

// DeleteByTag 删除特定标签的所有缓存
// 参数：
//
//	ctx: 上下文（本实现中未使用，但保留以兼容接口）
//	tag: 标签名称
//
// 返回：
//
//	error: 操作过程中遇到的错误，本实现始终返回nil
//
// 性能：
//   - 标签索引提供O(1)时间复杂度的标签查找
//   - 实际删除时间与使用该标签的缓存项数量成正比
//
// 示例：
//
//	// 删除所有带"admin"标签的缓存
//	cache.DeleteByTag(context.Background(), "admin")
func (c *MemoryCache) DeleteByTag(_ context.Context, tag string) error {
	c.tagMu.RLock()
	keys, exists := c.tagIndex[tag]
	c.tagMu.RUnlock()

	if !exists {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 复制键列表，避免在遍历时修改map
	keysToDelete := make([]string, 0, len(keys))
	for key := range keys {
		keysToDelete = append(keysToDelete, key)
	}

	// 删除所有标记的键
	for _, key := range keysToDelete {
		if item, found := c.items[key]; found {
			c.removeItemFromTagIndex(key, item.Tags)
			delete(c.items, key)
		}
	}

	return nil
}

// Exists 检查键是否存在
func (c *MemoryCache) Exists(_ context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return false, nil
	}

	// 检查是否过期
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return false, nil
	}

	return true, nil
}

// Increment 增加数值
func (c *MemoryCache) Increment(_ context.Context, key string, value int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	if !found {
		// 如果不存在，创建新项
		c.items[key] = &memoryItem{
			Value:      value,
			Expiration: 0,
			Tags:       []string{},
		}
		return value, nil
	}

	// 检查是否过期
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		// 如果过期，创建新项
		c.items[key] = &memoryItem{
			Value:      value,
			Expiration: 0,
			Tags:       []string{},
		}
		return value, nil
	}

	// 尝试转换为整数类型并增加
	var newValue int64
	switch v := item.Value.(type) {
	case int:
		newValue = int64(v) + value
	case int8:
		newValue = int64(v) + value
	case int16:
		newValue = int64(v) + value
	case int32:
		newValue = int64(v) + value
	case int64:
		newValue = v + value
	case uint:
		newValue = int64(v) + value
	case uint8:
		newValue = int64(v) + value
	case uint16:
		newValue = int64(v) + value
	case uint32:
		newValue = int64(v) + value
	case uint64:
		newValue = int64(v) + value
	case float32:
		newValue = int64(v) + value
	case float64:
		newValue = int64(v) + value
	default:
		return 0, fmt.Errorf("value is not a number: %v", item.Value)
	}

	// 更新值
	item.Value = newValue
	return newValue, nil
}

// Decrement 减少数值
func (c *MemoryCache) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return c.Increment(ctx, key, -value)
}

// Flush 清空所有缓存
func (c *MemoryCache) Flush(_ context.Context) error {
	c.mu.Lock()
	c.items = make(map[string]*memoryItem)
	c.mu.Unlock()

	c.tagMu.Lock()
	c.tagIndex = make(map[string]map[string]struct{})
	c.tagMu.Unlock()

	return nil
}

// Close 关闭缓存连接
func (c *MemoryCache) Close() error {
	if c.janitor != nil {
		c.janitor.stop()
	}
	return nil
}

// 从标签索引中删除项
func (c *MemoryCache) removeItemFromTagIndex(key string, tags []string) {
	c.tagMu.Lock()
	defer c.tagMu.Unlock()

	for _, tag := range tags {
		if keys, exists := c.tagIndex[tag]; exists {
			delete(keys, key)
			// 如果标签没有关联的键，删除该标签
			if len(keys) == 0 {
				delete(c.tagIndex, tag)
			}
		}
	}
}

// 更新标签索引
func (c *MemoryCache) updateTagIndex(key string, tags []string) {
	c.tagMu.Lock()
	defer c.tagMu.Unlock()

	for _, tag := range tags {
		if _, exists := c.tagIndex[tag]; !exists {
			c.tagIndex[tag] = make(map[string]struct{})
		}
		c.tagIndex[tag][key] = struct{}{}
	}
}

// 删除过期项
func (c *MemoryCache) deleteExpired() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if item.Expiration > 0 && now > item.Expiration {
			c.removeItemFromTagIndex(key, item.Tags)
			delete(c.items, key)
		}
	}
}

// janitor 负责定期清理过期项
type janitor struct {
	interval time.Duration
	stopChan chan bool
}

func newJanitor(interval time.Duration) *janitor {
	return &janitor{
		interval: interval,
		stopChan: make(chan bool),
	}
}

func (j *janitor) run(c *MemoryCache) {
	ticker := time.NewTicker(j.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.deleteExpired()
			case <-j.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (j *janitor) stop() {
	j.stopChan <- true
}
