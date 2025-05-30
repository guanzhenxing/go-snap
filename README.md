# go-snap

本脚手架旨在提供一个标准化、模块化、可扩展的 Go 应用开发框架，适用于构建从小型服务到大型企业级应用的各种场景。它结合了业界最佳实践和经验，为开发团队提供一个一致的开发环境和架构方案。遵循以下设计理念：

1. **简洁高效**：提供简洁的 API，减少样板代码，提高开发效率
2. **模块解耦**：框架核心与应用分离，模块间松耦合，便于按需引入
3. **约定优于配置**：提供合理默认值，减少配置工作，同时保留灵活性和通过配置覆盖默认行为的能力
4. **面向接口**：通过接口定义组件行为，降低模块间依赖，提高可测试性
5. **包容生态**：不重复造轮子，整合 Go 生态中优秀的库和工具（如 Gin、GORM、go-redis、zap 等）
6. **渐进式架构**：核心功能可以独立使用，组件可以按需引入，支持从简单应用到复杂系统的平滑过渡
7. **自动化依赖管理**：使用依赖注入框架简化组件依赖关系，提高代码可测试性和可维护性
8. **反应式设计**：支持异步处理和事件驱动模型，提高系统响应能力和吞吐量

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [架构概览](#架构概览)
- [核心模块](#核心模块)
  - [Errors 模块](#errors-模块)
  - [Logger 模块](#logger-模块)
  - [Config 模块](#config-模块)
  - [缓存模块](#缓存模块)
  - [分布式锁](#分布式锁)
  - [DBStore 模块](#dbstore)
  - [Web 模块](#web-模块)
  - [Boot 模块](#boot-模块)
- [扩展指南](#扩展指南)
- [性能考虑](#性能考虑)
- [版本兼容性](#版本兼容性)
- [常见问题](#常见问题)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

## 安装

### 前置要求

- Go 1.16 或更高版本
- 如果使用 MySQL，需要安装 MySQL 客户端库
- 如果使用 Redis 缓存，需要安装 Redis 服务器

### 安装框架

```bash
# 安装 go-snap 库
go get -u github.com/guanzhenxing/go-snap

# 或者克隆仓库进行开发
git clone https://github.com/guanzhenxing/go-snap.git
cd go-snap
go mod download
```

### 创建新项目

```bash
# 创建项目目录
mkdir myproject
cd myproject

# 初始化 Go 模块
go mod init github.com/yourusername/myproject

# 添加 go-snap 依赖
go get -u github.com/guanzhenxing/go-snap
```

## 快速开始

### 基本应用示例

下面是一个最小化的 Go-Snap 应用示例，展示了如何使用框架构建一个简单的 Web 服务：

```go
package main

import (
	"github.com/guanzhenxing/go-snap/boot"
	"github.com/guanzhenxing/go-snap/web"
	"github.com/guanzhenxing/go-snap/web/response"
	"github.com/gin-gonic/gin"
)

func main() {
	// 创建启动器
	app := boot.NewBoot()
	
	// 添加自定义组件
	app.AddComponent(&MyWebComponent{})
	
	// 运行应用
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// MyWebComponent 自定义Web组件
type MyWebComponent struct{}

func (c *MyWebComponent) Name() string {
	return "myWebComponent"
}

func (c *MyWebComponent) Initialize(registry *boot.ComponentRegistry) error {
	// 获取Web服务器组件
	webServer, _ := registry.GetComponent("webServer")
	server := webServer.(*web.Server)
	
	// 注册路由
	server.GET("/hello", func(c *gin.Context) {
		response.Success(c, map[string]interface{}{
			"message": "Hello from Go-Snap!",
		})
	})
	
	return nil
}
```

### 配置文件示例

在 `configs/application.yml` 中添加配置：

```yaml
app:
  name: myapp
  version: 1.0.0
  
web:
  host: 0.0.0.0
  port: 8080
  mode: debug
  
logger:
  level: info
  format: json
  output: stdout
  
database:
  driver: mysql
  dsn: "user:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
  max_idle_conns: 10
  max_open_conns: 100
  
cache:
  type: redis
  redis:
    addr: "localhost:6379"
```

### 运行应用

```bash
# 直接运行
go run main.go

# 或者编译后运行
go build -o myapp
./myapp
```

访问 http://localhost:8080/hello 应该能看到 JSON 响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "Hello from Go-Snap!"
  },
  "request_id": "c4dbe23a-f54d-4c9a-b730-65b700c54982",
  "timestamp": 1634567890123
}
```

## 架构概览

Go-Snap 采用模块化架构，遵循关注点分离原则，各个模块之间通过接口进行交互。

```
                     ┌───────────────────────────────────────────┐
                     │              应用层 (Application)           │
                     └───────────────────────────────────────────┘
                                          │
                                          ▼
┌───────────────────────────────────────────────────────────────────────────────┐
│                                 启动模块 (Boot)                                 │
└───────────────────────────────────────────────────────────────────────────────┘
         │                │               │                │               │
         ▼                ▼               ▼                ▼               ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  Web 模块   │  │  缓存模块   │  │ 数据库模块  │  │  日志模块   │  │  配置模块   │
└─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘
         │                │               │                │               │
         └────────────────┴───────────────┴────────────────┴───────────────┘
                                          │
                                          ▼
                          ┌───────────────────────────┐
                          │      错误处理 (Errors)     │
                          └───────────────────────────┘
```

### 核心概念

1. **组件 (Component)**：框架中的功能单元，如日志、缓存、Web服务器等
2. **属性源 (PropertySource)**：配置数据的抽象，提供统一的配置访问接口
3. **组件注册表 (ComponentRegistry)**：管理组件生命周期和依赖关系
4. **事件总线 (EventBus)**：组件间通信的消息系统
5. **自动配置 (AutoConfig)**：基于配置自动初始化和配置组件
6. **插件 (Plugin)**：扩展框架功能的可插拔单元

## 核心模块

### Errors 模块

Errors 模块提供了强大而灵活的错误处理机制，使开发者能更好地管理和处理应用中的错误：

- **错误链**：支持错误包装和链式传播，保留完整的错误上下文和调用栈
- **错误类型**：提供常见错误类型（如验证错误、业务错误、系统错误等）
- **错误码**：统一的错误码系统，便于错误分类和客户端处理
- **国际化**：支持错误信息的多语言翻译
- **安全处理**：自动过滤敏感信息，避免在错误信息中泄露敏感数据

**示例：**

```go
// 创建新错误
err := errors.New("something went wrong")

// 包装错误
if err != nil {
    return errors.Wrap(err, "failed to process request")
}

// 使用错误码
return errors.WithCode(UserNotFound, "user %s not found", username)

// 获取根本原因
rootErr := errors.Cause(err)

// 检查错误码
if errors.IsErrorCode(err, NotFound) {
    // 处理未找到错误
}
```

### Logger 模块

Logger 模块是一个高性能、功能丰富的日志系统，帮助开发者记录、监控和调试应用：

- **结构化日志**：支持字段化日志记录，便于后续分析和处理
- **多级别日志**：提供 Debug、Info、Warn、Error、Fatal 等多个日志级别
- **多输出目标**：支持同时输出到控制台、文件，支持 JSON 格式
- **数据掩码**：自动掩码敏感信息（如密码、信用卡号等）
- **异步日志**：高性能异步记录，避免阻塞主业务流程
- **上下文集成**：支持从上下文（如 HTTP 请求）中提取信息
- **文件轮转**：支持按大小、时间自动轮转日志文件

**示例：**

```go
// 创建日志器
log := logger.New()

// 记录不同级别的日志
log.Debug("调试信息")
log.Info("用户登录", logger.String("user_id", userId), logger.Int("login_count", 5))
log.Warn("配置即将过期", logger.Time("expiry", expiryTime))
log.Error("操作失败", logger.String("reason", "数据库连接错误"), logger.Error(err))

// 创建子日志器
requestLog := log.With(logger.String("request_id", requestID))
requestLog.Info("处理请求开始")
```

### Config 模块

Config 模块提供了一个灵活、强大的配置管理解决方案，基于Viper库实现，简化应用配置管理：

- **多格式支持**：支持YAML、TOML、JSON、环境变量等多种配置格式
- **多环境配置**：轻松切换开发、测试、预生产、生产环境配置
- **热重载**：支持配置动态更新和监听配置变更
- **配置验证**：内置配置参数校验，支持多种验证方式
- **默认值处理**：提供合理的默认配置值
- **配置优先级**：遵循命令行参数 > 环境变量 > 配置文件 > 默认值的优先顺序
- **配置变更通知**：支持组件订阅配置变更事件

**示例：**

```go
// 创建配置提供器
provider, err := config.NewProvider("./configs", "application")
if err != nil {
    panic(err)
}

// 读取配置值
serverPort := provider.GetInt("web.port", 8080)
databaseURL := provider.GetString("database.url", "")

// 绑定配置到结构体
var dbConfig DatabaseConfig
if err := provider.Unmarshal(&dbConfig); err != nil {
    panic(err)
}

// 监听配置变更
provider.OnChange(func() {
    log.Info("配置已更新")
    // 重新加载配置...
})
```

### 缓存模块

缓存模块提供了统一的缓存接口和多种实现，支持本地内存缓存和Redis缓存，以及多级缓存机制。

- **统一缓存接口**：通过`Cache`接口抽象不同的缓存实现
- **多级缓存**：支持本地缓存和分布式缓存的组合使用
- **Redis集成**：支持Redis的单机模式、哨兵模式和集群模式
- **自动序列化**：内置JSON和Gob序列化器，自动处理对象的序列化和反序列化
- **缓存策略**：支持TTL、标签等缓存管理策略
- **分布式锁**：提供基于Redis的分布式锁实现

**示例：**

```go
// 创建内存缓存
memCache := cache.NewMemoryCache()

// 设置缓存
ctx := context.Background()
memCache.Set(ctx, "user:123", user, time.Hour)

// 获取缓存
if value, found := memCache.Get(ctx, "user:123"); found {
    user := value.(*User)
    // 使用user...
}

// 删除缓存
memCache.Delete(ctx, "user:123")

// 使用标签
item := &cache.Item{
    Value:      user,
    Expiration: time.Hour,
    Tags:       []string{"user", "premium"},
}
memCache.SetItem(ctx, "user:123", item)

// 按标签删除
memCache.DeleteByTag(ctx, "premium")
```

### 分布式锁

分布式锁模块提供了基于Redis的分布式锁实现，用于协调分布式系统中的并发操作。

- **简单API**：提供简洁易用的锁接口
- **Redis实现**：基于Redis的可靠分布式锁实现，使用Lua脚本确保原子性操作
- **自动重试**：支持获取锁时的自动重试机制
- **锁刷新**：支持延长锁的过期时间
- **安全释放**：只有持有锁的客户端才能释放锁
- **缓存集成**：与cache模块无缝集成

**示例：**

```go
// 创建分布式锁
lock := lock.NewRedisLock(redisClient, "my-resource", lock.WithExpiry(time.Second*10))

// 尝试获取锁
ctx := context.Background()
if err := lock.Acquire(ctx); err != nil {
    // 获取锁失败
    return err
}

// 确保释放锁
defer lock.Release(ctx)

// 执行需要锁保护的操作
// ...

// 可选：如果操作时间较长，可以延长锁过期时间
if err := lock.Refresh(ctx, time.Second*10); err != nil {
    // 延长失败，可能锁已经过期
    return err
}
```

### DBStore

`dbstore` 包提供了基于 [GORM](https://gorm.io/) 的数据库操作封装，简化项目中的数据库操作，并集成了Go-Snap项目的日志、配置和错误处理系统。

- 支持多种数据库（MySQL、PostgreSQL、SQLite）
- 支持从配置文件加载配置
- 支持连接池管理和优化
- 集成项目日志系统
- 支持事务操作
- 提供分页查询
- 提供通用仓储接口，简化CRUD操作

**示例：**

```go
// 创建数据库连接
db, err := dbstore.New(dbConfig)
if err != nil {
    panic(err)
}

// 使用通用仓储
type User struct {
    ID   uint   `gorm:"primarykey"`
    Name string `gorm:"size:100"`
    Age  int
}

userRepo := dbstore.NewRepository[User](db)

// 创建
user := &User{Name: "John", Age: 30}
if err := userRepo.Create(ctx, user); err != nil {
    return err
}

// 查询
user, err := userRepo.FindByID(ctx, 1)
if err != nil {
    return err
}

// 更新
user.Age = 31
if err := userRepo.Update(ctx, user); err != nil {
    return err
}

// 删除
if err := userRepo.Delete(ctx, user); err != nil {
    return err
}

// 分页查询
page, err := userRepo.FindPage(ctx, dbstore.PageQuery{
    Page:     1,
    PageSize: 10,
    OrderBy:  "created_at desc",
    Where:    "age > ?",
    Args:     []interface{}{18},
})
```

### Web 模块

`web` 模块是 go-snap 框架的 HTTP 服务组件，基于 Gin 构建，提供了丰富的功能和简洁的 API，帮助你快速构建高性能的 Web 应用和 RESTful API。

- **路由管理**：支持 RESTful API 路由注册和分组
- **中间件机制**：丰富的中间件支持（CORS、JWT 验证、日志、限流等）
- **中间件洋葱模型**：请求 -> 中间件1(前) -> 中间件2(前) -> ... -> 处理函数 -> ... -> 中间件2(后) -> 中间件1(后) -> 响应
- **中间件分组**：支持全局中间件、路由组中间件和单个路由中间件的灵活组合
- **统一响应格式**：标准化响应格式（包含 code、message、data、request_id、timestamp 字段）
- **Context 对象池**：使用 sync.Pool 管理请求上下文对象，减少 GC 压力
- **参数验证**：请求参数校验和绑定，基于 validator/v10
- **WebSocket**：内置 WebSocket 支持
- **Swagger 集成**：自动生成 API 文档

**示例：**

```go
// 创建Web服务器
server := web.New(web.DefaultConfig())

// 添加全局中间件
server.Use(middleware.Recovery())
server.Use(middleware.Logger(logger))

// 添加路由
server.GET("/ping", func(c *gin.Context) {
    response.Success(c, "pong")
})

// 路由分组
api := server.Group("/api/v1")
api.Use(middleware.JWT(jwtSecret))

// 用户路由
users := api.Group("/users")
users.GET("", listUsers)
users.POST("", createUser)
users.GET("/:id", getUser)
users.PUT("/:id", updateUser)
users.DELETE("/:id", deleteUser)

// 启动服务器
if err := server.Start(); err != nil {
    panic(err)
}
```

### Boot 模块

`boot` 模块是 go-snap 框架的启动模块，负责应用程序的初始化、组件管理和生命周期控制。它实现了一个类似 Spring Boot 的自动配置和依赖注入系统，使应用程序的开发变得更加简单和标准化。

- **组件生命周期管理**：统一管理组件的初始化、启动和关闭
- **自动配置**：基于配置自动创建和配置组件
- **依赖注入**：自动解析和注入组件依赖
- **事件机制**：提供应用程序生命周期事件通知
- **启动顺序控制**：基于依赖关系自动确定组件启动顺序
- **健康检查**：内置组件健康检查机制
- **优雅关闭**：支持应用程序的优雅关闭，确保资源正确释放

**示例：**

```go
// 创建应用启动器
app := boot.NewBoot()

// 设置配置路径
app.SetConfigPath("./configs")

// 添加自定义组件
app.AddComponent(&MyComponent{})

// 添加插件
app.AddPlugin(&MyPlugin{})

// 添加配置器
app.AddConfigurer(&MyConfigurer{})

// 运行应用
if err := app.Run(); err != nil {
    panic(err)
}
```

## 扩展指南

### 创建自定义组件

1. 实现 `Component` 接口

```go
type MyComponent struct {
    // 组件状态和依赖
    db  *dbstore.DB
    log logger.Logger
}

// Name 返回组件名称
func (c *MyComponent) Name() string {
    return "myComponent"
}

// Initialize 初始化组件
func (c *MyComponent) Initialize(registry *boot.ComponentRegistry) error {
    // 获取依赖
    if db, exists := registry.GetComponent("db"); exists {
        c.db = db.(*dbstore.DB)
    }
    
    if log, exists := registry.GetComponent("logger"); exists {
        c.log = log.(logger.Logger)
    }
    
    // 初始化逻辑
    return nil
}

// Start 启动组件
func (c *MyComponent) Start() error {
    c.log.Info("Starting MyComponent")
    // 启动逻辑
    return nil
}

// Stop 停止组件
func (c *MyComponent) Stop() error {
    c.log.Info("Stopping MyComponent")
    // 停止逻辑，释放资源
    return nil
}
```

2. 注册组件

```go
app := boot.NewBoot()
app.AddComponent(&MyComponent{})
```

### 创建自定义中间件

```go
// 创建中间件
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理
        startTime := time.Now()
        
        // 继续处理请求
        c.Next()
        
        // 后置处理
        duration := time.Since(startTime)
        log.Info("Request processed", 
            logger.String("path", c.Request.URL.Path),
            logger.Int("status", c.Writer.Status()),
            logger.Duration("duration", duration),
        )
    }
}

// 使用中间件
server.Use(MyMiddleware())
```

### 创建自定义错误码

```go
// 定义错误码常量
const (
    // 用户相关错误码 (1000-1099)
    UserNotFound = 1000 + iota
    UserAlreadyExists
    InvalidUserData
)

// 注册错误码
func init() {
    errors.RegisterErrorCode(UserNotFound, http.StatusNotFound, "用户不存在", "")
    errors.RegisterErrorCode(UserAlreadyExists, http.StatusConflict, "用户已存在", "")
    errors.RegisterErrorCode(InvalidUserData, http.StatusBadRequest, "无效的用户数据", "")
}

// 使用错误码
func GetUser(id string) (*User, error) {
    user, err := userRepository.FindByID(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.WithCode(UserNotFound, "用户 %s 不存在", id)
        }
        return nil, errors.Wrap(err, "查询用户失败")
    }
    return user, nil
}
```

## 性能考虑

Go-Snap 框架在设计时充分考虑了性能因素：

1. **对象池**：使用对象池减少内存分配和垃圾回收压力
2. **异步处理**：日志、事件处理等支持异步模式
3. **连接池**：数据库、Redis等使用连接池优化资源使用
4. **缓存友好**：数据结构设计考虑CPU缓存友好性
5. **并发控制**：使用适当的并发控制机制，避免不必要的锁竞争
6. **按需加载**：组件和功能按需加载，减少不必要的资源消耗

### 性能优化建议

1. 在高并发场景中使用异步日志
2. 合理配置连接池大小
3. 使用内存缓存减少数据库访问
4. 根据实际需求调整组件配置
5. 使用性能分析工具定位瓶颈

## 版本兼容性

- Go 版本：1.16 或更高
- 依赖库版本兼容性：
  - gin: v1.8.0+
  - gorm: v1.23.0+
  - zap: v1.21.0+
  - viper: v1.10.0+
  - redis: v8.0.0+

## 常见问题

### Q: 如何管理多环境配置？

A: 使用环境特定的配置文件，如 `application-dev.yml`, `application-prod.yml`，并通过环境变量 `GO_ENV` 指定当前环境。

### Q: 如何处理数据库迁移？

A: Go-Snap 目前没有内置数据库迁移工具，建议使用 [golang-migrate](https://github.com/golang-migrate/migrate) 或 [goose](https://github.com/pressly/goose) 等工具管理数据库迁移。

### Q: 如何实现自定义验证规则？

A: Go-Snap 使用 validator/v10 进行参数验证，可以通过注册自定义验证器扩展验证规则。

```go
import "github.com/go-playground/validator/v10"

// 注册自定义验证器
validate := validator.New()
validate.RegisterValidation("is_valid_code", func(fl validator.FieldLevel) bool {
    code := fl.Field().String()
    // 验证逻辑
    return regexp.MustCompile(`^[A-Z0-9]{6}$`).MatchString(code)
})
```

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 提交问题和功能建议
- 提交 Pull Request 修复 bug 或添加新功能
- 改进文档
- 分享使用经验

### 贡献流程

1. Fork 仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 代码规范

- 遵循 Go 标准代码风格
- 添加适当的注释和文档
- 编写单元测试
- 确保所有测试通过
- 遵循语义化版本规范

## 许可证

本项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。