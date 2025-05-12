package validator

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// 测试用的结构体
type TestUser struct {
	Name     string `json:"name" form:"name" validate:"required"`
	Email    string `json:"email" form:"email" validate:"required,email"`
	Age      int    `json:"age" form:"age" validate:"required,min=18"`
	Mobile   string `json:"mobile" form:"mobile" validate:"required,mobile"`
	Password string `json:"password" form:"password" validate:"required,min=6"`
}

func TestSetup(t *testing.T) {
	Setup()
	// 验证是否成功设置
	assert.NotNil(t, binding.Validator.Engine())
}

func TestValidateJSON(t *testing.T) {
	Setup()
	gin.SetMode(gin.TestMode)

	// 测试用例
	tests := []struct {
		name       string
		json       string
		expectPass bool
	}{
		{
			name: "valid data",
			json: `{
				"name": "张三",
				"email": "zhangsan@example.com",
				"age": 20,
				"mobile": "13800138000",
				"password": "123456"
			}`,
			expectPass: true,
		},
		{
			name: "invalid email",
			json: `{
				"name": "张三",
				"email": "invalid-email",
				"age": 20,
				"mobile": "13800138000",
				"password": "123456"
			}`,
			expectPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New() // 为每个测试创建新的路由实例
			router.POST("/test", func(c *gin.Context) {
				var user TestUser
				err := ValidateJSON(c, &user)
				if tt.expectPass {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
					if err != nil {
						assert.Contains(t, err.Error(), "email")
					}
				}
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(tt.json))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
		})
	}
}

func TestValidateQuery(t *testing.T) {
	Setup()
	gin.SetMode(gin.TestMode)

	// 测试用例
	tests := []struct {
		name       string
		query      string
		expectPass bool
	}{
		{
			name:       "valid query",
			query:      "name=张三&email=zhangsan@example.com&age=20&mobile=13800138000&password=123456",
			expectPass: true,
		},
		{
			name:       "invalid email",
			query:      "name=张三&email=invalid-email&age=20&mobile=13800138000&password=123456",
			expectPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New() // 为每个测试创建新的路由实例
			router.GET("/test", func(c *gin.Context) {
				var user TestUser
				err := ValidateQuery(c, &user)
				if tt.expectPass {
					assert.NoError(t, err)
					// 额外断言字段绑定
					assert.Equal(t, "张三", user.Name)
					assert.Equal(t, "zhangsan@example.com", user.Email)
					assert.Equal(t, 20, user.Age)
					assert.Equal(t, "13800138000", user.Mobile)
					assert.Equal(t, "123456", user.Password)
				} else {
					assert.Error(t, err)
					if err != nil {
						assert.Contains(t, err.Error(), "email")
					}
				}
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test?"+tt.query, nil)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(w, req)
		})
	}
}

func TestValidateMiddleware(t *testing.T) {
	Setup()
	gin.SetMode(gin.TestMode)

	// 测试用例
	tests := []struct {
		name       string
		json       string
		expectCode int
	}{
		{
			name: "valid data",
			json: `{
				"name": "张三",
				"email": "zhangsan@example.com",
				"age": 20,
				"mobile": "13800138000",
				"password": "123456"
			}`,
			expectCode: 200,
		},
		{
			name: "invalid data",
			json: `{
				"name": "张三",
				"email": "invalid-email",
				"age": 16,
				"mobile": "123",
				"password": "123"
			}`,
			expectCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New() // 每个子测试新建 router 实例
			router.POST("/test", ValidateMiddleware(&TestUser{}), func(c *gin.Context) {
				c.Status(200)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(tt.json))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectCode, w.Code)
		})
	}
}

func TestGetValidated(t *testing.T) {
	Setup()
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/test", ValidateMiddleware(&TestUser{}), func(c *gin.Context) {
		validated := GetValidated(c)
		assert.NotNil(t, validated)
		user, ok := validated.(*TestUser)
		assert.True(t, ok)
		assert.Equal(t, "张三", user.Name)
		c.Status(200)
	})

	w := httptest.NewRecorder()
	json := `{
		"name": "张三",
		"email": "zhangsan@example.com",
		"age": 20,
		"mobile": "13800138000",
		"password": "123456"
	}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(json))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestStruct(t *testing.T) {
	Setup()
	// 测试用例
	tests := []struct {
		name       string
		user       TestUser
		expectPass bool
	}{
		{
			name: "valid user",
			user: TestUser{
				Name:     "张三",
				Email:    "zhangsan@example.com",
				Age:      20,
				Mobile:   "13800138000",
				Password: "123456",
			},
			expectPass: true,
		},
		{
			name: "invalid user",
			user: TestUser{
				Name:     "张三",
				Email:    "invalid-email",
				Age:      16,
				Mobile:   "123",
				Password: "123",
			},
			expectPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Struct(tt.user)
			if tt.expectPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateForm(t *testing.T) {
	Setup()
	gin.SetMode(gin.TestMode)

	// 测试用例
	tests := []struct {
		name       string
		form       string
		expectPass bool
	}{
		{
			name:       "valid form",
			form:       "name=张三&email=zhangsan@example.com&age=20&mobile=13800138000&password=123456",
			expectPass: true,
		},
		{
			name:       "invalid form",
			form:       "name=张三&email=invalid-email&age=20&mobile=13800138000&password=123456",
			expectPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New() // 为每个测试创建新的路由实例
			router.POST("/test", func(c *gin.Context) {
				var user TestUser
				err := ValidateForm(c, &user)
				if tt.expectPass {
					assert.NoError(t, err)
					assert.Equal(t, "张三", user.Name)
				} else {
					assert.Error(t, err)
					if err != nil {
						assert.Contains(t, err.Error(), "email")
					}
				}
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(tt.form))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(w, req)
		})
	}
}

func TestValidateURI(t *testing.T) {
	Setup()
	gin.SetMode(gin.TestMode)

	type URIParams struct {
		ID   string `uri:"id" validate:"required"`
		Type string `uri:"type" validate:"required,oneof=user admin"`
	}

	// 测试用例
	tests := []struct {
		name       string
		uri        string
		expectPass bool
	}{
		{
			name:       "valid uri",
			uri:        "/test/123/user",
			expectPass: true,
		},
		{
			name:       "invalid uri",
			uri:        "/test/123/guest",
			expectPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New() // 为每个测试创建新的路由实例
			router.GET("/test/:id/:type", func(c *gin.Context) {
				var params URIParams
				err := ValidateURI(c, &params)
				if tt.expectPass {
					assert.NoError(t, err)
					assert.Equal(t, "123", params.ID)
					assert.Equal(t, "user", params.Type)
				} else {
					assert.Error(t, err)
				}
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.uri, nil)
			router.ServeHTTP(w, req)
		})
	}
}

// 测试更多验证错误消息
func TestGetValidationErrorMsg(t *testing.T) {
	Setup()

	// 更全面测试 getValidationErrorMsg 中的条件分支
	// 创建一个复杂的测试用结构体
	type ComplexTestStruct struct {
		Required string `validate:"required"`
		Length   string `validate:"len=5"`
		Min      int    `validate:"min=10"`
		Max      int    `validate:"max=20"`
		Equal    int    `validate:"eq=15"`
		NotEqual int    `validate:"ne=15"`
		OneOf    string `validate:"oneof=a b c"`
	}

	// 多种情况验证错误
	tests := []struct {
		name     string
		testData ComplexTestStruct
		field    string
		contains string
	}{
		{
			name:     "required error",
			testData: ComplexTestStruct{},
			field:    "Required",
			contains: "必填字段",
		},
		{
			name:     "length error",
			testData: ComplexTestStruct{Required: "test", Length: "1234"},
			field:    "Length",
			contains: "长度必须等于 5",
		},
		{
			name:     "min error",
			testData: ComplexTestStruct{Required: "test", Length: "12345", Min: 5},
			field:    "Min",
			contains: "长度或值必须大于或等于 10",
		},
		{
			name:     "max error",
			testData: ComplexTestStruct{Required: "test", Length: "12345", Min: 10, Max: 25},
			field:    "Max",
			contains: "长度或值必须小于或等于 20",
		},
		{
			name:     "eq error",
			testData: ComplexTestStruct{Required: "test", Length: "12345", Min: 10, Max: 15, Equal: 10},
			field:    "Equal",
			contains: "值必须等于 15",
		},
		{
			name:     "ne error",
			testData: ComplexTestStruct{Required: "test", Length: "12345", Min: 10, Max: 15, Equal: 15, NotEqual: 15},
			field:    "NotEqual",
			contains: "值不能等于 15",
		},
		{
			name:     "oneof error",
			testData: ComplexTestStruct{Required: "test", Length: "12345", Min: 10, Max: 15, Equal: 15, NotEqual: 10, OneOf: "d"},
			field:    "OneOf",
			contains: "值必须是 [a b c] 其中之一",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Struct(tt.testData)
			assert.Error(t, err)

			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				found := false
				for _, e := range validationErrors {
					if e.Field() == tt.field {
						found = true
						msg := getValidationErrorMsg(e)
						assert.Contains(t, msg, tt.contains)
						break
					}
				}
				assert.True(t, found, "未找到字段 %s 的验证错误", tt.field)
			}
		})
	}
}
