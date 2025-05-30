package components

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guanzhenxing/go-snap/boot"
	"github.com/guanzhenxing/go-snap/examples/admin/internal/handlers"
	"github.com/guanzhenxing/go-snap/logger"
	"github.com/guanzhenxing/go-snap/web"
	"github.com/guanzhenxing/go-snap/web/middleware"
	"github.com/guanzhenxing/go-snap/web/response"
)

// AdminComponent 管理面板组件
type AdminComponent struct {
	name        string
	webServer   *web.Server
	log         logger.Logger
	config      *AdminConfig
	status      string
	initialized bool
	metrics     map[string]interface{}
}

// AdminConfig 管理面板配置
type AdminConfig struct {
	Host          string `json:"host"`
	Port          int    `json:"port"`
	BasePath      string `json:"base_path"`
	EnableSwagger bool   `json:"enable_swagger"`
	SecretKey     string `json:"secret_key"`
}

// NewAdminComponent 创建管理面板组件
func NewAdminComponent() *AdminComponent {
	return &AdminComponent{
		name:    "admin",
		status:  "created",
		metrics: make(map[string]interface{}),
	}
}

// Name 返回组件名称
func (c *AdminComponent) Name() string {
	return c.name
}

// Type 返回组件类型
func (c *AdminComponent) Type() boot.ComponentType {
	return boot.ComponentTypeWeb
}

// GetStatus 返回组件状态 (实现boot.Component接口)
func (c *AdminComponent) GetStatus() boot.ComponentStatus {
	switch c.status {
	case "created":
		return boot.ComponentStatusCreated
	case "initialized":
		return boot.ComponentStatusInitialized
	case "running":
		return boot.ComponentStatusStarted
	case "stopped":
		return boot.ComponentStatusStopped
	case "failed":
		return boot.ComponentStatusFailed
	default:
		return boot.ComponentStatusUnknown
	}
}

// Status 返回组件状态
func (c *AdminComponent) Status() string {
	return c.status
}

// getLoggerFromComponent 从组件中提取Logger接口
// 这个辅助函数通过接口分离解决接口冲突问题
func getLoggerFromComponent(component interface{}) (logger.Logger, bool) {
	// 尝试将组件转换为logger.Logger接口
	log, ok := component.(logger.Logger)
	return log, ok
}

// Initialize 初始化组件
func (c *AdminComponent) Initialize(ctx context.Context) error {
	// 获取应用
	app, ok := ctx.Value("application").(*boot.Application)
	if !ok || app == nil {
		return fmt.Errorf("无法从上下文获取应用实例")
	}

	// 获取日志组件
	loggerComponent, found := app.GetComponent("logger")
	if !found {
		return fmt.Errorf("找不到logger组件")
	}

	// 通过辅助函数提取Logger接口
	loggerInstance, ok := getLoggerFromComponent(loggerComponent)
	if !ok {
		return fmt.Errorf("logger组件不是有效的logger.Logger实例")
	}
	c.log = loggerInstance

	// 加载配置
	props := app.GetPropertySource()
	c.config = &AdminConfig{
		Host:          props.GetString("admin.host", "127.0.0.1"),
		Port:          props.GetInt("admin.port", 8080),
		BasePath:      props.GetString("admin.base_path", "/admin"),
		EnableSwagger: props.GetBool("admin.swagger.enabled", true),
		SecretKey:     props.GetString("admin.security.secret_key", "admin-secret-key"),
	}

	// 创建Web服务配置
	webConfig := web.DefaultConfig()
	webConfig.Host = c.config.Host
	webConfig.Port = c.config.Port
	webConfig.EnableSwagger = c.config.EnableSwagger
	webConfig.LogRequests = true

	// 创建Web服务
	c.webServer = web.New(webConfig, web.WithLogger(c.log))

	// 设置健康指标
	c.metrics["host"] = c.config.Host
	c.metrics["port"] = c.config.Port

	c.status = "initialized"
	c.initialized = true
	c.log.Info("管理面板组件初始化完成")
	return nil
}

// Start 启动组件
func (c *AdminComponent) Start(ctx context.Context) error {
	// 注册路由
	c.registerRoutes()

	// 启动服务器
	go func() {
		c.log.Info(fmt.Sprintf("管理面板正在启动，地址: %s:%d%s", c.config.Host, c.config.Port, c.config.BasePath))
		if err := c.webServer.Start(); err != nil && err != http.ErrServerClosed {
			c.log.Error(fmt.Sprintf("管理面板启动失败: %v", err))
			c.status = "failed"
		}
	}()

	c.status = "running"
	return nil
}

// Stop 停止组件
func (c *AdminComponent) Stop(ctx context.Context) error {
	// 停止Web服务
	serverCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.webServer.Stop(serverCtx); err != nil {
		c.log.Error(fmt.Sprintf("停止管理面板服务失败: %v", err))
		return err
	}

	c.status = "stopped"
	c.log.Info("管理面板组件已停止")
	return nil
}

// HealthCheck 健康检查
func (c *AdminComponent) HealthCheck() error {
	if !c.initialized || c.status != "running" {
		return fmt.Errorf("管理面板组件未运行")
	}

	return nil
}

// Dependencies 返回组件依赖
func (c *AdminComponent) Dependencies() []string {
	return []string{"logger"}
}

// GetMetrics 获取组件指标
func (c *AdminComponent) GetMetrics() map[string]interface{} {
	return c.metrics
}

// registerRoutes 注册路由
func (c *AdminComponent) registerRoutes() {
	// 基础路由组
	baseGroup := c.webServer.Group(c.config.BasePath)

	// 健康检查路由
	baseGroup.GET("/health", func(c *gin.Context) {
		ctx := response.GetContext(c)
		ctx.Success(map[string]string{"status": "ok"})
	})

	// 创建处理器
	h := handlers.NewAdminHandler(c.log)

	// 注册管理处理器
	c.registerAdminHandlers(baseGroup, h)
}

// 注册管理处理器
func (c *AdminComponent) registerAdminHandlers(baseGroup *web.RouteGroup, h *handlers.AdminHandler) {
	// 创建认证中间件
	authMiddleware := middleware.JWT(c.config.SecretKey)

	// 登录路由（无需认证）
	baseGroup.POST("/login", h.HandleLogin)

	// 受保护路由组
	adminGroup := baseGroup.Group("", authMiddleware)
	{
		// 仪表盘
		adminGroup.GET("/dashboard", h.HandleDashboard)

		// 用户管理
		users := adminGroup.Group("/users")
		{
			users.GET("", h.HandleListUsers)
			users.POST("", h.HandleCreateUser)
			users.GET("/:id", h.HandleGetUser)
			users.PUT("/:id", h.HandleUpdateUser)
			users.DELETE("/:id", h.HandleDeleteUser)
		}

		// 系统管理
		system := adminGroup.Group("/system")
		{
			// 系统指标
			system.GET("/metrics", h.HandleSystemMetrics)

			// 健康状态
			system.GET("/health", h.HandleSystemHealth)

			// 组件管理
			components := system.Group("/components")
			{
				components.GET("", h.HandleListComponents)
				components.GET("/:name", h.HandleGetComponent)
				components.POST("/:name/toggle", h.HandleToggleComponent)
			}
		}
	}
}
