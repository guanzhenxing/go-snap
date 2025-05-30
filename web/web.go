// Package web 提供了基于Gin构建的HTTP服务框架
// 封装了路由、中间件、请求处理和服务器生命周期管理
//
// # Web框架架构
//
// 本包基于Gin框架构建，提供了更高级别的抽象和扩展。主要组件包括：
//
// 1. Server：核心服务器组件，管理HTTP服务器生命周期
// 2. RouteGroup：路由分组，用于组织API路由结构
// 3. 中间件系统：提供请求处理管道
// 4. 请求上下文：提供请求处理的统一接口
// 5. 响应工具：标准化API响应格式
//
// 架构特点：
//
// - 基于DI（依赖注入）的设计，便于测试和扩展
// - 中间件链模式，灵活组合请求处理逻辑
// - 标准化错误处理和响应格式
// - 内置多种实用中间件
// - 与框架其他组件（如日志、配置）无缝集成
//
// # 中间件设计
//
// 中间件是处理HTTP请求的可组合单元，遵循洋葱模型：
//
//	传入请求 → [中间件1 → [中间件2 → [处理器] → 中间件2] → 中间件1] → 响应
//
// 框架提供以下内置中间件：
//
// - Recovery：捕获panic，确保服务器稳定性
// - Logger：请求日志记录
// - CORS：跨域资源共享
// - JWT：身份验证
// - RateLimit：请求频率限制
// - RequestSizeLimiter：请求体大小限制
// - Timeout：请求超时控制
// - RequestID：请求标识生成
//
// # 路由系统
//
// 路由系统支持RESTful API设计，提供以下功能：
//
// - 路由分组：按功能或版本组织API
// - 路由参数：支持URL参数和查询参数
// - HTTP方法绑定：支持GET、POST、PUT、DELETE等方法
// - 中间件绑定：全局、分组和路由级别中间件
//
// # 使用示例
//
// 1. 创建基本服务器：
//
//	// 使用默认配置创建服务器
//	server := web.New(web.DefaultConfig())
//
//	// 添加全局中间件
//	server.Use(middleware.Recovery())
//	server.Use(middleware.Logger(logger))
//
//	// 定义路由
//	server.GET("/ping", func(c *gin.Context) {
//	    response.Success(c, "pong")
//	})
//
//	// 启动服务器
//	server.Start()
//
// 2. 使用路由分组：
//
//	// 创建API分组
//	api := server.Group("/api/v1")
//
//	// 添加分组中间件
//	api.Use(middleware.JWT(config.JWTSecret))
//
//	// 定义分组路由
//	api.GET("/users", userHandler.List)
//	api.POST("/users", userHandler.Create)
//	api.GET("/users/:id", userHandler.Get)
//
//	// 创建嵌套分组
//	admin := api.Group("/admin")
//	admin.Use(middleware.Role("admin"))
//	admin.GET("/stats", adminHandler.GetStats)
//
// 3. 使用配置提供器：
//
//	// 从配置创建服务器
//	server, err := web.NewFromProvider(configProvider, "web")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// 4. 完整请求处理流程：
//
//	// 处理函数示例
//	func GetUser(c *gin.Context) {
//	    // 获取参数
//	    id := c.Param("id")
//
//	    // 业务逻辑
//	    user, err := userService.GetUser(id)
//	    if err != nil {
//	        response.Error(c, err)
//	        return
//	    }
//
//	    // 成功响应
//	    response.Success(c, user)
//	}
//
// # 最佳实践
//
// 1. 路由组织：按功能域或资源类型组织路由
// 2. 中间件使用：全局中间件应处理横切关注点，特定中间件应用于特定路由组
// 3. 错误处理：统一使用response包处理错误和响应
// 4. 优雅关闭：在应用退出时调用server.Stop()确保请求处理完成
// 5. 请求上下文：使用context传递请求级数据，避免全局状态
// 6. 路由定义：集中定义路由，提高可维护性
// 7. 参数验证：使用结构体标签进行输入验证
package web

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
	"github.com/guanzhenxing/go-snap/web/middleware"
	"github.com/guanzhenxing/go-snap/web/response"
)

// Config Web服务配置结构体，定义Web服务器的各种配置选项
// 配置项通常从配置文件或环境变量加载，也可以通过代码设置
// 所有字段都有合理的默认值，可以通过DefaultConfig()获取
type Config struct {
	// Host 服务器监听的主机地址，如"0.0.0.0"表示监听所有网络接口
	// 默认值："0.0.0.0"
	// 也可以指定具体IP如"127.0.0.1"，限制只接受本地连接
	Host string `json:"host" validate:"required"`

	// Port 服务器监听的端口号，有效范围1-65535
	// 默认值：8080
	// 常用端口：80(HTTP)、443(HTTPS)、8080(开发)等
	Port int `json:"port" validate:"required,min=1,max=65535"`

	// Mode Gin运行模式：debug、release或test
	// 默认值：release
	// - debug: 详细日志，适合开发
	// - release: 精简日志，适合生产
	// - test: 用于测试，不写入日志
	Mode string `json:"mode" validate:"required,oneof=debug release test"`

	// BasePath API的基础路径前缀，例如"/api/v1"
	// 默认值：""（空字符串，无前缀）
	// 用于统一为所有路由添加前缀，便于API版本管理
	BasePath string `json:"base_path"`

	// BodyLimit 请求体大小限制，例如"1MB"
	// 默认值："1MB"
	// 格式支持："B"、"KB"、"MB"、"GB"等
	// 防止超大请求导致服务器内存压力或DOS攻击
	BodyLimit string `json:"body_limit" validate:"required"`

	// ReadTimeout 请求读取超时时间
	// 默认值：15秒
	// 从接受连接到读取完整请求的最大时间
	// 防止慢客户端占用连接资源
	ReadTimeout time.Duration `json:"read_timeout" validate:"required,min=1ms"`

	// WriteTimeout 响应写入超时时间
	// 默认值：15秒
	// 从请求处理完成到写入最后一个字节的最大时间
	// 防止慢客户端占用连接资源
	WriteTimeout time.Duration `json:"write_timeout" validate:"required,min=1ms"`

	// EnableSwagger 是否启用Swagger API文档
	// 默认值：true
	// 生产环境中可能需要禁用
	EnableSwagger bool `json:"enable_swagger"`

	// EnableProfiling 是否启用性能分析接口
	// 默认值：false
	// 启用后可以通过/debug/pprof/访问性能分析数据
	// 通常只在开发或排查性能问题时启用
	EnableProfiling bool `json:"enable_profiling"`

	// EnableCORS 是否启用跨域资源共享
	// 默认值：true
	// 启用后允许来自不同域的前端应用访问API
	EnableCORS bool `json:"enable_cors"`

	// LogRequests 是否记录请求日志
	// 默认值：true
	// 记录每个HTTP请求的基本信息
	LogRequests bool `json:"log_requests"`

	// LogResponses 是否记录响应日志
	// 默认值：false
	// 是否记录HTTP响应内容，可能包含敏感数据
	// 高流量系统中不建议启用，会产生大量日志
	LogResponses bool `json:"log_responses"`

	// TrustedProxies 受信任的代理服务器IP列表
	// 默认值：["127.0.0.1"]
	// 当应用部署在代理（如Nginx）后面时，用于正确获取客户端IP
	TrustedProxies []string `json:"trusted_proxies"`
}

// DefaultConfig 返回默认Web服务配置
// 返回：
//
//	Config: 包含合理默认值的配置实例
//
// 通常在没有特定配置需求时使用，或作为自定义配置的基础
// 这些默认值适合开发环境，生产环境可能需要调整
func DefaultConfig() Config {
	return Config{
		Host:            "0.0.0.0",
		Port:            8080,
		Mode:            gin.ReleaseMode,
		BasePath:        "",
		BodyLimit:       "1MB",
		ReadTimeout:     time.Second * 15,
		WriteTimeout:    time.Second * 15,
		EnableSwagger:   true,
		EnableProfiling: false,
		EnableCORS:      true,
		LogRequests:     true,
		LogResponses:    false,
		TrustedProxies:  []string{"127.0.0.1"},
	}
}

// LoadFromProvider 从配置提供器加载Web服务配置
// 参数：
//
//	p: 配置提供器实例
//
// 返回：
//
//	error: 加载过程中遇到的错误，如果加载成功则返回nil
//
// 用于从配置系统（文件、环境变量等）加载Web服务配置
// 配置路径通常是"web"或"server"，取决于配置文件结构
func (c *Config) LoadFromProvider(p config.Provider) error {
	if err := p.Unmarshal(c); err != nil {
		return errors.Wrap(err, "unmarshal web config failed")
	}
	return nil
}

// Server Web服务器实例，封装了HTTP服务器和路由管理
// 是Web包的核心组件，负责HTTP服务器生命周期和请求处理
// 不直接使用，而是通过New()或NewFromProvider()创建
type Server struct {
	// config 服务器配置
	// 包含服务器所有配置选项
	config Config

	// engine 底层Gin引擎实例
	// 处理HTTP请求路由和中间件
	engine *gin.Engine

	// router 根路由组
	// 所有路由都基于此路由组创建
	router *gin.RouterGroup

	// log 日志记录器
	// 用于记录服务器日志
	log logger.Logger

	// middlewares 全局中间件列表
	// 应用于所有请求
	middlewares []gin.HandlerFunc

	// httpServer 底层HTTP服务器
	// 封装的标准库http.Server
	httpServer *http.Server

	// routeGroups 路由组映射，键为路径
	// 缓存已创建的路由组
	routeGroups map[string]*RouteGroup

	// pool 响应上下文对象池，用于减少内存分配
	// 提高性能，减少GC压力
	pool sync.Pool
}

// RouteGroup 路由组，用于管理分组路由和中间件
// 路由组可以嵌套，允许构建复杂的API结构
// 每个路由组可以有自己的中间件链
type RouteGroup struct {
	// group 底层Gin路由组
	// 处理实际的路由注册
	group *gin.RouterGroup

	// middlewares 路由组特定的中间件
	// 仅应用于此路由组
	middlewares []gin.HandlerFunc

	// server 所属的服务器实例
	// 用于访问服务器资源和状态
	server *Server
}

// Option 定义Server的配置选项，用于函数式选项模式
// 使用函数式选项模式可以灵活配置Server
// 可以组合多个选项以自定义服务器行为
type Option func(*Server)

// WithLogger 设置自定义日志器
// 参数：
//
//	log: 自定义的日志记录器
//
// 返回：
//
//	Option: 服务器配置选项函数
//
// 用于覆盖默认日志器，通常用于与应用主日志系统集成
// 所有Web服务器日志将通过此日志器记录
func WithLogger(log logger.Logger) Option {
	return func(s *Server) {
		s.log = log
	}
}

// WithMiddleware 添加全局中间件
// 参数：
//
//	middleware: 要添加的一个或多个中间件函数
//
// 返回：
//
//	Option: 服务器配置选项函数
//
// 用于在服务器创建时添加全局中间件
// 这些中间件将按添加顺序应用于所有请求
func WithMiddleware(middleware ...gin.HandlerFunc) Option {
	return func(s *Server) {
		s.middlewares = append(s.middlewares, middleware...)
	}
}

// New 创建一个新的Web服务器实例
// 参数：
//
//	cfg: 服务器配置
//	opts: 可选的配置选项
//
// 返回：
//
//	*Server: 配置好的服务器实例
//
// 创建并配置一个新的Web服务器，但不启动它
// 需要调用Start()方法才会实际开始监听请求
func New(cfg Config, opts ...Option) *Server {
	// 设置Gin模式
	gin.SetMode(cfg.Mode)

	// 创建Gin引擎
	engine := gin.New()

	// 设置信任代理
	if len(cfg.TrustedProxies) > 0 {
		if err := engine.SetTrustedProxies(cfg.TrustedProxies); err != nil {
			panic(fmt.Sprintf("failed to set trusted proxies: %v", err))
		}
	}

	// 创建根路由组
	var baseRouter *gin.RouterGroup
	if cfg.BasePath != "" {
		baseRouter = engine.Group(cfg.BasePath)
	} else {
		baseRouter = engine.Group("")
	}

	// 创建服务器实例
	server := &Server{
		config:      cfg,
		engine:      engine,
		router:      baseRouter,
		log:         logger.New(),
		middlewares: []gin.HandlerFunc{},
		routeGroups: make(map[string]*RouteGroup),
		pool: sync.Pool{
			New: func() interface{} {
				return &response.Context{}
			},
		},
	}

	// 应用选项
	for _, opt := range opts {
		opt(server)
	}

	// 初始化HTTP服务器
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// 注册默认中间件
	server.registerDefaultMiddleware()

	return server
}

// NewFromProvider 从配置提供器创建一个新的Web服务器实例
// 参数：
//
//	p: 配置提供器
//	configPath: 配置路径，如果为空则直接解析整个配置
//	opts: 可选的配置选项
//
// 返回：
//
//	*Server: 配置好的服务器实例
//	error: 创建过程中遇到的错误
func NewFromProvider(p config.Provider, configPath string, opts ...Option) (*Server, error) {
	cfg := DefaultConfig()

	// 获取配置
	if configPath != "" {
		// 使用子配置路径
		if err := p.UnmarshalKey(configPath, &cfg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal web config")
		}
	} else {
		// 直接解析
		if err := p.Unmarshal(&cfg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal web config")
		}
	}

	return New(cfg, opts...), nil
}

// registerDefaultMiddleware 注册默认中间件
// 根据配置添加恢复、日志记录和CORS等中间件
func (s *Server) registerDefaultMiddleware() {
	// 添加Recovery中间件
	s.Use(middleware.Recovery())

	// 如果启用请求日志
	if s.config.LogRequests {
		s.Use(middleware.Logger(s.log))
	}

	// 如果启用CORS
	if s.config.EnableCORS {
		s.Use(middleware.CORS())
	}

	// 添加用户自定义的全局中间件
	s.Use(s.middlewares...)
}

// Engine 返回底层的Gin引擎
// 返回：
//
//	*gin.Engine: 底层Gin引擎实例
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// Use 添加全局中间件
func (s *Server) Use(middleware ...gin.HandlerFunc) *Server {
	s.router.Use(middleware...)
	return s
}

// Group 创建一个新的路由组
func (s *Server) Group(path string, middleware ...gin.HandlerFunc) *RouteGroup {
	// 如果路由组已存在，直接返回
	if group, exists := s.routeGroups[path]; exists {
		group.middlewares = append(group.middlewares, middleware...)
		return group
	}

	// 创建新的路由组
	group := &RouteGroup{
		group:       s.router.Group(path),
		middlewares: middleware,
		server:      s,
	}

	// 应用中间件
	if len(middleware) > 0 {
		group.group.Use(middleware...)
	}

	// 保存路由组
	s.routeGroups[path] = group
	return group
}

// GET 注册GET请求处理器
func (s *Server) GET(path string, handlers ...gin.HandlerFunc) *Server {
	s.router.GET(path, s.wrapHandlers(handlers)...)
	return s
}

// POST 注册POST请求处理器
func (s *Server) POST(path string, handlers ...gin.HandlerFunc) *Server {
	s.router.POST(path, s.wrapHandlers(handlers)...)
	return s
}

// PUT 注册PUT请求处理器
func (s *Server) PUT(path string, handlers ...gin.HandlerFunc) *Server {
	s.router.PUT(path, s.wrapHandlers(handlers)...)
	return s
}

// DELETE 注册DELETE请求处理器
func (s *Server) DELETE(path string, handlers ...gin.HandlerFunc) *Server {
	s.router.DELETE(path, s.wrapHandlers(handlers)...)
	return s
}

// PATCH 注册PATCH请求处理器
func (s *Server) PATCH(path string, handlers ...gin.HandlerFunc) *Server {
	s.router.PATCH(path, s.wrapHandlers(handlers)...)
	return s
}

// OPTIONS 注册OPTIONS请求处理器
func (s *Server) OPTIONS(path string, handlers ...gin.HandlerFunc) *Server {
	s.router.OPTIONS(path, s.wrapHandlers(handlers)...)
	return s
}

// HEAD 注册HEAD请求处理器
func (s *Server) HEAD(path string, handlers ...gin.HandlerFunc) *Server {
	s.router.HEAD(path, s.wrapHandlers(handlers)...)
	return s
}

// 路由组方法
// GET 注册GET请求处理器
func (g *RouteGroup) GET(path string, handlers ...gin.HandlerFunc) *RouteGroup {
	g.group.GET(path, g.server.wrapHandlers(handlers)...)
	return g
}

// POST 注册POST请求处理器
func (g *RouteGroup) POST(path string, handlers ...gin.HandlerFunc) *RouteGroup {
	g.group.POST(path, g.server.wrapHandlers(handlers)...)
	return g
}

// PUT 注册PUT请求处理器
func (g *RouteGroup) PUT(path string, handlers ...gin.HandlerFunc) *RouteGroup {
	g.group.PUT(path, g.server.wrapHandlers(handlers)...)
	return g
}

// DELETE 注册DELETE请求处理器
func (g *RouteGroup) DELETE(path string, handlers ...gin.HandlerFunc) *RouteGroup {
	g.group.DELETE(path, g.server.wrapHandlers(handlers)...)
	return g
}

// PATCH 注册PATCH请求处理器
func (g *RouteGroup) PATCH(path string, handlers ...gin.HandlerFunc) *RouteGroup {
	g.group.PATCH(path, g.server.wrapHandlers(handlers)...)
	return g
}

// OPTIONS 注册OPTIONS请求处理器
func (g *RouteGroup) OPTIONS(path string, handlers ...gin.HandlerFunc) *RouteGroup {
	g.group.OPTIONS(path, g.server.wrapHandlers(handlers)...)
	return g
}

// HEAD 注册HEAD请求处理器
func (g *RouteGroup) HEAD(path string, handlers ...gin.HandlerFunc) *RouteGroup {
	g.group.HEAD(path, g.server.wrapHandlers(handlers)...)
	return g
}

// Group 创建子路由组
func (g *RouteGroup) Group(path string, middleware ...gin.HandlerFunc) *RouteGroup {
	group := &RouteGroup{
		group:       g.group.Group(path),
		middlewares: middleware,
		server:      g.server,
	}

	// 应用中间件
	if len(middleware) > 0 {
		group.group.Use(middleware...)
	}

	// 使用完整路径作为键保存路由组
	fullPath := g.group.BasePath() + path
	g.server.routeGroups[fullPath] = group
	return group
}

// Use 添加路由组中间件
func (g *RouteGroup) Use(middleware ...gin.HandlerFunc) *RouteGroup {
	g.middlewares = append(g.middlewares, middleware...)
	g.group.Use(middleware...)
	return g
}

// wrapHandlers 包装处理器，添加上下文对象并处理响应
func (s *Server) wrapHandlers(handlers []gin.HandlerFunc) []gin.HandlerFunc {
	wrappedHandlers := make([]gin.HandlerFunc, len(handlers))

	for i, handler := range handlers {
		wrappedHandlers[i] = func(ginCtx *gin.Context) {
			// 从对象池获取上下文
			ctx := s.pool.Get().(*response.Context)
			defer s.pool.Put(ctx)

			// 初始化上下文
			ctx.Init(ginCtx)

			// 设置上下文
			ginCtx.Set("ctx", ctx)

			// 执行原始处理器
			handler(ginCtx)

			// 如果启用了响应日志
			if s.config.LogResponses && !ctx.IsStreaming() {
				s.log.Info(fmt.Sprintf("Response: %d %s", ctx.GetStatus(), ctx.GetBodyString()))
			}
		}
	}

	return wrappedHandlers
}

// Start 启动Web服务器
func (s *Server) Start() error {
	s.log.Info(fmt.Sprintf("Starting web server on %s:%d", s.config.Host, s.config.Port))

	// 如果启用Swagger
	if s.config.EnableSwagger {
		s.registerSwagger()
	}

	// 如果启用性能分析
	if s.config.EnableProfiling {
		s.registerProfiling()
	}

	// 启动HTTP服务器
	return s.httpServer.ListenAndServe()
}

// Stop 优雅地关闭Web服务器
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("Shutting down web server")
	return s.httpServer.Shutdown(ctx)
}
