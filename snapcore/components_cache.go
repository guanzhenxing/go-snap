package snapcore

import (
	"context"
	"time"

	"github.com/guanzhenxing/go-snap/cache"
	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
)

// CacheConfig 缓存配置结构
type CacheConfig struct {
	// 缓存类型：memory, redis, multi_level
	Type string `json:"type"`

	// 内存缓存配置
	Memory struct {
		DefaultTTL      string `json:"default_ttl"`
		CleanupInterval string `json:"cleanup_interval"`
	} `json:"memory"`

	// Redis缓存配置
	Redis struct {
		Mode            string   `json:"mode"`
		Addr            string   `json:"addr"`
		Username        string   `json:"username"`
		Password        string   `json:"password"`
		DB              int      `json:"db"`
		MasterName      string   `json:"master_name"`
		SentinelAddrs   []string `json:"sentinel_addrs"`
		ClusterAddrs    []string `json:"cluster_addrs"`
		MaxRetries      int      `json:"max_retries"`
		DialTimeout     string   `json:"dial_timeout"`
		ReadTimeout     string   `json:"read_timeout"`
		WriteTimeout    string   `json:"write_timeout"`
		PoolSize        int      `json:"pool_size"`
		MinIdleConns    int      `json:"min_idle_conns"`
		MaxIdleConns    int      `json:"max_idle_conns"`
		ConnMaxIdleTime string   `json:"conn_max_idle_time"`
		ConnMaxLifetime string   `json:"conn_max_lifetime"`
		KeyPrefix       string   `json:"key_prefix"`
		DefaultTTL      string   `json:"default_ttl"`
	} `json:"redis"`

	// 多级缓存配置
	MultiLevel struct {
		WriteMode string `json:"write_mode"`
		LocalTTL  string `json:"local_ttl"`
	} `json:"multi_level"`
}

// CacheComponent 缓存组件适配器
type CacheComponent struct {
	name   string
	cache  cache.Cache
	config config.Provider
	logger logger.Logger
}

// NewCacheComponent 创建缓存组件
func NewCacheComponent() *CacheComponent {
	return &CacheComponent{
		name: "cache",
	}
}

// Initialize 初始化缓存组件
func (c *CacheComponent) Initialize(ctx context.Context) error {
	// 从配置中读取缓存配置
	var cacheConfig CacheConfig
	if err := c.config.UnmarshalKey("cache", &cacheConfig); err != nil {
		return errors.Wrap(err, "unmarshal cache config failed")
	}

	c.logger.Info("Initializing cache component",
		logger.String("type", cacheConfig.Type),
	)

	var err error

	// 根据缓存类型创建缓存实例
	switch cacheConfig.Type {
	case "memory":
		c.cache, err = c.createMemoryCache(cacheConfig)
	case "redis":
		c.cache, err = c.createRedisCache(cacheConfig)
	case "multi_level":
		c.cache, err = c.createMultiLevelCache(cacheConfig)
	default:
		// 默认使用内存缓存
		c.logger.Warn("Unknown cache type, using memory cache as default",
			logger.String("type", cacheConfig.Type),
		)
		c.cache, err = c.createMemoryCache(cacheConfig)
	}

	if err != nil {
		return errors.Wrap(err, "create cache failed")
	}

	return nil
}

// createMemoryCache 创建内存缓存
func (c *CacheComponent) createMemoryCache(cfg CacheConfig) (cache.Cache, error) {
	options := cache.DefaultOptions()

	// 解析配置项
	if cfg.Memory.DefaultTTL != "" {
		ttl, err := time.ParseDuration(cfg.Memory.DefaultTTL)
		if err == nil {
			options.DefaultTTL = ttl
		}
	}

	if cfg.Memory.CleanupInterval != "" {
		interval, err := time.ParseDuration(cfg.Memory.CleanupInterval)
		if err == nil {
			options.CleanupInterval = interval
		}
	}

	// 创建内存缓存
	return cache.NewCache(cache.CacheTypeMemory, options)
}

// createRedisCache 创建Redis缓存
func (c *CacheComponent) createRedisCache(cfg CacheConfig) (cache.Cache, error) {
	options := cache.DefaultRedisOptions()

	// 解析Redis配置
	if cfg.Redis.Mode != "" {
		switch cfg.Redis.Mode {
		case "standalone":
			options.Mode = cache.RedisModeStandalone
		case "sentinel":
			options.Mode = cache.RedisModeSentinel
		case "cluster":
			options.Mode = cache.RedisModeCluster
		}
	}

	// 基本连接信息
	if cfg.Redis.Addr != "" {
		options.Addr = cfg.Redis.Addr
	}
	options.Username = cfg.Redis.Username
	options.Password = cfg.Redis.Password
	if cfg.Redis.DB > 0 {
		options.DB = cfg.Redis.DB
	}

	// 哨兵模式配置
	options.MasterName = cfg.Redis.MasterName
	if len(cfg.Redis.SentinelAddrs) > 0 {
		options.SentinelAddrs = cfg.Redis.SentinelAddrs
	}

	// 集群模式配置
	if len(cfg.Redis.ClusterAddrs) > 0 {
		options.ClusterAddrs = cfg.Redis.ClusterAddrs
	}

	// 连接池配置
	if cfg.Redis.MaxRetries > 0 {
		options.MaxRetries = cfg.Redis.MaxRetries
	}
	if cfg.Redis.PoolSize > 0 {
		options.PoolSize = cfg.Redis.PoolSize
	}
	if cfg.Redis.MinIdleConns > 0 {
		options.MinIdleConns = cfg.Redis.MinIdleConns
	}
	if cfg.Redis.MaxIdleConns > 0 {
		options.MaxIdleConns = cfg.Redis.MaxIdleConns
	}

	// 超时配置
	if cfg.Redis.DialTimeout != "" {
		timeout, err := time.ParseDuration(cfg.Redis.DialTimeout)
		if err == nil {
			options.DialTimeout = timeout
		}
	}
	if cfg.Redis.ReadTimeout != "" {
		timeout, err := time.ParseDuration(cfg.Redis.ReadTimeout)
		if err == nil {
			options.ReadTimeout = timeout
		}
	}
	if cfg.Redis.WriteTimeout != "" {
		timeout, err := time.ParseDuration(cfg.Redis.WriteTimeout)
		if err == nil {
			options.WriteTimeout = timeout
		}
	}

	// 连接生命周期
	if cfg.Redis.ConnMaxIdleTime != "" {
		duration, err := time.ParseDuration(cfg.Redis.ConnMaxIdleTime)
		if err == nil {
			options.ConnMaxIdleTime = duration
		}
	}
	if cfg.Redis.ConnMaxLifetime != "" {
		duration, err := time.ParseDuration(cfg.Redis.ConnMaxLifetime)
		if err == nil {
			options.ConnMaxLifetime = duration
		}
	}

	// 其他选项
	options.KeyPrefix = cfg.Redis.KeyPrefix
	if cfg.Redis.DefaultTTL != "" {
		ttl, err := time.ParseDuration(cfg.Redis.DefaultTTL)
		if err == nil {
			options.DefaultTTL = ttl
		}
	}

	// 创建Redis缓存
	return cache.NewCache(cache.CacheTypeRedis, options)
}

// createMultiLevelCache 创建多级缓存
func (c *CacheComponent) createMultiLevelCache(cfg CacheConfig) (cache.Cache, error) {
	// 创建本地缓存（内存缓存）
	localCache, err := c.createMemoryCache(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create local cache failed")
	}

	// 创建远程缓存（Redis缓存）
	remoteCache, err := c.createRedisCache(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create remote cache failed")
	}

	// 创建多级缓存选项
	options := cache.DefaultMultiLevelOptions()

	// 解析多级缓存配置
	if cfg.MultiLevel.WriteMode == "write_back" {
		options.WriteMode = cache.WriteModeWriteBack
	} else {
		options.WriteMode = cache.WriteModeWriteThrough
	}

	if cfg.MultiLevel.LocalTTL != "" {
		ttl, err := time.ParseDuration(cfg.MultiLevel.LocalTTL)
		if err == nil {
			options.LocalTTL = ttl
		}
	}

	// 使用Builder创建多级缓存
	builder := cache.NewBuilder().
		WithType(cache.CacheTypeMultiLevel).
		WithLocalCache(localCache).
		WithOptions(struct {
			RemoteCache cache.Cache
			Options     cache.MultiLevelOptions
		}{
			RemoteCache: remoteCache,
			Options:     options,
		})

	return builder.Build()
}

// Start 启动缓存组件
func (c *CacheComponent) Start(ctx context.Context) error {
	c.logger.Info("Starting cache component")
	return nil
}

// Stop 停止缓存组件
func (c *CacheComponent) Stop(ctx context.Context) error {
	c.logger.Info("Stopping cache component")
	if c.cache != nil {
		return c.cache.Close()
	}
	return nil
}

// Name 获取组件名称
func (c *CacheComponent) Name() string {
	return c.name
}

// Type 获取组件类型
func (c *CacheComponent) Type() ComponentType {
	return ComponentTypeDataSource
}

// Dependencies 获取组件依赖
func (c *CacheComponent) Dependencies() []string {
	return []string{"config", "logger"}
}

// GetCache 获取缓存实例
func (c *CacheComponent) GetCache() cache.Cache {
	return c.cache
}

// SetConfig 设置配置提供器
func (c *CacheComponent) SetConfig(config config.Provider) {
	c.config = config
}

// SetLogger 设置日志器
func (c *CacheComponent) SetLogger(logger logger.Logger) {
	c.logger = logger
}
