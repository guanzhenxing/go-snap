// Package validator 提供请求参数验证功能
package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/web/response"
)

// 验证器实例
var validate = validator.New()

// Setup 设置参数验证
func Setup() {
	// 将validator库设置为Gin的默认验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证器
		registerCustomValidators(v)
		// 使用JSON tag作为字段名
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return fld.Name
			}
			return name
		})
	}
}

// registerCustomValidators 注册自定义验证器
func registerCustomValidators(v *validator.Validate) {
	// 手机号验证器
	_ = v.RegisterValidation("mobile", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) == 11
	})

	// 可以添加更多自定义验证器...
}

// ValidateJSON 验证JSON请求体
func ValidateJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return handleValidationError(err)
	}
	return nil
}

// ValidateQuery 验证查询参数
func ValidateQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return handleValidationError(err)
	}
	return nil
}

// ValidateForm 验证表单参数
func ValidateForm(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBind(obj); err != nil {
		return handleValidationError(err)
	}
	return nil
}

// ValidateURI 验证URI参数
func ValidateURI(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindUri(obj); err != nil {
		return handleValidationError(err)
	}
	return nil
}

// handleValidationError 处理验证错误
func handleValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		// 收集所有验证错误
		errorMsgs := make([]string, 0)
		for _, e := range validationErrors {
			errorMsgs = append(errorMsgs, fmt.Sprintf("字段 '%s' 验证失败: %s", e.Field(), getValidationErrorMsg(e)))
		}
		return errors.New(strings.Join(errorMsgs, "; "))
	}
	return err
}

// getValidationErrorMsg 根据验证标签获取错误消息
func getValidationErrorMsg(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "必填字段"
	case "email":
		return "必须是有效的电子邮件地址"
	case "min":
		return fmt.Sprintf("长度或值必须大于或等于 %s", e.Param())
	case "max":
		return fmt.Sprintf("长度或值必须小于或等于 %s", e.Param())
	case "len":
		return fmt.Sprintf("长度必须等于 %s", e.Param())
	case "eq":
		return fmt.Sprintf("值必须等于 %s", e.Param())
	case "ne":
		return fmt.Sprintf("值不能等于 %s", e.Param())
	case "oneof":
		return fmt.Sprintf("值必须是 [%s] 其中之一", e.Param())
	case "mobile":
		return "必须是有效的手机号码"
	default:
		return fmt.Sprintf("验证不通过 (%s)", e.Tag())
	}
}

// ValidateMiddleware 参数验证中间件
func ValidateMiddleware(obj interface{}, bindType ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取自定义上下文
		ctx := response.GetContext(c)

		// 默认使用JSON绑定
		bType := "json"
		if len(bindType) > 0 {
			bType = bindType[0]
		}

		var err error
		switch bType {
		case "json":
			err = ValidateJSON(c, obj)
		case "query":
			err = ValidateQuery(c, obj)
		case "form":
			err = ValidateForm(c, obj)
		case "uri":
			err = ValidateURI(c, obj)
		default:
			err = errors.New("unsupported bind type")
		}

		if err != nil {
			if ctx != nil {
				ctx.BadRequest(err)
			} else {
				c.AbortWithStatusJSON(
					400,
					response.NewErrorResponse(
						response.CodeParamError,
						err.Error(),
						c.GetHeader("X-Request-ID"),
					),
				)
			}
			c.Abort()
			return
		}

		// 将验证通过的对象存储到上下文中
		c.Set("validated", obj)
		c.Next()
	}
}

// GetValidated 从上下文中获取已验证的对象
func GetValidated(c *gin.Context) interface{} {
	val, exists := c.Get("validated")
	if !exists {
		return nil
	}
	return val
}

// Struct 验证结构体
func Struct(s interface{}) error {
	return validate.Struct(s)
}
