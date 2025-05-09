// Package web 提供了基于Gin构建的HTTP服务框架
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

// Config Web服务配置
type Config struct {
	Host            string        `json:"host" validate:"required"`
	Port            int           `json:"port" validate:"required,min=1,max=65535"`
	Mode            string        `json:"mode" validate:"required,oneof=debug release test"`
	BasePath        string        `json:"base_path"`
	BodyLimit       string        `json:"body_limit" validate:"required"`
	ReadTimeout     time.Duration `json:"read_timeout" validate:"required,min=1ms"`
	WriteTimeout    time.Duration `json:"write_timeout" validate:"required,min=1ms"`
	EnableSwagger   bool          `json:"enable_swagger"`
	EnableProfiling bool          `json:"enable_profiling"`
	EnableCORS      bool          `json:"enable_cors"`
	LogRequests     bool          `json:"log_requests"`
	LogResponses    bool          `json:"log_responses"`
	TrustedProxies  []string      `json:"trusted_proxies"`
}

// DefaultConfig 返回默认Web服务配置
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
func (c *Config) LoadFromProvider(p config.Provider) error {
	if err := p.Unmarshal(c); err != nil {
		return errors.Wrap(err, "unmarshal web config failed")
	}
	return nil
}

// Server Web服务器实例
type Server struct {
	config      Config
	engine      *gin.Engine
	router      *gin.RouterGroup
	log         logger.Logger
	middlewares []gin.HandlerFunc
	httpServer  *http.Server
	routeGroups map[string]*RouteGroup
	pool        sync.Pool
}

// RouteGroup 路由组，用于管理分组路由
type RouteGroup struct {
	group       *gin.RouterGroup
	middlewares []gin.HandlerFunc
	server      *Server
}

// Option 定义Server的配置选项
type Option func(*Server)

// WithLogger 设置自定义日志器
func WithLogger(log logger.Logger) Option {
	return func(s *Server) {
		s.log = log
	}
}

// WithMiddleware 添加全局中间件
func WithMiddleware(middleware ...gin.HandlerFunc) Option {
	return func(s *Server) {
		s.middlewares = append(s.middlewares, middleware...)
	}
}

// New 创建一个新的Web服务器实例
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

// 注册默认中间件
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
