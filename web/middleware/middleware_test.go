package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guanzhenxing/go-snap/logger"
	"github.com/stretchr/testify/assert"
)

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 添加恢复中间件
	router.Use(Recovery())

	// 测试panic恢复
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 创建测试日志器
	testLogger := logger.New()

	// 添加日志中间件
	router.Use(Logger(testLogger))

	// 测试请求
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 添加CORS中间件
	router.Use(CORS())

	// 测试OPTIONS请求
	router.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
}

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 添加限流中间件
	router.Use(RateLimit(2, time.Second))

	// 测试请求
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 发送3个请求，第3个应该被限流
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if i < 2 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 添加JWT中间件
	router.Use(JWT("test-secret"))

	// 测试请求
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 测试无token请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 添加超时中间件
	router.Use(Timeout(100 * time.Millisecond))

	// 测试超时请求
	router.GET("/timeout", func(c *gin.Context) {
		time.Sleep(200 * time.Millisecond)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/timeout", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestTimeout, w.Code)
}

func TestRequestSizeLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 添加请求大小限制中间件
	router.Use(RequestSizeLimiter("1KB"))

	// 测试请求
	router.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
