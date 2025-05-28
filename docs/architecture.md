# Go-Snap 架构设计

## 设计理念

Go-Snap 框架的设计遵循以下核心理念：

### 1. 简洁高效
- 提供简洁的 API，减少样板代码
- 高性能的组件设计，最小化资源开销
- 快速的应用启动和响应时间

### 2. 模块解耦
- 框架核心与应用分离
- 模块间松耦合设计
- 支持按需引入组件

### 3. 约定优于配置
- 提供合理的默认值配置
- 支持配置覆盖和自定义
- 最小化必需配置项

### 4. 面向接口
- 通过接口定义组件行为
- 降低模块间依赖
- 提高代码可测试性

### 5. 包容生态
- 集成 Go 生态优秀库
- 不重复造轮子
- 保持与标准库的兼容性

## 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        Application                          │
├─────────────────────────────────────────────────────────────┤
│                     Boot Framework                          │
├─────────────────────────────────────────────────────────────┤
│  Component System  │  Event System  │  Configuration      │
├─────────────────────────────────────────────────────────────┤
│   Core Modules                                             │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────────────┐ │
│  │ Logger  │ Config  │ Cache   │ DBStore │ Web / Lock /... │ │
│  └─────────┴─────────┴─────────┴─────────┴─────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                   Foundation Layer                         │
│  ┌─────────────────┬─────────────────┬─────────────────────┐ │
│  │ Errors Handling │ Property Source │ Dependency Injection│ │
│  └─────────────────┴─────────────────┴─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. Boot Framework

#### Boot 启动器
- **职责**: 应用启动和生命周期管理
- **特性**:
  - 应用创建和初始化
  - 组件注册和配置
  - 插件系统支持
  - 优雅关闭机制

#### Application 应用实例
- **职责**: 运行时应用管理
- **特性**:
  - 组件生命周期管理
  - 健康检查系统
  - 指标收集和监控
  - 事件发布和订阅

### 2. Component System

#### Component Registry
- **职责**: 组件注册和依赖管理
- **特性**:
  - 并发安全的组件注册
  - 拓扑排序的依赖解析
  - 组件工厂模式
  - 循环依赖检测

#### Component Lifecycle
```
Created → Initialized → Started → Running ⟷ HealthCheck
   ↓           ↓          ↓         ↓            ↓
Failed ←── Failed ←── Failed ←── Stopping → Stopped
```

### 3. Configuration System

#### Property Source
- **职责**: 配置源抽象和管理
- **支持格式**: YAML, JSON, TOML, 环境变量
- **优先级**: 命令行 > 环境变量 > 配置文件 > 默认值

#### Auto Configuration
- **职责**: 自动配置组件
- **特性**:
  - 条件化配置
  - 配置验证
  - 配置模式定义

### 4. Event System

#### Event Bus
- **职责**: 应用内事件通信
- **特性**:
  - 同步和异步事件发布
  - 事件订阅和取消订阅
  - 事件监听器管理

#### 内置事件类型
- `application.initialized` - 应用初始化完成
- `application.started` - 应用启动完成
- `application.stopping` - 应用开始停止
- `application.stopped` - 应用停止完成
- `application.state.changed` - 应用状态变更
- `application.health_check.passed` - 健康检查通过
- `application.health_check.failed` - 健康检查失败
- `component.stop.error` - 组件停止错误

## 模块架构

### Logger 模块

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Logger API    │    │   Zap Logger    │    │  Output Targets │
│                 │    │                 │    │                 │
│ • Info()        │───▶│ • Structured    │───▶│ • Console       │
│ • Error()       │    │ • Leveled       │    │ • File          │
│ • Debug()       │    │ • Fast          │    │ • JSON Format   │
│ • With()        │    │ • Sampled       │    │ • Rotation      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Config 模块

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Config API     │    │  Viper Core     │    │ Config Sources  │
│                 │    │                 │    │                 │
│ • Get()         │───▶│ • Multi-format  │───▶│ • YAML Files    │
│ • Set()         │    │ • Watch         │    │ • ENV Variables │
│ • IsSet()       │    │ • Merge         │    │ • Command Args  │
│ • Sub()         │    │ • Validate      │    │ • Remote Config │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Cache 模块

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Cache API     │    │ Cache Provider  │    │ Cache Backends  │
│                 │    │                 │    │                 │
│ • Get()         │───▶│ • Memory Cache  │───▶│ • In-Memory     │
│ • Set()         │    │ • Redis Cache   │    │ • Redis         │
│ • Delete()      │    │ • Multi-Level   │    │ • Multi-Level   │
│ • Exists()      │    │ • Serializer    │    │ • Distributed   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Web 模块

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web API       │    │   Gin Engine    │    │   Features      │
│                 │    │                 │    │                 │
│ • Router        │───▶│ • HTTP Server   │───▶│ • Middleware    │
│ • Middleware    │    │ • Route Groups  │    │ • CORS          │
│ • Context       │    │ • JSON Binding  │    │ • JWT Auth      │
│ • Response      │    │ • Validation    │    │ • Rate Limiting │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### DBStore 模块

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  DBStore API    │    │   GORM Core     │    │   Databases     │
│                 │    │                 │    │                 │
│ • Repository    │───▶│ • ORM Features  │───▶│ • MySQL         │
│ • Transaction  │    │ • Migration     │    │ • PostgreSQL    │
│ • Pagination   │    │ • Connection    │    │ • SQLite        │
│ • Query        │    │ • Pool          │    │ • SQL Server    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 设计模式

### 1. 依赖注入 (Dependency Injection)

```go
// 组件工厂负责创建组件实例
type ComponentFactory interface {
    Create(ctx context.Context, props PropertySource) (Component, error)
    Dependencies() []string
    ValidateConfig(props PropertySource) error
    GetConfigSchema() ConfigSchema
}

// 注册表管理组件依赖关系
registry.RegisterFactory("myComponent", &MyComponentFactory{})
```

### 2. 工厂模式 (Factory Pattern)

```go
// 每个组件都有对应的工厂
type LoggerComponentFactory struct{}

func (f *LoggerComponentFactory) Create(ctx context.Context, props PropertySource) (Component, error) {
    // 根据配置创建具体的日志组件实例
    return &LoggerComponent{logger: logger.New(opts...)}, nil
}
```

### 3. 观察者模式 (Observer Pattern)

```go
// 事件总线实现观察者模式
eventBus.Subscribe("application.started", func(eventName string, eventData interface{}) {
    // 处理应用启动事件
})

eventBus.Publish("application.started", app)
```

### 4. 策略模式 (Strategy Pattern)

```go
// 不同的缓存实现策略
type Cache interface {
    Get(ctx context.Context, key string) (interface{}, bool)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

// 内存缓存策略
type MemoryCache struct { ... }

// Redis缓存策略
type RedisCache struct { ... }
```

### 5. 模板方法模式 (Template Method Pattern)

```go
// BaseComponent 提供通用的生命周期模板
type BaseComponent struct {
    // 通用字段和方法
}

func (c *BaseComponent) Initialize(ctx context.Context) error {
    // 通用初始化逻辑
    c.SetStatus(ComponentStatusInitialized)
    c.SetMetric("initialized_at", time.Now())
    return nil
}

// 具体组件继承并扩展
type LoggerComponent struct {
    *BaseComponent
    logger logger.Logger
}
```

## 并发安全

### 1. 组件注册表并发安全

使用双重检查锁定模式确保组件创建的并发安全：

```go
func (r *ComponentRegistry) GetComponent(name string) (Component, bool) {
    // 第一次检查（读锁）
    r.mutex.RLock()
    component, exists := r.components[name]
    if exists {
        r.mutex.RUnlock()
        return component, true
    }
    r.mutex.RUnlock()
    
    // 第二次检查（写锁）
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    // 再次检查，防止重复创建
    if component, exists := r.components[name]; exists {
        return component, true
    }
    
    // 创建组件...
}
```

### 2. 组件状态并发安全

每个组件都有独立的读写锁保护状态：

```go
type BaseComponent struct {
    status   ComponentStatus
    statusMu sync.RWMutex
    metrics  map[string]interface{}
    metricsMu sync.RWMutex
}
```

### 3. 事件系统并发安全

事件总线使用读写锁保护监听器列表：

```go
type EventBus struct {
    listeners map[string][]EventListener
    mutex     sync.RWMutex
}
```

## 性能优化

### 1. 启动性能

- **拓扑排序**: 依赖解析使用拓扑排序算法，时间复杂度 O(V+E)
- **并行初始化**: 无依赖的组件可以并行初始化
- **延迟加载**: 支持组件的延迟创建和初始化

### 2. 运行时性能

- **异步处理**: 健康检查和事件处理采用异步模式
- **连接池**: 数据库和缓存使用连接池
- **内存复用**: 使用对象池减少GC压力

### 3. 内存优化

- **组件复用**: 单例模式避免重复创建
- **配置缓存**: 配置解析结果缓存
- **指标聚合**: 指标数据定期聚合清理

## 扩展机制

### 1. 自定义组件

```go
// 实现Component接口
type CustomComponent struct {
    *boot.BaseComponent
}

// 实现ComponentFactory接口
type CustomComponentFactory struct{}
```

### 2. 插件系统

```go
// 实现Plugin接口
type CustomPlugin struct{}

func (p *CustomPlugin) Register(app *boot.Application) error {
    // 注册插件逻辑
    return nil
}
```

### 3. 配置器扩展

```go
// 实现AutoConfigurer接口
type CustomConfigurer struct{}

func (c *CustomConfigurer) Configure(registry *boot.ComponentRegistry, props boot.PropertySource) error {
    // 自定义配置逻辑
    return nil
}
```

## 错误处理

### 错误类型层次

```
ConfigError (基础错误)
├── ComponentError (组件错误)
├── DependencyError (依赖错误)
└── ValidationError (验证错误)
```

### 错误上下文

- **组件信息**: 错误发生的组件名称
- **操作信息**: 错误发生的操作类型
- **时间戳**: 错误发生时间
- **错误链**: 支持错误包装和链式传播

## 监控和观测

### 1. 健康检查

- **组件级健康检查**: 每个组件实现自己的健康检查逻辑
- **应用级健康检查**: 定期检查所有组件健康状态
- **健康检查事件**: 健康检查结果通过事件总线发布

### 2. 指标收集

- **应用指标**: 启动时间、运行时间、组件数量等
- **组件指标**: 组件状态、初始化时间、错误计数等
- **注册表指标**: 依赖解析时间、健康检查次数等

### 3. 事件追踪

- **生命周期事件**: 应用和组件的生命周期事件
- **状态变更事件**: 状态变更时的事件通知
- **错误事件**: 错误发生时的事件记录

## 最佳实践

### 1. 组件设计

- 组件应该职责单一，功能内聚
- 组件间通过接口交互，减少直接依赖
- 组件应该支持优雅启动和关闭
- 实现完整的健康检查逻辑

### 2. 配置管理

- 使用有意义的配置键名
- 提供合理的默认值
- 对关键配置进行验证
- 支持不同环境的配置

### 3. 错误处理

- 使用结构化的错误类型
- 提供详细的错误上下文
- 支持错误链和错误包装
- 记录错误日志和指标

### 4. 性能考虑

- 避免在热路径上进行重复计算
- 使用适当的并发控制
- 合理使用缓存机制
- 定期监控和优化性能

这种架构设计确保了 Go-Snap 框架的可扩展性、可维护性和高性能，为构建企业级Go应用提供了坚实的基础。 