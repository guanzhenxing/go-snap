# 基础使用示例

本文档提供 Go-Snap 框架的基础使用示例，帮助您快速上手框架的核心功能。

## 🚀 Hello World 应用

### 1. 最简单的应用

```go
package main

import (
    "log"
    "github.com/guanzhenxing/go-snap/boot"
)

func main() {
    // 创建并运行应用
    app := boot.NewBoot()
    if err := app.Run(); err != nil {
        log.Fatalf("应用启动失败: %v", err)
    }
}
```

### 2. 带配置的应用

```go
package main

import (
    "log"
    "github.com/guanzhenxing/go-snap/boot"
)

func main() {
    app := boot.NewBoot().
        SetConfigPath("configs").           // 设置配置文件路径
        SetConfigName("application").       // 设置配置文件名
        SetConfigType("yaml")              // 设置配置文件类型
    
    if err := app.Run(); err != nil {
        log.Fatalf("应用启动失败: %v", err)
    }
}
```

配置文件 `configs/application.yaml`:

```yaml
app:
  name: "hello-world-app"
  version: "1.0.0"
  env: "development"

logger:
  enabled: true
  level: "info"
  json: false
```

## 📝 日志使用示例

### 1. 基础日志记录

```go
package main

import (
    "github.com/guanzhenxing/go-snap/boot"
    "github.com/guanzhenxing/go-snap/logger"
)

func main() {
    app := boot.NewBoot()
    application, err := app.Initialize()
    if err != nil {
        panic(err)
    }
    
    // 获取日志组件
    if loggerComp, found := application.GetComponent("logger"); found {
        if lc, ok := loggerComp.(*boot.LoggerComponent); ok {
            log := lc.GetLogger()
            
            // 基础日志记录
            log.Info("应用启动成功")
            log.Debug("调试信息")
            log.Warn("警告信息")
            log.Error("错误信息")
            
            // 结构化日志
            log.Info("用户登录",
                logger.String("username", "john"),
                logger.Int("user_id", 123),
                logger.Duration("login_time", time.Second*2),
            )
        }
    }
    
    // 启动应用
    if err := application.Start(); err != nil {
        panic(err)
    }
}
```

### 2. 自定义日志配置

配置文件:

```yaml
logger:
  enabled: true
  level: "debug"
  json: true
  file:
    enabled: true
    filename: "logs/app.log"
    max_size: 100
    max_backups: 3
    max_age: 28
    compress: true
  sampling:
    enabled: true
    initial: 100
    thereafter: 100
```

## ⚙️ 配置管理示例

### 1. 读取配置

```go
package main

import (
    "fmt"
    "github.com/guanzhenxing/go-snap/boot"
)

type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Username string `mapstructure:"username"`
    Password string `mapstructure:"password"`
    Database string `mapstructure:"database"`
}

func main() {
    app := boot.NewBoot().SetConfigPath("configs")
    application, _ := app.Initialize()
    
    // 获取配置组件
    if configComp, found := application.GetComponent("config"); found {
        if cc, ok := configComp.(*boot.ConfigComponent); ok {
            config := cc.GetConfig()
            
            // 读取简单配置
            appName := config.GetString("app.name")
            appPort := config.GetInt("server.port")
            
            fmt.Printf("应用名称: %s\n", appName)
            fmt.Printf("端口: %d\n", appPort)
            
            // 读取复杂配置
            var dbConfig DatabaseConfig
            if err := config.UnmarshalKey("database", &dbConfig); err != nil {
                panic(err)
            }
            
            fmt.Printf("数据库配置: %+v\n", dbConfig)
        }
    }
}
```

配置文件:

```yaml
app:
  name: "my-app"
  
server:
  port: 8080
  
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  database: "myapp"
```

### 2. 环境变量覆盖

```bash
# 通过环境变量覆盖配置
export APP_NAME="production-app"
export SERVER_PORT=9090
export DATABASE_HOST="prod-db.example.com"
```

## 💾 缓存使用示例

### 1. 基础缓存操作

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
            
            // 检查是否存在
            exists := cache.Exists(ctx, "user:123")
            fmt.Printf("缓存存在: %t\n", exists)
            
            // 删除缓存
            cache.Delete(ctx, "user:123")
        }
    }
}
```

### 2. 复杂对象缓存

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func cacheUserExample(cache cache.Cache) {
    ctx := context.Background()
    
    user := &User{
        ID:    123,
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    // 缓存用户对象
    err := cache.Set(ctx, "user:123", user, time.Hour)
    if err != nil {
        panic(err)
    }
    
    // 获取用户对象
    value, found := cache.Get(ctx, "user:123")
    if found {
        if cachedUser, ok := value.(*User); ok {
            fmt.Printf("缓存的用户: %+v\n", cachedUser)
        }
    }
}
```

缓存配置:

```yaml
cache:
  enabled: true
  type: "memory"
  memory:
    max_entries: 10000
    cleanup_interval: "5m"
```

## 🌐 Web 服务示例

### 1. 简单的 HTTP 服务器

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/guanzhenxing/go-snap/boot"
)

func main() {
    app := boot.NewBoot().
        SetConfigPath("configs").
        AddConfigurer(func(application *boot.Application) error {
            // 配置 Web 路由
            if webComp, found := application.GetComponent("web"); found {
                if wc, ok := webComp.(*boot.WebComponent); ok {
                    router := wc.GetRouter()
                    
                    // 添加路由
                    router.GET("/", func(c *gin.Context) {
                        c.JSON(http.StatusOK, gin.H{
                            "message": "Hello, Go-Snap!",
                        })
                    })
                    
                    router.GET("/health", func(c *gin.Context) {
                        c.JSON(http.StatusOK, gin.H{
                            "status": "healthy",
                        })
                    })
                }
            }
            return nil
        })
    
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

Web 配置:

```yaml
web:
  enabled: true
  port: 8080
  mode: "debug"
```

### 2. RESTful API 示例

```go
package main

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/guanzhenxing/go-snap/boot"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

var users = []User{
    {ID: 1, Name: "Alice"},
    {ID: 2, Name: "Bob"},
}

func main() {
    app := boot.NewBoot().
        SetConfigPath("configs").
        AddConfigurer(func(application *boot.Application) error {
            if webComp, found := application.GetComponent("web"); found {
                if wc, ok := webComp.(*boot.WebComponent); ok {
                    router := wc.GetRouter()
                    
                    // 用户 API 路由组
                    userAPI := router.Group("/api/users")
                    {
                        userAPI.GET("", getUsers)
                        userAPI.GET("/:id", getUser)
                        userAPI.POST("", createUser)
                    }
                }
            }
            return nil
        })
    
    if err := app.Run(); err != nil {
        panic(err)
    }
}

func getUsers(c *gin.Context) {
    c.JSON(http.StatusOK, users)
}

func getUser(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    
    for _, user := range users {
        if user.ID == id {
            c.JSON(http.StatusOK, user)
            return
        }
    }
    
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func createUser(c *gin.Context) {
    var newUser User
    if err := c.ShouldBindJSON(&newUser); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    newUser.ID = len(users) + 1
    users = append(users, newUser)
    
    c.JSON(http.StatusCreated, newUser)
}
```

## 🔧 自定义组件示例

### 1. 创建自定义组件

```go
package main

import (
    "context"
    "fmt"
    "github.com/guanzhenxing/go-snap/boot"
)

// 自定义服务组件
type EmailService struct {
    *boot.BaseComponent
    smtpHost string
    smtpPort int
}

func NewEmailService() *EmailService {
    return &EmailService{
        BaseComponent: boot.NewBaseComponent("emailService", boot.ComponentTypeService),
    }
}

func (s *EmailService) Initialize(ctx context.Context) error {
    // 从配置中读取 SMTP 设置
    s.smtpHost = "smtp.example.com"
    s.smtpPort = 587
    
    s.SetStatus(boot.ComponentStatusInitialized)
    return nil
}

func (s *EmailService) Start(ctx context.Context) error {
    fmt.Println("邮件服务启动")
    s.SetStatus(boot.ComponentStatusRunning)
    return nil
}

func (s *EmailService) Stop(ctx context.Context) error {
    fmt.Println("邮件服务停止")
    s.SetStatus(boot.ComponentStatusStopped)
    return nil
}

func (s *EmailService) SendEmail(to, subject, body string) error {
    fmt.Printf("发送邮件到 %s: %s\n", to, subject)
    return nil
}

// 自定义组件工厂
type EmailServiceFactory struct{}

func (f *EmailServiceFactory) Create(config interface{}) (boot.Component, error) {
    return NewEmailService(), nil
}

func (f *EmailServiceFactory) ValidateConfig(config interface{}) error {
    return nil
}

func (f *EmailServiceFactory) GetConfigSchema() *boot.ConfigSchema {
    return &boot.ConfigSchema{
        Type: "object",
        Properties: map[string]*boot.PropertySchema{
            "smtp_host": {Type: "string", Description: "SMTP 服务器地址"},
            "smtp_port": {Type: "integer", Description: "SMTP 端口"},
        },
    }
}

func main() {
    app := boot.NewBoot().
        SetConfigPath("configs").
        AddComponent("emailService", &EmailServiceFactory{}).
        AddConfigurer(func(application *boot.Application) error {
            // 使用自定义组件
            if emailComp, found := application.GetComponent("emailService"); found {
                if es, ok := emailComp.(*EmailService); ok {
                    es.SendEmail("user@example.com", "欢迎", "欢迎使用我们的服务！")
                }
            }
            return nil
        })
    
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

### 2. 组件依赖注入

```go
// 依赖其他组件的服务
type UserService struct {
    *boot.BaseComponent
    logger logger.Logger
    cache  cache.Cache
}

func NewUserService() *UserService {
    return &UserService{
        BaseComponent: boot.NewBaseComponent("userService", boot.ComponentTypeService),
    }
}

func (s *UserService) Initialize(ctx context.Context) error {
    // 这里可以注入依赖的组件
    s.SetStatus(boot.ComponentStatusInitialized)
    return nil
}

func (s *UserService) SetDependencies(logger logger.Logger, cache cache.Cache) {
    s.logger = logger
    s.cache = cache
}

func (s *UserService) GetUser(userID string) (*User, error) {
    // 先从缓存获取
    if value, found := s.cache.Get(context.Background(), "user:"+userID); found {
        if user, ok := value.(*User); ok {
            s.logger.Info("从缓存获取用户", logger.String("user_id", userID))
            return user, nil
        }
    }
    
    // 从数据库获取（模拟）
    user := &User{ID: 1, Name: "John"}
    
    // 缓存用户信息
    s.cache.Set(context.Background(), "user:"+userID, user, time.Hour)
    
    s.logger.Info("从数据库获取用户", logger.String("user_id", userID))
    return user, nil
}
```

## 🔍 错误处理示例

### 1. 基础错误处理

```go
package main

import (
    "fmt"
    "github.com/guanzhenxing/go-snap/errors"
)

func getUserExample(userID string) (*User, error) {
    if userID == "" {
        return nil, errors.NewValidationError("user_id", "用户ID不能为空")
    }
    
    // 模拟数据库查询
    if userID == "999" {
        return nil, errors.NewUserError("用户不存在", errors.CodeUserNotFound)
    }
    
    return &User{ID: 1, Name: "John"}, nil
}

func main() {
    // 测试错误处理
    user, err := getUserExample("")
    if err != nil {
        fmt.Printf("错误类型: %T\n", err)
        fmt.Printf("错误码: %s\n", errors.GetCode(err))
        fmt.Printf("错误消息: %s\n", err.Error())
    }
    
    user, err = getUserExample("999")
    if err != nil {
        // 根据错误类型处理
        switch errors.GetCode(err) {
        case errors.CodeUserNotFound:
            fmt.Println("用户不存在，创建新用户")
        case errors.CodeValidation:
            fmt.Println("输入验证失败")
        default:
            fmt.Printf("未知错误: %v\n", err)
        }
    }
}
```

### 2. 错误包装和上下文

```go
func processUserOrder(userID, orderID string) error {
    user, err := getUserExample(userID)
    if err != nil {
        return errors.WithContext(err, map[string]interface{}{
            "user_id":  userID,
            "order_id": orderID,
            "operation": "process_order",
        })
    }
    
    // 处理订单逻辑
    if orderID == "invalid" {
        return errors.NewOrderError("订单无效", errors.CodeValidation)
    }
    
    return nil
}
```

## 📊 健康检查示例

### 1. 应用健康检查

```go
func main() {
    app := boot.NewBoot().
        SetConfigPath("configs").
        AddConfigurer(func(application *boot.Application) error {
            if webComp, found := application.GetComponent("web"); found {
                if wc, ok := webComp.(*boot.WebComponent); ok {
                    router := wc.GetRouter()
                    
                    // 健康检查端点
                    router.GET("/health", func(c *gin.Context) {
                        healthStatus := application.GetHealthStatus()
                        
                        status := "healthy"
                        httpStatus := http.StatusOK
                        
                        if healthStatus.Status != boot.HealthStatusHealthy {
                            status = "unhealthy"
                            httpStatus = http.StatusServiceUnavailable
                        }
                        
                        c.JSON(httpStatus, gin.H{
                            "status": status,
                            "components": healthStatus.Components,
                            "timestamp": healthStatus.Timestamp,
                        })
                    })
                }
            }
            return nil
        })
    
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

## 🚀 完整应用示例

### 项目结构

```
my-app/
├── main.go
├── configs/
│   ├── application.yaml
│   ├── application-dev.yaml
│   └── application-prod.yaml
├── internal/
│   ├── service/
│   │   └── user_service.go
│   └── handler/
│       └── user_handler.go
└── go.mod
```

### main.go

```go
package main

import (
    "log"
    "github.com/guanzhenxing/go-snap/boot"
    "my-app/internal/service"
    "my-app/internal/handler"
)

func main() {
    app := boot.NewBoot().
        SetConfigPath("configs").
        AddComponent("userService", &service.UserServiceFactory{}).
        AddConfigurer(func(application *boot.Application) error {
            return handler.SetupRoutes(application)
        })
    
    if err := app.Run(); err != nil {
        log.Fatalf("应用启动失败: %v", err)
    }
}
```

### configs/application.yaml

```yaml
app:
  name: "user-management-api"
  version: "1.0.0"

web:
  enabled: true
  port: 8080

logger:
  enabled: true
  level: "info"
  json: false

cache:
  enabled: true
  type: "memory"

database:
  enabled: true
  driver: "sqlite"
  dsn: "users.db"
```

这个基础使用示例涵盖了 Go-Snap 框架的主要功能，帮助您快速上手。更多高级功能请参考 [高级使用示例](advanced-usage.md)。 