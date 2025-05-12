package web

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guanzhenxing/go-snap/logger"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, gin.ReleaseMode, cfg.Mode)
	assert.Equal(t, "", cfg.BasePath)
	assert.Equal(t, "1MB", cfg.BodyLimit)
	assert.Equal(t, time.Second*15, cfg.ReadTimeout)
	assert.Equal(t, time.Second*15, cfg.WriteTimeout)
	assert.True(t, cfg.EnableSwagger)
	assert.False(t, cfg.EnableProfiling)
	assert.True(t, cfg.EnableCORS)
	assert.True(t, cfg.LogRequests)
	assert.False(t, cfg.LogResponses)
	assert.Equal(t, []string{"127.0.0.1"}, cfg.TrustedProxies)
}

func TestNewServer(t *testing.T) {
	cfg := DefaultConfig()
	server := New(cfg)

	assert.NotNil(t, server)
	assert.NotNil(t, server.engine)
	assert.NotNil(t, server.router)
	assert.NotNil(t, server.httpServer)
	assert.NotNil(t, server.routeGroups)
}

func TestServerGroup(t *testing.T) {
	cfg := DefaultConfig()
	server := New(cfg)

	// 测试创建路由组
	group := server.Group("/api")
	assert.NotNil(t, group)
	assert.Equal(t, "/api", group.group.BasePath())

	// 测试路由组注册
	group.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	// 验证路由是否注册成功
	routes := server.engine.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/api/test" && route.Method == "GET" {
			found = true
			break
		}
	}
	assert.True(t, found, "Route /api/test should be registered")
}

func TestServerStartStop(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Port = 8081 // 使用不同的端口避免冲突
	server := New(cfg)

	// 启动服务器
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// 等待服务器启动
	time.Sleep(time.Second)

	// 停止服务器
	ctx := context.Background()
	err := server.Stop(ctx)
	assert.NoError(t, err)

	// 检查启动错误
	err = <-errChan
	assert.Contains(t, err.Error(), "Server closed")
}

func TestServerMiddleware(t *testing.T) {
	cfg := DefaultConfig()
	server := New(cfg)

	// 测试添加中间件
	customMiddleware := func(c *gin.Context) {
		c.Set("test", "middleware")
		c.Next()
	}

	server.Use(customMiddleware)

	// 注册测试路由
	server.GET("/middleware-test", func(c *gin.Context) {
		value, exists := c.Get("test")
		assert.True(t, exists)
		assert.Equal(t, "middleware", value)
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务器进行测试
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// 等待服务器启动
	time.Sleep(time.Second)

	// 停止服务器
	ctx := context.Background()
	err := server.Stop(ctx)
	assert.NoError(t, err)

	// 检查启动错误
	err = <-errChan
	assert.Contains(t, err.Error(), "Server closed")
}

func TestServerHTTPMethods(t *testing.T) {
	config := DefaultConfig()
	server := New(config)
	assert.NotNil(t, server)

	// 测试 HTTP 方法
	server.POST("/post", func(c *gin.Context) {})
	server.PUT("/put", func(c *gin.Context) {})
	server.DELETE("/delete", func(c *gin.Context) {})
	server.PATCH("/patch", func(c *gin.Context) {})
	server.OPTIONS("/options", func(c *gin.Context) {})
	server.HEAD("/head", func(c *gin.Context) {})

	// 验证路由是否已注册
	routes := server.engine.Routes()
	methodMap := make(map[string]bool)
	for _, route := range routes {
		methodMap[route.Method] = true
	}

	assert.True(t, methodMap["POST"])
	assert.True(t, methodMap["PUT"])
	assert.True(t, methodMap["DELETE"])
	assert.True(t, methodMap["PATCH"])
	assert.True(t, methodMap["OPTIONS"])
	assert.True(t, methodMap["HEAD"])
}

func TestServerGroupHTTPMethods(t *testing.T) {
	config := DefaultConfig()
	server := New(config)
	assert.NotNil(t, server)

	// 测试Group中的HTTP方法
	group := server.Group("/api")
	group.POST("/post", func(c *gin.Context) {})
	group.PUT("/put", func(c *gin.Context) {})
	group.DELETE("/delete", func(c *gin.Context) {})
	group.PATCH("/patch", func(c *gin.Context) {})
	group.OPTIONS("/options", func(c *gin.Context) {})
	group.HEAD("/head", func(c *gin.Context) {})

	// 验证路由是否已注册
	routes := server.engine.Routes()
	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route.Method+route.Path] = true
	}

	assert.True(t, routeMap["POST/api/post"])
	assert.True(t, routeMap["PUT/api/put"])
	assert.True(t, routeMap["DELETE/api/delete"])
	assert.True(t, routeMap["PATCH/api/patch"])
	assert.True(t, routeMap["OPTIONS/api/options"])
	assert.True(t, routeMap["HEAD/api/head"])
}

func TestGroupNested(t *testing.T) {
	config := DefaultConfig()
	server := New(config)
	assert.NotNil(t, server)

	// 嵌套的Group
	v1 := server.Group("/v1")
	users := v1.Group("/users")
	users.GET("/:id", func(c *gin.Context) {})
	users.Use(func(c *gin.Context) {})

	// 验证路由是否已注册
	routes := server.engine.Routes()
	found := false
	for _, route := range routes {
		if route.Method == "GET" && route.Path == "/v1/users/:id" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestWithLogger(t *testing.T) {
	// 创建测试logger
	testLog := logger.New()

	// 创建服务器并应用logger选项
	config := DefaultConfig()
	server := New(config, WithLogger(testLog))

	// 检查logger是否已设置
	assert.Equal(t, testLog, server.log)
}

func TestWithMiddleware(t *testing.T) {
	// 添加自定义中间件
	testMiddleware := func(c *gin.Context) {
		c.Set("test-key", "test-value")
		c.Next()
	}

	// 创建服务器并应用中间件选项
	config := DefaultConfig()
	server := New(config, WithMiddleware(testMiddleware))

	// 检查中间件是否已添加
	assert.Len(t, server.middlewares, 1)
}
