# Errors 模块

Errors 模块是 Go-Snap 框架的统一错误处理组件，提供结构化的错误类型定义、错误包装、错误链追踪等功能，帮助开发者构建健壮的错误处理机制。

## 概述

Errors 模块设计用于标准化应用中的错误处理，提供了丰富的错误类型、详细的错误上下文和完整的错误链支持。它与框架的其他组件无缝集成，为应用提供一致的错误处理体验。

### 核心特性

- ✅ **结构化错误** - 提供层次化的错误类型定义
- ✅ **错误包装** - 支持错误链和错误上下文传递
- ✅ **错误码支持** - 内置错误码和错误分类
- ✅ **堆栈追踪** - 自动记录错误发生的调用栈
- ✅ **多语言支持** - 支持错误消息的国际化
- ✅ **JSON 序列化** - 错误对象可直接序列化为JSON
- ✅ **HTTP 集成** - 自动转换为合适的HTTP状态码
- ✅ **日志集成** - 与日志系统无缝集成

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/guanzhenxing/go-snap/errors"
)

func main() {
    // 创建简单错误
    err := errors.New("something went wrong")
    fmt.Printf("错误: %v\n", err)
    
    // 创建带错误码的错误
    err = errors.NewWithCode(errors.CodeValidation, "输入验证失败")
    fmt.Printf("错误码: %s, 消息: %s\n", err.Code(), err.Error())
    
    // 包装错误
    originalErr := fmt.Errorf("原始错误")
    wrappedErr := errors.Wrap(originalErr, "包装错误消息")
    fmt.Printf("包装错误: %v\n", wrappedErr)
}
```

### 错误类型层次

```go
// 框架定义的错误类型层次
ConfigError              // 配置相关错误
├── ComponentError       // 组件错误
├── DependencyError      // 依赖错误
└── ValidationError      // 验证错误

BusinessError            // 业务逻辑错误
├── UserError           // 用户相关错误
├── OrderError          // 订单相关错误
└── PaymentError        // 支付相关错误

SystemError             // 系统错误
├── DatabaseError       // 数据库错误
├── NetworkError        // 网络错误
└── IOError            // IO错误
```

## 错误类型

### 基础错误

```go
import "github.com/guanzhenxing/go-snap/errors"

// 创建基础错误
err := errors.New("基础错误消息")

// 创建格式化错误
err = errors.Errorf("用户 %s 不存在", username)

// 创建带错误码的错误
err = errors.NewWithCode(errors.CodeNotFound, "资源未找到")
```

### 配置错误

```go
// 配置错误 - 配置文件解析、验证失败等
configErr := errors.NewConfigError("config.yaml", "解析配置文件失败", originalErr)

// 组件错误 - 组件初始化、启动失败等
componentErr := errors.NewComponentError("logger", "initialize", "初始化日志组件失败", originalErr)

// 依赖错误 - 组件依赖关系错误
depErr := errors.NewDependencyError("userService", []string{"database", "cache"}, "依赖组件不可用")

// 验证错误 - 配置参数验证失败
validationErr := errors.NewValidationError("database.port", "端口号必须在1-65535之间")
```

### 业务错误

```go
// 用户错误
userErr := errors.NewUserError("用户名已存在", errors.CodeAlreadyExists)

// 订单错误
orderErr := errors.NewOrderError("订单状态不允许取消", errors.CodeInvalidState)

// 支付错误
paymentErr := errors.NewPaymentError("余额不足", errors.CodeInsufficientFunds)
```

### 系统错误

```go
// 数据库错误
dbErr := errors.NewDatabaseError("users", "insert", "插入用户记录失败", originalErr)

// 网络错误
netErr := errors.NewNetworkError("api.example.com", "连接超时", originalErr)

// IO错误
ioErr := errors.NewIOError("/path/to/file", "read", "读取文件失败", originalErr)
```

## 错误码系统

### 预定义错误码

```go
const (
    // 成功
    CodeSuccess = "SUCCESS"
    
    // 客户端错误 (4xx)
    CodeBadRequest          = "BAD_REQUEST"           // 400
    CodeUnauthorized        = "UNAUTHORIZED"          // 401
    CodeForbidden          = "FORBIDDEN"              // 403
    CodeNotFound           = "NOT_FOUND"              // 404
    CodeMethodNotAllowed   = "METHOD_NOT_ALLOWED"     // 405
    CodeConflict          = "CONFLICT"                // 409
    CodeValidation        = "VALIDATION_ERROR"        // 422
    CodeTooManyRequests   = "TOO_MANY_REQUESTS"      // 429
    
    // 服务器错误 (5xx)
    CodeInternalError     = "INTERNAL_ERROR"          // 500
    CodeNotImplemented    = "NOT_IMPLEMENTED"         // 501
    CodeServiceUnavailable = "SERVICE_UNAVAILABLE"    // 503
    CodeTimeout           = "TIMEOUT"                 // 504
    
    // 业务错误
    CodeUserNotFound      = "USER_NOT_FOUND"
    CodeUserAlreadyExists = "USER_ALREADY_EXISTS"
    CodeInvalidCredentials = "INVALID_CREDENTIALS"
    CodeInsufficientFunds = "INSUFFICIENT_FUNDS"
    CodeOrderNotFound     = "ORDER_NOT_FOUND"
    CodeInvalidState      = "INVALID_STATE"
    
    // 系统错误
    CodeDatabaseError     = "DATABASE_ERROR"
    CodeCacheError        = "CACHE_ERROR"
    CodeNetworkError      = "NETWORK_ERROR"
    CodeConfigError       = "CONFIG_ERROR"
)
```

### 自定义错误码

```go
// 定义自定义错误码
const (
    CodeCustomBusiness = "CUSTOM_BUSINESS_ERROR"
    CodeCustomSystem   = "CUSTOM_SYSTEM_ERROR"
)

// 注册错误码（可选，用于验证和文档生成）
errors.RegisterCode(CodeCustomBusiness, "自定义业务错误", 400)
errors.RegisterCode(CodeCustomSystem, "自定义系统错误", 500)
```

## 错误包装和链

### 错误包装

```go
// 包装错误，添加上下文信息
originalErr := fmt.Errorf("数据库连接失败")
wrappedErr := errors.Wrap(originalErr, "初始化用户服务失败")

// 包装错误并添加错误码
wrappedErr = errors.WrapWithCode(originalErr, errors.CodeDatabaseError, "用户服务启动失败")

// 格式化包装
wrappedErr = errors.Wrapf(originalErr, "处理用户 %s 的请求失败", username)
```

### 错误链遍历

```go
// 检查错误链中是否包含特定错误
if errors.Is(err, ErrUserNotFound) {
    // 处理用户未找到错误
}

// 从错误链中提取特定类型的错误
var configErr *errors.ConfigError
if errors.As(err, &configErr) {
    fmt.Printf("配置文件: %s, 错误: %s\n", configErr.ConfigFile, configErr.Error())
}

// 展开错误链
cause := errors.Unwrap(err)
for cause != nil {
    fmt.Printf("原因: %v\n", cause)
    cause = errors.Unwrap(cause)
}
```

## 错误上下文

### 添加错误上下文

```go
// 创建带上下文的错误
err := errors.NewWithContext(map[string]interface{}{
    "user_id":    123,
    "operation":  "create_order",
    "request_id": "req-12345",
    "timestamp":  time.Now(),
}, "创建订单失败")

// 为现有错误添加上下文
err = errors.WithContext(originalErr, map[string]interface{}{
    "retry_count": 3,
    "last_error":  lastErr.Error(),
})
```

### 获取错误上下文

```go
// 获取错误上下文
if ctxErr, ok := err.(*errors.ContextError); ok {
    context := ctxErr.Context()
    userID := context["user_id"]
    operation := context["operation"]
}

// 获取所有上下文（包括错误链中的上下文）
allContext := errors.GetAllContext(err)
```

## 堆栈追踪

### 启用堆栈追踪

```go
// 创建带堆栈信息的错误
err := errors.NewWithStack("发生错误")

// 为现有错误添加堆栈信息
err = errors.WithStack(originalErr)

// 在特定位置记录堆栈
err = errors.WithStackSkip(originalErr, 1) // 跳过1层调用栈
```

### 获取堆栈信息

```go
// 获取格式化的堆栈信息
if stackErr, ok := err.(errors.StackTracer); ok {
    stack := stackErr.StackTrace()
    fmt.Printf("堆栈信息:\n%+v\n", stack)
}

// 获取简化的堆栈信息
stackTrace := errors.GetStackTrace(err)
for _, frame := range stackTrace {
    fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
}
```

## HTTP 集成

### 错误码到HTTP状态码映射

```go
// 自动转换错误码为HTTP状态码
httpStatus := errors.ToHTTPStatus(err)

// 预定义的映射关系
var defaultStatusMap = map[string]int{
    CodeSuccess:           200,
    CodeBadRequest:        400,
    CodeUnauthorized:      401,
    CodeForbidden:         403,
    CodeNotFound:          404,
    CodeValidation:        422,
    CodeInternalError:     500,
    CodeServiceUnavailable: 503,
}

// 自定义映射关系
errors.SetStatusMapping(CodeCustomError, 418) // I'm a teapot
```

### HTTP错误响应

```go
// HTTP错误响应结构
type ErrorResponse struct {
    Code      string                 `json:"code"`
    Message   string                 `json:"message"`
    Details   interface{}           `json:"details,omitempty"`
    Context   map[string]interface{} `json:"context,omitempty"`
    Timestamp time.Time             `json:"timestamp"`
    RequestID string                `json:"request_id,omitempty"`
}

// 转换错误为HTTP响应
func ToHTTPResponse(err error, requestID string) *ErrorResponse {
    return errors.ToHTTPResponse(err, requestID)
}
```

### Gin 集成示例

```go
// Gin 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            // 转换为HTTP响应
            response := errors.ToHTTPResponse(err, c.GetString("request_id"))
            status := errors.ToHTTPStatus(err)
            
            c.JSON(status, response)
        }
    }
}

// 在处理函数中使用
func GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    user, err := userService.GetUser(userID)
    if err != nil {
        // 直接返回错误，由中间件处理
        c.Error(err)
        return
    }
    
    c.JSON(200, user)
}
```

## 日志集成

### 结构化错误日志

```go
import "github.com/guanzhenxing/go-snap/logger"

// 记录错误日志
func LogError(log logger.Logger, err error) {
    fields := []logger.Field{
        logger.String("error_code", errors.GetCode(err)),
        logger.String("error_message", err.Error()),
    }
    
    // 添加错误上下文
    if context := errors.GetAllContext(err); len(context) > 0 {
        fields = append(fields, logger.Any("error_context", context))
    }
    
    // 添加堆栈信息（仅在调试模式下）
    if log.Level() <= logger.DebugLevel {
        if stack := errors.GetStackTrace(err); len(stack) > 0 {
            fields = append(fields, logger.Any("stack_trace", stack))
        }
    }
    
    log.Error("应用错误", fields...)
}
```

### 错误监控

```go
// 错误统计
type ErrorStats struct {
    Count    int64                    `json:"count"`
    LastSeen time.Time               `json:"last_seen"`
    Codes    map[string]int64        `json:"codes"`
    Types    map[string]int64        `json:"types"`
}

// 错误监控器
type ErrorMonitor struct {
    stats map[string]*ErrorStats
    mutex sync.RWMutex
}

func (m *ErrorMonitor) Record(err error) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    code := errors.GetCode(err)
    errType := errors.GetType(err)
    
    if m.stats[code] == nil {
        m.stats[code] = &ErrorStats{
            Codes: make(map[string]int64),
            Types: make(map[string]int64),
        }
    }
    
    stats := m.stats[code]
    stats.Count++
    stats.LastSeen = time.Now()
    stats.Codes[code]++
    stats.Types[errType]++
}
```

## 最佳实践

### 1. 错误创建

```go
// ✅ 好的做法：使用具体的错误类型
func GetUser(id string) (*User, error) {
    if id == "" {
        return nil, errors.NewValidationError("user_id", "用户ID不能为空")
    }
    
    user, err := userRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.NewUserError("用户不存在", errors.CodeUserNotFound)
        }
        return nil, errors.WrapWithCode(err, errors.CodeDatabaseError, "查询用户失败")
    }
    
    return user, nil
}

// ❌ 避免的做法：使用模糊的错误消息
func GetUser(id string) (*User, error) {
    user, err := userRepo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("error") // 信息不明确
    }
    return user, nil
}
```

### 2. 错误处理

```go
// ✅ 根据错误类型进行不同处理
func HandleUserOperation(userID string) error {
    user, err := GetUser(userID)
    if err != nil {
        // 根据错误码进行不同处理
        switch errors.GetCode(err) {
        case errors.CodeUserNotFound:
            // 用户不存在，记录日志但不返回错误
            log.Info("尝试访问不存在的用户", logger.String("user_id", userID))
            return nil
        case errors.CodeDatabaseError:
            // 数据库错误，需要重试
            return errors.Wrap(err, "用户操作失败，请重试")
        default:
            // 其他错误，直接返回
            return err
        }
    }
    
    // 处理用户逻辑
    return nil
}
```

### 3. 错误传播

```go
// ✅ 在错误传播时添加上下文
func ProcessOrder(orderID string) error {
    order, err := orderService.GetOrder(orderID)
    if err != nil {
        return errors.WithContext(err, map[string]interface{}{
            "order_id":  orderID,
            "operation": "process_order",
        })
    }
    
    // 处理订单逻辑
    if err := paymentService.Process(order.PaymentID); err != nil {
        return errors.Wrapf(err, "处理订单 %s 的支付失败", orderID)
    }
    
    return nil
}
```

### 4. 错误测试

```go
// ✅ 测试错误处理逻辑
func TestGetUser_NotFound(t *testing.T) {
    userService := NewUserService(mockRepo)
    
    // 模拟用户不存在
    mockRepo.EXPECT().FindByID("non-existent").Return(nil, gorm.ErrRecordNotFound)
    
    user, err := userService.GetUser("non-existent")
    
    // 验证返回的错误类型和错误码
    assert.Nil(t, user)
    assert.Error(t, err)
    
    var userErr *errors.UserError
    assert.True(t, errors.As(err, &userErr))
    assert.Equal(t, errors.CodeUserNotFound, userErr.Code())
}
```

## 国际化支持

### 多语言错误消息

```go
// 定义错误消息模板
var errorMessages = map[string]map[string]string{
    "en": {
        CodeUserNotFound:      "User not found",
        CodeValidation:        "Validation failed",
        CodeInternalError:     "Internal server error",
    },
    "zh": {
        CodeUserNotFound:      "用户不存在",
        CodeValidation:        "验证失败", 
        CodeInternalError:     "服务器内部错误",
    },
}

// 本地化错误消息
func LocalizeError(err error, lang string) string {
    code := errors.GetCode(err)
    if messages, ok := errorMessages[lang]; ok {
        if message, ok := messages[code]; ok {
            return message
        }
    }
    return err.Error() // 降级到原始消息
}
```

## 性能考虑

### 堆栈追踪优化

```go
// 在生产环境中可以禁用堆栈追踪以提高性能
func init() {
    if os.Getenv("GO_ENV") == "production" {
        errors.DisableStackTrace()
    }
}

// 或者只在特定错误级别启用堆栈追踪
errors.SetStackTraceLevel(errors.LevelError)
```

### 错误缓存

```go
// 缓存常用错误对象以减少内存分配
var (
    ErrUserNotFound = errors.NewUserError("用户不存在", errors.CodeUserNotFound)
    ErrValidation   = errors.New("验证失败", errors.CodeValidation)
)

// 使用预定义错误
func GetUser(id string) (*User, error) {
    if id == "" {
        return nil, ErrValidation
    }
    // ...
}
```

## 参考资料

- [Go Error Handling](https://blog.golang.org/error-handling-and-go)
- [Go-Snap Boot 模块](boot.md)
- [Go-Snap Logger 模块](logger.md)
- [Go-Snap 架构设计](../architecture.md) 