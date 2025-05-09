// Package middleware 提供了Web服务的中间件集合
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
func RequestSizeLimiter(maxSize string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用Gin内置的MaxBytes中间件可以实现，暂不自行实现
		c.Next()
	}
}

// Timeout 超时中间件
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
