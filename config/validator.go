package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// ErrConfigValidationFailed 配置验证失败错误
	ErrConfigValidationFailed = errors.New("config validation failed")
)

// Validator 结构体验证器实例
var Validator = validator.New()

// 初始化验证器
func init() {
	// 注册结构体标签名称处理，使用json标签作为字段名
	Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateStruct 验证结构体是否符合要求
func ValidateStruct(s interface{}) error {
	err := Validator.Struct(s)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return fmt.Errorf("invalid validation error: %w", err)
		}

		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			errorMsgs := make([]string, 0, len(validationErrors))
			for _, e := range validationErrors {
				errorMsgs = append(errorMsgs, fmt.Sprintf(
					"Field '%s' failed validation: '%s'",
					e.Field(),
					e.Tag(),
				))
			}
			return fmt.Errorf("%w: %s", ErrConfigValidationFailed, strings.Join(errorMsgs, "; "))
		}

		return fmt.Errorf("validation error: %w", err)
	}

	return nil
}

// CreateStructValidator 创建结构体验证器
func CreateStructValidator(structPtr interface{}) ConfigValidator {
	return func(p Provider) error {
		// 创建一个新的结构体实例，而不是直接使用传入的指针
		// 这样可以避免修改原始结构体
		v := reflect.New(reflect.TypeOf(structPtr).Elem()).Interface()

		if err := p.Unmarshal(v); err != nil {
			return fmt.Errorf("failed to unmarshal config into validation struct: %w", err)
		}

		if err := ValidateStruct(v); err != nil {
			return err
		}

		return nil
	}
}

// RequiredConfigValidator 创建必需配置项验证器
func RequiredConfigValidator(keys ...string) ConfigValidator {
	return func(p Provider) error {
		var missingKeys []string

		for _, key := range keys {
			if !p.IsSet(key) {
				missingKeys = append(missingKeys, key)
			}
		}

		if len(missingKeys) > 0 {
			return fmt.Errorf("%w: missing required configuration keys: %s",
				ErrConfigValidationFailed, strings.Join(missingKeys, ", "))
		}

		return nil
	}
}

// RangeValidator 创建数值范围验证器
func RangeValidator(key string, min, max int) ConfigValidator {
	return func(p Provider) error {
		if !p.IsSet(key) {
			return nil // 如果配置项不存在，不进行验证
		}

		value := p.GetInt(key)
		if value < min || value > max {
			return fmt.Errorf("%w: %s must be between %d and %d, got %d",
				ErrConfigValidationFailed, key, min, max, value)
		}

		return nil
	}
}

// EnumValidator 创建枚举值验证器
func EnumValidator(key string, validValues ...string) ConfigValidator {
	return func(p Provider) error {
		if !p.IsSet(key) {
			return nil
		}

		value := p.GetString(key)

		for _, valid := range validValues {
			if value == valid {
				return nil
			}
		}

		return fmt.Errorf("%w: %s must be one of [%s], got '%s'",
			ErrConfigValidationFailed, key, strings.Join(validValues, ", "), value)
	}
}

// URLValidator 创建URL格式验证器
func URLValidator(key string) ConfigValidator {
	return func(p Provider) error {
		if !p.IsSet(key) {
			return nil
		}

		value := p.GetString(key)

		// 使用validator库中的URL验证标签
		if err := Validator.Var(value, "url"); err != nil {
			return fmt.Errorf("%w: %s must be a valid URL, got '%s'",
				ErrConfigValidationFailed, key, value)
		}

		return nil
	}
}

// EmailValidator 创建Email格式验证器
func EmailValidator(key string) ConfigValidator {
	return func(p Provider) error {
		if !p.IsSet(key) {
			return nil
		}

		value := p.GetString(key)

		// 使用validator库中的email验证标签
		if err := Validator.Var(value, "email"); err != nil {
			return fmt.Errorf("%w: %s must be a valid email address, got '%s'",
				ErrConfigValidationFailed, key, value)
		}

		return nil
	}
}

// EnvironmentValidator 创建环境名称验证器
func EnvironmentValidator() ConfigValidator {
	return EnumValidator("app.env",
		string(EnvDevelopment),
		string(EnvTesting),
		string(EnvStaging),
		string(EnvProduction),
	)
}

// CombineValidators 组合多个验证器
func CombineValidators(validators ...ConfigValidator) ConfigValidator {
	return func(p Provider) error {
		for _, validator := range validators {
			if err := validator(p); err != nil {
				return err
			}
		}
		return nil
	}
}
