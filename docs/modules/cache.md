# Cache 模块

Cache 模块是 Go-Snap 框架的高性能缓存组件，提供统一的缓存接口，支持多种缓存后端包括内存缓存、Redis、多级缓存等，为应用提供快速的数据存取能力。

## 概述

Cache 模块设计为一个可插拔的缓存系统，通过统一的接口支持不同的缓存实现。它提供了丰富的缓存操作、自动序列化、TTL 管理、缓存统计等功能，帮助开发者构建高性能的应用系统。

### 核心特性

- ✅ **统一接口** - 提供一致的缓存操作接口
- ✅ **多种后端** - 支持内存、Redis、多级缓存等
- ✅ **自动序列化** - 支持复杂对象的自动序列化/反序列化
- ✅ **TTL 支持** - 灵活的过期时间设置
- ✅ **缓存统计** - 提供命中率、操作计数等统计信息
- ✅ **并发安全** - 所有操作都是并发安全的
- ✅ **配置驱动** - 通过配置文件灵活配置缓存行为
- ✅ **事件监听** - 支持缓存事件的监听和处理
- ✅ **批量操作** - 支持批量读写操作提高性能

## 快速开始

### 基础用法

```go
package main

import (
    "context"
    "time"
    "github.com/guanzhenxing/go-snap/boot"
)

func main() {
    // 启动应用（会自动配置缓存组件）
    app := boot.NewBoot()
    application, _ := app.Initialize()
    
    // 获取缓存组件
    if cacheComp, found := application.GetComponent("cache"); found {
        if cc, ok := cacheComp.(*boot.CacheComponent); ok {
            cache := cc.GetCache()
            ctx := context.Background()
            
            // 设置缓存
            err := cache.Set(ctx, "user:123", "John Doe", time.Hour)
            if err != nil {
                panic(err)
            }
            
            // 获取缓存
            value, found := cache.Get(ctx, "user:123")
            if found {
                fmt.Printf("用户: %s\n", value)
            }
            
            // 删除缓存
            cache.Delete(ctx, "user:123")
        }
    }
}
```

### 直接使用Cache

```go
import "github.com/guanzhenxing/go-snap/cache"

// 创建内存缓存
memCache := cache.NewMemoryCache()

// 创建Redis缓存
redisCache := cache.NewRedisCache(&cache.RedisOptions{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

// 使用缓存
ctx := context.Background()
memCache.Set(ctx, "key", "value", time.Minute*10)
value, found := memCache.Get(ctx, "key")
```

## 配置

### 配置文件示例

```yaml
# 缓存配置
cache:
  enabled: true                    # 是否启用缓存
  type: "memory"                   # 缓存类型: memory, redis, multi
  prefix: "myapp:"                 # 缓存键前缀
  
  # 内存缓存配置
  memory:
    max_entries: 10000             # 最大条目数
    cleanup_interval: "5m"         # 清理间隔
    
  # Redis缓存配置
  redis:
    addr: "localhost:6379"         # Redis地址
    password: ""                   # Redis密码
    db: 0                          # 数据库编号
    pool_size: 10                  # 连接池大小
    min_idle_conns: 5              # 最小空闲连接数
    dial_timeout: "5s"             # 连接超时
    read_timeout: "3s"             # 读取超时
    write_timeout: "3s"            # 写入超时
    
  # 多级缓存配置
  multi:
    l1: "memory"                   # 一级缓存类型
    l2: "redis"                    # 二级缓存类型
    l1_ttl: "5m"                   # 一级缓存TTL
    l2_ttl: "1h"                   # 二级缓存TTL
    
  # 序列化配置
  serializer: "json"               # 序列化方式: json, gob, msgpack
  
  # 统计配置
  stats:
    enabled: true                  # 是否启用统计
    interval: "1m"                 # 统计报告间隔
```

### 环境变量配置

```bash
# 通过环境变量覆盖配置
export CACHE_TYPE=redis
export CACHE_REDIS_ADDR=redis-server:6379
export CACHE_REDIS_PASSWORD=secret
export CACHE_PREFIX=prod:
```

## API 参考

### 缓存接口

```go
type Cache interface {
    // 基础操作
    Get(ctx context.Context, key string) (interface{}, bool)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) bool
    Clear(ctx context.Context) error
    
    // 批量操作
    GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)
    SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
    DeleteMulti(ctx context.Context, keys []string) error
    
    // 原子操作
    Increment(ctx context.Context, key string, delta int64) (int64, error)
    Decrement(ctx context.Context, key string, delta int64) (int64, error)
    
    // TTL管理
    Expire(ctx context.Context, key string, ttl time.Duration) error
    TTL(ctx context.Context, key string) (time.Duration, error)
    
    // 统计信息
    Stats() CacheStats
    
    // 关闭
    Close() error
}
```

### 基础操作

```go
ctx := context.Background()

// 设置缓存
err := cache.Set(ctx, "user:123", user, time.Hour)

// 获取缓存
value, found := cache.Get(ctx, "user:123")
if found {
    user := value.(*User)
    fmt.Printf("用户: %+v\n", user)
}

// 检查是否存在
if cache.Exists(ctx, "user:123") {
    fmt.Println("缓存存在")
}

// 删除缓存
err = cache.Delete(ctx, "user:123")

// 清空所有缓存
err = cache.Clear(ctx)
```

### 批量操作

```go
// 批量获取
keys := []string{"user:1", "user:2", "user:3"}
results, err := cache.GetMulti(ctx, keys)
for key, value := range results {
    fmt.Printf("%s: %v\n", key, value)
}

// 批量设置
items := map[string]interface{}{
    "user:1": user1,
    "user:2": user2,
    "user:3": user3,
}
err = cache.SetMulti(ctx, items, time.Hour)

// 批量删除
err = cache.DeleteMulti(ctx, keys)
```

### 原子操作

```go
// 递增
newValue, err := cache.Increment(ctx, "counter", 1)
fmt.Printf("新值: %d\n", newValue)

// 递减
newValue, err = cache.Decrement(ctx, "counter", 1)
fmt.Printf("新值: %d\n", newValue)
```

### TTL 管理

```go
// 设置过期时间
err := cache.Expire(ctx, "temp_data", time.Minute*30)

// 获取剩余时间
ttl, err := cache.TTL(ctx, "temp_data")
fmt.Printf("剩余时间: %v\n", ttl)
```

### 缓存统计

```go
// 获取统计信息
stats := cache.Stats()
fmt.Printf("命中次数: %d\n", stats.Hits)
fmt.Printf("丢失次数: %d\n", stats.Misses)
fmt.Printf("命中率: %.2f%%\n", stats.HitRate*100)
fmt.Printf("总操作数: %d\n", stats.Operations)
```

## 缓存类型

### 内存缓存

内存缓存适用于单实例应用，提供最快的访问速度。

```go
// 创建内存缓存
memCache := cache.NewMemoryCache(&cache.MemoryOptions{
    MaxEntries:      10000,              // 最大条目数
    CleanupInterval: time.Minute * 5,    // 清理间隔
    OnEvicted: func(key string, value interface{}) {
        fmt.Printf("缓存被驱逐: %s\n", key)
    },
})
```

**特性**:
- 最快的访问速度
- 支持LRU驱逐策略
- 内存占用可控
- 单进程内共享

**适用场景**:
- 单实例应用
- 频繁访问的小数据
- 临时缓存

### Redis缓存

Redis缓存适用于分布式应用，支持持久化和高可用。

```go
// 创建Redis缓存
redisCache := cache.NewRedisCache(&cache.RedisOptions{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     10,
    MinIdleConns: 5,
    DialTimeout:  time.Second * 5,
    ReadTimeout:  time.Second * 3,
    WriteTimeout: time.Second * 3,
})
```

**特性**:
- 分布式共享
- 数据持久化
- 高可用支持
- 丰富的数据类型

**适用场景**:
- 分布式应用
- 会话存储
- 需要持久化的缓存
- 大容量缓存

### 多级缓存

多级缓存结合了内存缓存和Redis缓存的优势，提供最佳的性能和可扩展性。

```go
// 创建多级缓存
multiCache := cache.NewMultiCache(&cache.MultiOptions{
    L1: memCache,           // 一级缓存（内存）
    L2: redisCache,         // 二级缓存（Redis）
    L1TTL: time.Minute * 5, // 一级缓存TTL
    L2TTL: time.Hour,       // 二级缓存TTL
})
```

**工作原理**:
1. 读取时先查询L1缓存（内存）
2. L1未命中则查询L2缓存（Redis）
3. L2命中则同时更新L1缓存
4. 写入时同时更新L1和L2缓存

**适用场景**:
- 高并发应用
- 热点数据访问
- 需要最佳性能的场景

## 序列化

Cache 模块支持多种序列化方式来存储复杂对象。

### JSON 序列化

```go
// 使用JSON序列化（默认）
cache := cache.NewRedisCache(options, cache.WithSerializer(cache.JSONSerializer))

// 存储复杂对象
user := &User{
    ID:   123,
    Name: "John Doe",
    Email: "john@example.com",
}
cache.Set(ctx, "user:123", user, time.Hour)

// 读取时需要类型断言
value, found := cache.Get(ctx, "user:123")
if found {
    user := value.(*User)
}
```

### Gob 序列化

```go
// 使用Gob序列化（Go特有，效率高）
cache := cache.NewRedisCache(options, cache.WithSerializer(cache.GobSerializer))
```

### MessagePack 序列化

```go
// 使用MessagePack序列化（跨语言，紧凑）
cache := cache.NewRedisCache(options, cache.WithSerializer(cache.MsgpackSerializer))
```

### 自定义序列化

```go
// 实现自定义序列化器
type CustomSerializer struct{}

func (s *CustomSerializer) Serialize(value interface{}) ([]byte, error) {
    // 自定义序列化逻辑
    return nil, nil
}

func (s *CustomSerializer) Deserialize(data []byte) (interface{}, error) {
    // 自定义反序列化逻辑
    return nil, nil
}

// 使用自定义序列化器
cache := cache.NewRedisCache(options, cache.WithSerializer(&CustomSerializer{}))
```

## 高级功能

### 缓存事件

```go
// 监听缓存事件
cache.OnHit(func(key string) {
    fmt.Printf("缓存命中: %s\n", key)
})

cache.OnMiss(func(key string) {
    fmt.Printf("缓存丢失: %s\n", key)
})

cache.OnSet(func(key string, value interface{}) {
    fmt.Printf("缓存设置: %s\n", key)
})

cache.OnDelete(func(key string) {
    fmt.Printf("缓存删除: %s\n", key)
})
```

### 缓存预热

```go
// 应用启动时预热缓存
func WarmupCache(cache cache.Cache) error {
    ctx := context.Background()
    
    // 预加载热点数据
    hotUsers := []int{1, 2, 3, 4, 5}
    for _, userID := range hotUsers {
        user, err := userService.GetUser(userID)
        if err != nil {
            continue
        }
        
        key := fmt.Sprintf("user:%d", userID)
        cache.Set(ctx, key, user, time.Hour)
    }
    
    return nil
}
```

### 缓存穿透保护

```go
// 使用空值缓存防止缓存穿透
func GetUserWithProtection(cache cache.Cache, userID int) (*User, error) {
    ctx := context.Background()
    key := fmt.Sprintf("user:%d", userID)
    
    // 先查缓存
    value, found := cache.Get(ctx, key)
    if found {
        if value == nil {
            return nil, ErrUserNotFound // 空值缓存
        }
        return value.(*User), nil
    }
    
    // 查数据库
    user, err := userService.GetUser(userID)
    if err != nil {
        if err == ErrUserNotFound {
            // 缓存空值，防止穿透
            cache.Set(ctx, key, nil, time.Minute*5)
        }
        return nil, err
    }
    
    // 缓存正常值
    cache.Set(ctx, key, user, time.Hour)
    return user, nil
}
```

### 缓存雪崩保护

```go
// 使用随机TTL防止缓存雪崩
func SetWithRandomTTL(cache cache.Cache, key string, value interface{}, baseTTL time.Duration) error {
    ctx := context.Background()
    
    // 在基础TTL上增加随机时间
    randomTTL := baseTTL + time.Duration(rand.Intn(300))*time.Second
    return cache.Set(ctx, key, value, randomTTL)
}
```

### 分布式锁

```go
// 使用Redis实现分布式锁
func WithDistributedLock(cache cache.Cache, key string, ttl time.Duration, fn func() error) error {
    ctx := context.Background()
    lockKey := "lock:" + key
    
    // 尝试获取锁
    success := cache.SetNX(ctx, lockKey, "locked", ttl)
    if !success {
        return ErrLockAcquisitionFailed
    }
    
    defer func() {
        // 释放锁
        cache.Delete(ctx, lockKey)
    }()
    
    // 执行业务逻辑
    return fn()
}
```

## 最佳实践

### 1. 缓存键设计

```go
// ✅ 好的做法：使用有意义的前缀和分隔符
const (
    UserCacheKeyPrefix = "user:"
    OrderCacheKeyPrefix = "order:"
    SessionCacheKeyPrefix = "session:"
)

func userCacheKey(userID int) string {
    return fmt.Sprintf("%s%d", UserCacheKeyPrefix, userID)
}

func orderCacheKey(orderID string) string {
    return fmt.Sprintf("%s%s", OrderCacheKeyPrefix, orderID)
}

// ❌ 避免的做法：使用不规范的键名
cache.Set(ctx, "u123", user, time.Hour)        // 不清晰
cache.Set(ctx, "user_data_123", user, time.Hour) // 不一致
```

### 2. TTL 设置策略

```go
// ✅ 根据数据特性设置合适的TTL
const (
    UserProfileTTL = time.Hour * 24      // 用户资料：1天
    SessionTTL     = time.Minute * 30    // 会话：30分钟
    ProductTTL     = time.Hour * 6       // 商品信息：6小时
    ConfigTTL      = time.Hour * 12      // 配置信息：12小时
)

// ✅ 为不同环境设置不同的TTL
func getTTL(env string, baseTTL time.Duration) time.Duration {
    switch env {
    case "development":
        return baseTTL / 10  // 开发环境短TTL，便于测试
    case "production":
        return baseTTL       // 生产环境正常TTL
    default:
        return baseTTL
    }
}
```

### 3. 错误处理

```go
// ✅ 优雅处理缓存错误
func GetUserWithFallback(cache cache.Cache, userID int) (*User, error) {
    ctx := context.Background()
    key := userCacheKey(userID)
    
    // 尝试从缓存获取
    if value, found := cache.Get(ctx, key); found {
        if user, ok := value.(*User); ok {
            return user, nil
        }
    }
    
    // 缓存未命中或错误，从数据库获取
    user, err := userService.GetUser(userID)
    if err != nil {
        return nil, err
    }
    
    // 异步更新缓存，不影响主流程
    go func() {
        if err := cache.Set(ctx, key, user, UserProfileTTL); err != nil {
            log.Printf("缓存更新失败: %v", err)
        }
    }()
    
    return user, nil
}
```

### 4. 缓存一致性

```go
// ✅ 写操作时同步更新缓存
func UpdateUser(cache cache.Cache, user *User) error {
    ctx := context.Background()
    
    // 更新数据库
    if err := userService.UpdateUser(user); err != nil {
        return err
    }
    
    // 更新缓存
    key := userCacheKey(user.ID)
    if err := cache.Set(ctx, key, user, UserProfileTTL); err != nil {
        log.Printf("缓存更新失败: %v", err)
        // 不返回错误，因为数据库已更新成功
    }
    
    return nil
}

// ✅ 删除操作时清理缓存
func DeleteUser(cache cache.Cache, userID int) error {
    ctx := context.Background()
    
    // 删除数据库记录
    if err := userService.DeleteUser(userID); err != nil {
        return err
    }
    
    // 清理缓存
    key := userCacheKey(userID)
    if err := cache.Delete(ctx, key); err != nil {
        log.Printf("缓存删除失败: %v", err)
    }
    
    return nil
}
```

### 5. 监控和告警

```go
// ✅ 定期监控缓存状态
func MonitorCache(cache cache.Cache) {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := cache.Stats()
        
        // 监控命中率
        if stats.HitRate < 0.8 {
            log.Printf("警告: 缓存命中率过低 %.2f%%", stats.HitRate*100)
        }
        
        // 监控错误率
        errorRate := float64(stats.Errors) / float64(stats.Operations)
        if errorRate > 0.01 {
            log.Printf("警告: 缓存错误率过高 %.2f%%", errorRate*100)
        }
    }
}
```

## 性能优化

### 1. 连接池配置

```yaml
cache:
  redis:
    pool_size: 20          # 连接池大小
    min_idle_conns: 5      # 最小空闲连接
    max_conn_age: "1h"     # 连接最大生存时间
    pool_timeout: "4s"     # 获取连接超时
    idle_timeout: "5m"     # 空闲连接超时
```

### 2. 批量操作

```go
// ✅ 使用批量操作减少网络开销
func GetMultipleUsers(cache cache.Cache, userIDs []int) (map[int]*User, error) {
    ctx := context.Background()
    
    // 构建缓存键
    keys := make([]string, len(userIDs))
    keyToID := make(map[string]int)
    for i, id := range userIDs {
        key := userCacheKey(id)
        keys[i] = key
        keyToID[key] = id
    }
    
    // 批量获取
    results, err := cache.GetMulti(ctx, keys)
    if err != nil {
        return nil, err
    }
    
    // 处理结果
    users := make(map[int]*User)
    for key, value := range results {
        if user, ok := value.(*User); ok {
            userID := keyToID[key]
            users[userID] = user
        }
    }
    
    return users, nil
}
```

### 3. 异步操作

```go
// ✅ 异步更新缓存，不阻塞主流程
func AsyncCacheUpdate(cache cache.Cache, key string, value interface{}, ttl time.Duration) {
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
        defer cancel()
        
        if err := cache.Set(ctx, key, value, ttl); err != nil {
            log.Printf("异步缓存更新失败: %v", err)
        }
    }()
}
```

## 故障排除

### 常见问题

#### 1. Redis 连接失败

**错误**: `dial tcp: connect: connection refused`

**解决方案**:
- 检查Redis服务是否运行
- 验证连接地址和端口
- 检查网络连通性
- 确认防火墙设置

#### 2. 序列化失败

**错误**: `json: cannot unmarshal`

**解决方案**:
- 检查数据类型是否匹配
- 确认序列化器配置
- 验证数据格式

#### 3. 内存缓存溢出

**错误**: 内存使用过高

**解决方案**:
- 调整MaxEntries参数
- 缩短TTL时间
- 启用LRU驱逐策略

### 调试技巧

```go
// 启用缓存调试日志
cache := cache.NewRedisCache(options, cache.WithDebug(true))

// 监控缓存操作
cache.OnOperation(func(op string, key string, duration time.Duration) {
    log.Printf("缓存操作: %s %s 耗时: %v", op, key, duration)
})

// 检查缓存状态
stats := cache.Stats()
log.Printf("缓存统计: %+v", stats)
```

## 参考资料

- [Redis 官方文档](https://redis.io/documentation)
- [Go-Snap Boot 模块](boot.md)
- [Go-Snap 配置模块](config.md)
- [Go-Snap 架构设计](../architecture.md) 