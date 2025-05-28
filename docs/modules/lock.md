# Lock 模块

Lock 模块是 Go-Snap 框架的分布式锁组件，提供可靠的分布式锁实现，支持 Redis 后端、锁超时机制、自动续租等功能，帮助开发者在分布式环境下保证数据一致性和操作原子性。

## 概述

Lock 模块提供了一个统一的分布式锁接口，当前主要基于 Redis 实现。它解决了在分布式系统中多个进程或服务同时访问共享资源时的并发控制问题，确保关键代码段在同一时刻只有一个进程能够执行。

### 核心特性

- ✅ **分布式锁** - 支持跨进程、跨服务的锁机制
- ✅ **Redis 后端** - 基于 Redis 的高性能锁实现
- ✅ **可重入锁** - 同一线程可以多次获取同一把锁
- ✅ **锁超时** - 自动锁超时机制，防止死锁
- ✅ **自动续租** - 自动延长锁的有效期
- ✅ **阻塞/非阻塞** - 支持阻塞和非阻塞的锁获取
- ✅ **锁监听** - 支持锁事件监听和回调
- ✅ **批量锁** - 支持同时获取多个锁
- ✅ **死锁检测** - 检测和处理死锁情况
- ✅ **高可用** - 支持 Redis 集群和哨兵模式

## 快速开始

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/guanzhenxing/go-snap/boot"
)

func main() {
    app := boot.NewBoot().SetConfigPath("configs")
    application, _ := app.Initialize()
    
    // 获取锁组件
    if lockComp, found := application.GetComponent("lock"); found {
        if lc, ok := lockComp.(*boot.LockComponent); ok {
            lockManager := lc.GetLockManager()
            ctx := context.Background()
            
            // 获取分布式锁
            lock, err := lockManager.NewLock("order:process", 30*time.Second)
            if err != nil {
                panic(err)
            }
            
            // 尝试获取锁
            acquired, err := lock.TryLock(ctx)
            if err != nil {
                panic(err)
            }
            
            if acquired {
                fmt.Println("成功获取锁，开始处理订单...")
                
                // 执行关键业务逻辑
                time.Sleep(time.Second * 5)
                
                // 释放锁
                lock.Unlock(ctx)
                fmt.Println("锁已释放")
            } else {
                fmt.Println("获取锁失败，可能有其他进程在处理")
            }
        }
    }
}
```

### 配置文件

```yaml
# configs/application.yaml
lock:
  enabled: true
  backend: "redis"
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
    pool_size: 10
  default_ttl: "30s"
  retry_interval: "100ms"
  max_retries: 3
```

## 配置选项

### 基础配置

```yaml
lock:
  enabled: true                    # 是否启用锁模块
  backend: "redis"                 # 锁后端: redis
  default_ttl: "30s"              # 默认锁超时时间
  retry_interval: "100ms"         # 重试间隔
  max_retries: 3                  # 最大重试次数
  
  # Redis 配置
  redis:
    addr: "localhost:6379"        # Redis 地址
    password: ""                  # Redis 密码
    db: 0                         # 数据库编号
    pool_size: 10                 # 连接池大小
    min_idle_conns: 5             # 最小空闲连接数
    dial_timeout: "5s"            # 连接超时
    read_timeout: "3s"            # 读取超时
    write_timeout: "3s"           # 写入超时
    
    # 集群配置
    cluster:
      enabled: false              # 是否启用集群模式
      addrs:                      # 集群节点地址
        - "localhost:7000"
        - "localhost:7001"
        - "localhost:7002"
    
    # 哨兵配置
    sentinel:
      enabled: false              # 是否启用哨兵模式
      master_name: "mymaster"     # 主节点名称
      addrs:                      # 哨兵节点地址
        - "localhost:26379"
        - "localhost:26380"
        - "localhost:26381"
```

## API 参考

### 锁管理器

#### LockManager

锁管理器接口。

```go
type LockManager interface {
    // 创建新锁
    NewLock(key string, ttl time.Duration) (Lock, error)
    
    // 创建可重入锁
    NewReentrantLock(key string, ttl time.Duration) (ReentrantLock, error)
    
    // 批量创建锁
    NewMultiLock(keys []string, ttl time.Duration) (MultiLock, error)
    
    // 获取锁信息
    GetLockInfo(key string) (*LockInfo, error)
    
    // 释放所有锁
    ReleaseAllLocks() error
    
    // 关闭管理器
    Close() error
}
```

### 锁接口

#### Lock

基础锁接口。

```go
type Lock interface {
    // 获取锁（阻塞）
    Lock(ctx context.Context) error
    
    // 尝试获取锁（非阻塞）
    TryLock(ctx context.Context) (bool, error)
    
    // 带超时的获取锁
    TryLockWithTimeout(ctx context.Context, timeout time.Duration) (bool, error)
    
    // 释放锁
    Unlock(ctx context.Context) error
    
    // 延长锁有效期
    Extend(ctx context.Context, ttl time.Duration) error
    
    // 获取锁信息
    Info() *LockInfo
    
    // 检查锁是否被当前实例持有
    IsHeldByCurrentThread() bool
}
```

#### ReentrantLock

可重入锁接口。

```go
type ReentrantLock interface {
    Lock
    
    // 获取重入次数
    GetHoldCount() int
    
    // 检查是否被锁定
    IsLocked() bool
    
    // 检查是否被当前线程锁定
    IsHeldByCurrentThread() bool
}
```

#### MultiLock

多重锁接口。

```go
type MultiLock interface {
    // 获取所有锁
    LockAll(ctx context.Context) error
    
    // 尝试获取所有锁
    TryLockAll(ctx context.Context) (bool, error)
    
    // 释放所有锁
    UnlockAll(ctx context.Context) error
    
    // 获取已获取的锁数量
    GetAcquiredCount() int
    
    // 获取锁状态
    GetLockStates() map[string]bool
}
```

### 锁信息

```go
type LockInfo struct {
    Key        string        `json:"key"`         // 锁键
    Owner      string        `json:"owner"`       // 锁持有者
    TTL        time.Duration `json:"ttl"`         // 剩余时间
    CreateTime time.Time     `json:"create_time"` // 创建时间
    IsLocked   bool         `json:"is_locked"`   // 是否被锁定
    HoldCount  int          `json:"hold_count"`  // 重入次数
}
```

## 使用示例

### 1. 基础锁使用

```go
func processOrder(lockManager lock.LockManager, orderID string) error {
    ctx := context.Background()
    lockKey := fmt.Sprintf("order:process:%s", orderID)
    
    // 创建锁
    orderLock, err := lockManager.NewLock(lockKey, time.Minute)
    if err != nil {
        return err
    }
    
    // 获取锁
    if err := orderLock.Lock(ctx); err != nil {
        return fmt.Errorf("获取锁失败: %v", err)
    }
    defer orderLock.Unlock(ctx)
    
    // 执行订单处理逻辑
    fmt.Printf("开始处理订单: %s\n", orderID)
    time.Sleep(time.Second * 10) // 模拟处理时间
    fmt.Printf("订单处理完成: %s\n", orderID)
    
    return nil
}
```

### 2. 非阻塞锁

```go
func tryProcessOrder(lockManager lock.LockManager, orderID string) error {
    ctx := context.Background()
    lockKey := fmt.Sprintf("order:process:%s", orderID)
    
    orderLock, err := lockManager.NewLock(lockKey, time.Minute)
    if err != nil {
        return err
    }
    
    // 尝试获取锁（非阻塞）
    acquired, err := orderLock.TryLock(ctx)
    if err != nil {
        return err
    }
    
    if !acquired {
        return fmt.Errorf("订单 %s 正在被其他进程处理", orderID)
    }
    defer orderLock.Unlock(ctx)
    
    // 处理订单
    return processOrderLogic(orderID)
}
```

### 3. 带超时的锁

```go
func processWithTimeout(lockManager lock.LockManager, resourceID string) error {
    ctx := context.Background()
    lockKey := fmt.Sprintf("resource:%s", resourceID)
    
    resourceLock, err := lockManager.NewLock(lockKey, time.Minute)
    if err != nil {
        return err
    }
    
    // 尝试在5秒内获取锁
    acquired, err := resourceLock.TryLockWithTimeout(ctx, 5*time.Second)
    if err != nil {
        return err
    }
    
    if !acquired {
        return fmt.Errorf("获取锁超时")
    }
    defer resourceLock.Unlock(ctx)
    
    // 执行资源处理逻辑
    return processResource(resourceID)
}
```

### 4. 可重入锁

```go
func processWithReentrantLock(lockManager lock.LockManager, taskID string) error {
    ctx := context.Background()
    lockKey := fmt.Sprintf("task:%s", taskID)
    
    // 创建可重入锁
    reentrantLock, err := lockManager.NewReentrantLock(lockKey, time.Minute)
    if err != nil {
        return err
    }
    
    // 第一次获取锁
    if err := reentrantLock.Lock(ctx); err != nil {
        return err
    }
    defer reentrantLock.Unlock(ctx)
    
    fmt.Printf("第一次获取锁，重入次数: %d\n", reentrantLock.GetHoldCount())
    
    // 在同一线程中再次获取锁（可重入）
    if err := reentrantLock.Lock(ctx); err != nil {
        return err
    }
    defer reentrantLock.Unlock(ctx)
    
    fmt.Printf("第二次获取锁，重入次数: %d\n", reentrantLock.GetHoldCount())
    
    return processTask(taskID)
}
```

### 5. 多重锁

```go
func processMultipleResources(lockManager lock.LockManager, resourceIDs []string) error {
    ctx := context.Background()
    
    // 生成锁键
    lockKeys := make([]string, len(resourceIDs))
    for i, id := range resourceIDs {
        lockKeys[i] = fmt.Sprintf("resource:%s", id)
    }
    
    // 创建多重锁
    multiLock, err := lockManager.NewMultiLock(lockKeys, time.Minute)
    if err != nil {
        return err
    }
    
    // 尝试获取所有锁
    acquired, err := multiLock.TryLockAll(ctx)
    if err != nil {
        return err
    }
    
    if !acquired {
        return fmt.Errorf("无法获取所有资源锁")
    }
    defer multiLock.UnlockAll(ctx)
    
    fmt.Printf("成功获取 %d 个锁\n", multiLock.GetAcquiredCount())
    
    // 处理所有资源
    return processAllResources(resourceIDs)
}
```

### 6. 锁自动续租

```go
func longRunningTask(lockManager lock.LockManager, taskID string) error {
    ctx := context.Background()
    lockKey := fmt.Sprintf("task:%s", taskID)
    
    taskLock, err := lockManager.NewLock(lockKey, 30*time.Second)
    if err != nil {
        return err
    }
    
    if err := taskLock.Lock(ctx); err != nil {
        return err
    }
    defer taskLock.Unlock(ctx)
    
    // 启动自动续租
    stopRenewal := make(chan bool)
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                if err := taskLock.Extend(ctx, 30*time.Second); err != nil {
                    log.Printf("续租失败: %v", err)
                }
            case <-stopRenewal:
                return
            }
        }
    }()
    
    // 执行长时间任务
    err = performLongRunningTask(taskID)
    
    // 停止续租
    close(stopRenewal)
    
    return err
}
```

## 高级功能

### 1. 锁事件监听

```go
// 锁事件类型
type LockEvent struct {
    Type      LockEventType `json:"type"`
    Key       string        `json:"key"`
    Owner     string        `json:"owner"`
    Timestamp time.Time     `json:"timestamp"`
    Error     error         `json:"error,omitempty"`
}

type LockEventType string

const (
    LockEventAcquired LockEventType = "acquired"
    LockEventReleased LockEventType = "released"
    LockEventExpired  LockEventType = "expired"
    LockEventFailed   LockEventType = "failed"
)

// 注册事件监听器
func setupLockEventListener(lockManager lock.LockManager) {
    lockManager.OnEvent(func(event *LockEvent) {
        switch event.Type {
        case LockEventAcquired:
            log.Printf("锁已获取: %s by %s", event.Key, event.Owner)
        case LockEventReleased:
            log.Printf("锁已释放: %s by %s", event.Key, event.Owner)
        case LockEventExpired:
            log.Printf("锁已过期: %s", event.Key)
        case LockEventFailed:
            log.Printf("锁操作失败: %s, 错误: %v", event.Key, event.Error)
        }
    })
}
```

### 2. 死锁检测

```go
// 死锁检测器
type DeadlockDetector struct {
    lockManager lock.LockManager
    timeout     time.Duration
}

func NewDeadlockDetector(lockManager lock.LockManager) *DeadlockDetector {
    return &DeadlockDetector{
        lockManager: lockManager,
        timeout:     time.Minute,
    }
}

func (d *DeadlockDetector) DetectDeadlock() error {
    // 获取所有锁信息
    allLocks, err := d.lockManager.GetAllLocks()
    if err != nil {
        return err
    }
    
    // 构建等待图
    waitGraph := d.buildWaitGraph(allLocks)
    
    // 检测循环依赖
    if cycle := d.detectCycle(waitGraph); cycle != nil {
        return fmt.Errorf("检测到死锁: %v", cycle)
    }
    
    return nil
}
```

### 3. 锁性能监控

```go
// 锁性能指标
type LockMetrics struct {
    TotalAcquired   int64         `json:"total_acquired"`
    TotalReleased   int64         `json:"total_released"`
    TotalFailed     int64         `json:"total_failed"`
    AverageHoldTime time.Duration `json:"average_hold_time"`
    MaxHoldTime     time.Duration `json:"max_hold_time"`
    ActiveLocks     int64         `json:"active_locks"`
}

func (lm *LockManager) GetMetrics() *LockMetrics {
    return &LockMetrics{
        TotalAcquired:   lm.stats.totalAcquired,
        TotalReleased:   lm.stats.totalReleased,
        TotalFailed:     lm.stats.totalFailed,
        AverageHoldTime: lm.stats.calculateAverageHoldTime(),
        MaxHoldTime:     lm.stats.maxHoldTime,
        ActiveLocks:     lm.stats.activeLocks,
    }
}
```

## 集成示例

### 与 Web 服务集成

```go
// 锁中间件
func LockMiddleware(lockManager lock.LockManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        lockKey := c.GetHeader("X-Lock-Key")
        if lockKey == "" {
            c.Next()
            return
        }
        
        lock, err := lockManager.NewLock(lockKey, time.Minute)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "创建锁失败"})
            c.Abort()
            return
        }
        
        acquired, err := lock.TryLockWithTimeout(c.Request.Context(), 5*time.Second)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "获取锁失败"})
            c.Abort()
            return
        }
        
        if !acquired {
            c.JSON(http.StatusConflict, gin.H{"error": "资源被锁定"})
            c.Abort()
            return
        }
        
        defer lock.Unlock(c.Request.Context())
        c.Next()
    }
}

// 使用锁中间件
func setupRoutes(lockManager lock.LockManager) *gin.Engine {
    r := gin.Default()
    
    api := r.Group("/api")
    api.Use(LockMiddleware(lockManager))
    {
        api.POST("/orders", createOrder)
        api.PUT("/orders/:id", updateOrder)
        api.DELETE("/orders/:id", deleteOrder)
    }
    
    return r
}
```

### 与任务调度集成

```go
// 分布式任务执行器
type DistributedTaskExecutor struct {
    lockManager lock.LockManager
    taskQueue   chan Task
}

type Task struct {
    ID       string
    Handler  func() error
    Timeout  time.Duration
}

func (e *DistributedTaskExecutor) ExecuteTask(task Task) error {
    lockKey := fmt.Sprintf("task:%s", task.ID)
    
    taskLock, err := e.lockManager.NewLock(lockKey, task.Timeout)
    if err != nil {
        return err
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
    defer cancel()
    
    // 尝试获取锁，确保任务只在一个节点执行
    acquired, err := taskLock.TryLock(ctx)
    if err != nil {
        return err
    }
    
    if !acquired {
        return fmt.Errorf("任务 %s 正在其他节点执行", task.ID)
    }
    defer taskLock.Unlock(ctx)
    
    // 执行任务
    return task.Handler()
}

// 定期任务调度
func (e *DistributedTaskExecutor) SchedulePeriodicTask(taskID string, interval time.Duration, handler func() error) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        
        for range ticker.C {
            task := Task{
                ID:      taskID,
                Handler: handler,
                Timeout: interval / 2, // 设置超时为间隔的一半
            }
            
            if err := e.ExecuteTask(task); err != nil {
                log.Printf("执行定期任务失败: %v", err)
            }
        }
    }()
}
```

## 最佳实践

### 1. 锁键命名规范

```go
// ✅ 好的做法：使用有意义的分层键名
const (
    OrderProcessLockPrefix = "order:process:"
    UserSessionLockPrefix  = "user:session:"
    ResourceLockPrefix     = "resource:"
)

func getOrderLockKey(orderID string) string {
    return fmt.Sprintf("%s%s", OrderProcessLockPrefix, orderID)
}

func getUserSessionLockKey(userID string) string {
    return fmt.Sprintf("%s%s", UserSessionLockPrefix, userID)
}

// ❌ 避免的做法：使用不规范的键名
func badLockKey(id string) string {
    return "lock_" + id // 不清晰的前缀
}
```

### 2. 锁超时设置

```go
// ✅ 根据业务场景设置合适的超时时间
const (
    QuickOperationTimeout = 5 * time.Second   // 快速操作
    NormalOperationTimeout = 30 * time.Second // 普通操作
    LongOperationTimeout = 5 * time.Minute    // 长时间操作
)

func processPayment(lockManager lock.LockManager, paymentID string) error {
    lock, _ := lockManager.NewLock(
        "payment:"+paymentID,
        NormalOperationTimeout, // 支付处理通常需要30秒
    )
    // ...
}

// ❌ 避免的做法：使用过长或过短的超时时间
func badTimeout(lockManager lock.LockManager, id string) error {
    lock, _ := lockManager.NewLock(
        "resource:"+id,
        24*time.Hour, // 过长的超时时间
    )
    // ...
}
```

### 3. 错误处理

```go
// ✅ 完善的错误处理
func safeProcess(lockManager lock.LockManager, resourceID string) error {
    ctx := context.Background()
    lock, err := lockManager.NewLock("resource:"+resourceID, time.Minute)
    if err != nil {
        return fmt.Errorf("创建锁失败: %v", err)
    }
    
    acquired, err := lock.TryLockWithTimeout(ctx, 5*time.Second)
    if err != nil {
        return fmt.Errorf("获取锁失败: %v", err)
    }
    
    if !acquired {
        return fmt.Errorf("资源 %s 当前被占用", resourceID)
    }
    
    defer func() {
        if err := lock.Unlock(ctx); err != nil {
            log.Printf("释放锁失败: %v", err)
        }
    }()
    
    // 业务逻辑处理
    return processResource(resourceID)
}

// ❌ 避免的做法：忽略错误
func unsafeProcess(lockManager lock.LockManager, resourceID string) {
    lock, _ := lockManager.NewLock("resource:"+resourceID, time.Minute)
    lock.Lock(context.Background()) // 忽略错误
    
    // 业务处理...
    
    lock.Unlock(context.Background()) // 忽略错误
}
```

### 4. 锁粒度控制

```go
// ✅ 合适的锁粒度
func processUserOrder(lockManager lock.LockManager, userID, orderID string) error {
    // 使用细粒度锁，只锁定特定订单
    orderLock, err := lockManager.NewLock(
        fmt.Sprintf("order:%s", orderID),
        time.Minute,
    )
    if err != nil {
        return err
    }
    
    ctx := context.Background()
    if err := orderLock.Lock(ctx); err != nil {
        return err
    }
    defer orderLock.Unlock(ctx)
    
    return processOrder(orderID)
}

// ❌ 避免的做法：锁粒度过粗
func processUserOrderBad(lockManager lock.LockManager, userID, orderID string) error {
    // 锁定整个用户，影响该用户的所有操作
    userLock, err := lockManager.NewLock(
        fmt.Sprintf("user:%s", userID),
        time.Minute,
    )
    // ...
}
```

### 5. 资源清理

```go
// ✅ 确保资源清理
func processWithCleanup(lockManager lock.LockManager, taskID string) error {
    ctx := context.Background()
    lock, err := lockManager.NewLock("task:"+taskID, time.Minute)
    if err != nil {
        return err
    }
    
    if err := lock.Lock(ctx); err != nil {
        return err
    }
    
    // 使用 defer 确保锁被释放
    defer func() {
        if err := lock.Unlock(ctx); err != nil {
            log.Printf("释放锁失败: %v", err)
        }
    }()
    
    // 设置超时上下文
    taskCtx, cancel := context.WithTimeout(ctx, 50*time.Second)
    defer cancel()
    
    return performTask(taskCtx, taskID)
}
```

## 性能优化

### 1. 连接池配置

```yaml
lock:
  redis:
    pool_size: 20          # 根据并发量调整
    min_idle_conns: 5      # 保持最小连接数
    max_conn_age: "1h"     # 连接最大生存时间
    pool_timeout: "4s"     # 获取连接超时
    idle_timeout: "5m"     # 空闲连接超时
```

### 2. 批量操作优化

```go
// 使用多重锁减少网络开销
func processMultipleOrdersOptimized(lockManager lock.LockManager, orderIDs []string) error {
    lockKeys := make([]string, len(orderIDs))
    for i, id := range orderIDs {
        lockKeys[i] = "order:" + id
    }
    
    multiLock, err := lockManager.NewMultiLock(lockKeys, time.Minute)
    if err != nil {
        return err
    }
    
    ctx := context.Background()
    if err := multiLock.LockAll(ctx); err != nil {
        return err
    }
    defer multiLock.UnlockAll(ctx)
    
    // 批量处理订单
    return processBatchOrders(orderIDs)
}
```

### 3. 锁重试策略

```go
// 指数退避重试
func acquireLockWithBackoff(lock lock.Lock, ctx context.Context) error {
    maxRetries := 5
    baseDelay := 100 * time.Millisecond
    
    for i := 0; i < maxRetries; i++ {
        acquired, err := lock.TryLock(ctx)
        if err != nil {
            return err
        }
        
        if acquired {
            return nil
        }
        
        // 指数退避
        delay := baseDelay * time.Duration(1<<uint(i))
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            continue
        }
    }
    
    return fmt.Errorf("获取锁重试次数耗尽")
}
```

## 故障排除

### 常见问题

#### 1. Redis 连接失败

**错误**: `dial tcp: connect: connection refused`

**解决方案**:
- 检查 Redis 服务是否运行
- 验证连接地址和端口
- 检查网络连通性
- 确认 Redis 认证信息

#### 2. 锁获取超时

**错误**: `context deadline exceeded`

**解决方案**:
- 检查锁的超时设置是否合理
- 分析锁竞争情况
- 优化业务逻辑执行时间
- 考虑使用非阻塞锁

#### 3. 死锁情况

**错误**: 多个进程互相等待对方释放锁

**解决方案**:
- 统一锁获取顺序
- 设置合理的锁超时时间
- 使用多重锁避免分步获取
- 启用死锁检测机制

### 调试技巧

```go
// 启用锁调试日志
func enableLockDebugging(lockManager lock.LockManager) {
    lockManager.SetLogLevel(lock.LogLevelDebug)
    
    lockManager.OnEvent(func(event *lock.LockEvent) {
        log.Printf("[LOCK] %s: %s by %s at %v",
            event.Type, event.Key, event.Owner, event.Timestamp)
    })
}

// 监控锁状态
func monitorLockStatus(lockManager lock.LockManager) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        metrics := lockManager.GetMetrics()
        log.Printf("锁指标: 活跃锁数=%d, 总获取数=%d, 失败数=%d",
            metrics.ActiveLocks, metrics.TotalAcquired, metrics.TotalFailed)
    }
}
```

## 参考资料

- [Redis 分布式锁原理](https://redis.io/topics/distlock)
- [Go-Snap Boot 模块](boot.md)
- [Go-Snap Config 模块](config.md)
- [Go-Snap 架构设计](../architecture.md) 