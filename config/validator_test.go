package config

import (
	"errors"
	"strings"
	"testing"
)

// 用于测试的结构体
type TestConfig struct {
	Name     string `json:"name" validate:"required"`
	Age      int    `json:"age" validate:"min=0,max=120"`
	Email    string `json:"email" validate:"email"`
	URL      string `json:"url" validate:"url"`
	Password string `json:"password" validate:"min=8"`
}

// 测试结构体验证函数
func TestValidateStruct(t *testing.T) {
	tests := []struct {
		name        string
		config      interface{}
		shouldError bool
		errorMsg    string
	}{
		{
			name: "Valid config",
			config: TestConfig{
				Name:     "Test User",
				Age:      30,
				Email:    "test@example.com",
				URL:      "https://example.com",
				Password: "password123",
			},
			shouldError: false,
		},
		{
			name: "Missing required field",
			config: TestConfig{
				Age:      30,
				Email:    "test@example.com",
				URL:      "https://example.com",
				Password: "password123",
			},
			shouldError: true,
			errorMsg:    "name' failed validation: 'required",
		},
		{
			name: "Invalid age (too high)",
			config: TestConfig{
				Name:     "Test User",
				Age:      150, // 超过最大值
				Email:    "test@example.com",
				URL:      "https://example.com",
				Password: "password123",
			},
			shouldError: true,
			errorMsg:    "age' failed validation: 'max",
		},
		{
			name: "Invalid email",
			config: TestConfig{
				Name:     "Test User",
				Age:      30,
				Email:    "invalid-email", // 无效的邮箱
				URL:      "https://example.com",
				Password: "password123",
			},
			shouldError: true,
			errorMsg:    "email' failed validation: 'email",
		},
		{
			name: "Invalid URL",
			config: TestConfig{
				Name:     "Test User",
				Age:      30,
				Email:    "test@example.com",
				URL:      "not-a-url", // 无效的URL
				Password: "password123",
			},
			shouldError: true,
			errorMsg:    "url' failed validation: 'url",
		},
		{
			name: "Password too short",
			config: TestConfig{
				Name:     "Test User",
				Age:      30,
				Email:    "test@example.com",
				URL:      "https://example.com",
				Password: "short", // 太短的密码
			},
			shouldError: true,
			errorMsg:    "password' failed validation: 'min",
		},
		{
			name:        "Invalid structure (nil)",
			config:      nil,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.config)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}

				// 针对nil情况的特殊处理
				if tt.name == "Invalid structure (nil)" {
					// 对于nil结构，只检查是否有错误返回，不检查错误类型
					return
				}

				if !errors.Is(err, ErrConfigValidationFailed) {
					t.Errorf("Expected error to wrap ErrConfigValidationFailed, but it didn't")
				}

				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%v'", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// 测试结构体验证器创建函数
func TestCreateStructValidator(t *testing.T) {
	// 创建测试结构体和验证器
	testConfig := &TestConfig{}
	validator := CreateStructValidator(testConfig)

	// 创建模拟提供者
	mockProvider := newMockProvider()

	// 测试有效配置
	mockProvider.values = map[string]interface{}{
		"name":     "Test User",
		"age":      30,
		"email":    "test@example.com",
		"url":      "https://example.com",
		"password": "password123",
	}

	mockProvider.unmarshalFunc = func(key string, val interface{}) error {
		if val, ok := val.(*TestConfig); ok {
			val.Name = "Test User"
			val.Age = 30
			val.Email = "test@example.com"
			val.URL = "https://example.com"
			val.Password = "password123"
			return nil
		}
		return nil
	}

	err := validator(mockProvider)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}

	// 测试无效配置
	mockProvider.unmarshalFunc = func(key string, val interface{}) error {
		if val, ok := val.(*TestConfig); ok {
			val.Age = 30
			val.Email = "invalid-email"
			val.URL = "not-a-url"
			val.Password = "short"
			return nil
		}
		return nil
	}

	err = validator(mockProvider)
	if err == nil {
		t.Error("Expected validation error but got nil")
	} else if !errors.Is(err, ErrConfigValidationFailed) {
		t.Errorf("Expected error to wrap ErrConfigValidationFailed, got: %v", err)
	}
}

// 测试必需配置项验证器
func TestRequiredConfigValidator(t *testing.T) {
	// 创建模拟提供者
	mockProvider := newMockProvider()

	// 测试所有必需项存在的情况
	mockProvider.values = map[string]interface{}{
		"app.name":    "test-app",
		"app.version": "1.0.0",
		"server.port": 8080,
	}

	validator := RequiredConfigValidator("app.name", "app.version", "server.port")
	err := validator(mockProvider)
	if err != nil {
		t.Errorf("Expected no error when all required keys exist, got: %v", err)
	}

	// 测试缺少必需项的情况
	mockProvider.values = map[string]interface{}{
		"app.name": "test-app",
		// app.version missing
		"server.port": 8080,
	}

	err = validator(mockProvider)
	if err == nil {
		t.Error("Expected error when required key is missing, got nil")
	} else if !errors.Is(err, ErrConfigValidationFailed) {
		t.Errorf("Expected error to wrap ErrConfigValidationFailed, got: %v", err)
	}
}

// 测试范围验证器
func TestRangeValidator(t *testing.T) {
	// 创建模拟提供者
	mockProvider := newMockProvider()

	// 测试有效范围
	mockProvider.values = map[string]interface{}{
		"server.port": 8080, // valid: 1-65535
	}

	validator := RangeValidator("server.port", 1, 65535)
	err := validator(mockProvider)
	if err != nil {
		t.Errorf("Expected no error for value in range, got: %v", err)
	}

	// 测试值太小
	mockProvider.values["server.port"] = 0 // 小于最小值
	err = validator(mockProvider)
	if err == nil {
		t.Error("Expected error for value below min, got nil")
	} else if !errors.Is(err, ErrConfigValidationFailed) {
		t.Errorf("Expected error to wrap ErrConfigValidationFailed, got: %v", err)
	}

	// 测试值太大
	mockProvider.values["server.port"] = 70000 // 大于最大值
	err = validator(mockProvider)
	if err == nil {
		t.Error("Expected error for value above max, got nil")
	} else if !errors.Is(err, ErrConfigValidationFailed) {
		t.Errorf("Expected error to wrap ErrConfigValidationFailed, got: %v", err)
	}

	// 测试键不存在的情况
	delete(mockProvider.values, "server.port")
	err = validator(mockProvider)
	if err != nil {
		t.Errorf("Expected no error when key doesn't exist, got: %v", err)
	}
}

// 测试枚举验证器
func TestEnumValidator(t *testing.T) {
	// 创建模拟提供者
	mockProvider := newMockProvider()

	// 有效的枚举值
	validValues := []string{"development", "testing", "staging", "production"}
	validator := EnumValidator("app.env", validValues...)

	// 测试有效值
	for _, validValue := range validValues {
		mockProvider.values = map[string]interface{}{
			"app.env": validValue,
		}

		err := validator(mockProvider)
		if err != nil {
			t.Errorf("Expected no error for valid enum value '%s', got: %v", validValue, err)
		}
	}

	// 测试无效值
	mockProvider.values["app.env"] = "invalid-env"
	err := validator(mockProvider)
	if err == nil {
		t.Error("Expected error for invalid enum value, got nil")
	} else if !errors.Is(err, ErrConfigValidationFailed) {
		t.Errorf("Expected error to wrap ErrConfigValidationFailed, got: %v", err)
	}

	// 测试键不存在的情况
	delete(mockProvider.values, "app.env")
	err = validator(mockProvider)
	if err != nil {
		t.Errorf("Expected no error when key doesn't exist, got: %v", err)
	}
}

// 测试URL验证器
func TestURLValidator(t *testing.T) {
	// 创建模拟提供者
	mockProvider := newMockProvider()
	validator := URLValidator("api.url")

	// 测试有效URL
	validURLs := []string{
		"http://example.com",
		"https://example.com",
		"http://localhost:8080",
		"https://sub.domain.example.com/path?query=value",
		// FTP URL 在某些验证库中被视为有效的URL协议
		"ftp://example.com",
	}

	for _, url := range validURLs {
		mockProvider.values = map[string]interface{}{
			"api.url": url,
		}

		err := validator(mockProvider)
		if err != nil {
			t.Errorf("Expected no error for valid URL '%s', got: %v", url, err)
		}
	}

	// 测试无效URL
	invalidURLs := []string{
		"not-a-url",
		"http://",
		"https://",
		"://example.com",
	}

	for _, url := range invalidURLs {
		mockProvider.values = map[string]interface{}{
			"api.url": url,
		}

		err := validator(mockProvider)
		if err == nil {
			t.Errorf("Expected error for invalid URL '%s', got nil", url)
		} else if !errors.Is(err, ErrConfigValidationFailed) {
			t.Errorf("Expected error to wrap ErrConfigValidationFailed, got: %v", err)
		}
	}

	// 测试键不存在的情况
	delete(mockProvider.values, "api.url")
	err := validator(mockProvider)
	if err != nil {
		t.Errorf("Expected no error when key doesn't exist, got: %v", err)
	}
}

// 测试Email验证器
func TestEmailValidator(t *testing.T) {
	// 创建模拟提供者
	mockProvider := newMockProvider()
	validator := EmailValidator("user.email")

	// 测试有效Email
	validEmails := []string{
		"user@example.com",
		"user.name@example.com",
		"user+tag@example.com",
		"user@sub.domain.example.com",
	}

	for _, email := range validEmails {
		mockProvider.values = map[string]interface{}{
			"user.email": email,
		}

		err := validator(mockProvider)
		if err != nil {
			t.Errorf("Expected no error for valid email '%s', got: %v", email, err)
		}
	}

	// 测试无效Email
	invalidEmails := []string{
		"not-an-email",
		"user@",
		"@example.com",
		"user@example.",
		"user@.com",
	}

	for _, email := range invalidEmails {
		mockProvider.values = map[string]interface{}{
			"user.email": email,
		}

		err := validator(mockProvider)
		if err == nil {
			t.Errorf("Expected error for invalid email '%s', got nil", email)
		} else if !errors.Is(err, ErrConfigValidationFailed) {
			t.Errorf("Expected error to wrap ErrConfigValidationFailed, got: %v", err)
		}
	}

	// 测试键不存在的情况
	delete(mockProvider.values, "user.email")
	err := validator(mockProvider)
	if err != nil {
		t.Errorf("Expected no error when key doesn't exist, got: %v", err)
	}
}

// 测试环境验证器
func TestEnvironmentValidator(t *testing.T) {
	// 创建模拟提供者
	mockProvider := newMockProvider()
	validator := EnvironmentValidator()

	// 测试有效环境
	validEnvs := []string{
		string(EnvDevelopment),
		string(EnvTesting),
		string(EnvStaging),
		string(EnvProduction),
	}

	for _, env := range validEnvs {
		mockProvider.values = map[string]interface{}{
			"app.env": env,
		}

		err := validator(mockProvider)
		if err != nil {
			t.Errorf("Expected no error for valid environment '%s', got: %v", env, err)
		}
	}

	// 测试无效环境
	mockProvider.values["app.env"] = "invalid-env"
	err := validator(mockProvider)
	if err == nil {
		t.Error("Expected error for invalid environment, got nil")
	} else if !errors.Is(err, ErrConfigValidationFailed) {
		t.Errorf("Expected error to wrap ErrConfigValidationFailed, got: %v", err)
	}

	// 测试键不存在的情况
	delete(mockProvider.values, "app.env")
	err = validator(mockProvider)
	if err != nil {
		t.Errorf("Expected no error when key doesn't exist, got: %v", err)
	}
}

// 测试验证器组合
func TestCombineValidators(t *testing.T) {
	// 创建模拟提供者
	mockProvider := newMockProvider()

	// 创建多个验证器
	reqValidator := RequiredConfigValidator("app.name", "app.env")
	envValidator := EnvironmentValidator()
	rangeValidator := RangeValidator("server.port", 1, 65535)

	// 组合验证器
	combinedValidator := CombineValidators(reqValidator, envValidator, rangeValidator)

	// 测试全部有效的情况
	mockProvider.values = map[string]interface{}{
		"app.name":    "test-app",
		"app.env":     "production",
		"server.port": 8080,
	}

	err := combinedValidator(mockProvider)
	if err != nil {
		t.Errorf("Expected no error for valid combined validation, got: %v", err)
	}

	// 测试第一个验证器失败
	delete(mockProvider.values, "app.name")
	err = combinedValidator(mockProvider)
	if err == nil {
		t.Error("Expected error when first validator fails, got nil")
	}

	// 恢复第一个验证器所需的值，测试第二个验证器失败
	mockProvider.values["app.name"] = "test-app"
	mockProvider.values["app.env"] = "invalid-env"
	err = combinedValidator(mockProvider)
	if err == nil {
		t.Error("Expected error when second validator fails, got nil")
	}

	// 恢复第二个验证器所需的值，测试第三个验证器失败
	mockProvider.values["app.env"] = "production"
	mockProvider.values["server.port"] = 70000 // 无效的端口号
	err = combinedValidator(mockProvider)
	if err == nil {
		t.Error("Expected error when third validator fails, got nil")
	}
}
