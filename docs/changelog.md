# 变更日志

所有重要变更都会记录在此文件中。

本项目遵循 [Semantic Versioning](https://semver.org/spec/v2.0.0.html) 规范。

## [未发布]

### 新增
- 无

### 变更
- 无

### 修复
- 无

### 移除
- 无

## [1.0.0] - 2024-01-15

### 新增

#### 🚀 核心框架
- **Boot 模块**: 完整的应用启动框架
  - 组件化架构设计
  - 依赖注入和自动配置
  - 应用生命周期管理
  - 健康检查系统
  - 事件驱动架构
  - 插件系统支持
  - 监控指标收集

#### 🛠️ 基础设施组件
- **Logger 模块**: 高性能结构化日志系统
  - 基于 Uber Zap 的零分配日志
  - 多级别日志支持 (Debug, Info, Warn, Error, Fatal)
  - 多输出目标 (控制台、文件、JSON 格式)
  - 异步日志记录
  - 自动文件轮转
  - 日志采样功能
  - 开发模式友好输出

- **Config 模块**: 强大的配置管理系统
  - 基于 Viper 的多格式支持 (YAML, JSON, TOML, INI, HCL)
  - 多环境配置支持
  - 配置优先级 (命令行 > 环境变量 > 配置文件 > 默认值)
  - 配置热重载
  - 环境变量自动映射
  - 配置验证机制
  - 远程配置中心支持

- **Errors 模块**: 统一错误处理组件
  - 结构化错误类型定义
  - 错误包装和错误链
  - 内置错误码系统
  - 堆栈追踪支持
  - 多语言错误消息
  - HTTP 状态码映射
  - 日志系统集成

#### 💾 数据和缓存
- **Cache 模块**: 统一缓存接口
  - 内存缓存实现 (支持 LRU 驱逐策略)
  - Redis 缓存支持
  - 多级缓存架构
  - 自动序列化 (JSON, Gob, MessagePack)
  - TTL 管理
  - 批量操作支持
  - 缓存统计和监控
  - 事件监听机制

- **DBStore 模块**: 数据库 ORM 抽象
  - 基于 GORM 的数据库操作
  - 多数据库支持 (MySQL, PostgreSQL, SQLite, SQL Server)
  - 连接池管理
  - 事务支持
  - 数据库迁移
  - 查询构建器
  - 分页支持
  - 软删除支持

- **Lock 模块**: 分布式锁组件
  - Redis 分布式锁实现
  - 可重入锁支持
  - 锁超时机制
  - 自动续租功能
  - 死锁检测

#### 🌐 Web 服务
- **Web 模块**: HTTP 服务器框架
  - 基于 Gin 的高性能 HTTP 服务器
  - 路由组和中间件支持
  - 请求/响应绑定和验证
  - CORS 支持
  - JWT 认证中间件
  - 限流中间件
  - 请求日志记录
  - 错误处理中间件

#### 🔧 核心功能增强
- **并发安全优化**
  - ComponentRegistry 双重检查锁定模式
  - 拓扑排序的依赖解析算法 (O(n+m) 复杂度)
  - 线程安全的组件状态管理
  - 事件总线并发安全

- **健康检查系统**
  - 组件级健康检查
  - 应用级健康检查
  - 定期健康检查调度
  - 健康检查事件通知
  - 健康状态聚合

- **监控和指标**
  - 应用运行时指标
  - 组件状态指标
  - 注册表操作指标
  - 错误统计指标
  - 性能监控指标

- **配置和验证增强**
  - ComponentFactory 配置验证
  - ConfigSchema 定义支持
  - PropertySchema 属性描述
  - 配置文档自动生成

#### 📚 文档和示例
- **完整文档体系**
  - 快速开始指南
  - 架构设计文档
  - 各模块详细文档
  - API 参考手册
  - 最佳实践指南
  - 故障排除指南

- **示例代码**
  - 基础使用示例
  - 高级功能示例
  - 集成测试示例
  - 性能基准测试

- **开发者资源**
  - 贡献指南
  - 代码规范
  - 测试指南
  - 发布流程

#### 🎯 企业级特性
- **错误处理增强**
  - 结构化错误类型 (ConfigError, ComponentError, DependencyError 等)
  - 错误上下文和元数据
  - 错误链和根因分析
  - 国际化错误消息

- **性能优化**
  - 启动性能优化 (拓扑排序、并行初始化)
  - 运行时性能优化 (异步处理、连接池)
  - 内存优化 (对象复用、指标聚合)

- **安全性**
  - 输入验证和清理
  - 敏感信息掩码
  - 安全的默认配置
  - 最小权限原则

### 技术规格

#### 支持的 Go 版本
- Go 1.21 或更高版本

#### 主要依赖
- `github.com/gin-gonic/gin` - HTTP Web 框架
- `github.com/spf13/viper` - 配置管理
- `go.uber.org/zap` - 高性能日志
- `gorm.io/gorm` - ORM 数据库框架
- `github.com/go-redis/redis/v8` - Redis 客户端
- `github.com/stretchr/testify` - 测试框架

#### 架构特点
- **模块化设计**: 每个功能作为独立模块，支持按需引入
- **依赖注入**: 自动解析和注入组件依赖关系
- **事件驱动**: 基于事件总线的松耦合架构
- **配置驱动**: 通过配置文件控制应用行为
- **接口优先**: 面向接口编程，易于测试和扩展

#### 性能指标
- **应用启动时间**: < 100ms (典型配置)
- **依赖解析**: O(n+m) 时间复杂度
- **日志性能**: 零分配结构化日志
- **缓存性能**: 支持高并发读写操作
- **HTTP 性能**: 基于 Gin 的高性能 HTTP 处理

#### 兼容性
- **向后兼容**: 遵循语义化版本控制
- **平台支持**: Linux, macOS, Windows
- **数据库支持**: MySQL, PostgreSQL, SQLite, SQL Server
- **缓存支持**: Memory, Redis, Multi-level

### 迁移指南

#### 从自定义框架迁移
1. **替换应用启动代码**
   ```go
   // 旧的方式
   app := &MyApp{}
   app.Init()
   app.Start()
   
   // 新的方式
   app := boot.NewBoot()
   app.Run()
   ```

2. **配置文件迁移**
   - 将现有配置转换为 YAML/JSON 格式
   - 使用嵌套结构组织配置
   - 添加环境特定配置文件

3. **日志系统迁移**
   ```go
   // 旧的方式
   log.Printf("User %s logged in", username)
   
   // 新的方式
   logger.Info("用户登录", logger.String("username", username))
   ```

4. **错误处理迁移**
   ```go
   // 旧的方式
   return fmt.Errorf("user not found: %s", userID)
   
   // 新的方式
   return errors.NewUserError("用户不存在", errors.CodeUserNotFound)
   ```

### 已知问题

#### 限制
- 分布式锁目前仅支持 Redis 后端
- 配置热重载不支持所有配置项
- 某些组件在 Windows 上的性能可能有差异

#### 计划改进
- 添加更多分布式锁后端 (etcd, consul)
- 增强配置热重载功能
- 优化 Windows 平台性能
- 添加更多缓存后端支持

### 致谢

感谢以下开源项目为 Go-Snap 提供的基础支持：
- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Zap](https://github.com/uber-go/zap) - 高性能日志
- [GORM](https://github.com/go-gorm/gorm) - ORM 框架
- [Redis](https://github.com/go-redis/redis) - Redis 客户端
- [Testify](https://github.com/stretchr/testify) - 测试框架

特别感谢所有贡献者和早期用户的反馈与建议！

---

## 版本说明

### 版本格式
遵循 [Semantic Versioning](https://semver.org/) 规范：
- **MAJOR**: 不兼容的 API 变更
- **MINOR**: 向后兼容的功能增加  
- **PATCH**: 向后兼容的问题修复

### 变更类型
- **新增**: 新功能
- **变更**: 现有功能的变更
- **修复**: 问题修复
- **移除**: 移除的功能
- **安全**: 安全相关的修复

### 符号说明
- 🚀 重大新功能
- ✨ 功能增强
- 🐛 问题修复
- 📚 文档更新
- 🔧 配置变更
- ⚡ 性能改进
- 🛡️ 安全修复
- 💥 破坏性变更

---

*更多信息请查看 [GitHub Releases](https://github.com/guanzhenxing/go-snap/releases)* 