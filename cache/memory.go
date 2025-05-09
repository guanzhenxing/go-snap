package cache

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"
)

// MemoryCache 实现基于内存的缓存
type MemoryCache struct {
	items    map[string]*memoryItem
	mu       sync.RWMutex
	janitor  *janitor
	options  Options
	tagIndex map[string]map[string]struct{} // 标签到键的映射
	tagMu    sync.RWMutex
}

type memoryItem struct {
	Value      interface{}
	Expiration int64 // Unix 纳秒时间戳
	Tags       []string
}

// NewMemoryCache 创建新的内存缓存实例
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
