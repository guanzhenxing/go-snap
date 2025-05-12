package response

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试用的结构体
type TestUser struct {
	Name string `json:"name" form:"name" validate:"required"`
}

func TestContextInit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)

		assert.NotNil(t, ctx.ginCtx)
		assert.Equal(t, http.StatusOK, ctx.status)
		assert.NotNil(t, ctx.body)
		assert.False(t, ctx.isStreaming)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}

func TestContextSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)

		data := map[string]interface{}{"message": "success"}
		ctx.Success(data)

		// 验证响应
		var response Response
		err := json.Unmarshal(ctx.body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, CodeSuccess, response.Code)
		assert.Equal(t, StatusMessages[CodeSuccess], response.Message)
		assert.Equal(t, data, response.Data)
		assert.NotEmpty(t, response.Timestamp)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}

func TestContextError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		code       int
		err        error
		expectCode int
	}{
		{
			name:       "param error",
			code:       CodeParamError,
			err:        errors.New("invalid parameter"),
			expectCode: CodeParamError,
		},
		{
			name:       "server error",
			code:       CodeServerError,
			err:        errors.New("internal server error"),
			expectCode: CodeServerError,
		},
		{
			name:       "unauthorized",
			code:       CodeUnauthorized,
			err:        errors.New("unauthorized access"),
			expectCode: CodeUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New() // 为每个测试创建新的路由实例
			router.GET("/test", func(c *gin.Context) {
				ctx := &Context{}
				ctx.Init(c)

				ctx.Error(tt.code, tt.err)

				// 验证响应
				var response Response
				err := json.Unmarshal(ctx.body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.expectCode, response.Code)
				assert.Equal(t, tt.err.Error(), response.Message)
				assert.NotEmpty(t, response.Timestamp)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
		})
	}
}

func TestContextBind(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name       string
		json       string
		expectPass bool
	}{
		{
			name:       "valid json",
			json:       `{"name": "test", "value": 123}`,
			expectPass: true,
		},
		{
			name:       "invalid json",
			json:       `{"name": "test", "value": "invalid"}`,
			expectPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New() // 每个子测试新建 router 实例
			router.POST("/test", func(c *gin.Context) {
				ctx := &Context{}
				ctx.Init(c)

				var data TestStruct
				err := ctx.BindJSON(&data)

				if tt.expectPass {
					assert.NoError(t, err)
					assert.Equal(t, "test", data.Name)
					assert.Equal(t, 123, data.Value)
				} else {
					assert.Error(t, err)
				}
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(tt.json))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
		})
	}
}

func TestContextQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)

		// 测试Query
		assert.Equal(t, "value", ctx.Query("key"))

		// 测试DefaultQuery
		assert.Equal(t, "default", ctx.DefaultQuery("nonexistent", "default"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?key=value", nil)
	router.ServeHTTP(w, req)
}

func TestContextStream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)

		ctx.Stream(http.StatusOK, "text/plain", func(w http.ResponseWriter) bool {
			w.Write([]byte("streaming data"))
			return true
		})

		assert.True(t, ctx.IsStreaming())
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, "streaming data", w.Body.String())
}

func TestNewErrorResponse(t *testing.T) {
	requestID := "test-request-id"
	response := NewErrorResponse(CodeParamError, "invalid parameter", requestID)

	assert.Equal(t, CodeParamError, response.Code)
	assert.Equal(t, "invalid parameter", response.Message)
	assert.Equal(t, requestID, response.RequestID)
	assert.NotEmpty(t, response.Timestamp)
}

func TestNewSuccessResponse(t *testing.T) {
	requestID := "test-request-id"
	data := map[string]string{"message": "success"}
	response := NewSuccessResponse(data, requestID)

	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, StatusMessages[CodeSuccess], response.Message)
	assert.Equal(t, data, response.Data)
	assert.Equal(t, requestID, response.RequestID)
	assert.NotEmpty(t, response.Timestamp)
}

func TestGetStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		ctx.Success("success")
		assert.Equal(t, 200, ctx.GetStatus())
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}

func TestGetBodyString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		ctx.Success("success")
		body := ctx.GetBodyString()
		assert.Contains(t, body, "success")
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}

func TestGinContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		assert.NotNil(t, ctx.GinContext())
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}

func TestParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test/:id", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		assert.Equal(t, "123", ctx.Param("id"))
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test/123", nil)
	router.ServeHTTP(w, req)
}

func TestBind(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		var user TestUser
		err := ctx.Bind(&user)
		assert.NoError(t, err)
		assert.Equal(t, "张三", user.Name)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{"name":"张三"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
}

func TestBindQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		var user TestUser
		err := ctx.BindQuery(&user)
		assert.NoError(t, err)
		assert.Equal(t, "张三", user.Name)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?name=张三", nil)
	router.ServeHTTP(w, req)
}

func TestString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		ctx.String(200, "success")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), "success")
	assert.Equal(t, 200, w.Code)
}

func TestHTML(t *testing.T) {
	// 先跳过这个测试，因为需要设置 HTML 模板才能运行
	t.Skip("Skipping HTML test as it requires template setup")
}

func TestFile(t *testing.T) {
	// 这个测试需要一个真实存在的文件，先跳过
	t.Skip("Skipping File test as it requires an actual file to exist")
}

func TestBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		ctx.BadRequest(errors.New("bad request"))
		assert.Equal(t, 400, ctx.GetStatus())
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}

func TestUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		ctx.Unauthorized(errors.New("unauthorized"))
		assert.Equal(t, 401, ctx.GetStatus())
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}

func TestForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		ctx.Forbidden(errors.New("forbidden"))
		assert.Equal(t, 403, ctx.GetStatus())
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}

func TestNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		ctx.NotFound(errors.New("not found"))
		assert.Equal(t, 404, ctx.GetStatus())
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}

func TestServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		ctx.ServerError(errors.New("server error"))
		assert.Equal(t, 500, ctx.GetStatus())
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}

func TestTooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		ctx.TooManyRequests(errors.New("too many requests"))
		assert.Equal(t, 429, ctx.GetStatus())
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}

func TestGetContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ctx := &Context{}
		ctx.Init(c)
		// 设置到 gin.Context 中
		c.Set("ctx", ctx)
		// 验证可以获取回来
		assert.NotNil(t, GetContext(c))
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
}
