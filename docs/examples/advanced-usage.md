# 高级使用示例

本文档提供 Go-Snap 框架的高级使用示例，涵盖企业级应用开发、微服务架构、性能优化等高级主题。

## 🏗️ 企业级应用架构

### 完整项目结构

```
enterprise-app/
├── cmd/
│   ├── api/
│   │   └── main.go              # API 服务入口
│   ├── worker/
│   │   └── main.go              # 后台任务处理器
│   └── migrate/
│       └── main.go              # 数据库迁移工具
├── internal/
│   ├── domain/                  # 领域模型
│   │   ├── user/
│   │   ├── order/
│   │   └── payment/
│   ├── service/                 # 业务服务层
│   ├── repository/              # 数据访问层
│   ├── handler/                 # HTTP 处理器
│   ├── middleware/              # 自定义中间件
│   └── config/                  # 配置管理
├── pkg/                         # 可复用包
│   ├── auth/                    # 认证包
│   ├── cache/                   # 缓存工具
│   └── utils/                   # 工具函数
├── deployments/                 # 部署配置
├── scripts/                     # 脚本文件
└── configs/                     # 配置文件
```

### 多服务启动器

```go
// cmd/api/main.go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/guanzhenxing/go-snap/boot"
    "enterprise-app/internal/config"
    "enterprise-app/internal/handler"
    "enterprise-app/internal/service"
    "enterprise-app/internal/repository"
)

func main() {
    app := boot.NewBoot().
        SetConfigPath("configs").
        AddComponent("userService", &service.UserServiceFactory{}).
        AddComponent("orderService", &service.OrderServiceFactory{}).
        AddComponent("paymentService", &service.PaymentServiceFactory{}).
        AddConfigurer(setupWebRoutes).
        AddConfigurer(setupMiddlewares)
    
    application, err := app.Initialize()
    if err != nil {
        log.Fatalf("应用初始化失败: %v", err)
    }
    
    // 启动应用
    go func() {
        if err := application.Start(); err != nil {
            log.Fatalf("应用启动失败: %v", err)
        }
    }()
    
    // 优雅关闭
    gracefulShutdown(application)
}

func gracefulShutdown(app *boot.Application) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("开始优雅关闭应用...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := app.Stop(ctx); err != nil {
        log.Printf("应用关闭失败: %v", err)
    }
    
    log.Println("应用已安全关闭")
}
```

### 领域驱动设计 (DDD)

```go
// internal/domain/user/entity.go
package user

import (
    "time"
    "github.com/guanzhenxing/go-snap/errors"
)

type User struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Status    Status    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Status int

const (
    StatusActive Status = iota + 1
    StatusInactive
    StatusSuspended
)

// 领域方法
func (u *User) Activate() error {
    if u.Status == StatusActive {
        return errors.NewUserError("用户已激活", errors.CodeInvalidState)
    }
    u.Status = StatusActive
    return nil
}

func (u *User) Suspend() error {
    if u.Status == StatusSuspended {
        return errors.NewUserError("用户已暂停", errors.CodeInvalidState)
    }
    u.Status = StatusSuspended
    return nil
}

// internal/domain/user/repository.go
type Repository interface {
    Save(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id uint) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Delete(ctx context.Context, id uint) error
}

// internal/domain/user/service.go
type Service struct {
    repo   Repository
    cache  cache.Cache
    logger logger.Logger
}

func NewService(repo Repository, cache cache.Cache, logger logger.Logger) *Service {
    return &Service{
        repo:   repo,
        cache:  cache,
        logger: logger.Named("user.service"),
    }
}

func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 业务规则验证
    if exists, err := s.repo.FindByEmail(ctx, req.Email); err == nil && exists != nil {
        return nil, errors.NewUserError("邮箱已存在", errors.CodeUserAlreadyExists)
    }
    
    user := &User{
        Username: req.Username,
        Email:    req.Email,
        Status:   StatusActive,
    }
    
    if err := s.repo.Save(ctx, user); err != nil {
        return nil, errors.WrapWithCode(err, errors.CodeDatabaseError, "保存用户失败")
    }
    
    // 缓存用户信息
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    s.cache.Set(ctx, cacheKey, user, time.Hour)
    
    s.logger.Info("用户创建成功",
        logger.String("user_id", fmt.Sprintf("%d", user.ID)),
        logger.String("email", user.Email),
    )
    
    return user, nil
}
```

## 🔧 自定义组件开发

### 高级自定义组件

```go
// internal/components/notification.go
package components

import (
    "context"
    "time"
    "github.com/guanzhenxing/go-snap/boot"
    "github.com/guanzhenxing/go-snap/logger"
)

type NotificationComponent struct {
    *boot.BaseComponent
    logger   logger.Logger
    channels map[string]NotificationChannel
    config   *NotificationConfig
}

type NotificationChannel interface {
    Send(ctx context.Context, message *Message) error
    Name() string
}

type Message struct {
    To      string                 `json:"to"`
    Subject string                 `json:"subject"`
    Content string                 `json:"content"`
    Type    string                 `json:"type"`
    Data    map[string]interface{} `json:"data"`
}

type NotificationConfig struct {
    Email EmailConfig `mapstructure:"email"`
    SMS   SMSConfig   `mapstructure:"sms"`
    Push  PushConfig  `mapstructure:"push"`
}

func NewNotificationComponent(logger logger.Logger) *NotificationComponent {
    return &NotificationComponent{
        BaseComponent: boot.NewBaseComponent("notification", boot.ComponentTypeService),
        logger:        logger.Named("notification"),
        channels:      make(map[string]NotificationChannel),
    }
}

func (c *NotificationComponent) Initialize(ctx context.Context) error {
    if err := c.BaseComponent.Initialize(ctx); err != nil {
        return err
    }
    
    // 初始化通知渠道
    if c.config.Email.Enabled {
        emailChannel := NewEmailChannel(c.config.Email)
        c.channels["email"] = emailChannel
    }
    
    if c.config.SMS.Enabled {
        smsChannel := NewSMSChannel(c.config.SMS)
        c.channels["sms"] = smsChannel
    }
    
    c.SetMetric("channels_count", len(c.channels))
    c.logger.Info("通知组件初始化完成", logger.Int("channels", len(c.channels)))
    
    return nil
}

func (c *NotificationComponent) SendNotification(ctx context.Context, channel string, message *Message) error {
    ch, exists := c.channels[channel]
    if !exists {
        return fmt.Errorf("通知渠道 %s 不存在", channel)
    }
    
    start := time.Now()
    err := ch.Send(ctx, message)
    duration := time.Since(start)
    
    // 记录指标
    c.SetMetric(fmt.Sprintf("%s_send_duration", channel), duration)
    if err != nil {
        c.SetMetric(fmt.Sprintf("%s_error_count", channel), 
            c.GetMetrics()[fmt.Sprintf("%s_error_count", channel)].(int64) + 1)
        c.logger.Error("发送通知失败",
            logger.String("channel", channel),
            logger.String("to", message.To),
            logger.Error(err),
        )
    } else {
        c.SetMetric(fmt.Sprintf("%s_success_count", channel),
            c.GetMetrics()[fmt.Sprintf("%s_success_count", channel)].(int64) + 1)
        c.logger.Info("通知发送成功",
            logger.String("channel", channel),
            logger.String("to", message.To),
        )
    }
    
    return err
}

// 组件工厂
type NotificationComponentFactory struct{}

func (f *NotificationComponentFactory) Create(ctx context.Context, props boot.PropertySource) (boot.Component, error) {
    logger, _ := ctx.Value("logger").(logger.Logger)
    
    var config NotificationConfig
    if err := props.UnmarshalKey("notification", &config); err != nil {
        return nil, err
    }
    
    component := NewNotificationComponent(logger)
    component.config = &config
    
    return component, nil
}

func (f *NotificationComponentFactory) Dependencies() []string {
    return []string{"logger"}
}

func (f *NotificationComponentFactory) ValidateConfig(props boot.PropertySource) error {
    if !props.GetBool("notification.enabled", false) {
        return nil
    }
    
    // 验证至少启用一个通知渠道
    emailEnabled := props.GetBool("notification.email.enabled", false)
    smsEnabled := props.GetBool("notification.sms.enabled", false)
    
    if !emailEnabled && !smsEnabled {
        return fmt.Errorf("至少需要启用一个通知渠道")
    }
    
    return nil
}
```

### 插件系统

```go
// pkg/plugins/metrics.go
package plugins

import (
    "time"
    "github.com/guanzhenxing/go-snap/boot"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsPlugin struct {
    registry     *prometheus.Registry
    httpRequests *prometheus.CounterVec
    httpDuration *prometheus.HistogramVec
}

func NewMetricsPlugin() *MetricsPlugin {
    registry := prometheus.NewRegistry()
    
    httpRequests := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpDuration := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "Duration of HTTP requests",
        },
        []string{"method", "path"},
    )
    
    registry.MustRegister(httpRequests, httpDuration)
    
    return &MetricsPlugin{
        registry:     registry,
        httpRequests: httpRequests,
        httpDuration: httpDuration,
    }
}

func (p *MetricsPlugin) Name() string {
    return "MetricsPlugin"
}

func (p *MetricsPlugin) Version() string {
    return "1.0.0"
}

func (p *MetricsPlugin) Register(app *boot.Application) error {
    // 注册 metrics 端点
    if webComp, found := app.GetComponent("web"); found {
        if wc, ok := webComp.(*boot.WebComponent); ok {
            router := wc.GetRouter()
            router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})))
            
            // 注册 metrics 中间件
            router.Use(p.metricsMiddleware())
        }
    }
    
    return nil
}

func (p *MetricsPlugin) metricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := fmt.Sprintf("%d", c.Writer.Status())
        
        p.httpRequests.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
        p.httpDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
    }
}
```

## 🚀 微服务集成

### 服务发现与注册

```go
// pkg/discovery/consul.go
package discovery

import (
    "fmt"
    "github.com/hashicorp/consul/api"
)

type ConsulRegistry struct {
    client *api.Client
    config *RegistryConfig
}

type RegistryConfig struct {
    Address string `mapstructure:"address"`
    Service string `mapstructure:"service"`
    Port    int    `mapstructure:"port"`
    Tags    []string `mapstructure:"tags"`
}

func NewConsulRegistry(config *RegistryConfig) (*ConsulRegistry, error) {
    clientConfig := api.DefaultConfig()
    clientConfig.Address = config.Address
    
    client, err := api.NewClient(clientConfig)
    if err != nil {
        return nil, err
    }
    
    return &ConsulRegistry{
        client: client,
        config: config,
    }, nil
}

func (r *ConsulRegistry) Register() error {
    registration := &api.AgentServiceRegistration{
        ID:   fmt.Sprintf("%s-%d", r.config.Service, r.config.Port),
        Name: r.config.Service,
        Port: r.config.Port,
        Tags: r.config.Tags,
        Check: &api.AgentServiceCheck{
            HTTP:                           fmt.Sprintf("http://localhost:%d/health", r.config.Port),
            Timeout:                        "3s",
            Interval:                       "10s",
            DeregisterCriticalServiceAfter: "30s",
        },
    }
    
    return r.client.Agent().ServiceRegister(registration)
}

func (r *ConsulRegistry) Deregister() error {
    serviceID := fmt.Sprintf("%s-%d", r.config.Service, r.config.Port)
    return r.client.Agent().ServiceDeregister(serviceID)
}

// 集成到启动器
func setupServiceDiscovery(app *boot.Application) error {
    if configComp, found := app.GetComponent("config"); found {
        if cc, ok := configComp.(*boot.ConfigComponent); ok {
            config := cc.GetConfig()
            
            var registryConfig RegistryConfig
            if err := config.UnmarshalKey("registry", &registryConfig); err != nil {
                return err
            }
            
            registry, err := NewConsulRegistry(&registryConfig)
            if err != nil {
                return err
            }
            
            // 注册服务
            if err := registry.Register(); err != nil {
                return err
            }
            
            // 在应用关闭时注销服务
            app.OnShutdown(func() error {
                return registry.Deregister()
            })
        }
    }
    
    return nil
}
```

### 分布式配置中心

```go
// pkg/config/remote.go
package config

import (
    "context"
    "time"
    "github.com/guanzhenxing/go-snap/boot"
    "go.etcd.io/etcd/clientv3"
)

type RemoteConfigSource struct {
    client *clientv3.Client
    prefix string
    cache  map[string]string
}

func NewRemoteConfigSource(endpoints []string, prefix string) (*RemoteConfigSource, error) {
    client, err := clientv3.New(clientv3.Config{
        Endpoints:   endpoints,
        DialTimeout: 5 * time.Second,
    })
    if err != nil {
        return nil, err
    }
    
    return &RemoteConfigSource{
        client: client,
        prefix: prefix,
        cache:  make(map[string]string),
    }, nil
}

func (r *RemoteConfigSource) Load() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    resp, err := r.client.Get(ctx, r.prefix, clientv3.WithPrefix())
    if err != nil {
        return err
    }
    
    for _, kv := range resp.Kvs {
        key := string(kv.Key)
        value := string(kv.Value)
        r.cache[key] = value
    }
    
    return nil
}

func (r *RemoteConfigSource) Watch() (<-chan map[string]string, error) {
    updates := make(chan map[string]string, 1)
    
    go func() {
        watchChan := r.client.Watch(context.Background(), r.prefix, clientv3.WithPrefix())
        
        for resp := range watchChan {
            changes := make(map[string]string)
            for _, event := range resp.Events {
                key := string(event.Kv.Key)
                value := string(event.Kv.Value)
                changes[key] = value
                r.cache[key] = value
            }
            
            if len(changes) > 0 {
                updates <- changes
            }
        }
    }()
    
    return updates, nil
}
```

## ⚡ 性能优化

### 连接池优化

```go
// pkg/pool/connection.go
package pool

import (
    "context"
    "sync"
    "time"
)

type ConnectionPool struct {
    factory     func() (interface{}, error)
    close       func(interface{}) error
    ping        func(interface{}) error
    connections chan interface{}
    maxIdle     int
    maxOpen     int
    idleTimeout time.Duration
    mu          sync.RWMutex
    opened      int
}

func NewConnectionPool(config PoolConfig) *ConnectionPool {
    return &ConnectionPool{
        factory:     config.Factory,
        close:       config.Close,
        ping:        config.Ping,
        connections: make(chan interface{}, config.MaxIdle),
        maxIdle:     config.MaxIdle,
        maxOpen:     config.MaxOpen,
        idleTimeout: config.IdleTimeout,
    }
}

func (p *ConnectionPool) Get() (interface{}, error) {
    // 尝试从池中获取连接
    select {
    case conn := <-p.connections:
        if p.ping != nil && p.ping(conn) != nil {
            p.close(conn)
            return p.createConnection()
        }
        return conn, nil
    default:
        return p.createConnection()
    }
}

func (p *ConnectionPool) Put(conn interface{}) {
    if conn == nil {
        return
    }
    
    select {
    case p.connections <- conn:
    default:
        // 池已满，关闭连接
        p.close(conn)
        p.mu.Lock()
        p.opened--
        p.mu.Unlock()
    }
}

func (p *ConnectionPool) createConnection() (interface{}, error) {
    p.mu.Lock()
    if p.opened >= p.maxOpen {
        p.mu.Unlock()
        // 等待可用连接
        return <-p.connections, nil
    }
    p.opened++
    p.mu.Unlock()
    
    return p.factory()
}
```

### 缓存策略优化

```go
// pkg/cache/multi_level.go
package cache

import (
    "context"
    "time"
    "github.com/guanzhenxing/go-snap/cache"
)

type MultiLevelCache struct {
    l1       cache.Cache // 内存缓存
    l2       cache.Cache // Redis缓存
    l1TTL    time.Duration
    l2TTL    time.Duration
    stats    *CacheStats
}

type CacheStats struct {
    L1Hits   int64 `json:"l1_hits"`
    L2Hits   int64 `json:"l2_hits"`
    Misses   int64 `json:"misses"`
    L1Size   int64 `json:"l1_size"`
}

func NewMultiLevelCache(l1, l2 cache.Cache, l1TTL, l2TTL time.Duration) *MultiLevelCache {
    return &MultiLevelCache{
        l1:    l1,
        l2:    l2,
        l1TTL: l1TTL,
        l2TTL: l2TTL,
        stats: &CacheStats{},
    }
}

func (c *MultiLevelCache) Get(ctx context.Context, key string) (interface{}, bool) {
    // 先查L1缓存
    if value, found := c.l1.Get(ctx, key); found {
        atomic.AddInt64(&c.stats.L1Hits, 1)
        return value, true
    }
    
    // 再查L2缓存
    if value, found := c.l2.Get(ctx, key); found {
        atomic.AddInt64(&c.stats.L2Hits, 1)
        // 回写到L1缓存
        c.l1.Set(ctx, key, value, c.l1TTL)
        return value, true
    }
    
    atomic.AddInt64(&c.stats.Misses, 1)
    return nil, false
}

func (c *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    // 同时写入L1和L2缓存
    c.l1.Set(ctx, key, value, c.l1TTL)
    return c.l2.Set(ctx, key, value, c.l2TTL)
}

func (c *MultiLevelCache) Delete(ctx context.Context, key string) error {
    c.l1.Delete(ctx, key)
    return c.l2.Delete(ctx, key)
}

func (c *MultiLevelCache) Stats() *CacheStats {
    return c.stats
}
```

## 🔍 监控和观测

### 链路追踪

```go
// pkg/tracing/opentelemetry.go
package tracing

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
    "github.com/gin-gonic/gin"
)

func TracingMiddleware(serviceName string) gin.HandlerFunc {
    tracer := otel.Tracer(serviceName)
    
    return func(c *gin.Context) {
        ctx, span := tracer.Start(c.Request.Context(), c.Request.URL.Path)
        defer span.End()
        
        // 设置 span 属性
        span.SetAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.String()),
            attribute.String("http.user_agent", c.Request.UserAgent()),
        )
        
        // 将 context 传递给下一个处理器
        c.Request = c.Request.WithContext(ctx)
        c.Next()
        
        // 设置响应状态
        span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))
        
        if c.Writer.Status() >= 400 {
            span.SetStatus(codes.Error, "HTTP Error")
        }
    }
}

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
    return otel.Tracer("go-snap").Start(ctx, name)
}
```

### 健康检查增强

```go
// pkg/health/advanced.go
package health

import (
    "context"
    "fmt"
    "time"
    "github.com/guanzhenxing/go-snap/boot"
)

type AdvancedHealthChecker struct {
    checks   map[string]HealthCheck
    timeout  time.Duration
    cache    map[string]*HealthResult
    cacheTTL time.Duration
}

type HealthCheck interface {
    Check(ctx context.Context) *HealthResult
    Name() string
}

type HealthResult struct {
    Status    HealthStatus           `json:"status"`
    Message   string                 `json:"message"`
    Timestamp time.Time              `json:"timestamp"`
    Duration  time.Duration          `json:"duration"`
    Details   map[string]interface{} `json:"details,omitempty"`
}

type HealthStatus string

const (
    StatusHealthy   HealthStatus = "healthy"
    StatusUnhealthy HealthStatus = "unhealthy"
    StatusDegraded  HealthStatus = "degraded"
)

func NewAdvancedHealthChecker() *AdvancedHealthChecker {
    return &AdvancedHealthChecker{
        checks:   make(map[string]HealthCheck),
        timeout:  time.Second * 30,
        cache:    make(map[string]*HealthResult),
        cacheTTL: time.Second * 10,
    }
}

func (h *AdvancedHealthChecker) AddCheck(check HealthCheck) {
    h.checks[check.Name()] = check
}

func (h *AdvancedHealthChecker) CheckAll(ctx context.Context) map[string]*HealthResult {
    results := make(map[string]*HealthResult)
    
    for name, check := range h.checks {
        // 检查缓存
        if cached, exists := h.cache[name]; exists {
            if time.Since(cached.Timestamp) < h.cacheTTL {
                results[name] = cached
                continue
            }
        }
        
        // 执行健康检查
        ctx, cancel := context.WithTimeout(ctx, h.timeout)
        result := check.Check(ctx)
        cancel()
        
        // 缓存结果
        h.cache[name] = result
        results[name] = result
    }
    
    return results
}

// 数据库健康检查
type DatabaseHealthCheck struct {
    db *gorm.DB
}

func (d *DatabaseHealthCheck) Name() string {
    return "database"
}

func (d *DatabaseHealthCheck) Check(ctx context.Context) *HealthResult {
    start := time.Now()
    
    var count int64
    err := d.db.WithContext(ctx).Raw("SELECT 1").Count(&count).Error
    
    duration := time.Since(start)
    
    if err != nil {
        return &HealthResult{
            Status:    StatusUnhealthy,
            Message:   fmt.Sprintf("数据库连接失败: %v", err),
            Timestamp: time.Now(),
            Duration:  duration,
        }
    }
    
    status := StatusHealthy
    if duration > time.Second {
        status = StatusDegraded
    }
    
    return &HealthResult{
        Status:    status,
        Message:   "数据库连接正常",
        Timestamp: time.Now(),
        Duration:  duration,
        Details: map[string]interface{}{
            "response_time_ms": duration.Milliseconds(),
        },
    }
}
```

## 🧪 测试策略

### 集成测试框架

```go
// test/integration/setup.go
package integration

import (
    "context"
    "testing"
    "github.com/guanzhenxing/go-snap/boot"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/mysql"
    "github.com/testcontainers/testcontainers-go/modules/redis"
)

type TestSuite struct {
    app           *boot.Application
    mysqlContainer testcontainers.Container
    redisContainer testcontainers.Container
    cleanup       []func()
}

func NewTestSuite(t *testing.T) *TestSuite {
    suite := &TestSuite{}
    
    // 启动 MySQL 容器
    mysqlContainer, err := mysql.RunContainer(context.Background(),
        mysql.WithDatabase("testdb"),
        mysql.WithUsername("test"),
        mysql.WithPassword("test"),
    )
    if err != nil {
        t.Fatalf("启动 MySQL 容器失败: %v", err)
    }
    suite.mysqlContainer = mysqlContainer
    
    // 启动 Redis 容器
    redisContainer, err := redis.RunContainer(context.Background())
    if err != nil {
        t.Fatalf("启动 Redis 容器失败: %v", err)
    }
    suite.redisContainer = redisContainer
    
    // 创建测试应用
    app := boot.NewBoot().
        SetConfigPath("testdata").
        AddComponent("testDB", &TestDatabaseComponent{container: mysqlContainer}).
        AddComponent("testCache", &TestCacheComponent{container: redisContainer})
    
    application, err := app.Initialize()
    if err != nil {
        t.Fatalf("初始化测试应用失败: %v", err)
    }
    
    suite.app = application
    return suite
}

func (s *TestSuite) Cleanup() {
    for _, cleanup := range s.cleanup {
        cleanup()
    }
    
    if s.mysqlContainer != nil {
        s.mysqlContainer.Terminate(context.Background())
    }
    
    if s.redisContainer != nil {
        s.redisContainer.Terminate(context.Background())
    }
}

// 使用示例
func TestUserService(t *testing.T) {
    suite := NewTestSuite(t)
    defer suite.Cleanup()
    
    // 获取用户服务
    userService := suite.app.GetComponent("userService").(*UserService)
    
    // 执行测试
    user, err := userService.CreateUser(context.Background(), &CreateUserRequest{
        Username: "test",
        Email:    "test@example.com",
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test", user.Username)
}
```

## 📋 部署和运维

### Docker 化部署

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./main"]
```

### Kubernetes 部署配置

```yaml
# deployments/k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-snap-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-snap-app
  template:
    metadata:
      labels:
        app: go-snap-app
    spec:
      containers:
      - name: app
        image: go-snap-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: CONFIG_PATH
          value: "/etc/config"
        - name: LOG_LEVEL
          value: "info"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: go-snap-app-service
spec:
  selector:
    app: go-snap-app
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

这些高级示例展示了 Go-Snap 框架在企业级应用开发中的强大能力，包括微服务架构、性能优化、监控观测等方面的最佳实践。通过这些示例，开发者可以构建出高性能、可扩展、易维护的现代化应用系统。 