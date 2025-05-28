# Boot 模块

Boot 模块是 Go-Snap 框架的核心启动模块，提供类似 Spring Boot 的应用启动体验。它负责应用的生命周期管理、组件注册与依赖注入、自动配置等核心功能。

## 概述

Boot 模块采用组件化架构，通过依赖注入、自动配置、事件驱动等机制，简化 Go 应用的开发和部署。

### 核心特性

- ✅ **应用生命周期管理** - 完整的启动、运行、关闭流程
- ✅ **依赖注入系统** - 自动解析和注入组件依赖
- ✅ **自动配置机制** - 基于配置的组件自动装配
- ✅ **健康检查系统** - 内置的应用和组件健康监控
- ✅ **事件驱动架构** - 支持事件发布订阅模式
- ✅ **插件系统** - 灵活的插件扩展机制
- ✅ **监控指标** - 内置的性能指标收集

## 核心概念

### 1. 应用启动器 (Boot)

应用启动器是应用的入口点，负责创建和配置应用实例。

```go
// 创建启动器
boot := boot.NewBoot()

// 配置启动器
boot.SetConfigPath("configs").
     AddComponent(customComponent).
     AddPlugin(customPlugin).
     AddConfigurer(customConfigurer)

// 运行应用
err := boot.Run()
```

### 2. 应用实例 (Application)

应用实例是运行时的应用管理器，负责组件生命周期、健康检查、事件处理等。

```go
// 初始化应用（不启动）
app, err := boot.Initialize()

// 获取组件
logger, found := app.GetComponent("logger")

// 获取健康状态
healthStatus := app.GetHealthStatus()

// 获取运行指标
metrics := app.GetMetrics()
```

### 3. 组件系统 (Component)

组件是应用的基本构建单元，每个组件都有完整的生命周期。

#### 组件接口

```go
type Component interface {
    Name() string
    Type() ComponentType
    Initialize(ctx context.Context) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    HealthCheck() error
    GetStatus() ComponentStatus
    GetMetrics() map[string]interface{}
}
```

#### 组件类型

```go
const (
    ComponentTypeInfrastructure ComponentType = iota // 基础设施组件
    ComponentTypeDataSource                          // 数据源组件
    ComponentTypeCore                                // 核心业务组件
    ComponentTypeWeb                                 // Web服务组件
)
```

#### 组件状态

```go
const (
    ComponentStatusCreated ComponentStatus = iota     // 已创建
    ComponentStatusInitialized                       // 已初始化
    ComponentStatusStarted                           // 已启动
    ComponentStatusStopped                           // 已停止
    ComponentStatusFailed                            // 失败
)
```

### 4. 组件工厂 (ComponentFactory)

组件工厂负责创建组件实例和管理依赖关系。

```go
type ComponentFactory interface {
    Create(ctx context.Context, props PropertySource) (Component, error)
    Dependencies() []string
    ValidateConfig(props PropertySource) error
    GetConfigSchema() ConfigSchema
}
```

### 5. 自动配置器 (AutoConfigurer)

自动配置器负责根据配置自动装配组件。

```go
type AutoConfigurer interface {
    Configure(registry *ComponentRegistry, props PropertySource) error
    Order() int
    GetName() string
}
```

## 使用指南

### 基础用法

#### 1. 简单应用

```go
package main

import (
    "log"
    "github.com/guanzhenxing/go-snap/boot"
)

func main() {
    app := boot.NewBoot()
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

#### 2. 配置路径

```go
app := boot.NewBoot().SetConfigPath("/path/to/configs")
```

#### 3. 添加自定义组件

```go
app := boot.NewBoot().AddComponent(myCustomComponent)
```

### 自定义组件开发

#### 1. 使用基础组件类

```go
type MyComponent struct {
    *boot.BaseComponent
    service MyService
}

func NewMyComponent() *MyComponent {
    return &MyComponent{
        BaseComponent: boot.NewBaseComponent("my-component", boot.ComponentTypeCore),
    }
}

func (c *MyComponent) Initialize(ctx context.Context) error {
    if err := c.BaseComponent.Initialize(ctx); err != nil {
        return err
    }
    
    // 自定义初始化逻辑
    c.service = NewMyService()
    c.SetMetric("service_type", "custom")
    
    return nil
}

func (c *MyComponent) Start(ctx context.Context) error {
    if err := c.BaseComponent.Start(ctx); err != nil {
        return err
    }
    
    // 启动服务
    return c.service.Start()
}

func (c *MyComponent) Stop(ctx context.Context) error {
    // 停止服务
    c.service.Stop()
    return c.BaseComponent.Stop(ctx)
}

func (c *MyComponent) HealthCheck() error {
    if err := c.BaseComponent.HealthCheck(); err != nil {
        return err
    }
    
    // 自定义健康检查
    return c.service.HealthCheck()
}
```

#### 2. 组件工厂

```go
type MyComponentFactory struct{}

func (f *MyComponentFactory) Create(ctx context.Context, props PropertySource) (Component, error) {
    component := NewMyComponent()
    
    // 根据配置初始化组件
    if props.GetBool("my.enabled", false) {
        // 配置组件
    }
    
    return component, nil
}

func (f *MyComponentFactory) Dependencies() []string {
    return []string{"logger", "config"}
}

func (f *MyComponentFactory) ValidateConfig(props PropertySource) error {
    if !props.GetBool("my.enabled", false) {
        return nil
    }
    
    // 验证必需配置
    if !props.HasProperty("my.required_config") {
        return fmt.Errorf("my.required_config is required")
    }
    
    return nil
}

func (f *MyComponentFactory) GetConfigSchema() ConfigSchema {
    return ConfigSchema{
        RequiredProperties: []string{"my.required_config"},
        Properties: map[string]PropertySchema{
            "my.enabled": {
                Type:         "bool",
                DefaultValue: false,
                Description:  "启用自定义组件",
                Required:     false,
            },
            "my.required_config": {
                Type:        "string",
                Description: "必需的配置项",
                Required:    true,
            },
        },
        Dependencies: []string{"logger", "config"},
    }
}
```

#### 3. 自动配置器

```go
type MyAutoConfigurer struct{}

func (c *MyAutoConfigurer) Configure(registry *ComponentRegistry, props PropertySource) error {
    // 检查是否启用
    if !props.GetBool("my.enabled", false) {
        return nil
    }
    
    // 注册组件工厂
    return registry.RegisterFactory("my-component", &MyComponentFactory{})
}

func (c *MyAutoConfigurer) Order() int {
    return 500 // 配置顺序
}

func (c *MyAutoConfigurer) GetName() string {
    return "MyAutoConfigurer"
}
```

#### 4. 注册自定义配置器

```go
app := boot.NewBoot().AddConfigurer(&MyAutoConfigurer{})
```

### 插件开发

```go
type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "MyPlugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Dependencies() []string {
    return []string{}
}

func (p *MyPlugin) Register(app *boot.Application) error {
    // 订阅事件
    eventBus := app.GetEventBus()
    
    eventBus.Subscribe("application.started", func(eventName string, eventData interface{}) {
        log.Println("应用已启动，插件开始工作")
    })
    
    eventBus.Subscribe("application.health_check.failed", func(eventName string, eventData interface{}) {
        // 处理健康检查失败
        if healthResults, ok := eventData.(map[string]error); ok {
            log.Printf("健康检查失败: %v", healthResults)
        }
    })
    
    return nil
}

// 注册插件
app := boot.NewBoot().AddPlugin(&MyPlugin{})
```

## 事件系统

### 内置事件

- `application.initialized` - 应用初始化完成
- `application.started` - 应用启动完成
- `application.stopping` - 应用开始停止
- `application.stopped` - 应用停止完成
- `application.state.changed` - 应用状态变更
- `application.health_check.passed` - 健康检查通过
- `application.health_check.failed` - 健康检查失败
- `component.stop.error` - 组件停止错误

### 事件订阅

```go
app, _ := boot.Initialize()
eventBus := app.GetEventBus()

// 订阅事件
eventBus.Subscribe("application.started", func(eventName string, eventData interface{}) {
    fmt.Println("应用已启动!")
})

// 发布事件
eventBus.Publish("custom.event", customData)

// 同步发布事件
eventBus.PublishSync("custom.event", customData)
```

## 健康检查

### 应用级健康检查

```go
app, _ := boot.Initialize()

// 获取健康状态
healthStatus := app.GetHealthStatus()
if len(healthStatus) == 0 {
    fmt.Println("所有组件健康")
} else {
    for component, err := range healthStatus {
        fmt.Printf("组件 %s 健康检查失败: %v\n", component, err)
    }
}
```

### 手动健康检查

```go
// 手动触发健康检查
registry := app.GetRegistry()
healthResults := registry.HealthCheck()
```

### 配置健康检查间隔

```yaml
app:
  health_check_interval: 30  # 秒
```

## 监控指标

### 应用指标

```go
app, _ := boot.Initialize()

// 获取应用指标
metrics := app.GetMetrics()
fmt.Printf("应用指标: %+v\n", metrics)

// 包含的指标：
// - app_name: 应用名称
// - app_version: 应用版本
// - app_state: 应用状态
// - start_time: 启动时间
// - uptime: 运行时间
// - component_count: 组件数量
// - health_check_count: 健康检查次数
// - error_count: 错误次数
```

### 注册表指标

```go
registry := app.GetRegistry()
registryMetrics := registry.GetMetrics()

// 包含的指标：
// - ComponentCount: 组件数量
// - FactoryCount: 工厂数量
// - DependencyResolvTime: 依赖解析时间
// - HealthCheckCount: 健康检查次数
// - FailedComponents: 失败组件列表
```

### 组件指标

```go
component, found := app.GetComponent("my-component")
if found {
    metrics := component.GetMetrics()
    fmt.Printf("组件指标: %+v\n", metrics)
}
```

## 配置选项

### 应用配置

```yaml
app:
  name: "my-app"                    # 应用名称
  version: "1.0.0"                  # 应用版本
  env: "production"                 # 运行环境
  shutdown_timeout: 30              # 关闭超时时间（秒）
  health_check_interval: 30         # 健康检查间隔（秒）
```

### 组件配置

每个组件都有自己的配置节，具体配置项请参考各组件的文档。

## 最佳实践

### 1. 组件设计原则

```go
// ✅ 好的做法
type GoodComponent struct {
    *boot.BaseComponent
    config MyConfig
    service MyService
}

func (c *GoodComponent) Initialize(ctx context.Context) error {
    // 调用基类方法
    if err := c.BaseComponent.Initialize(ctx); err != nil {
        return err
    }
    
    // 初始化配置
    c.config = LoadConfig()
    
    // 初始化服务
    c.service = NewMyService(c.config)
    
    // 设置指标
    c.SetMetric("config_loaded", true)
    
    return nil
}

// ❌ 避免的做法
type BadComponent struct {
    // 不继承BaseComponent
    name string
}

func (c *BadComponent) Initialize(ctx context.Context) error {
    // 没有状态管理
    // 没有指标收集
    // 没有错误处理
    return nil
}
```

### 2. 依赖管理

```go
// ✅ 声明明确的依赖
func (f *MyComponentFactory) Dependencies() []string {
    return []string{"logger", "config", "cache"}
}

// ❌ 避免循环依赖
// A 依赖 B，B 又依赖 A
```

### 3. 配置验证

```go
func (f *MyComponentFactory) ValidateConfig(props PropertySource) error {
    // ✅ 验证必需配置
    if !props.HasProperty("my.required_field") {
        return fmt.Errorf("my.required_field is required")
    }
    
    // ✅ 验证配置值
    port := props.GetInt("my.port", 8080)
    if port <= 0 || port > 65535 {
        return fmt.Errorf("invalid port: %d", port)
    }
    
    return nil
}
```

### 4. 健康检查实现

```go
func (c *MyComponent) HealthCheck() error {
    // ✅ 先检查基础状态
    if err := c.BaseComponent.HealthCheck(); err != nil {
        return err
    }
    
    // ✅ 检查具体功能
    if !c.service.IsReady() {
        return fmt.Errorf("service not ready")
    }
    
    // ✅ 检查外部依赖
    if err := c.service.Ping(); err != nil {
        return fmt.Errorf("external dependency check failed: %v", err)
    }
    
    return nil
}
```

### 5. 错误处理

```go
func (c *MyComponent) Start(ctx context.Context) error {
    if err := c.BaseComponent.Start(ctx); err != nil {
        return err
    }
    
    // ✅ 使用结构化错误
    if err := c.service.Start(); err != nil {
        return boot.NewComponentError(
            c.Name(),
            "start",
            "启动服务失败",
            err,
        )
    }
    
    return nil
}
```

## 故障排除

### 常见问题

#### 1. 组件依赖循环

**错误**: `发现循环依赖: [A, B, A]`

**解决方案**:
- 检查组件工厂的 `Dependencies()` 方法
- 重构组件设计，消除循环依赖
- 使用事件系统替代直接依赖

#### 2. 组件初始化失败

**错误**: `组件初始化失败: component_name`

**解决方案**:
- 检查组件的 `Initialize()` 方法实现
- 验证组件所需的配置是否正确
- 确认依赖组件已正确初始化

#### 3. 健康检查失败

**错误**: 组件健康检查持续失败

**解决方案**:
- 检查组件的 `HealthCheck()` 方法实现
- 验证外部依赖（数据库、缓存等）是否可用
- 查看组件日志了解具体错误

### 调试技巧

#### 1. 启用调试日志

```yaml
logger:
  level: "debug"
```

#### 2. 查看组件状态

```go
components := app.GetRegistry().GetAllComponents()
for name, component := range components {
    fmt.Printf("组件 %s 状态: %s\n", name, component.GetStatus())
}
```

#### 3. 监控应用指标

```go
metrics := app.GetMetrics()
fmt.Printf("应用运行时间: %s\n", metrics["uptime"])
fmt.Printf("组件数量: %d\n", metrics["component_count"])
fmt.Printf("错误次数: %d\n", metrics["error_count"])
```

## 性能优化

### 1. 启动性能优化

- 减少不必要的组件
- 优化组件初始化逻辑
- 使用并行初始化（框架自动处理）

### 2. 运行时性能优化

- 合理设置健康检查间隔
- 避免在热路径上进行重复计算
- 使用组件缓存机制

### 3. 内存优化

- 及时清理不需要的指标数据
- 避免在组件中持有大量内存
- 使用对象池减少GC压力

## 版本兼容性

Boot 模块保持向后兼容性，但建议使用最新版本以获得最佳性能和功能。

### 迁移指南

从旧版本迁移时，请注意：

1. 组件接口新增了方法，需要实现新的接口方法
2. 配置结构可能有变化，请更新配置文件
3. 事件名称可能有调整，请更新事件监听器

## 参考资料

- [架构设计](../architecture.md)
- [快速开始指南](../getting-started.md)
- [配置模块文档](config.md)
- [示例代码](../examples/) 