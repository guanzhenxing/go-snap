package appcore

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

// 确保CacheComponent实现了CacheProvider接口
var _ CacheProvider = (*CacheComponent)(nil)

// NewCacheComponent 创建缓存组件
func NewCacheComponent() *CacheComponent {
	return &CacheComponent{
		name: "cache",
	}
}

// Initialize 初始化缓存组件
func (c *CacheComponent) Initialize(ctx context.Context) error {
	// 确保配置已设置
	if c.config == nil {
		return errors.New("缓存组件需要配置")
	}

	// 从配置中获取缓存配置
	var cacheConfig struct {
		Type       string `json:"type"`
		DefaultTTL string `json:"default_ttl"`

		// Redis配置
		Redis struct {
			Addr       string `json:"addr"`
			Password   string `json:"password"`
			DB         int    `json:"db"`
			MaxRetries int    `json:"max_retries"`
			KeyPrefix  string `json:"key_prefix"`
		} `json:"redis"`
	}

	if err := c.config.UnmarshalKey("cache", &cacheConfig); err != nil {
		if c.logger != nil {
			c.logger.Warn("解析缓存配置失败，将使用默认配置",
				logger.String("error", err.Error()),
			)
		}
		// 使用默认配置 - 内存缓存
		cacheConfig.Type = string(cache.CacheTypeMemory)
	}

	// 确定缓存类型
	cacheType := cache.CacheType(cacheConfig.Type)
	if cacheType == "" {
		cacheType = cache.CacheTypeMemory
	}

	// 创建缓存配置选项
	var cacheOptions interface{}

	switch cacheType {
	case cache.CacheTypeMemory:
		opts := cache.DefaultOptions()
		if cacheConfig.DefaultTTL != "" {
			if ttl, err := time.ParseDuration(cacheConfig.DefaultTTL); err == nil {
				opts.DefaultTTL = ttl
			}
		}
		cacheOptions = opts

	case cache.CacheTypeRedis:
		opts := cache.DefaultRedisOptions()
		if cacheConfig.Redis.Addr != "" {
			opts.Addr = cacheConfig.Redis.Addr
		}
		opts.Password = cacheConfig.Redis.Password
		opts.DB = cacheConfig.Redis.DB
		opts.KeyPrefix = cacheConfig.Redis.KeyPrefix
		if cacheConfig.Redis.MaxRetries > 0 {
			opts.MaxRetries = cacheConfig.Redis.MaxRetries
		}
		if cacheConfig.DefaultTTL != "" {
			if ttl, err := time.ParseDuration(cacheConfig.DefaultTTL); err == nil {
				opts.DefaultTTL = ttl
			}
		}
		cacheOptions = opts
	}

	// 创建缓存实例
	cacheInstance, err := cache.NewCache(cacheType, cacheOptions)
	if err != nil {
		return errors.Wrap(err, "创建缓存实例失败")
	}

	c.cache = cacheInstance

	if c.logger != nil {
		c.logger.Info("缓存组件初始化成功",
			logger.String("type", string(cacheType)),
		)
	}

	return nil
}

// Start 启动缓存组件
func (c *CacheComponent) Start(ctx context.Context) error {
	if c.logger != nil {
		c.logger.Info("缓存组件已启动")
	}
	return nil
}

// Stop 停止缓存组件
func (c *CacheComponent) Stop(ctx context.Context) error {
	if c.logger != nil {
		c.logger.Info("缓存组件正在停止")
	}

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

// GetCache 获取缓存实例
func (c *CacheComponent) GetCache() interface{} {
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

// NewCacheProvider 提供缓存组件
func NewCacheProvider(config config.Provider, logger logger.Logger) (CacheProvider, error) {
	component := NewCacheComponent()
	component.SetConfig(config)
	component.SetLogger(logger)
	if err := component.Initialize(context.Background()); err != nil {
		return nil, err
	}
	return component, nil
}
