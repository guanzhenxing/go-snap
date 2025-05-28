# Go-Snap 快速开始指南

## 简介

Go-Snap 是一个标准化、模块化、可扩展的 Go 应用开发框架，提供了类似 Spring Boot 的开发体验。本指南将帮助您快速上手 Go-Snap 框架。

## 安装

### 环境要求

- Go 1.21 或更高版本
- Git

### 获取框架

```bash
go mod init your-project-name
go get github.com/guanzhenxing/go-snap
```

## 第一个应用

### 1. 基础应用

创建一个简单的应用：

```go
package main

import (
    "log"
    "github.com/guanzhenxing/go-snap/boot"
)

func main() {
    // 创建应用启动器
    app := boot.NewBoot().
        SetConfigPath("configs")

    // 运行应用
    if err := app.Run(); err != nil {
        log.Fatalf("应用启动失败: %v", err)
    }
}
```

### 2. 创建配置文件

在 `configs/application.yaml` 中添加基础配置：

```yaml
# 应用配置
app:
  name: "my-first-app"
  version: "1.0.0"
  env: "development"

# 日志配置
logger:
  enabled: true
  level: "info"
  json: false

# 缓存配置
cache:
  enabled: true
  type: "memory"
```

### 3. 启用 Web 服务

如果需要 Web 功能，在配置中添加：

```yaml
# Web服务配置
web:
  enabled: true
  host: "0.0.0.0"
  port: 8080
```

然后运行应用：

```bash
go run main.go
```

## 核心概念

### 应用启动流程

1. **创建启动器** - 使用 `boot.NewBoot()` 创建应用启动器
2. **配置组件** - 添加配置器、组件和插件
3. **初始化** - 解析依赖、初始化组件
4. **启动** - 启动所有组件
5. **运行** - 应用进入运行状态
6. **关闭** - 优雅关闭所有组件

### 组件系统

Go-Snap 基于组件系统构建，主要组件类型：

- **基础设施组件** - 日志、配置等基础功能
- **数据源组件** - 数据库、缓存等数据存储
- **核心业务组件** - 业务逻辑组件
- **Web服务组件** - HTTP服务、路由等

### 配置管理

支持多种配置源：
- YAML/JSON/TOML 配置文件
- 环境变量
- 命令行参数
- 配置优先级：命令行 > 环境变量 > 配置文件 > 默认值

## 常用功能

### 日志使用

```go
// 获取日志组件
if loggerComp, found := app.GetComponent("logger"); found {
    if lc, ok := loggerComp.(*boot.LoggerComponent); ok {
        logger := lc.GetLogger()
        logger.Info("应用已启动")
        logger.Error("发生错误", zap.Error(err))
    }
}
```

### 缓存使用

```go
// 获取缓存组件
if cacheComp, found := app.GetComponent("cache"); found {
    if cc, ok := cacheComp.(*boot.CacheComponent); ok {
        cache := cc.GetCache()
        cache.Set(ctx, "key", "value", time.Hour)
        value, found := cache.Get(ctx, "key")
    }
}
```

### 健康检查

```go
// 获取应用健康状态
healthStatus := app.GetHealthStatus()
if len(healthStatus) == 0 {
    fmt.Println("所有组件健康")
} else {
    fmt.Printf("健康检查失败: %+v", healthStatus)
}
```

### 应用指标

```go
// 获取应用运行指标
metrics := app.GetMetrics()
fmt.Printf("应用指标: %+v", metrics)
```

## 自定义组件

### 创建自定义组件

```go
type MyComponent struct {
    *boot.BaseComponent
    // 自定义字段
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
    return nil
}

func (c *MyComponent) Start(ctx context.Context) error {
    if err := c.BaseComponent.Start(ctx); err != nil {
        return err
    }
    // 自定义启动逻辑
    return nil
}

func (c *MyComponent) HealthCheck() error {
    if err := c.BaseComponent.HealthCheck(); err != nil {
        return err
    }
    // 自定义健康检查逻辑
    return nil
}
```

### 注册自定义组件

```go
app := boot.NewBoot().
    AddComponent(NewMyComponent())
```

## 环境配置

### 开发环境

```yaml
app:
  env: "development"
logger:
  level: "debug"
  json: false
```

### 生产环境

```yaml
app:
  env: "production"
logger:
  level: "info"
  json: true
  file:
    path: "/var/log/app.log"
```

## 下一步

- 查看 [架构设计](architecture.md) 了解框架设计理念
- 阅读 [模块文档](modules/) 深入了解各个模块
- 参考 [示例代码](examples/) 学习最佳实践
- 查看 [API 参考](api/api-reference.md) 了解完整API

## 常见问题

### Q: 如何调试应用启动问题？

A: 设置日志级别为 debug，查看详细的启动日志：

```yaml
logger:
  level: "debug"
```

### Q: 如何自定义配置文件路径？

A: 使用 `SetConfigPath()` 方法：

```go
app := boot.NewBoot().SetConfigPath("/path/to/configs")
```

### Q: 如何禁用某个组件？

A: 在配置文件中设置对应的 enabled 为 false：

```yaml
database:
  enabled: false
```

## 技术支持

- GitHub Issues: [提交问题](https://github.com/guanzhenxing/go-snap/issues)
- 文档: [完整文档](https://github.com/guanzhenxing/go-snap/tree/main/docs)
- 示例: [示例代码](https://github.com/guanzhenxing/go-snap/tree/main/examples) 