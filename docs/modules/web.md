# Web 模块

Web 模块是 Go-Snap 框架的 HTTP 服务器组件，基于 [Gin](https://gin-gonic.com/) 构建，提供高性能的 Web 服务、RESTful API、中间件支持等功能，帮助开发者快速构建现代化的 Web 应用。

## 概述

Web 模块提供了一个完整的 HTTP 服务器解决方案，集成了路由管理、中间件系统、请求处理、响应渲染等功能。它与框架的其他组件无缝集成，提供统一的 Web 开发体验。

### 核心特性

- ✅ **高性能** - 基于 Gin 的高性能 HTTP 服务器
- ✅ **RESTful API** - 完整的 REST API 支持
- ✅ **路由管理** - 灵活的路由配置和分组
- ✅ **中间件系统** - 丰富的内置中间件和自定义支持
- ✅ **请求绑定** - 自动请求参数绑定和验证
- ✅ **响应渲染** - 多种响应格式支持
- ✅ **静态文件服务** - 静态资源服务
- ✅ **模板引擎** - HTML 模板渲染
- ✅ **CORS 支持** - 跨域资源共享
- ✅ **认证授权** - JWT、Session 等认证方式
- ✅ **限流保护** - API 限流和防护
- ✅ **健康检查** - 内置健康检查端点

## 快速开始

### 基础用法

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/guanzhenxing/go-snap/boot"
)

func main() {
    app := boot.NewBoot().SetConfigPath("configs")
    
    // 添加路由配置
    app.AddConfigurer(func(application *boot.Application) error {
        // 获取 Web 组件
        if webComp, found := application.GetComponent("web"); found {
            if wc, ok := webComp.(*boot.WebComponent); ok {
                router := wc.GetRouter()
                
                // 基础路由
                router.GET("/", func(c *gin.Context) {
                    c.JSON(http.StatusOK, gin.H{
                        "message": "Hello, Go-Snap Web!",
                        "version": "1.0.0",
                    })
                })
                
                // API 路由组
                api := router.Group("/api/v1")
                {
                    api.GET("/health", healthCheck)
                    api.GET("/users", getUsers)
                    api.POST("/users", createUser)
                }
            }
        }
        return nil
    })
    
    if err := app.Run(); err != nil {
        panic(err)
    }
}

func healthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func getUsers(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"users": []string{"Alice", "Bob"}})
}

func createUser(c *gin.Context) {
    c.JSON(http.StatusCreated, gin.H{"message": "User created"})
}
```

### 配置文件

```yaml
# configs/application.yaml
web:
  enabled: true
  host: "0.0.0.0"
  port: 8080
  mode: "debug"                    # debug, release, test
  read_timeout: "30s"
  write_timeout: "30s"
  max_header_bytes: 1048576        # 1MB
  
  # TLS 配置
  tls:
    enabled: false
    cert_file: "server.crt"
    key_file: "server.key"
  
  # 中间件配置
  middleware:
    cors:
      enabled: true
      allow_origins: ["*"]
      allow_methods: ["GET", "POST", "PUT", "DELETE"]
      allow_headers: ["*"]
    
    rate_limit:
      enabled: true
      requests_per_second: 100
      burst: 200
    
    auth:
      enabled: true
      jwt_secret: "your-secret-key"
      
    logging:
      enabled: true
      skip_paths: ["/health", "/metrics"]
```

## 配置选项

### 基础配置

```yaml
web:
  enabled: true                    # 是否启用 Web 服务
  host: "0.0.0.0"                 # 监听地址
  port: 8080                      # 监听端口
  mode: "debug"                   # 运行模式: debug, release, test
  
  # 服务器配置
  read_timeout: "30s"             # 读取超时
  write_timeout: "30s"            # 写入超时
  idle_timeout: "60s"             # 空闲超时
  max_header_bytes: 1048576       # 最大请求头大小
  
  # TLS/HTTPS 配置
  tls:
    enabled: false                # 是否启用 TLS
    cert_file: "server.crt"       # 证书文件路径
    key_file: "server.key"        # 私钥文件路径
    auto_cert: false              # 是否自动申请证书
    
  # 静态文件配置
  static:
    enabled: true                 # 是否启用静态文件服务
    path: "/static"               # URL 路径
    root: "./static"              # 文件系统路径
    
  # 模板配置
  templates:
    enabled: false                # 是否启用模板引擎
    pattern: "templates/*"        # 模板文件模式
    
  # 上传配置
  upload:
    max_memory: 33554432          # 最大内存使用 (32MB)
    allowed_types: ["image/jpeg", "image/png"]
    max_file_size: 10485760       # 最大文件大小 (10MB)
```

### 中间件配置

```yaml
web:
  middleware:
    # CORS 跨域配置
    cors:
      enabled: true
      allow_origins: ["*"]
      allow_methods: ["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"]
      allow_headers: ["*"]
      allow_credentials: false
      max_age: 86400
      
    # 限流配置
    rate_limit:
      enabled: true
      requests_per_second: 100    # 每秒请求数
      burst: 200                  # 突发请求数
      key_func: "ip"              # 限流键: ip, user, custom
      
    # 认证配置
    auth:
      enabled: true
      type: "jwt"                 # 认证类型: jwt, session, basic
      jwt_secret: "your-secret"
      jwt_expire: "24h"
      exclude_paths: ["/login", "/register", "/health"]
      
    # 日志中间件
    logging:
      enabled: true
      format: "default"           # 日志格式: default, json, custom
      skip_paths: ["/health", "/metrics", "/favicon.ico"]
      
    # 压缩中间件
    compression:
      enabled: true
      level: 6                    # 压缩级别 1-9
      min_length: 1024            # 最小压缩长度
      
    # 恢复中间件
    recovery:
      enabled: true
      stack: true                 # 是否打印堆栈信息
      
    # 安全中间件
    security:
      enabled: true
      content_type_nosniff: true
      frame_deny: true
      xss_protection: true
```

## API 参考

### Web 组件

#### WebComponent

Web 组件接口。

```go
type WebComponent interface {
    boot.Component
    
    // 获取 Gin 引擎
    GetRouter() *gin.Engine
    
    // 获取 HTTP 服务器
    GetServer() *http.Server
    
    // 注册路由组
    RegisterRouteGroup(path string, handlers ...gin.HandlerFunc) *gin.RouterGroup
    
    // 注册中间件
    RegisterMiddleware(middleware gin.HandlerFunc)
    
    // 获取服务器地址
    GetAddress() string
    
    // 优雅关闭
    Shutdown(ctx context.Context) error
}
```

### 路由管理

#### 基础路由

```go
// HTTP 方法路由
router.GET("/path", handler)
router.POST("/path", handler)
router.PUT("/path", handler)
router.DELETE("/path", handler)
router.PATCH("/path", handler)
router.HEAD("/path", handler)
router.OPTIONS("/path", handler)

// 通用路由
router.Any("/path", handler)        // 匹配所有 HTTP 方法
router.NoRoute(handler)             // 404 处理
router.NoMethod(handler)            // 405 处理
```

#### 路由参数

```go
// 路径参数
router.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"user_id": id})
})

// 通配符参数
router.GET("/files/*filepath", func(c *gin.Context) {
    filepath := c.Param("filepath")
    c.JSON(200, gin.H{"filepath": filepath})
})

// 查询参数
router.GET("/search", func(c *gin.Context) {
    query := c.Query("q")           // 获取查询参数
    page := c.DefaultQuery("page", "1") // 带默认值
    c.JSON(200, gin.H{"query": query, "page": page})
})
```

#### 路由组

```go
// 创建路由组
v1 := router.Group("/api/v1")
{
    v1.GET("/users", getUsers)
    v1.POST("/users", createUser)
    
    // 嵌套路由组
    admin := v1.Group("/admin")
    admin.Use(authMiddleware()) // 组级中间件
    {
        admin.GET("/stats", getStats)
        admin.DELETE("/users/:id", deleteUser)
    }
}

// 带中间件的路由组
authorized := router.Group("/")
authorized.Use(authMiddleware())
{
    authorized.POST("/submit", submitData)
    authorized.GET("/secret", getSecret)
}
```

### 请求处理

#### 请求绑定

```go
type User struct {
    ID   int    `json:"id" binding:"required"`
    Name string `json:"name" binding:"required,min=2,max=50"`
    Email string `json:"email" binding:"required,email"`
    Age  int    `json:"age" binding:"min=18,max=100"`
}

func createUser(c *gin.Context) {
    var user User
    
    // JSON 绑定
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 表单绑定
    if err := c.ShouldBind(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // URI 绑定
    if err := c.ShouldBindUri(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, user)
}
```

#### 文件上传

```go
func uploadFile(c *gin.Context) {
    // 单文件上传
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
        return
    }
    defer file.Close()
    
    // 保存文件
    filename := header.Filename
    if err := c.SaveUploadedFile(header, "./uploads/"+filename); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"filename": filename})
}

func uploadMultipleFiles(c *gin.Context) {
    // 多文件上传
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    files := form.File["files"]
    for _, file := range files {
        if err := c.SaveUploadedFile(file, "./uploads/"+file.Filename); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save " + file.Filename})
            return
        }
    }
    
    c.JSON(http.StatusOK, gin.H{"count": len(files)})
}
```

### 响应处理

#### JSON 响应

```go
func jsonResponse(c *gin.Context) {
    // 简单 JSON
    c.JSON(http.StatusOK, gin.H{
        "message": "success",
        "data": map[string]interface{}{
            "id": 1,
            "name": "John",
        },
    })
    
    // 结构体 JSON
    user := User{ID: 1, Name: "John", Email: "john@example.com"}
    c.JSON(http.StatusOK, user)
    
    // 美化 JSON
    c.IndentedJSON(http.StatusOK, user)
    
    // JSONP
    c.JSONP(http.StatusOK, user)
}
```

#### 其他响应格式

```go
func variousResponses(c *gin.Context) {
    // XML 响应
    c.XML(http.StatusOK, gin.H{"message": "Hello XML"})
    
    // YAML 响应
    c.YAML(http.StatusOK, gin.H{"message": "Hello YAML"})
    
    // 纯文本
    c.String(http.StatusOK, "Hello, %s!", "World")
    
    // HTML 响应
    c.HTML(http.StatusOK, "index.html", gin.H{
        "title": "Main website",
    })
    
    // 重定向
    c.Redirect(http.StatusMovedPermanently, "https://www.google.com")
    
    // 文件下载
    c.File("./files/report.pdf")
    c.FileAttachment("./files/report.pdf", "monthly-report.pdf")
    
    // 数据流
    c.Data(http.StatusOK, "application/octet-stream", []byte("binary data"))
}
```

## 中间件系统

### 内置中间件

#### 认证中间件

```go
func JWTMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
            c.Abort()
            return
        }
        
        // 移除 "Bearer " 前缀
        if strings.HasPrefix(token, "Bearer ") {
            token = token[7:]
        }
        
        // 验证 JWT token
        claims, err := validateJWT(token, secret)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // 设置用户信息到上下文
        c.Set("user_id", claims["user_id"])
        c.Set("username", claims["username"])
        
        c.Next()
    }
}

// 使用认证中间件
router.Use(JWTMiddleware("your-secret-key"))
```

#### CORS 中间件

```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

#### 限流中间件

```go
import "golang.org/x/time/rate"

func RateLimitMiddleware(r rate.Limit, b int) gin.HandlerFunc {
    limiter := rate.NewLimiter(r, b)
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}

// 使用限流中间件
router.Use(RateLimitMiddleware(rate.Limit(10), 20)) // 每秒10个请求，突发20个
```

#### 请求日志中间件

```go
func LoggerMiddleware() gin.HandlerFunc {
    return gin.LoggerWithConfig(gin.LoggerConfig{
        Formatter: func(param gin.LogFormatterParams) string {
            return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
                param.ClientIP,
                param.TimeStamp.Format(time.RFC1123),
                param.Method,
                param.Path,
                param.Request.Proto,
                param.StatusCode,
                param.Latency,
                param.Request.UserAgent(),
                param.ErrorMessage,
            )
        },
        Output: os.Stdout,
        SkipPaths: []string{"/health", "/metrics"},
    })
}
```

### 自定义中间件

```go
// 请求 ID 中间件
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = generateRequestID()
        }
        
        c.Set("request_id", requestID)
        c.Writer.Header().Set("X-Request-ID", requestID)
        
        c.Next()
    }
}

// 性能监控中间件
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        
        // 记录指标
        recordMetrics(c.Request.Method, c.FullPath(), c.Writer.Status(), duration)
    }
}

// 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            
            switch err.Type {
            case gin.ErrorTypeBind:
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
            case gin.ErrorTypePublic:
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            default:
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            }
        }
    }
}
```

## 实际应用示例

### RESTful API 设计

```go
type UserController struct {
    userService *UserService
}

func NewUserController(userService *UserService) *UserController {
    return &UserController{userService: userService}
}

// GET /users
func (uc *UserController) GetUsers(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    
    users, total, err := uc.userService.GetUsers(page, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "users": users,
        "total": total,
        "page":  page,
        "limit": limit,
    })
}

// GET /users/:id
func (uc *UserController) GetUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    
    user, err := uc.userService.GetUser(uint(id))
    if err != nil {
        if err.Error() == "user not found" {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, user)
}

// POST /users
func (uc *UserController) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    user, err := uc.userService.CreateUser(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, user)
}

// PUT /users/:id
func (uc *UserController) UpdateUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    
    var req UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    user, err := uc.userService.UpdateUser(uint(id), &req)
    if err != nil {
        if err.Error() == "user not found" {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, user)
}

// DELETE /users/:id
func (uc *UserController) DeleteUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    
    err = uc.userService.DeleteUser(uint(id))
    if err != nil {
        if err.Error() == "user not found" {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusNoContent, nil)
}

// 注册路由
func (uc *UserController) RegisterRoutes(router *gin.RouterGroup) {
    users := router.Group("/users")
    {
        users.GET("", uc.GetUsers)
        users.GET("/:id", uc.GetUser)
        users.POST("", uc.CreateUser)
        users.PUT("/:id", uc.UpdateUser)
        users.DELETE("/:id", uc.DeleteUser)
    }
}
```

### 认证和授权

```go
// JWT 工具函数
func generateJWT(userID uint, username string, secret string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id":  userID,
        "username": username,
        "exp":      time.Now().Add(time.Hour * 24).Unix(),
    })
    
    return token.SignedString([]byte(secret))
}

func validateJWT(tokenString, secret string) (jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })
    
    if err != nil || !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }
    
    return token.Claims.(jwt.MapClaims), nil
}

// 认证控制器
type AuthController struct {
    userService *UserService
    jwtSecret   string
}

func (ac *AuthController) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    user, err := ac.userService.Authenticate(req.Username, req.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }
    
    token, err := generateJWT(user.ID, user.Username, ac.jwtSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "token": token,
        "user":  user,
    })
}

func (ac *AuthController) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    user, err := ac.userService.CreateUser(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    token, err := generateJWT(user.ID, user.Username, ac.jwtSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "token": token,
        "user":  user,
    })
}

// 权限检查中间件
func RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("user_id")
        
        user, err := getUserByID(userID)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
            c.Abort()
            return
        }
        
        if !hasRole(user, role) {
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 文件上传和处理

```go
type FileController struct {
    uploadPath string
    maxSize    int64
}

func NewFileController(uploadPath string, maxSize int64) *FileController {
    return &FileController{
        uploadPath: uploadPath,
        maxSize:    maxSize,
    }
}

func (fc *FileController) UploadFile(c *gin.Context) {
    // 限制文件大小
    c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, fc.maxSize)
    
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
        return
    }
    defer file.Close()
    
    // 验证文件类型
    contentType := header.Header.Get("Content-Type")
    if !isAllowedFileType(contentType) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "File type not allowed"})
        return
    }
    
    // 生成文件名
    ext := filepath.Ext(header.Filename)
    filename := fmt.Sprintf("%d%s", time.Now().Unix(), ext)
    filepath := filepath.Join(fc.uploadPath, filename)
    
    // 保存文件
    if err := c.SaveUploadedFile(header, filepath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "filename": filename,
        "size":     header.Size,
        "url":      "/uploads/" + filename,
    })
}

func (fc *FileController) DownloadFile(c *gin.Context) {
    filename := c.Param("filename")
    filepath := filepath.Join(fc.uploadPath, filename)
    
    // 检查文件是否存在
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
        return
    }
    
    c.File(filepath)
}

func isAllowedFileType(contentType string) bool {
    allowedTypes := []string{
        "image/jpeg",
        "image/png",
        "image/gif",
        "text/plain",
        "application/pdf",
    }
    
    for _, t := range allowedTypes {
        if t == contentType {
            return true
        }
    }
    return false
}
```

## 最佳实践

### 1. 项目结构

```
project/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口
├── internal/
│   ├── controller/              # 控制器层
│   │   ├── user_controller.go
│   │   └── auth_controller.go
│   ├── service/                 # 服务层
│   │   ├── user_service.go
│   │   └── auth_service.go
│   ├── repository/              # 数据访问层
│   │   └── user_repository.go
│   ├── middleware/              # 中间件
│   │   ├── auth.go
│   │   └── cors.go
│   ├── model/                   # 数据模型
│   │   └── user.go
│   └── dto/                     # 数据传输对象
│       ├── request.go
│       └── response.go
├── configs/                     # 配置文件
├── static/                      # 静态文件
├── templates/                   # 模板文件
└── docs/                        # API 文档
```

### 2. 错误处理

```go
// 统一错误响应格式
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

// 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            
            var code string
            var status int
            
            switch err.Type {
            case gin.ErrorTypeBind:
                code = "INVALID_REQUEST"
                status = http.StatusBadRequest
            case gin.ErrorTypePublic:
                code = "SERVER_ERROR"
                status = http.StatusInternalServerError
            default:
                code = "UNKNOWN_ERROR"
                status = http.StatusInternalServerError
            }
            
            c.JSON(status, ErrorResponse{
                Code:    code,
                Message: err.Error(),
            })
        }
    }
}

// 在控制器中使用
func (uc *UserController) GetUser(c *gin.Context) {
    id := c.Param("id")
    
    user, err := uc.userService.GetUser(id)
    if err != nil {
        // 记录错误到上下文
        c.Error(err)
        return
    }
    
    c.JSON(http.StatusOK, user)
}
```

### 3. 输入验证

```go
// 使用验证器标签
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50,alphanum"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Age      int    `json:"age" binding:"min=18,max=120"`
}

// 自定义验证器
func customValidator() {
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("username", validateUsername)
    }
}

func validateUsername(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    // 自定义验证逻辑
    return regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(username)
}

// ✅ 好的做法：完整的输入验证
func (uc *UserController) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
        return
    }
    
    // 业务逻辑验证
    if exists, _ := uc.userService.UsernameExists(req.Username); exists {
        c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
        return
    }
    
    user, err := uc.userService.CreateUser(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, user)
}
```

### 4. 响应格式标准化

```go
// 标准化响应结构
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

type Meta struct {
    Page       int `json:"page,omitempty"`
    Limit      int `json:"limit,omitempty"`
    Total      int `json:"total,omitempty"`
    TotalPages int `json:"total_pages,omitempty"`
}

// 响应帮助函数
func SuccessResponse(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Data:    data,
    })
}

func ErrorResponse(c *gin.Context, status int, code, message string) {
    c.JSON(status, APIResponse{
        Success: false,
        Error: &APIError{
            Code:    code,
            Message: message,
        },
    })
}

func PaginatedResponse(c *gin.Context, data interface{}, page, limit, total int) {
    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Data:    data,
        Meta: &Meta{
            Page:       page,
            Limit:      limit,
            Total:      total,
            TotalPages: (total + limit - 1) / limit,
        },
    })
}
```

### 5. 安全最佳实践

```go
// 安全中间件
func SecurityMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 设置安全头
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Content-Security-Policy", "default-src 'self'")
        
        c.Next()
    }
}

// 输入清理
func sanitizeInput(input string) string {
    // 移除潜在的恶意字符
    input = strings.ReplaceAll(input, "<", "&lt;")
    input = strings.ReplaceAll(input, ">", "&gt;")
    input = strings.ReplaceAll(input, "\"", "&quot;")
    input = strings.ReplaceAll(input, "'", "&#x27;")
    return input
}

// SQL 注入防护（使用参数化查询）
func (ur *UserRepository) GetUserByID(id uint) (*User, error) {
    var user User
    // ✅ 使用参数化查询
    err := ur.db.Where("id = ?", id).First(&user).Error
    return &user, err
}
```

## 性能优化

### 1. 压缩和缓存

```go
// 压缩中间件
func CompressionMiddleware() gin.HandlerFunc {
    return gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{
        "/metrics",
        "/health",
    }))
}

// 缓存中间件
func CacheMiddleware(duration time.Duration) gin.HandlerFunc {
    cache := make(map[string]cacheItem)
    var mutex sync.RWMutex
    
    return func(c *gin.Context) {
        key := c.Request.URL.String()
        
        mutex.RLock()
        if item, found := cache[key]; found && time.Now().Before(item.expiry) {
            mutex.RUnlock()
            c.Data(http.StatusOK, item.contentType, item.data)
            return
        }
        mutex.RUnlock()
        
        // 创建响应写入器
        writer := &responseWriter{c.Writer, bytes.NewBuffer(nil)}
        c.Writer = writer
        
        c.Next()
        
        // 缓存响应
        if c.Writer.Status() == http.StatusOK {
            mutex.Lock()
            cache[key] = cacheItem{
                data:        writer.buffer.Bytes(),
                contentType: c.Writer.Header().Get("Content-Type"),
                expiry:      time.Now().Add(duration),
            }
            mutex.Unlock()
        }
    }
}
```

### 2. 连接池和超时配置

```yaml
web:
  # 服务器性能配置
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"
  max_header_bytes: 1048576
  
  # 连接保持
  keep_alive: true
  keep_alive_timeout: "60s"
  
  # 并发限制
  max_connections: 1000
```

### 3. 异步处理

```go
// 异步任务处理
func AsyncHandler() gin.HandlerFunc {
    taskQueue := make(chan Task, 1000)
    
    // 启动工作协程
    for i := 0; i < 10; i++ {
        go taskWorker(taskQueue)
    }
    
    return func(c *gin.Context) {
        task := Task{
            ID:   generateTaskID(),
            Data: extractTaskData(c),
        }
        
        select {
        case taskQueue <- task:
            c.JSON(http.StatusAccepted, gin.H{"task_id": task.ID})
        default:
            c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Task queue full"})
        }
    }
}

func taskWorker(taskQueue <-chan Task) {
    for task := range taskQueue {
        processTask(task)
    }
}
```

## 测试

### 单元测试

```go
import (
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    router := gin.New()
    controller := NewUserController(mockUserService)
    controller.RegisterRoutes(router.Group("/api/v1"))
    
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/v1/users/1", nil)
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    // 验证响应内容
}
```

### 集成测试

```go
func TestUserAPI(t *testing.T) {
    // 设置测试数据库
    db := setupTestDB()
    defer cleanupTestDB(db)
    
    // 创建应用
    app := boot.NewBoot().
        SetConfigPath("testdata").
        AddComponent("database", &DatabaseComponent{db: db})
    
    application, _ := app.Initialize()
    
    // 获取 Web 组件
    webComp := application.GetComponent("web").(*WebComponent)
    server := httptest.NewServer(webComp.GetRouter())
    defer server.Close()
    
    // 执行 API 测试
    resp, err := http.Get(server.URL + "/api/v1/users")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## 参考资料

- [Gin 官方文档](https://gin-gonic.com/)
- [Go HTTP 服务器最佳实践](https://golang.org/doc/articles/wiki/)
- [Go-Snap Boot 模块](boot.md)
- [Go-Snap 架构设计](../architecture.md) 