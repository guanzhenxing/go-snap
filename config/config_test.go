package config

import (
	"os"
	"testing"
	"time"
)

// 模拟配置提供者，用于测试
type mockProvider struct {
	values        map[string]interface{}
	unmarshalFunc func(key string, val interface{}) error
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		values:        make(map[string]interface{}),
		unmarshalFunc: nil,
	}
}

func (m *mockProvider) Get(key string) interface{} {
	return m.values[key]
}

func (m *mockProvider) GetString(key string) string {
	if v, ok := m.values[key].(string); ok {
		return v
	}
	return ""
}

func (m *mockProvider) GetBool(key string) bool {
	if v, ok := m.values[key].(bool); ok {
		return v
	}
	return false
}

func (m *mockProvider) GetInt(key string) int {
	if v, ok := m.values[key].(int); ok {
		return v
	}
	return 0
}

func (m *mockProvider) GetInt64(key string) int64 {
	if v, ok := m.values[key].(int64); ok {
		return v
	}
	return 0
}

func (m *mockProvider) GetFloat64(key string) float64 {
	if v, ok := m.values[key].(float64); ok {
		return v
	}
	return 0
}

func (m *mockProvider) GetTime(key string) time.Time {
	if v, ok := m.values[key].(time.Time); ok {
		return v
	}
	return time.Time{}
}

func (m *mockProvider) GetDuration(key string) time.Duration {
	if v, ok := m.values[key].(time.Duration); ok {
		return v
	}
	return 0
}

func (m *mockProvider) GetStringSlice(key string) []string {
	if v, ok := m.values[key].([]string); ok {
		return v
	}
	return nil
}

func (m *mockProvider) GetStringMap(key string) map[string]interface{} {
	if v, ok := m.values[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}

func (m *mockProvider) GetStringMapString(key string) map[string]string {
	if v, ok := m.values[key].(map[string]string); ok {
		return v
	}
	return nil
}

func (m *mockProvider) IsSet(key string) bool {
	_, ok := m.values[key]
	return ok
}

func (m *mockProvider) Set(key string, value interface{}) {
	m.values[key] = value
}

func (m *mockProvider) UnmarshalKey(key string, rawVal interface{}) error {
	if m.unmarshalFunc != nil {
		return m.unmarshalFunc(key, rawVal)
	}
	return nil
}

func (m *mockProvider) Unmarshal(rawVal interface{}) error {
	if m.unmarshalFunc != nil {
		return m.unmarshalFunc("", rawVal)
	}
	return nil
}

func (m *mockProvider) LoadConfig() error {
	return nil
}

func (m *mockProvider) WatchConfig() {
}

func (m *mockProvider) OnConfigChange(run func()) {
}

func (m *mockProvider) ValidateConfig() error {
	return nil
}

// 测试环境判断函数
func TestEnvironmentFunctions(t *testing.T) {
	// 保存原始配置以便测试后恢复
	originalConfig := Config
	defer func() { Config = originalConfig }()

	// 初始化模拟配置提供者
	mockConfig := newMockProvider()
	Config = mockConfig

	// 测试默认环境（未设置时）
	env := GetCurrentEnvironment()
	if env != EnvDevelopment {
		t.Errorf("Expected default environment to be development, got %s", env)
	}

	// 测试各种环境设置
	testCases := []struct {
		envName       string
		expectedEnv   Environment
		isDevelopment bool
		isTesting     bool
		isStaging     bool
		isProduction  bool
	}{
		{
			envName:       "development",
			expectedEnv:   EnvDevelopment,
			isDevelopment: true,
			isTesting:     false,
			isStaging:     false,
			isProduction:  false,
		},
		{
			envName:       "testing",
			expectedEnv:   EnvTesting,
			isDevelopment: false,
			isTesting:     true,
			isStaging:     false,
			isProduction:  false,
		},
		{
			envName:       "staging",
			expectedEnv:   EnvStaging,
			isDevelopment: false,
			isTesting:     false,
			isStaging:     true,
			isProduction:  false,
		},
		{
			envName:       "production",
			expectedEnv:   EnvProduction,
			isDevelopment: false,
			isTesting:     false,
			isStaging:     false,
			isProduction:  true,
		},
		{
			envName:       "invalid",
			expectedEnv:   EnvDevelopment, // 默认为开发环境
			isDevelopment: true,
			isTesting:     false,
			isStaging:     false,
			isProduction:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.envName), func(t *testing.T) {
			mockConfig.Set("app.env", tc.envName)

			env := GetCurrentEnvironment()
			if env != tc.expectedEnv {
				t.Errorf("Expected environment to be %s, got %s", tc.expectedEnv, env)
			}

			if IsDevelopment() != tc.isDevelopment {
				t.Errorf("IsDevelopment() returned %v, expected %v", IsDevelopment(), tc.isDevelopment)
			}

			if IsTesting() != tc.isTesting {
				t.Errorf("IsTesting() returned %v, expected %v", IsTesting(), tc.isTesting)
			}

			if IsStaging() != tc.isStaging {
				t.Errorf("IsStaging() returned %v, expected %v", IsStaging(), tc.isStaging)
			}

			if IsProduction() != tc.isProduction {
				t.Errorf("IsProduction() returned %v, expected %v", IsProduction(), tc.isProduction)
			}
		})
	}
}

// 测试配置初始化函数
func TestInitConfig(t *testing.T) {
	// 保存原始配置以便测试后恢复
	originalConfig := Config
	defer func() { Config = originalConfig }()

	// 设置测试配置文件
	configFilePath := "testdata/app.yaml"

	// 确保测试文件存在
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		t.Skipf("Skipping test because test config file does not exist: %s", configFilePath)
	}

	// 初始化配置
	err := InitConfig(
		WithConfigName("app"),
		WithConfigType("yaml"),
		WithConfigPath("./testdata"),
		WithDefaultValue("test.key", "test_value"),
	)

	if err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}

	// 验证配置已正确初始化
	if Config == nil {
		t.Error("Config is nil after initialization")
	}

	// 验证可以获取默认值
	if Config.GetString("test.key") != "test_value" {
		t.Errorf("Expected default value 'test_value', got '%s'", Config.GetString("test.key"))
	}

	// 验证可以获取配置文件中的值
	appName := Config.GetString("app.name")
	if appName != "go-snap-app" {
		t.Errorf("Expected app.name to be 'go-snap-app', got '%s'", appName)
	}
}
