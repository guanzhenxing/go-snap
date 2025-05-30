// Package middleware 提供了Web服务的中间件集合
// 包含常用的HTTP中间件实现，如恢复处理、日志记录、CORS、限流、JWT认证等
// 所有中间件都遵循Gin的中间件接口规范，可直接与Gin框架集成使用
//
// 中间件顺序推荐：
// 1. Recovery - 保证发生panic时能够恢复并返回500错误
// 2. Logger - 记录请求信息，便于追踪和调试
// 3. CORS - 处理跨域请求
// 4. RateLimit - 限制请求频率，防止滥用
// 5. RequestSizeLimiter - 限制请求体大小，防止大请求DOS攻击
// 6. JWT/认证 - 验证用户身份
// 7. 业务中间件 - 应用特定的业务逻辑
//
// 使用示例：
//
//	router := gin.New()
//
//	// 添加恢复和日志中间件
//	router.Use(middleware.Recovery())
//	router.Use(middleware.Logger(myLogger))
//
//	// 添加认证中间件到特定路由组
//	authorized := router.Group("/api")
//	authorized.Use(middleware.JWT("your-secret-key"))
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
	"github.com/guanzhenxing/go-snap/web/response"
)

// Recovery 恢复中间件，捕获panic并返回500错误
// 此中间件应该作为第一个中间件添加，以确保所有panic都能被捕获和处理
// 功能：
//   - 捕获处理链中任何位置的panic
//   - 将panic转换为友好的JSON错误响应
//   - 防止服务器因未处理的panic而崩溃
//
// 使用场景：
//   - 建议在所有生产环境中启用
//   - 对于高可用性要求的API服务尤为重要
//
// 示例：
//
//	router := gin.New() // 不使用gin.Default()，因为它已包含恢复中间件
//	router.Use(middleware.Recovery())
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch x := r.(type) {
				case string:
					err = errors.New(x)
				case error:
					err = x
				default:
					err = errors.New("unknown panic")
				}

				// 获取自定义上下文
				ctx := response.GetContext(c)
				if ctx != nil {
					ctx.ServerError(err)
				} else {
					// 直接返回JSON
					c.AbortWithStatusJSON(
						http.StatusInternalServerError,
						response.NewErrorResponse(
							response.CodeServerError,
							err.Error(),
							c.GetHeader("X-Request-ID"),
						),
					)
				}

				c.Abort()
			}
		}()

		c.Next()
	}
}

// Logger 日志中间件，记录请求日志
// 功能：
//   - 记录请求方法、路径、状态码和响应时间
//   - 生成并传递唯一请求ID (X-Request-ID)
//   - 在响应头中返回请求ID，便于客户端追踪
//
// 参数：
//
//	log: 日志记录器实例，用于写入日志
//
// 性能影响：
//   - 对每个请求增加少量开销，用于计时和日志写入
//   - 使用异步日志记录器可以减少对请求处理时间的影响
//
// 示例：
//
//	// 创建日志记录器
//	log := logger.New(logger.WithConsole(true))
//
//	// 添加日志中间件
//	router.Use(middleware.Logger(log))
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 生成请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("X-Request-ID", requestID)
		c.Header("X-Request-ID", requestID)

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 状态码
		statusCode := c.Writer.Status()
		path := c.Request.URL.Path

		// 记录请求日志
		log.Info(
			fmt.Sprintf("[%s] %s %s %d %v",
				requestID,
				c.Request.Method,
				path,
				statusCode,
				latency,
			),
		)
	}
}

// CORS 跨域中间件
// 功能：
//   - 处理跨域资源共享（CORS）
//   - 配置允许的源、方法、头部等
//   - 处理预检请求（OPTIONS）
//
// 安全注意事项：
//   - 默认配置允许所有源（*），生产环境中应根据需要限制
//   - 允许敏感头部如Authorization可能存在安全风险
//
// 性能：
//   - 对预检请求的处理可能增加额外的HTTP请求
//
// 示例：
//
//	// 使用默认配置的CORS中间件
//	router.Use(middleware.CORS())
//
//	// 自定义CORS配置（如需更严格的控制）
//	config := cors.Config{
//	    AllowOrigins: []string{"https://example.com"},
//	    AllowMethods: []string{"GET", "POST"},
//	}
//	router.Use(cors.New(config))
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Cache-Control", "X-Requested-With", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// RateLimit 限流中间件
// 功能：
//   - 基于令牌桶算法限制请求频率
//   - 按客户端IP地址进行限流
//   - 超过限制时返回429状态码
//
// 参数：
//
//	limit: 令牌桶容量和初始令牌数
//	duration: 生成一个新令牌的时间间隔
//
// 局限性：
//   - 使用内存存储限流状态，重启后重置
//   - 在多实例部署时不共享限流状态
//   - 仅基于IP地址限流，可能不适合所有场景
//
// 示例：
//
//	// 限制每个IP每秒最多5个请求
//	router.Use(middleware.RateLimit(5, time.Second))
//
//	// 对特定路由组应用更严格的限流
//	apiGroup := router.Group("/api")
//	apiGroup.Use(middleware.RateLimit(2, time.Second))
func RateLimit(limit int, duration time.Duration) gin.HandlerFunc {
	// 创建令牌桶限流器
	type TokenBucket struct {
		tokens     int
		capacity   int
		rate       time.Duration
		lastRefill time.Time
	}

	buckets := make(map[string]*TokenBucket)

	return func(c *gin.Context) {
		// 获取请求IP作为限流标识
		ip := c.ClientIP()

		// 获取或创建该IP的令牌桶
		bucket, exists := buckets[ip]
		if !exists {
			bucket = &TokenBucket{
				tokens:     limit,
				capacity:   limit,
				rate:       duration,
				lastRefill: time.Now(),
			}
			buckets[ip] = bucket
		}

		// 重新填充令牌
		now := time.Now()
		elapsed := now.Sub(bucket.lastRefill)
		tokensToAdd := int(elapsed / bucket.rate)

		if tokensToAdd > 0 {
			bucket.tokens = bucket.tokens + tokensToAdd
			if bucket.tokens > bucket.capacity {
				bucket.tokens = bucket.capacity
			}
			bucket.lastRefill = now
		}

		// 没有令牌时拒绝请求
		if bucket.tokens <= 0 {
			ctx := response.GetContext(c)
			if ctx != nil {
				ctx.TooManyRequests(errors.New("rate limit exceeded"))
			} else {
				c.AbortWithStatusJSON(
					http.StatusTooManyRequests,
					response.NewErrorResponse(
						response.CodeTooManyReqs,
						"请求过于频繁，请稍后再试",
						c.GetHeader("X-Request-ID"),
					),
				)
			}
			c.Abort()
			return
		}

		// 消耗一个令牌
		bucket.tokens--

		c.Next()
	}
}

// JWT JWT验证中间件
// 功能：
//   - 验证请求头中的JWT令牌
//   - 拒绝未授权或令牌无效的请求
//
// 参数：
//
//	secret: JWT签名密钥
//
// 安全注意事项：
//   - 应使用足够强度的密钥
//   - 生产环境应考虑令牌轮换和过期策略
//   - 敏感操作应验证令牌中的声明内容
//
// 注意：当前实现是一个框架，实际JWT验证逻辑需要完成
// 示例：
//
//	// 创建受保护的路由组
//	protected := router.Group("/api")
//	protected.Use(middleware.JWT("your-strong-secret-key"))
//
//	// 添加受保护的路由
//	protected.GET("/profile", handlers.GetProfile)
func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			ctx := response.GetContext(c)
			if ctx != nil {
				ctx.Unauthorized(errors.New("missing authorization header"))
			} else {
				c.AbortWithStatusJSON(
					http.StatusUnauthorized,
					response.NewErrorResponse(
						response.CodeUnauthorized,
						"未提供授权信息",
						c.GetHeader("X-Request-ID"),
					),
				)
			}
			c.Abort()
			return
		}

		// TODO: 实现JWT验证逻辑

		c.Next()
	}
}

// RequestSizeLimiter 请求大小限制中间件
// 功能：
//   - 限制HTTP请求体的最大大小
//   - 防止大请求导致服务器内存压力或DOS攻击
//
// 参数：
//
//	maxSize: 请求体最大大小，如"10M"、"1G"等
//
// 安全注意事项：
//   - 应根据API需求设置合理的限制大小
//   - 上传接口需要设置更大的限制，普通API可设置较小的限制
//
// 注意：
//   - 当前实现是框架，具体实现将使用Gin的MaxBytes中间件
//
// 示例：
//
//	// 限制请求体最大为8MB
//	router.Use(middleware.RequestSizeLimiter("8M"))
//
//	// 为上传接口单独设置更大的限制
//	uploadGroup := router.Group("/upload")
//	uploadGroup.Use(middleware.RequestSizeLimiter("50M"))
func RequestSizeLimiter(maxSize string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用Gin内置的MaxBytes中间件可以实现，暂不自行实现
		c.Next()
	}
}

// Timeout 超时中间件
// 功能：
//   - 为请求处理设置最大执行时间
//   - 超过时间限制时中断请求并返回超时响应
//   - 防止长时间运行的请求占用服务器资源
//
// 参数：
//
//	timeout: 请求处理的最大允许时间
//
// 性能注意事项：
//   - 每个请求会创建一个额外的goroutine
//   - 使用context.WithTimeout可能导致部分三方库无法正确响应取消信号
//
// 边界条件：
//   - 对于流式响应或WebSocket等长连接请求不应使用此中间件
//   - 文件上传等长时间操作应设置更长的超时时间
//
// 示例：
//
//	// 为所有请求设置5秒超时
//	router.Use(middleware.Timeout(5 * time.Second))
//
//	// 为特定API设置不同的超时
//	reportGroup := router.Group("/reports")
//	reportGroup.Use(middleware.Timeout(30 * time.Second)) // 报表生成需要更长时间
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建具有超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 使用新的上下文替换原始请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 使用通道来同步
		done := make(chan struct{})

		// 放在goroutine中执行后续处理器
		go func() {
			c.Next()
			done <- struct{}{}
		}()

		// 等待完成或超时
		select {
		case <-done:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			if ctx.Err() == context.DeadlineExceeded {
				rctx := response.GetContext(c)
				if rctx != nil {
					rctx.Error(response.CodeTimeout, errors.New("request timeout"))
				} else {
					c.AbortWithStatusJSON(
						http.StatusRequestTimeout,
						response.NewErrorResponse(
							response.CodeTimeout,
							"请求超时",
							c.GetHeader("X-Request-ID"),
						),
					)
				}
				c.Abort()
			}
		}
	}
}
