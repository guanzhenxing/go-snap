// Package cache 提供统一的缓存接口和多种实现
// 支持内存缓存、Redis缓存以及多级缓存策略
//
// # 缓存系统架构
//
// 本包设计了一个灵活的缓存抽象层，允许不同的缓存实现无缝切换。主要组件：
//
// 1. Cache接口：所有缓存实现的统一接口
// 2. Item结构：表示带有元数据的缓存项
// 3. Options结构：缓存配置选项
// 4. 实现类：
//   - MemoryCache：本地内存缓存，适用于单实例应用
//   - RedisCache：基于Redis的分布式缓存
//   - MultiLevelCache：组合多个缓存，实现多级缓存策略
//
// # 缓存实现比较
//
// | 特性           | MemoryCache     | RedisCache      | MultiLevelCache |
// |---------------|----------------|----------------|-----------------|
// | 存储位置        | 本地内存         | Redis服务器      | 多级组合         |
// | 持久性          | 重启后丢失       | 持久化（可配置）   | 取决于配置       |
// | 分布式支持      | 不支持          | 支持            | 支持            |
// | 性能            | 极快           | 较快（网络延迟）  | 分层性能        |
// | 内存占用        | 高             | 低（客户端）     | 可控            |
// | 适用场景        | 单实例、高性能   | 分布式系统       | 复杂系统        |
//
// # 主要功能
//
// - 基本的Get/Set操作
// - 支持TTL（生存时间）
// - 标签系统用于批量操作
// - 模式匹配删除
// - 数值增减操作
// - 缓存项元数据
//
// # 使用示例
//
// 1. 基本使用：
//
//	// 创建内存缓存
//	cache := cache.NewMemoryCache()
//
//	// 设置缓存项
//	cache.Set(context.Background(), "user:123", userObj, time.Hour)
//
//	// 获取缓存项
//	if value, found := cache.Get(context.Background(), "user:123"); found {
//	    user := value.(*User)
//	    // 使用用户数据
//	}
//
// 2. 使用标签：
//
//	// 创建带标签的缓存项
//	item := &cache.Item{
//	    Value: userObj,
//	    Expiration: time.Hour,
//	    Tags: []string{"user", "premium"},
//	}
//	cache.SetItem(context.Background(), "user:123", item)
//
//	// 删除所有带特定标签的缓存项
//	cache.DeleteByTag(context.Background(), "premium")
//
// 3. 多级缓存：
//
//	// 创建一级缓存（内存）和二级缓存（Redis）
//	l1 := cache.NewMemoryCache()
//	l2, _ := cache.NewRedisCache(redisClient)
//
//	// 创建多级缓存
//	multiCache := cache.NewMultiLevelCache(l1, l2)
//
//	// 使用多级缓存
//	multiCache.Set(context.Background(), "key", value, time.Minute)
//
// # 性能考虑
//
// - MemoryCache适合高频读取操作，但会占用应用内存
// - RedisCache适合分布式环境，但有网络开销
// - 对于大型系统，推荐使用MultiLevelCache组合两者优势
// - 合理设置TTL和清理间隔以优化内存使用
// - 对于高并发场景，建议使用带缓冲的异步写入策略
//
// # 线程安全性
//
// 所有实现都保证线程安全，可以在并发环境中安全使用
package cache

import (
	"context"
	"time"
)

// Item 表示缓存项，包含值和元数据
// 缓存项封装了要存储的值以及与之相关的元数据，如过期时间和标签
// 使用Item而不是直接使用值可以提供更细粒度的控制
type Item struct {
	// Value 缓存的实际值
	// 可以是任何类型的数据，使用时需要进行类型断言
	// 示例: user := item.Value.(*User)
	Value interface{}

	// Expiration 过期时间，0表示使用默认过期时间，负数表示永不过期
	// 以time.Duration表示，例如time.Hour表示1小时后过期
	// 特殊值:
	// - 0: 使用缓存实现的默认过期时间
	// - 负数: 永不过期
	Expiration time.Duration

	// Tags 缓存标签，用于批量操作，如批量删除特定标签的缓存
	// 标签可以用于逻辑分组缓存项，便于管理和操作
	// 例如: ["user", "admin", "active"]
	Tags []string
}

// Cache 定义缓存的通用接口
// 所有缓存实现（内存、Redis等）都需要实现此接口
// 此接口设计遵循以下原则：
// 1. 简单性：提供最常用的缓存操作
// 2. 一致性：所有方法行为在不同实现中保持一致
// 3. 上下文感知：支持通过context进行超时控制和取消
// 4. 元数据支持：支持TTL和标签等元数据
type Cache interface {
	// Get 获取缓存值
	// 参数：
	//   ctx: 上下文，可用于传递超时和取消信号
	//   key: 缓存键名
	// 返回：
	//   interface{}: 缓存值，如果未找到则为nil
	//   bool: 是否找到缓存，true表示找到，false表示未找到
	// 注意：
	//   - 返回的interface{}需要进行类型断言才能使用
	//   - 已过期的项会返回未找到
	//   - 此方法不会延长缓存项的生命周期
	Get(ctx context.Context, key string) (interface{}, bool)

	// GetWithTTL 获取缓存值和剩余生存时间
	// 参数：
	//   ctx: 上下文，可用于传递超时和取消信号
	//   key: 缓存键名
	// 返回：
	//   interface{}: 缓存值，如果未找到则为nil
	//   time.Duration: 剩余生存时间，如果未找到则为0，负数表示永不过期
	//   bool: 是否找到缓存，true表示找到，false表示未找到
	// 使用场景：
	//   - 需要了解缓存项还有多久过期
	//   - 需要基于剩余时间决定是否刷新缓存
	GetWithTTL(ctx context.Context, key string) (interface{}, time.Duration, bool)

	// Set 设置缓存值及其生存时间
	// 参数：
	//   ctx: 上下文，可用于传递超时和取消信号
	//   key: 缓存键名
	//   value: 要缓存的值
	//   ttl: 生存时间，0表示使用默认生存时间，负数表示永不过期
	// 返回：
	//   error: 设置过程中遇到的错误，如果设置成功则为nil
	// 注意：
	//   - 如果键已存在，将被覆盖
	//   - value可以是任何类型，包括指针和结构体
	//   - 实现应确保value的序列化和反序列化
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// SetItem 设置带完整选项的缓存项
	// 参数：
	//   ctx: 上下文，可用于传递超时和取消信号
	//   key: 缓存键名
	//   item: 包含值、过期时间和标签的缓存项
	// 返回：
	//   error: 设置过程中遇到的错误，如果设置成功则为nil
	// 使用场景：
	//   - 需要为缓存项设置标签
	//   - 需要细粒度控制缓存选项
	// 示例：
	//   cache.SetItem(ctx, "user:123", &cache.Item{
	//       Value: user,
	//       Expiration: time.Hour,
	//       Tags: []string{"user", "premium"},
	//   })
	SetItem(ctx context.Context, key string, item *Item) error

	// Delete 删除指定键的缓存
	// 参数：
	//   ctx: 上下文，可用于传递超时和取消信号
	//   key: 要删除的缓存键名
	// 返回：
	//   error: 删除过程中遇到的错误，如果删除成功则为nil
	// 行为：
	//   - 如果键不存在，通常不会报错
	//   - 操作是幂等的，多次删除同一键是安全的
	Delete(ctx context.Context, key string) error

	// DeleteByPattern 通过模式匹配删除缓存
	// 参数：
	//   ctx: 上下文，可用于传递超时和取消信号
	//   pattern: 匹配模式，格式取决于具体实现（如Redis使用glob风格模式）
	// 返回：
	//   error: 删除过程中遇到的错误，如果删除成功则为nil
	// 注意：
	//   某些实现可能不支持此功能，会返回不支持的错误
	// 性能：
	//   - 此操作可能会扫描大量键，在大型缓存中可能很慢
	//   - 建议在非关键路径中使用
	// 示例模式：
	//   - "user:*" - 所有以"user:"开头的键
	//   - "session:[0-9]*" - 所有以"session:"开头后跟数字的键
	DeleteByPattern(ctx context.Context, pattern string) error

	// DeleteByTag 删除带特定标签的所有缓存
	// 参数：
	//   ctx: 上下文，可用于传递超时和取消信号
	//   tag: 标签名称
	// 返回：
	//   error: 删除过程中遇到的错误，如果删除成功则为nil
	// 使用场景：
	//   - 批量失效相关缓存项
	//   - 基于业务事件清理缓存组
	// 示例：
	//   - 用户更新时：cache.DeleteByTag(ctx, "user:123")
	//   - 产品更新时：cache.DeleteByTag(ctx, "product")
	DeleteByTag(ctx context.Context, tag string) error

	// Exists 检查键是否存在
	// 参数：
	//   ctx: 上下文，可用于传递超时和取消信号
	//   key: 要检查的缓存键名
	// 返回：
	//   bool: 键是否存在，true表示存在，false表示不存在
	//   error: 检查过程中遇到的错误，如果检查成功则为nil
	// 注意：
	//   - 已过期的项会被视为不存在
	//   - 此方法通常比Get更高效，因为不需要传输值
	Exists(ctx context.Context, key string) (bool, error)

	// Increment 增加数值类型缓存的值
	// 参数：
	//   ctx: 上下文，可用于传递超时和取消信号
	//   key: 缓存键名
	//   value: 要增加的数值
	// 返回：
	//   int64: 增加后的新值
	//   error: 操作过程中遇到的错误，如果操作成功则为nil
	// 注意：
	//   - 如果键不存在，通常会创建并设置为初始值
	//   - 如果现有值不是数值类型，通常会返回错误
	// 使用场景：
	//   - 计数器（如访问次数、限流器）
	//   - 原子增量操作
	Increment(ctx context.Context, key string, value int64) (int64, error)

	// Decrement 减少数值类型缓存的值
	// 参数：
	//   ctx: 上下文，可用于传递超时和取消信号
	//   key: 缓存键名
	//   value: 要减少的数值
	// 返回：
	//   int64: 减少后的新值
	//   error: 操作过程中遇到的错误，如果操作成功则为nil
	// 注意：
	//   - 如果键不存在，通常会创建并设置为初始值的负数
	//   - 此方法通常是对Increment的简单包装
	Decrement(ctx context.Context, key string, value int64) (int64, error)

	// Flush 清空所有缓存
	// 参数：
	//   ctx: 上下文，可用于传递超时和取消信号
	// 返回：
	//   error: 清空过程中遇到的错误，如果清空成功则为nil
	// 警告：
	//   此操作会删除所有缓存数据，请谨慎使用
	// 使用场景：
	//   - 测试环境重置
	//   - 紧急情况下清空缓存
	//   - 缓存迁移前的准备
	Flush(ctx context.Context) error

	// Close 关闭缓存连接并释放资源
	// 返回：
	//   error: 关闭过程中遇到的错误，如果关闭成功则为nil
	// 注意：
	//   - 应用退出前应该调用此方法以确保资源正确释放
	//   - 关闭后的缓存实例不应再被使用
	// 资源释放：
	//   - 内存缓存：停止清理协程
	//   - Redis缓存：关闭连接池
	//   - 多级缓存：关闭所有级别缓存
	Close() error
}

// Options 缓存配置选项
// 用于配置缓存实现的行为和性能特性
// 不同的缓存实现可能会使用不同的选项子集
type Options struct {
	// DefaultTTL 默认缓存生存时间
	// 如果为0则使用实现特定的默认值
	// 如果为负数则永不过期
	// 此值用于Set方法的ttl参数为0时
	DefaultTTL time.Duration

	// CleanupInterval 过期项清理间隔
	// 主要用于内存缓存实现，定期清理过期的缓存项
	// 如果为0则使用实现特定的默认值
	// 较小的间隔会更及时清理过期项，但会增加CPU开销
	CleanupInterval time.Duration
}

// DefaultOptions 返回默认缓存选项
// 返回：
//
//	Options: 默认配置的Options实例
//	  - DefaultTTL: 1小时
//	  - CleanupInterval: 10分钟
//
// 这些默认值适用于大多数场景，提供了合理的性能和资源使用平衡
func DefaultOptions() Options {
	return Options{
		DefaultTTL:      time.Hour,
		CleanupInterval: 10 * time.Minute,
	}
}
