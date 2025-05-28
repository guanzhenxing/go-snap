# API 参考手册

本文档提供 Go-Snap 框架的完整 API 参考，包括所有公开接口、类型定义和方法说明。

## 目录

- [Boot 模块 API](#boot-模块-api)
- [Logger 模块 API](#logger-模块-api)
- [Config 模块 API](#config-模块-api)
- [Cache 模块 API](#cache-模块-api)
- [Errors 模块 API](#errors-模块-api)

## Boot 模块 API

### 核心接口

#### Boot

应用启动器接口。

```go
type Boot interface {
    // 设置配置文件路径
    SetConfigPath(path string) Boot
    
    // 设置配置文件名（不含扩展名）
    SetConfigName(name string) Boot
    
    // 设置配置文件类型
    SetConfigType(configType string) Boot
    
    // 添加组件工厂
    AddComponent(name string, factory ComponentFactory) Boot
    
    // 添加插件
    AddPlugin(plugin Plugin) Boot
    
    // 添加配置器
    AddConfigurer(configurer func(*Application) error) Boot
    
    // 初始化应用（不启动）
    Initialize() (*Application, error)
    
    // 运行应用（初始化并启动）
    Run() error
}
```

#### Application

应用实例接口。

```go
type Application interface {
    // 获取应用名称
    GetName() string
    
    // 获取应用版本
    GetVersion() string
    
    // 获取组件
    GetComponent(name string) (Component, bool)
    
    // 获取所有组件
    GetComponents() map[string]Component
    
    // 启动应用
    Start() error
    
    // 停止应用
    Stop() error
    
    // 获取健康状态
    GetHealthStatus() *HealthStatus
    
    // 获取运行指标
    GetMetrics() *ApplicationMetrics
    
    // 发布事件
    PublishEvent(event Event) error
    
    // 订阅事件
    SubscribeEvent(eventType string, handler EventHandler) error
}
```

#### Component

组件接口。

```go
type Component interface {
    // 获取组件名称
    GetName() string
    
    // 获取组件类型
    GetType() ComponentType
    
    // 获取组件状态
    GetStatus() ComponentStatus
    
    // 获取依赖组件列表
    GetDependencies() []string
    
    // 初始化组件
    Initialize(ctx context.Context) error
    
    // 启动组件
    Start(ctx context.Context) error
    
    // 停止组件
    Stop(ctx context.Context) error
    
    // 健康检查
    HealthCheck(ctx context.Context) error
    
    // 获取组件指标
    GetMetrics() map[string]interface{}
}
```

#### ComponentFactory

组件工厂接口。

```go
type ComponentFactory interface {
    // 创建组件实例
    Create(config interface{}) (Component, error)
    
    // 验证配置
    ValidateConfig(config interface{}) error
    
    // 获取配置模式
    GetConfigSchema() *ConfigSchema
}
```

### 类型定义

#### ComponentType

组件类型枚举。

```go
type ComponentType string

const (
    ComponentTypeCore       ComponentType = "core"        // 核心组件
    ComponentTypeDataSource ComponentType = "datasource"  // 数据源组件
    ComponentTypeService    ComponentType = "service"     // 服务组件
    ComponentTypeWeb        ComponentType = "web"         // Web组件
    ComponentTypePlugin     ComponentType = "plugin"      // 插件组件
)
```

#### ComponentStatus

组件状态枚举。

```go
type ComponentStatus string

const (
    ComponentStatusCreated     ComponentStatus = "created"     // 已创建
    ComponentStatusInitialized ComponentStatus = "initialized" // 已初始化
    ComponentStatusStarting    ComponentStatus = "starting"    // 启动中
    ComponentStatusRunning     ComponentStatus = "running"     // 运行中
    ComponentStatusStopping    ComponentStatus = "stopping"    // 停止中
    ComponentStatusStopped     ComponentStatus = "stopped"     // 已停止
    ComponentStatusFailed      ComponentStatus = "failed"      // 失败
)
```

#### HealthStatus

健康状态结构。

```go
type HealthStatus struct {
    Status     HealthStatusType           `json:"status"`
    Components map[string]ComponentHealth `json:"components"`
    Timestamp  time.Time                  `json:"timestamp"`
}

type HealthStatusType string

const (
    HealthStatusHealthy   HealthStatusType = "healthy"
    HealthStatusUnhealthy HealthStatusType = "unhealthy"
    HealthStatusUnknown   HealthStatusType = "unknown"
)

type ComponentHealth struct {
    Status    HealthStatusType `json:"status"`
    Message   string           `json:"message,omitempty"`
    Timestamp time.Time        `json:"timestamp"`
}
```

### 基础组件类

#### BaseComponent

基础组件实现。

```go
type BaseComponent struct {
    name         string
    componentType ComponentType
    status       ComponentStatus
    dependencies []string
    metrics      map[string]interface{}
    mutex        sync.RWMutex
}

// 创建基础组件
func NewBaseComponent(name string, componentType ComponentType) *BaseComponent

// 设置组件状态
func (c *BaseComponent) SetStatus(status ComponentStatus)

// 添加依赖
func (c *BaseComponent) AddDependency(dependency string)

// 设置指标
func (c *BaseComponent) SetMetric(key string, value interface{})
```

### 内置组件

#### LoggerComponent

日志组件。

```go
type LoggerComponent struct {
    *BaseComponent
    logger logger.Logger
}

// 获取日志实例
func (c *LoggerComponent) GetLogger() logger.Logger
```

#### ConfigComponent

配置组件。

```go
type ConfigComponent struct {
    *BaseComponent
    config config.Config
}

// 获取配置实例
func (c *ConfigComponent) GetConfig() config.Config
```

#### CacheComponent

缓存组件。

```go
type CacheComponent struct {
    *BaseComponent
    cache cache.Cache
}

// 获取缓存实例
func (c *CacheComponent) GetCache() cache.Cache
```

## Logger 模块 API

### 核心接口

#### Logger

日志记录器接口。

```go
type Logger interface {
    // 日志级别方法
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
    
    // 格式化日志方法
    Debugf(template string, args ...interface{})
    Infof(template string, args ...interface{})
    Warnf(template string, args ...interface{})
    Errorf(template string, args ...interface{})
    Fatalf(template string, args ...interface{})
    
    // 带上下文的日志方法
    DebugContext(ctx context.Context, msg string, fields ...Field)
    InfoContext(ctx context.Context, msg string, fields ...Field)
    WarnContext(ctx context.Context, msg string, fields ...Field)
    ErrorContext(ctx context.Context, msg string, fields ...Field)
    
    // 创建子日志器
    With(fields ...Field) Logger
    Named(name string) Logger
    
    // 获取日志级别
    Level() Level
    
    // 同步日志缓冲区
    Sync() error
}
```

### 字段类型

#### Field

日志字段类型。

```go
type Field struct {
    Key   string
    Type  FieldType
    Value interface{}
}

// 字段构造函数
func String(key, val string) Field
func Int(key string, val int) Field
func Int64(key string, val int64) Field
func Float64(key string, val float64) Field
func Bool(key string, val bool) Field
func Duration(key string, val time.Duration) Field
func Time(key string, val time.Time) Field
func Error(err error) Field
func Any(key string, val interface{}) Field
```

### 日志级别

```go
type Level int8

const (
    DebugLevel Level = iota - 1
    InfoLevel
    WarnLevel
    ErrorLevel
    FatalLevel
)
```

### 配置结构

```go
type Config struct {
    Enabled bool        `mapstructure:"enabled"`
    Level   string      `mapstructure:"level"`
    JSON    bool        `mapstructure:"json"`
    File    FileConfig  `mapstructure:"file"`
    Sampling SamplingConfig `mapstructure:"sampling"`
}

type FileConfig struct {
    Enabled    bool   `mapstructure:"enabled"`
    Filename   string `mapstructure:"filename"`
    MaxSize    int    `mapstructure:"max_size"`
    MaxBackups int    `mapstructure:"max_backups"`
    MaxAge     int    `mapstructure:"max_age"`
    Compress   bool   `mapstructure:"compress"`
}

type SamplingConfig struct {
    Enabled    bool `mapstructure:"enabled"`
    Initial    int  `mapstructure:"initial"`
    Thereafter int  `mapstructure:"thereafter"`
}
```

## Config 模块 API

### 核心接口

#### Config

配置管理器接口。

```go
type Config interface {
    // 基础读取方法
    Get(key string) interface{}
    GetString(key string) string
    GetBool(key string) bool
    GetInt(key string) int
    GetInt64(key string) int64
    GetFloat64(key string) float64
    GetDuration(key string) time.Duration
    GetTime(key string) time.Time
    GetStringSlice(key string) []string
    GetStringMap(key string) map[string]interface{}
    
    // 带默认值的读取方法
    GetStringDefault(key, defaultValue string) string
    GetBoolDefault(key string, defaultValue bool) bool
    GetIntDefault(key string, defaultValue int) int
    
    // 检查键是否存在
    IsSet(key string) bool
    
    // 反序列化到结构体
    Unmarshal(rawVal interface{}) error
    UnmarshalKey(key string, rawVal interface{}) error
    
    // 设置值
    Set(key string, value interface{})
    
    // 监听配置变更
    WatchConfig()
    OnConfigChange(run func(in fsnotify.Event))
    
    // 获取所有配置
    AllSettings() map[string]interface{}
}
```

### 配置选项

```go
type Options struct {
    ConfigPath  string   // 配置文件路径
    ConfigName  string   // 配置文件名
    ConfigType  string   // 配置文件类型
    ConfigFiles []string // 配置文件列表
    EnvPrefix   string   // 环境变量前缀
    AutomaticEnv bool    // 自动读取环境变量
}
```

## Cache 模块 API

### 核心接口

#### Cache

缓存接口。

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
    
    // TTL 管理
    Expire(ctx context.Context, key string, ttl time.Duration) error
    TTL(ctx context.Context, key string) (time.Duration, error)
    
    // 统计信息
    Stats() CacheStats
    
    // 关闭缓存
    Close() error
}
```

### 缓存统计

```go
type CacheStats struct {
    Hits       int64   `json:"hits"`        // 命中次数
    Misses     int64   `json:"misses"`      // 丢失次数
    Operations int64   `json:"operations"`  // 总操作数
    Errors     int64   `json:"errors"`      // 错误次数
    HitRate    float64 `json:"hit_rate"`    // 命中率
    LastReset  time.Time `json:"last_reset"` // 上次重置时间
}
```

### 配置选项

#### MemoryOptions

内存缓存配置。

```go
type MemoryOptions struct {
    MaxEntries      int           // 最大条目数
    CleanupInterval time.Duration // 清理间隔
    OnEvicted       func(key string, value interface{}) // 驱逐回调
}
```

#### RedisOptions

Redis缓存配置。

```go
type RedisOptions struct {
    Addr         string        // Redis地址
    Password     string        // 密码
    DB           int           // 数据库编号
    PoolSize     int           // 连接池大小
    MinIdleConns int           // 最小空闲连接数
    DialTimeout  time.Duration // 连接超时
    ReadTimeout  time.Duration // 读取超时
    WriteTimeout time.Duration // 写入超时
}
```

### 序列化器

#### Serializer

序列化器接口。

```go
type Serializer interface {
    Serialize(value interface{}) ([]byte, error)
    Deserialize(data []byte) (interface{}, error)
}
```

内置序列化器：
- `JSONSerializer` - JSON序列化
- `GobSerializer` - Gob序列化  
- `MsgpackSerializer` - MessagePack序列化

## Errors 模块 API

### 核心接口

#### Error

扩展的错误接口。

```go
type Error interface {
    error
    
    // 获取错误码
    Code() string
    
    // 获取错误类型
    Type() string
    
    // 获取错误上下文
    Context() map[string]interface{}
    
    // 获取时间戳
    Timestamp() time.Time
    
    // 是否包装了其他错误
    Unwrap() error
}
```

### 错误类型

#### ConfigError

配置错误。

```go
type ConfigError struct {
    ConfigFile string                 // 配置文件
    Component  string                 // 组件名称
    Message    string                 // 错误消息
    Cause      error                  // 原因错误
    Timestamp  time.Time              // 时间戳
    Context    map[string]interface{} // 上下文
}

func NewConfigError(configFile, message string, cause error) *ConfigError
```

#### ComponentError

组件错误。

```go
type ComponentError struct {
    Component string                 // 组件名称
    Operation string                 // 操作类型
    Message   string                 // 错误消息
    Cause     error                  // 原因错误
    Timestamp time.Time              // 时间戳
    Context   map[string]interface{} // 上下文
}

func NewComponentError(component, operation, message string, cause error) *ComponentError
```

#### DependencyError

依赖错误。

```go
type DependencyError struct {
    Component    string                 // 组件名称
    Dependencies []string               // 依赖列表
    Message      string                 // 错误消息
    Timestamp    time.Time              // 时间戳
    Context      map[string]interface{} // 上下文
}

func NewDependencyError(component string, dependencies []string, message string) *DependencyError
```

### 错误码常量

```go
const (
    // 成功
    CodeSuccess = "SUCCESS"
    
    // 客户端错误 (4xx)
    CodeBadRequest       = "BAD_REQUEST"
    CodeUnauthorized     = "UNAUTHORIZED"
    CodeForbidden        = "FORBIDDEN"
    CodeNotFound         = "NOT_FOUND"
    CodeValidation       = "VALIDATION_ERROR"
    
    // 服务器错误 (5xx)
    CodeInternalError    = "INTERNAL_ERROR"
    CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
    CodeTimeout          = "TIMEOUT"
    
    // 业务错误
    CodeUserNotFound     = "USER_NOT_FOUND"
    CodeUserAlreadyExists = "USER_ALREADY_EXISTS"
    
    // 系统错误
    CodeDatabaseError    = "DATABASE_ERROR"
    CodeCacheError       = "CACHE_ERROR"
    CodeNetworkError     = "NETWORK_ERROR"
    CodeConfigError      = "CONFIG_ERROR"
)
```

### 工具函数

#### 错误创建

```go
// 创建基础错误
func New(message string) error
func Errorf(format string, args ...interface{}) error
func NewWithCode(code, message string) error

// 包装错误
func Wrap(err error, message string) error
func WrapWithCode(err error, code, message string) error
func Wrapf(err error, format string, args ...interface{}) error

// 带上下文的错误
func NewWithContext(context map[string]interface{}, message string) error
func WithContext(err error, context map[string]interface{}) error

// 带堆栈的错误
func NewWithStack(message string) error
func WithStack(err error) error
```

#### 错误检查

```go
// 检查错误链
func Is(err, target error) bool
func As(err error, target interface{}) bool
func Unwrap(err error) error

// 获取错误信息
func GetCode(err error) string
func GetType(err error) string
func GetAllContext(err error) map[string]interface{}
func GetStackTrace(err error) []StackFrame
```

#### HTTP 集成

```go
// 转换为HTTP状态码
func ToHTTPStatus(err error) int

// 转换为HTTP响应
func ToHTTPResponse(err error, requestID string) *ErrorResponse

type ErrorResponse struct {
    Code      string                 `json:"code"`
    Message   string                 `json:"message"`
    Details   interface{}           `json:"details,omitempty"`
    Context   map[string]interface{} `json:"context,omitempty"`
    Timestamp time.Time             `json:"timestamp"`
    RequestID string                `json:"request_id,omitempty"`
}
```

## 使用示例

### 基础应用创建

```go
package main

import (
    "github.com/guanzhenxing/go-snap/boot"
)

func main() {
    app := boot.NewBoot().
        SetConfigPath("configs").
        SetConfigName("application").
        SetConfigType("yaml")
    
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

### 自定义组件注册

```go
type MyServiceFactory struct{}

func (f *MyServiceFactory) Create(config interface{}) (boot.Component, error) {
    return NewMyService(), nil
}

func (f *MyServiceFactory) ValidateConfig(config interface{}) error {
    return nil
}

func (f *MyServiceFactory) GetConfigSchema() *boot.ConfigSchema {
    return &boot.ConfigSchema{
        Type: "object",
        Properties: map[string]*boot.PropertySchema{
            "enabled": {Type: "boolean", Description: "是否启用服务"},
        },
    }
}

func main() {
    app := boot.NewBoot().
        AddComponent("myService", &MyServiceFactory{})
    
    app.Run()
}
```

### 日志使用

```go
import "github.com/guanzhenxing/go-snap/logger"

func useLogger(log logger.Logger) {
    log.Info("用户登录",
        logger.String("username", "john"),
        logger.Int("user_id", 123),
        logger.Duration("response_time", time.Millisecond*50),
    )
    
    log.Error("数据库连接失败",
        logger.String("database", "users"),
        logger.Error(err),
    )
}
```

### 缓存使用

```go
import "github.com/guanzhenxing/go-snap/cache"

func useCache(c cache.Cache) {
    ctx := context.Background()
    
    // 设置缓存
    c.Set(ctx, "user:123", user, time.Hour)
    
    // 获取缓存
    if value, found := c.Get(ctx, "user:123"); found {
        user := value.(*User)
        // 使用用户数据
    }
    
    // 批量操作
    keys := []string{"user:1", "user:2", "user:3"}
    results, _ := c.GetMulti(ctx, keys)
}
```

### 错误处理

```go
import "github.com/guanzhenxing/go-snap/errors"

func handleUser(userID string) error {
    if userID == "" {
        return errors.NewValidationError("user_id", "用户ID不能为空")
    }
    
    user, err := getUserFromDB(userID)
    if err != nil {
        return errors.WrapWithCode(err, errors.CodeDatabaseError, "获取用户失败")
    }
    
    return nil
}
```

---

更多详细信息请参考各模块的具体文档：
- [Boot 模块](../modules/boot.md)
- [Logger 模块](../modules/logger.md)
- [Config 模块](../modules/config.md)
- [Cache 模块](../modules/cache.md)
- [Errors 模块](../modules/errors.md) 