package cache

import (
	"fmt"
)

// CacheType 表示缓存类型
type CacheType string

const (
	// CacheTypeMemory 内存缓存
	CacheTypeMemory CacheType = "memory"
	// CacheTypeRedis Redis缓存
	CacheTypeRedis CacheType = "redis"
	// CacheTypeMultiLevel 多级缓存
	CacheTypeMultiLevel CacheType = "multi_level"
)

// NewCache 创建指定类型的缓存
func NewCache(cacheType CacheType, options interface{}) (Cache, error) {
	switch cacheType {
	case CacheTypeMemory:
		// 创建内存缓存
		var opts Options
		if options != nil {
			if o, ok := options.(Options); ok {
				opts = o
			} else {
				return nil, fmt.Errorf("invalid options type for memory cache: %T", options)
			}
		} else {
			opts = DefaultOptions()
		}
		return NewMemoryCache(opts), nil

	case CacheTypeRedis:
		// 创建Redis缓存
		var opts RedisOptions
		var serializer Serializer

		if options != nil {
			switch o := options.(type) {
			case RedisOptions:
				opts = o
			case struct {
				Options    RedisOptions
				Serializer Serializer
			}:
				opts = o.Options
				serializer = o.Serializer
			default:
				return nil, fmt.Errorf("invalid options type for redis cache: %T", options)
			}
		} else {
			opts = DefaultRedisOptions()
		}

		return NewRedisCache(opts, serializer)

	case CacheTypeMultiLevel:
		// 创建多级缓存
		multiOpts, ok := options.(struct {
			LocalCache  Cache
			RemoteCache Cache
			Options     MultiLevelOptions
		})
		if !ok {
			return nil, fmt.Errorf("invalid options type for multi-level cache: %T", options)
		}

		return NewMultiLevelCache(multiOpts.LocalCache, multiOpts.RemoteCache, multiOpts.Options)

	default:
		return nil, fmt.Errorf("unsupported cache type: %s", cacheType)
	}
}

// Builder 缓存构建器模式
type Builder struct {
	cacheType  CacheType
	options    interface{}
	serializer Serializer
	localCache Cache
	localTTL   interface{}
}

// NewBuilder 创建新的缓存构建器
func NewBuilder() *Builder {
	return &Builder{}
}

// WithType 设置缓存类型
func (b *Builder) WithType(cacheType CacheType) *Builder {
	b.cacheType = cacheType
	return b
}

// WithOptions 设置缓存选项
func (b *Builder) WithOptions(options interface{}) *Builder {
	b.options = options
	return b
}

// WithSerializer 设置序列化器
func (b *Builder) WithSerializer(serializer Serializer) *Builder {
	b.serializer = serializer
	return b
}

// WithLocalCache 设置本地缓存（用于多级缓存）
func (b *Builder) WithLocalCache(cache Cache) *Builder {
	b.localCache = cache
	return b
}

// Build 构建缓存实例
func (b *Builder) Build() (Cache, error) {
	switch b.cacheType {
	case CacheTypeMemory:
		if b.options == nil {
			b.options = DefaultOptions()
		}
		opts, ok := b.options.(Options)
		if !ok {
			return nil, fmt.Errorf("invalid options type for memory cache: %T", b.options)
		}
		return NewMemoryCache(opts), nil

	case CacheTypeRedis:
		if b.options == nil {
			b.options = DefaultRedisOptions()
		}
		opts, ok := b.options.(RedisOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for redis cache: %T", b.options)
		}
		return NewRedisCache(opts, b.serializer)

	case CacheTypeMultiLevel:
		if b.localCache == nil {
			return nil, fmt.Errorf("local cache is required for multi-level cache")
		}
		if b.options == nil {
			b.options = DefaultMultiLevelOptions()
		}

		var remoteCache Cache
		var err error

		// 检查远程缓存选项
		switch opts := b.options.(type) {
		case struct {
			RemoteOptions RedisOptions
			Options       MultiLevelOptions
		}:
			remoteCache, err = NewRedisCache(opts.RemoteOptions, b.serializer)
			if err != nil {
				return nil, err
			}
			return NewMultiLevelCache(b.localCache, remoteCache, opts.Options)

		case struct {
			RemoteCache Cache
			Options     MultiLevelOptions
		}:
			return NewMultiLevelCache(b.localCache, opts.RemoteCache, opts.Options)

		case MultiLevelOptions:
			// 如果只提供了多级缓存选项，需要先检查是否已经设置了远程缓存
			if b.options == nil {
				return nil, fmt.Errorf("remote cache or options is required for multi-level cache")
			}
			remoteOpts, ok := b.options.(RedisOptions)
			if !ok {
				return nil, fmt.Errorf("invalid remote options type: %T", b.options)
			}
			remoteCache, err = NewRedisCache(remoteOpts, b.serializer)
			if err != nil {
				return nil, err
			}
			return NewMultiLevelCache(b.localCache, remoteCache, opts)

		default:
			return nil, fmt.Errorf("invalid options type for multi-level cache: %T", b.options)
		}

	default:
		return nil, fmt.Errorf("unsupported cache type: %s", b.cacheType)
	}
}
