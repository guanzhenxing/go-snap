# Logger 模块

Logger 模块是 Go-Snap 框架的高性能日志组件，基于 [Uber Zap](https://github.com/uber-go/zap) 构建，提供结构化、高性能的日志记录功能。

## 概述

Logger 模块提供了企业级的日志解决方案，支持多级别日志、结构化输出、异步记录、文件轮转等功能，并与框架的配置系统和组件系统无缝集成。

### 核心特性

- ✅ **高性能日志** - 基于 Zap 的零分配日志记录
- ✅ **结构化日志** - 支持字段化日志输出
- ✅ **多级别日志** - Debug、Info、Warn、Error、Fatal 等级别
- ✅ **多输出目标** - 控制台、文件、JSON 格式等
- ✅ **异步日志** - 高性能异步记录，避免阻塞主业务
- ✅ **文件轮转** - 自动按大小、时间轮转日志文件
- ✅ **上下文集成** - 支持从上下文中提取信息
- ✅ **数据掩码** - 自动掩码敏感信息
- ✅ **配置驱动** - 通过配置文件灵活控制日志行为

## 快速开始

### 基础用法

```go
package main

import (
    "github.com/guanzhenxing/go-snap/boot"
    "github.com/guanzhenxing/go-snap/logger"
)

func main() {
    // 启动应用（会自动配置日志组件）
    app := boot.NewBoot()
    application, _ := app.Initialize()
    
    // 获取日志组件
    if loggerComp, found := application.GetComponent("logger"); found {
        if lc, ok := loggerComp.(*boot.LoggerComponent); ok {
            log := lc.GetLogger()
            
            // 使用日志
            log.Info("应用已启动")
            log.Error("发生错误", logger.String("module", "main"))
        }
    }
}
```

### 直接使用Logger

```go
import "github.com/guanzhenxing/go-snap/logger"

// 创建默认日志器
log := logger.New()

// 基础日志记录
log.Info("这是一条信息日志")
log.Error("这是一条错误日志")

// 结构化日志
log.Info("用户登录", 
    logger.String("username", "john"),
    logger.Int("user_id", 123),
    logger.Duration("response_time", time.Millisecond*50),
)
```

## 配置

### 配置文件示例

```yaml
# 日志配置
logger:
  enabled: true                    # 是否启用日志
  level: "info"                   # 日志级别: debug, info, warn, error, fatal
  json: false                     # 是否使用JSON格式输出
  
  # 控制台输出配置
  console:
    enabled: true                 # 是否启用控制台输出
    color: true                   # 是否启用颜色输出
    
  # 文件输出配置
  file:
    enabled: true                 # 是否启用文件输出
    path: "logs/app.log"         # 日志文件路径
    max_size: 100                # 单个文件最大大小(MB)
    max_backups: 10              # 保留的备份文件数量
    max_age: 30                  # 保留天数
    compress: true               # 是否压缩备份文件
    
  # 开发模式配置
  development: false             # 是否启用开发模式
  
  # 采样配置（生产环境推荐）
  sampling:
    enabled: false               # 是否启用采样
    initial: 100                 # 初始采样数
    thereafter: 100              # 后续采样间隔
```

### 环境变量配置

```bash
# 通过环境变量覆盖配置
export LOGGER_LEVEL=debug
export LOGGER_JSON=true
export LOGGER_FILE_PATH=/var/log/myapp.log
```

## API 参考

### 日志级别

```go
// 日志级别常量
const (
    DebugLevel = zap.DebugLevel   // 调试级别
    InfoLevel  = zap.InfoLevel    // 信息级别
    WarnLevel  = zap.WarnLevel    // 警告级别
    ErrorLevel = zap.ErrorLevel   // 错误级别
    FatalLevel = zap.FatalLevel   // 致命错误级别
)

// 解析日志级别
level, err := logger.ParseLevel("info")
```

### 基础日志方法

```go
log := logger.New()

// 基础方法
log.Debug("调试信息")
log.Info("普通信息")
log.Warn("警告信息")
log.Error("错误信息")
log.Fatal("致命错误") // 会调用 os.Exit(1)

// 格式化方法
log.Debugf("调试信息: %s", value)
log.Infof("用户 %s 登录成功", username)
log.Errorf("处理请求失败: %v", err)
```

### 结构化日志

```go
// 字段类型
logger.String("key", "value")           // 字符串字段
logger.Int("count", 42)                 // 整数字段
logger.Int64("timestamp", time.Now().Unix()) // 64位整数
logger.Float64("price", 99.99)          // 浮点数字段
logger.Bool("success", true)            // 布尔字段
logger.Duration("latency", time.Millisecond*100) // 时间间隔
logger.Time("created_at", time.Now())   // 时间字段
logger.Error(err)                       // 错误字段
logger.Any("data", complexObject)      // 任意类型

// 使用示例
log.Info("处理请求",
    logger.String("method", "GET"),
    logger.String("path", "/api/users"),
    logger.Int("status_code", 200),
    logger.Duration("response_time", time.Millisecond*50),
    logger.String("user_agent", "Go-Client/1.0"),
)
```

### 上下文日志

```go
// 创建带上下文的日志器
contextLogger := log.With(
    logger.String("request_id", "12345"),
    logger.String("user_id", "user123"),
)

// 使用上下文日志器
contextLogger.Info("开始处理请求")
contextLogger.Info("查询数据库")
contextLogger.Info("返回响应")
```

### 子日志器

```go
// 创建子日志器
userLogger := log.Named("user")
orderLogger := log.Named("order")

userLogger.Info("用户操作")   // 输出: [user] 用户操作
orderLogger.Info("订单操作") // 输出: [order] 订单操作
```

## 高级功能

### 自定义配置

```go
// 创建自定义配置的日志器
log := logger.New(
    logger.WithLevel(logger.DebugLevel),
    logger.WithJSON(true),
    logger.WithFilename("/var/log/app.log"),
    logger.WithMaxSize(100),
    logger.WithMaxBackups(10),
    logger.WithMaxAge(30),
    logger.WithCompress(true),
    logger.WithColorConsole(false),
)
```

### 配置选项说明

```go
// 可用的配置选项
logger.WithLevel(level)              // 设置日志级别
logger.WithJSON(enable)              // 启用JSON格式
logger.WithJSONConsole(enable)       // 控制台JSON输出
logger.WithFilename(path)            // 设置日志文件路径
logger.WithMaxSize(mb)               // 设置文件最大大小
logger.WithMaxBackups(count)         // 设置备份文件数量
logger.WithMaxAge(days)              // 设置文件保留天数
logger.WithCompress(enable)          // 启用文件压缩
logger.WithColorConsole(enable)      // 启用控制台颜色
logger.WithDevelopment(enable)       // 启用开发模式
logger.WithSampling(initial, thereafter) // 配置采样
```

### 文件轮转

```go
// 配置文件轮转
log := logger.New(
    logger.WithFilename("logs/app.log"),
    logger.WithMaxSize(100),        // 100MB后轮转
    logger.WithMaxBackups(10),      // 保留10个备份文件
    logger.WithMaxAge(30),          // 保留30天
    logger.WithCompress(true),      // 压缩旧文件
)
```

### 采样配置

```go
// 在高并发环境下使用采样减少日志量
log := logger.New(
    logger.WithSampling(100, 100), // 前100条全记录，之后每100条记录1条
)
```

### 开发模式

```go
// 开发模式：更友好的输出格式
log := logger.New(
    logger.WithDevelopment(true),
    logger.WithLevel(logger.DebugLevel),
)
```

## 最佳实践

### 1. 结构化日志

```go
// ✅ 好的做法：使用结构化字段
log.Info("用户登录",
    logger.String("username", username),
    logger.String("ip", clientIP),
    logger.Duration("login_time", loginDuration),
)

// ❌ 避免的做法：字符串拼接
log.Infof("用户 %s 从 %s 登录，耗时 %v", username, clientIP, loginDuration)
```

### 2. 错误日志

```go
// ✅ 好的做法：记录错误详情
if err := processUser(userID); err != nil {
    log.Error("处理用户失败",
        logger.String("user_id", userID),
        logger.String("operation", "process"),
        logger.Error(err),
        logger.String("stack", string(debug.Stack())), // 在需要时添加堆栈
    )
    return err
}

// ❌ 避免的做法：缺少上下文
log.Error("处理失败", logger.Error(err))
```

### 3. 性能敏感场景

```go
// ✅ 使用采样减少日志量
log := logger.New(logger.WithSampling(100, 100))

// ✅ 使用条件日志
if log.Level() <= logger.DebugLevel {
    log.Debug("详细调试信息", logger.Any("data", expensiveOperation()))
}

// ✅ 避免在循环中记录大量日志
for i, item := range items {
    if i%1000 == 0 { // 每1000条记录一次
        log.Info("处理进度", logger.Int("processed", i), logger.Int("total", len(items)))
    }
}
```

### 4. 敏感信息处理

```go
// ✅ 掩码敏感信息
log.Info("用户注册",
    logger.String("email", maskEmail(email)),
    logger.String("phone", maskPhone(phone)),
    logger.String("password", "***"), // 不记录密码
)

// 掩码函数示例
func maskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    username := parts[0]
    if len(username) > 2 {
        username = username[:2] + "***"
    }
    return username + "@" + parts[1]
}
```

### 5. 上下文传递

```go
// ✅ 在HTTP处理中传递上下文
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    requestID := r.Header.Get("X-Request-ID")
    
    // 创建带请求ID的日志器
    reqLogger := h.logger.With(logger.String("request_id", requestID))
    
    reqLogger.Info("开始处理获取用户请求")
    
    // 传递给服务层
    user, err := h.userService.GetUser(r.Context(), userID, reqLogger)
    if err != nil {
        reqLogger.Error("获取用户失败", logger.Error(err))
        http.Error(w, "Internal Error", 500)
        return
    }
    
    reqLogger.Info("成功获取用户", logger.String("user_id", user.ID))
}
```

## 集成示例

### HTTP 中间件

```go
func LoggingMiddleware(logger logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 处理请求
        c.Next()
        
        // 记录请求日志
        logger.Info("HTTP请求",
            logger.String("method", c.Request.Method),
            logger.String("path", c.Request.URL.Path),
            logger.Int("status", c.Writer.Status()),
            logger.Duration("latency", time.Since(start)),
            logger.String("client_ip", c.ClientIP()),
            logger.String("user_agent", c.Request.UserAgent()),
        )
    }
}
```

### 数据库日志

```go
// GORM日志集成
import "gorm.io/gorm/logger"

type GormLogger struct {
    logger logger.Logger
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
    return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
    l.logger.Info(msg, logger.Any("data", data))
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
    l.logger.Warn(msg, logger.Any("data", data))
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
    l.logger.Error(msg, logger.Any("data", data))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()
    
    fields := []logger.Field{
        logger.Duration("elapsed", elapsed),
        logger.String("sql", sql),
        logger.Int64("rows", rows),
    }
    
    if err != nil {
        fields = append(fields, logger.Error(err))
        l.logger.Error("数据库查询错误", fields...)
    } else {
        l.logger.Debug("数据库查询", fields...)
    }
}
```

### 业务组件集成

```go
type UserService struct {
    logger logger.Logger
    db     *gorm.DB
}

func NewUserService(logger logger.Logger, db *gorm.DB) *UserService {
    return &UserService{
        logger: logger.Named("user_service"),
        db:     db,
    }
}

func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    s.logger.Info("开始创建用户",
        logger.String("username", user.Username),
        logger.String("email", maskEmail(user.Email)),
    )
    
    if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
        s.logger.Error("创建用户失败",
            logger.String("username", user.Username),
            logger.Error(err),
        )
        return err
    }
    
    s.logger.Info("成功创建用户",
        logger.String("user_id", user.ID),
        logger.String("username", user.Username),
    )
    
    return nil
}
```

## 性能优化

### 1. 采样配置

```yaml
logger:
  sampling:
    enabled: true
    initial: 100      # 前100条全记录
    thereafter: 100   # 之后每100条记录1条
```

### 2. 异步写入

Logger 模块默认使用异步写入，无需额外配置。

### 3. 生产环境优化

```yaml
# 生产环境推荐配置
logger:
  level: "info"
  json: true
  development: false
  sampling:
    enabled: true
    initial: 100
    thereafter: 100
  file:
    enabled: true
    path: "/var/log/app.log"
    max_size: 100
    max_backups: 10
    max_age: 7
    compress: true
```

## 故障排除

### 常见问题

#### 1. 日志文件没有生成

**检查项**:
- 确认文件路径有写入权限
- 检查配置中的 `file.enabled` 是否为 `true`
- 验证日志级别设置

#### 2. 日志输出格式异常

**检查项**:
- 验证 `json` 配置项
- 检查是否启用了开发模式
- 确认控制台颜色设置

#### 3. 日志量过大

**解决方案**:
- 调整日志级别为 `info` 或更高
- 启用采样功能
- 配置合适的文件轮转参数

### 调试技巧

```go
// 检查日志器配置
log := logger.New()
fmt.Printf("当前日志级别: %s\n", log.Level())

// 测试不同级别的日志
log.Debug("这是调试日志")
log.Info("这是信息日志")
log.Warn("这是警告日志")
log.Error("这是错误日志")
```

## 参考资料

- [Uber Zap 官方文档](https://pkg.go.dev/go.uber.org/zap)
- [Lumberjack 日志轮转](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2)
- [Go-Snap 配置文档](config.md)
- [Go-Snap 架构设计](../architecture.md) 