# Go-Snap 框架文档

欢迎来到 Go-Snap 框架文档！Go-Snap 是一个标准化、模块化、可扩展的 Go 应用开发框架，提供了类似 Spring Boot 的开发体验。

## 📚 文档导航

### 🚀 快速开始
- [快速开始指南](getting-started.md) - 5分钟快速上手 Go-Snap
- [架构设计](architecture.md) - 深入了解框架设计理念和架构

### 🛠️ 核心模块

#### 启动和核心
- [Boot 模块](modules/boot.md) - 应用启动框架，依赖注入，生命周期管理
- [Errors 模块](modules/errors.md) - 统一错误处理和错误类型定义

#### 基础设施
- [Logger 模块](modules/logger.md) - 高性能结构化日志系统
- [Config 模块](modules/config.md) - 强大的配置管理组件

#### 数据和缓存
- [Cache 模块](modules/cache.md) - 统一缓存接口，支持多种后端
- [DBStore 模块](modules/dbstore.md) - 数据库ORM和存储抽象
- [Lock 模块](modules/lock.md) - 分布式锁组件

#### Web 和网络
- [Web 模块](modules/web.md) - HTTP 服务器和 REST API 框架

### 📖 使用指南

#### 基础教程
- [基础使用示例](examples/basic-usage.md) - 从零开始构建应用
- [高级使用示例](examples/advanced-usage.md) - 高级功能和最佳实践

#### API 文档
- [API 参考手册](api/api-reference.md) - 完整的 API 文档

### 🤝 参与贡献
- [贡献指南](contributing.md) - 如何参与 Go-Snap 开发
- [变更日志](changelog.md) - 版本更新记录

## 🎯 主要特性

### 🏗️ 模块化架构
- **组件化设计** - 每个功能都是独立的组件
- **依赖注入** - 自动解析和注入组件依赖
- **插件系统** - 灵活的插件扩展机制

### ⚡ 高性能
- **零配置启动** - 开箱即用的默认配置
- **异步处理** - 非阻塞的异步操作
- **连接池** - 高效的资源管理

### 🔧 配置驱动
- **多环境支持** - 轻松切换开发、测试、生产环境
- **热重载** - 配置文件变更自动生效
- **环境变量** - 支持环境变量覆盖配置

### 📊 企业级功能
- **健康检查** - 内置的应用和组件健康监控
- **指标收集** - 丰富的运行时指标
- **优雅关闭** - 安全的应用关闭机制

## 🏃‍♂️ 5分钟快速体验

### 1. 创建项目

```bash
mkdir my-go-snap-app
cd my-go-snap-app
go mod init my-go-snap-app
go get github.com/guanzhenxing/go-snap
```

### 2. 编写应用

```go
package main

import (
    "log"
    "github.com/guanzhenxing/go-snap/boot"
)

func main() {
    app := boot.NewBoot().
        SetConfigPath("configs")
    
    if err := app.Run(); err != nil {
        log.Fatalf("应用启动失败: %v", err)
    }
}
```

### 3. 创建配置

```yaml
# configs/application.yaml
app:
  name: "my-first-app"
  version: "1.0.0"

web:
  enabled: true
  port: 8080

logger:
  level: "info"
```

### 4. 运行应用

```bash
go run main.go
```

🎉 恭喜！你的第一个 Go-Snap 应用已经运行起来了！

## 📋 模块概览

| 模块 | 描述 | 状态 |
|------|------|------|
| **Boot** | 应用启动框架 | ✅ 稳定 |
| **Logger** | 高性能日志系统 | ✅ 稳定 |
| **Config** | 配置管理 | ✅ 稳定 |
| **Cache** | 缓存组件 | ✅ 稳定 |
| **DBStore** | 数据库ORM | ✅ 稳定 |
| **Web** | HTTP服务器 | ✅ 稳定 |
| **Lock** | 分布式锁 | ✅ 稳定 |
| **Errors** | 错误处理 | ✅ 稳定 |

## 🔗 相关链接

- **GitHub**: [https://github.com/guanzhenxing/go-snap](https://github.com/guanzhenxing/go-snap)
- **问题反馈**: [Issues](https://github.com/guanzhenxing/go-snap/issues)
- **讨论区**: [Discussions](https://github.com/guanzhenxing/go-snap/discussions)

## 📜 许可证

Go-Snap 采用 [MIT 许可证](../LICENSE)。

---

**需要帮助？** 查看我们的[快速开始指南](getting-started.md)或在 [GitHub Issues](https://github.com/guanzhenxing/go-snap/issues) 中提问。 