package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
	"github.com/guanzhenxing/go-snap/web"
	"github.com/guanzhenxing/go-snap/web/middleware"
	"github.com/guanzhenxing/go-snap/web/response"
	"github.com/guanzhenxing/go-snap/web/validator"
	"github.com/guanzhenxing/go-snap/web/websocket"
)

// 用户请求参数
type UserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"required,min=18,max=120"`
}

// 用户响应
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
}

// 这个示例展示了如何创建和使用Web服务
func Example() {
	// 初始化验证器
	validator.Setup()

	// 创建Web服务配置
	cfg := web.DefaultConfig()
	cfg.Host = "127.0.0.1"
	cfg.Port = 8080
	cfg.Mode = gin.DebugMode
	cfg.EnableSwagger = true
	cfg.EnableCORS = true
	cfg.LogRequests = true

	// 创建日志器
	log := logger.New()

	// 创建Web服务
	server := web.New(cfg, web.WithLogger(log))

	// 注册路由
	registerRoutes(server)

	// 在goroutine中启动服务器，以便可以优雅关闭
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Error(fmt.Sprintf("Failed to start web server: %v", err))
		}
	}()

	// 在实际应用中，这里会等待中断信号
	time.Sleep(500 * time.Millisecond)

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Error(fmt.Sprintf("Failed to stop web server: %v", err))
	}

	fmt.Println("Server stopped gracefully")
}

// 注册示例路由
func registerRoutes(server *web.Server) {
	// 注册健康检查路由
	server.GET("/health", func(c *gin.Context) {
		ctx := response.GetContext(c)
		ctx.Success(map[string]string{"status": "ok"})
	})

	// 创建API路由组
	apiGroup := server.Group("/api", middleware.RateLimit(100, time.Minute))

	// 注册V1版本API
	v1 := apiGroup.Group("/v1")
	{
		// 用户路由组
		users := v1.Group("/users")
		{
			// 获取用户列表
			users.GET("", listUsers)

			// 创建用户
			users.POST("", validator.ValidateMiddleware(UserRequest{}), createUser)

			// 获取用户详情
			users.GET("/:id", getUser)

			// 更新用户
			users.PUT("/:id", validator.ValidateMiddleware(UserRequest{}), updateUser)

			// 删除用户
			users.DELETE("/:id", deleteUser)
		}

		// WebSocket 示例
		ws := v1.Group("/ws")
		{
			wsHandler := websocket.NewHandler(
				websocket.WithPingInterval(30*time.Second),
				websocket.WithMaxMessageSize(1024*10),
			)
			ws.GET("/echo", wsHandler.Middleware())
		}
	}

	// 受保护的路由组 (需要JWT验证)
	protectedGroup := apiGroup.Group("/protected", middleware.JWT("your-secret-key"))
	{
		protectedGroup.GET("/profile", getProfile)
	}
}

// 处理函数示例
func listUsers(c *gin.Context) {
	ctx := response.GetContext(c)

	// 演示分页
	page := 1
	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := parseInt(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 10
	if pageSizeStr := ctx.Query("page_size"); pageSizeStr != "" {
		if ps, err := parseInt(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// 模拟数据库查询
	users := []UserResponse{
		{ID: "1", Username: "user1", Email: "user1@example.com", Age: 25},
		{ID: "2", Username: "user2", Email: "user2@example.com", Age: 30},
	}

	ctx.Success(map[string]interface{}{
		"users":      users,
		"page":       page,
		"page_size":  pageSize,
		"total":      2,
		"total_page": 1,
	})
}

func createUser(c *gin.Context) {
	ctx := response.GetContext(c)

	// 获取已验证的请求数据
	req := validator.GetValidated(c).(*UserRequest)

	// 创建用户响应
	user := &UserResponse{
		ID:       "3", // 模拟生成的ID
		Username: req.Username,
		Email:    req.Email,
		Age:      req.Age,
	}

	ctx.Success(user)
}

func getUser(c *gin.Context) {
	ctx := response.GetContext(c)
	id := ctx.Param("id")

	// 模拟数据库查询
	user := &UserResponse{
		ID:       id,
		Username: "user" + id,
		Email:    "user" + id + "@example.com",
		Age:      25,
	}

	ctx.Success(user)
}

func updateUser(c *gin.Context) {
	ctx := response.GetContext(c)
	id := ctx.Param("id")

	// 获取已验证的请求数据
	req := validator.GetValidated(c).(*UserRequest)

	// 更新用户
	user := &UserResponse{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
		Age:      req.Age,
	}

	ctx.Success(user)
}

func deleteUser(c *gin.Context) {
	ctx := response.GetContext(c)
	id := ctx.Param("id")

	// 模拟删除操作
	ctx.Success(map[string]interface{}{
		"id":      id,
		"deleted": true,
	})
}

func getProfile(c *gin.Context) {
	ctx := response.GetContext(c)

	// 获取认证用户信息
	// userID := c.GetString("user_id")

	profile := map[string]interface{}{
		"user_id":    "auth-user-id",
		"username":   "authenticated-user",
		"email":      "auth-user@example.com",
		"created_at": time.Now().Format(time.RFC3339),
	}

	ctx.Success(profile)
}

// 辅助函数
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse int")
	}
	return i, nil
}
