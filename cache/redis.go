package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/guanzhenxing/go-snap/errors"

	"github.com/redis/go-redis/v9"
)

// RedisMode 表示Redis的不同工作模式
type RedisMode string

const (
	// RedisModeStandalone 单机模式
	RedisModeStandalone RedisMode = "standalone"
	// RedisModeSentinel 哨兵模式
	RedisModeSentinel RedisMode = "sentinel"
	// RedisModeCluster 集群模式
	RedisModeCluster RedisMode = "cluster"
)

// RedisOptions Redis缓存选项
type RedisOptions struct {
	// 工作模式：standalone(单机)、sentinel(哨兵)或cluster(集群)
	Mode RedisMode

	// -------- 单机模式配置 --------
	Addr     string // 单机模式：redis地址，格式为 host:port
	Username string // 用户名，Redis 6.0+支持
	Password string // 密码

	// -------- 哨兵模式配置 --------
	MasterName    string   // 哨兵模式：master名称
	SentinelAddrs []string // 哨兵模式：哨兵地址列表

	// -------- 集群模式配置 --------
	ClusterAddrs []string // 集群模式：集群节点地址列表

	// -------- 通用选项 --------
	DB              int           // 数据库编号，仅适用于单机和哨兵模式
	MaxRetries      int           // 命令重试最大次数
	MinRetryBackoff time.Duration // 最小重试间隔
	MaxRetryBackoff time.Duration // 最大重试间隔
	DialTimeout     time.Duration // 连接超时
	ReadTimeout     time.Duration // 读取超时
	WriteTimeout    time.Duration // 写入超时
	PoolSize        int           // 连接池大小
	MinIdleConns    int           // 最小空闲连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxIdleTime time.Duration // 连接最大空闲时间
	ConnMaxLifetime time.Duration // 连接最大生存时间

	// -------- 缓存键前缀 --------
	KeyPrefix string // 所有键的统一前缀

	// -------- 默认过期时间 --------
	DefaultTTL time.Duration // 默认TTL
}

// DefaultRedisOptions 返回默认Redis配置
func DefaultRedisOptions() RedisOptions {
	return RedisOptions{
		Mode:            RedisModeStandalone,
		Addr:            "localhost:6379",
		DB:              0,
		MaxRetries:      3,
		MinRetryBackoff: time.Millisecond * 8,
		MaxRetryBackoff: time.Millisecond * 512,
		DialTimeout:     time.Second * 5,
		ReadTimeout:     time.Second * 3,
		WriteTimeout:    time.Second * 3,
		PoolSize:        10,
		MinIdleConns:    2,
		DefaultTTL:      time.Hour,
	}
}

// RedisCache 实现基于Redis的缓存
type RedisCache struct {
	client     redis.UniversalClient
	options    RedisOptions
	serializer Serializer
}

// NewRedisCache 创建新的Redis缓存
func NewRedisCache(opts RedisOptions, serializer Serializer) (*RedisCache, error) {
	if serializer == nil {
		serializer = DefaultSerializer()
	}

	var client redis.UniversalClient

	switch opts.Mode {
	case RedisModeStandalone:
		// 单机模式
		client = redis.NewClient(&redis.Options{
			Addr:            opts.Addr,
			Username:        opts.Username,
			Password:        opts.Password,
			DB:              opts.DB,
			MaxRetries:      opts.MaxRetries,
			MinRetryBackoff: opts.MinRetryBackoff,
			MaxRetryBackoff: opts.MaxRetryBackoff,
			DialTimeout:     opts.DialTimeout,
			ReadTimeout:     opts.ReadTimeout,
			WriteTimeout:    opts.WriteTimeout,
			PoolSize:        opts.PoolSize,
			MinIdleConns:    opts.MinIdleConns,
		})
	case RedisModeSentinel:
		// 哨兵模式
		if opts.MasterName == "" {
			return nil, errors.New("master name is required for sentinel mode")
		}
		if len(opts.SentinelAddrs) == 0 {
			return nil, errors.New("sentinel addresses are required for sentinel mode")
		}

		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:      opts.MasterName,
			SentinelAddrs:   opts.SentinelAddrs,
			Username:        opts.Username,
			Password:        opts.Password,
			DB:              opts.DB,
			MaxRetries:      opts.MaxRetries,
			MinRetryBackoff: opts.MinRetryBackoff,
			MaxRetryBackoff: opts.MaxRetryBackoff,
			DialTimeout:     opts.DialTimeout,
			ReadTimeout:     opts.ReadTimeout,
			WriteTimeout:    opts.WriteTimeout,
			PoolSize:        opts.PoolSize,
			MinIdleConns:    opts.MinIdleConns,
		})
	case RedisModeCluster:
		// 集群模式
		if len(opts.ClusterAddrs) == 0 {
			return nil, errors.New("cluster addresses are required for cluster mode")
		}

		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:           opts.ClusterAddrs,
			Username:        opts.Username,
			Password:        opts.Password,
			MaxRetries:      opts.MaxRetries,
			MinRetryBackoff: opts.MinRetryBackoff,
			MaxRetryBackoff: opts.MaxRetryBackoff,
			DialTimeout:     opts.DialTimeout,
			ReadTimeout:     opts.ReadTimeout,
			WriteTimeout:    opts.WriteTimeout,
			PoolSize:        opts.PoolSize,
			MinIdleConns:    opts.MinIdleConns,
		})
	default:
		return nil, fmt.Errorf("unsupported Redis mode: %s", opts.Mode)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), opts.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client:     client,
		options:    opts,
		serializer: serializer,
	}, nil
}

// 为键添加前缀
func (c *RedisCache) prefixKey(key string) string {
	if c.options.KeyPrefix == "" {
		return key
	}
	return c.options.KeyPrefix + ":" + key
}

// Get 从缓存中获取值
func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, bool) {
	prefixedKey := c.prefixKey(key)
	data, err := c.client.Get(ctx, prefixedKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		return nil, false
	}

	var value interface{}
	if err := c.serializer.Unmarshal(data, &value); err != nil {
		return nil, false
	}

	return value, true
}

// GetWithTTL 获取值和剩余TTL
func (c *RedisCache) GetWithTTL(ctx context.Context, key string) (interface{}, time.Duration, bool) {
	prefixedKey := c.prefixKey(key)

	// 获取值
	data, err := c.client.Get(ctx, prefixedKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, 0, false
		}
		return nil, 0, false
	}

	// 获取TTL
	ttl, err := c.client.TTL(ctx, prefixedKey).Result()
	if err != nil {
		return nil, 0, false
	}

	var value interface{}
	if err := c.serializer.Unmarshal(data, &value); err != nil {
		return nil, 0, false
	}

	return value, ttl, true
}

// Set 设置缓存
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	prefixedKey := c.prefixKey(key)

	data, err := c.serializer.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	// 如果没有指定TTL，使用默认TTL
	if ttl == 0 {
		ttl = c.options.DefaultTTL
	}

	err = c.client.Set(ctx, prefixedKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// SetItem 设置带完整选项的缓存项
func (c *RedisCache) SetItem(ctx context.Context, key string, item *Item) error {
	if item == nil {
		return errors.New("cache item cannot be nil")
	}

	prefixedKey := c.prefixKey(key)

	data, err := c.serializer.Marshal(item.Value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	// 如果没有指定TTL，使用默认TTL
	ttl := item.Expiration
	if ttl == 0 {
		ttl = c.options.DefaultTTL
	}

	// 使用管道以减少网络往返
	pipe := c.client.Pipeline()
	pipe.Set(ctx, prefixedKey, data, ttl)

	// 处理标签索引
	if len(item.Tags) > 0 {
		for _, tag := range item.Tags {
			tagKey := c.prefixKey("tag:" + tag)
			pipe.SAdd(ctx, tagKey, key)
			// 如果设置了过期时间，也为标签索引设置相同的过期时间
			if ttl > 0 {
				pipe.Expire(ctx, tagKey, ttl)
			}
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set cache item: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	prefixedKey := c.prefixKey(key)

	// 获取键关联的标签
	tagsKeyPattern := c.prefixKey("tag:*")
	var cursor uint64
	var tags []string

	for {
		var keys []string
		var err error

		// 扫描所有标签
		keys, cursor, err = c.client.Scan(ctx, cursor, tagsKeyPattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan tags: %w", err)
		}

		// 检查每个标签集合是否包含要删除的键
		for _, tagKey := range keys {
			isMember, err := c.client.SIsMember(ctx, tagKey, key).Result()
			if err != nil {
				continue
			}
			if isMember {
				// 提取标签名
				tag := strings.TrimPrefix(tagKey, c.prefixKey("tag:"))
				tags = append(tags, tag)
			}
		}

		if cursor == 0 {
			break
		}
	}

	// 使用管道删除键和标签引用
	pipe := c.client.Pipeline()
	pipe.Del(ctx, prefixedKey)

	// 从标签集合中移除键
	for _, tag := range tags {
		tagKey := c.prefixKey("tag:" + tag)
		pipe.SRem(ctx, tagKey, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	// 检查并删除空标签
	for _, tag := range tags {
		tagKey := c.prefixKey("tag:" + tag)
		count, err := c.client.SCard(ctx, tagKey).Result()
		if err == nil && count == 0 {
			c.client.Del(ctx, tagKey)
		}
	}

	return nil
}

// DeleteByPattern 根据模式删除缓存
func (c *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	prefixedPattern := c.prefixKey(pattern) + "*"

	// 使用SCAN命令查找匹配的键
	var cursor uint64
	var allKeys []string

	for {
		var keys []string
		var err error
		keys, cursor, err = c.client.Scan(ctx, cursor, prefixedPattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys: %w", err)
		}

		allKeys = append(allKeys, keys...)

		if cursor == 0 {
			break
		}
	}

	// 如果找到匹配的键，删除它们
	if len(allKeys) > 0 {
		err := c.client.Del(ctx, allKeys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

// DeleteByTag 删除带特定标签的所有缓存
func (c *RedisCache) DeleteByTag(ctx context.Context, tag string) error {
	tagKey := c.prefixKey("tag:" + tag)

	// 获取标签关联的所有键
	keys, err := c.client.SMembers(ctx, tagKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("failed to get tag members: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// 预处理键名称
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}

	// 使用管道删除所有键和标签索引
	pipe := c.client.Pipeline()
	pipe.Del(ctx, prefixedKeys...)
	pipe.Del(ctx, tagKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete keys by tag: %w", err)
	}

	return nil
}

// Exists 检查键是否存在
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	prefixedKey := c.prefixKey(key)
	n, err := c.client.Exists(ctx, prefixedKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return n > 0, nil
}

// Increment 增加数值
func (c *RedisCache) Increment(ctx context.Context, key string, value int64) (int64, error) {
	prefixedKey := c.prefixKey(key)
	result, err := c.client.IncrBy(ctx, prefixedKey, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment value: %w", err)
	}
	return result, nil
}

// Decrement 减少数值
func (c *RedisCache) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return c.Increment(ctx, key, -value)
}

// Flush 清空所有缓存
func (c *RedisCache) Flush(ctx context.Context) error {
	if c.options.KeyPrefix == "" {
		// 如果没有前缀，清空整个数据库
		_, err := c.client.FlushDB(ctx).Result()
		if err != nil {
			return fmt.Errorf("failed to flush cache: %w", err)
		}
		return nil
	}

	// 有前缀，只清除前缀对应的键
	prefixPattern := c.prefixKey("*")
	var cursor uint64
	var allKeys []string

	for {
		var keys []string
		var err error
		keys, cursor, err = c.client.Scan(ctx, cursor, prefixPattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys with prefix: %w", err)
		}

		allKeys = append(allKeys, keys...)

		if cursor == 0 {
			break
		}
	}

	if len(allKeys) > 0 {
		err := c.client.Del(ctx, allKeys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete prefixed keys: %w", err)
		}
	}

	return nil
}

// Close 关闭缓存连接
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// GetClient 返回底层的Redis客户端实例
// 注：此方法主要用于与分布式锁模块集成
func (c *RedisCache) GetClient() redis.UniversalClient {
	return c.client
}
