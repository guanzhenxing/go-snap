package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 测试默认选项
func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	// 检查默认值是否正确
	if opts.ConfigName != "config" {
		t.Errorf("Expected default ConfigName to be 'config', got '%s'", opts.ConfigName)
	}

	if opts.ConfigType != "yaml" {
		t.Errorf("Expected default ConfigType to be 'yaml', got '%s'", opts.ConfigType)
	}

	if len(opts.ConfigPaths) != 1 || opts.ConfigPaths[0] != "./configs" {
		t.Errorf("Expected default ConfigPaths to be ['./configs'], got %v", opts.ConfigPaths)
	}

	if opts.EnvPrefix != "APP" {
		t.Errorf("Expected default EnvPrefix to be 'APP', got '%s'", opts.EnvPrefix)
	}

	if !opts.AutomaticEnv {
		t.Errorf("Expected default AutomaticEnv to be true")
	}

	if opts.WatchConfigFile {
		t.Errorf("Expected default WatchConfigFile to be false")
	}

	// 检查默认值映射
	expectedDefaults := map[string]interface{}{
		"app.env":   "development",
		"app.debug": true,
		"app.name":  "go-snap-app",
	}

	for key, expectedValue := range expectedDefaults {
		value, exists := opts.DefaultValues[key]
		if !exists {
			t.Errorf("Expected default value for '%s' to exist", key)
			continue
		}

		if value != expectedValue {
			t.Errorf("Expected default value for '%s' to be '%v', got '%v'", key, expectedValue, value)
		}
	}
}

// 测试选项设置函数
func TestOptionFunctions(t *testing.T) {
	// 创建用于测试的临时目录
	testDir, err := os.MkdirTemp("", "config-options-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	tests := []struct {
		name     string
		option   Option
		validate func(t *testing.T, opts *Options)
	}{
		{
			name:   "WithConfigName",
			option: WithConfigName("test-config"),
			validate: func(t *testing.T, opts *Options) {
				if opts.ConfigName != "test-config" {
					t.Errorf("Expected ConfigName to be 'test-config', got '%s'", opts.ConfigName)
				}
			},
		},
		{
			name:   "WithConfigType",
			option: WithConfigType("json"),
			validate: func(t *testing.T, opts *Options) {
				if opts.ConfigType != "json" {
					t.Errorf("Expected ConfigType to be 'json', got '%s'", opts.ConfigType)
				}
			},
		},
		{
			name:   "WithConfigPath",
			option: WithConfigPath(testDir),
			validate: func(t *testing.T, opts *Options) {
				if len(opts.ConfigPaths) != 1 || opts.ConfigPaths[0] != testDir {
					t.Errorf("Expected ConfigPaths to be ['%s'], got %v", testDir, opts.ConfigPaths)
				}
			},
		},
		{
			name:   "WithConfigPaths",
			option: WithConfigPaths([]string{testDir, "./configs"}),
			validate: func(t *testing.T, opts *Options) {
				expected := []string{testDir, "./configs"}
				if len(opts.ConfigPaths) != len(expected) {
					t.Errorf("Expected ConfigPaths length to be %d, got %d", len(expected), len(opts.ConfigPaths))
				} else {
					for i, path := range expected {
						if opts.ConfigPaths[i] != path {
							t.Errorf("Expected ConfigPaths[%d] to be '%s', got '%s'", i, path, opts.ConfigPaths[i])
						}
					}
				}
			},
		},
		{
			name:   "AddConfigPath",
			option: AddConfigPath("./extra-configs"),
			validate: func(t *testing.T, opts *Options) {
				found := false
				for _, path := range opts.ConfigPaths {
					if path == "./extra-configs" {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected './extra-configs' to be in ConfigPaths, got %v", opts.ConfigPaths)
				}
			},
		},
		{
			name:   "WithEnvPrefix",
			option: WithEnvPrefix("TEST"),
			validate: func(t *testing.T, opts *Options) {
				if opts.EnvPrefix != "TEST" {
					t.Errorf("Expected EnvPrefix to be 'TEST', got '%s'", opts.EnvPrefix)
				}
			},
		},
		{
			name:   "WithAutomaticEnv",
			option: WithAutomaticEnv(false),
			validate: func(t *testing.T, opts *Options) {
				if opts.AutomaticEnv {
					t.Errorf("Expected AutomaticEnv to be false")
				}
			},
		},
		{
			name:   "WithEnvKeyReplacer",
			option: WithEnvKeyReplacer(strings.NewReplacer("-", "_")),
			validate: func(t *testing.T, opts *Options) {
				if opts.EnvKeyReplacer == nil {
					t.Errorf("Expected EnvKeyReplacer to be set")
				} else {
					// 测试替换器的功能
					replaced := opts.EnvKeyReplacer.Replace("test-key")
					expected := "test_key"
					if replaced != expected {
						t.Errorf("Expected replaced string to be '%s', got '%s'", expected, replaced)
					}
				}
			},
		},
		{
			name:   "WithWatchConfigFile",
			option: WithWatchConfigFile(true),
			validate: func(t *testing.T, opts *Options) {
				if !opts.WatchConfigFile {
					t.Errorf("Expected WatchConfigFile to be true")
				}
			},
		},
		{
			name:   "WithDefaultValue",
			option: WithDefaultValue("test.key", "test-value"),
			validate: func(t *testing.T, opts *Options) {
				value, exists := opts.DefaultValues["test.key"]
				if !exists {
					t.Errorf("Expected default value for 'test.key' to exist")
				} else if value != "test-value" {
					t.Errorf("Expected default value for 'test.key' to be 'test-value', got '%v'", value)
				}
			},
		},
		{
			name: "WithDefaultValues",
			option: WithDefaultValues(map[string]interface{}{
				"test.key1": "value1",
				"test.key2": 42,
			}),
			validate: func(t *testing.T, opts *Options) {
				value1, exists1 := opts.DefaultValues["test.key1"]
				value2, exists2 := opts.DefaultValues["test.key2"]

				if !exists1 {
					t.Errorf("Expected default value for 'test.key1' to exist")
				} else if value1 != "value1" {
					t.Errorf("Expected default value for 'test.key1' to be 'value1', got '%v'", value1)
				}

				if !exists2 {
					t.Errorf("Expected default value for 'test.key2' to exist")
				} else if value2 != 42 {
					t.Errorf("Expected default value for 'test.key2' to be 42, got '%v'", value2)
				}
			},
		},
		{
			name:   "WithConfigValidator",
			option: WithConfigValidator(func(p Provider) error { return nil }),
			validate: func(t *testing.T, opts *Options) {
				if len(opts.ConfigValidators) == 0 {
					t.Errorf("Expected ConfigValidators to have at least one validator")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions()
			tt.option(&opts)
			tt.validate(t, &opts)
		})
	}
}

// 测试查找配置文件
func TestFindConfigFile(t *testing.T) {
	// 创建测试目录结构
	testDir, err := os.MkdirTemp("", "config-find-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建测试配置文件
	configPath := filepath.Join(testDir, "configs")
	err = os.MkdirAll(configPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create configs directory: %v", err)
	}

	// 标准配置文件
	standardConfigFile := filepath.Join(configPath, "app.yaml")
	err = os.WriteFile(standardConfigFile, []byte("test: value"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// 环境特定配置文件
	envConfigFile := filepath.Join(configPath, "app.production.yaml")
	err = os.WriteFile(envConfigFile, []byte("test: prod-value"), 0644)
	if err != nil {
		t.Fatalf("Failed to create env config file: %v", err)
	}

	// 测试不同场景
	tests := []struct {
		name           string
		opts           Options
		expectedResult string
		setupEnv       func()
	}{
		{
			name: "Standard config file",
			opts: Options{
				ConfigName:  "app",
				ConfigType:  "yaml",
				ConfigPaths: []string{configPath},
			},
			expectedResult: standardConfigFile,
			setupEnv: func() {
				// 重置环境为默认
				mockConfig := newMockProvider()
				mockConfig.Set("app.env", "development")
				Config = mockConfig
			},
		},
		{
			name: "Env-specific config file with production env",
			opts: Options{
				ConfigName:  "app",
				ConfigType:  "yaml",
				ConfigPaths: []string{configPath},
			},
			expectedResult: envConfigFile,
			setupEnv: func() {
				// 设置为生产环境
				mockConfig := newMockProvider()
				mockConfig.Set("app.env", "production")
				Config = mockConfig
			},
		},
		{
			name: "Non-existent config file",
			opts: Options{
				ConfigName:  "nonexistent",
				ConfigType:  "yaml",
				ConfigPaths: []string{configPath},
			},
			expectedResult: "",
			setupEnv: func() {
				// 重置环境为默认
				mockConfig := newMockProvider()
				mockConfig.Set("app.env", "development")
				Config = mockConfig
			},
		},
	}

	// 保存原始环境以便后续恢复
	originalConfig := Config
	defer func() { Config = originalConfig }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置环境
			if tt.setupEnv != nil {
				tt.setupEnv()
			}

			result := FindConfigFile(tt.opts)

			if result != tt.expectedResult {
				t.Errorf("Expected file path '%s', got '%s'", tt.expectedResult, result)
			}
		})
	}
}

// 测试文件存在函数
func TestFileExists(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "file-exists-test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	// 测试存在的文件
	if !fileExists(tmpFilePath) {
		t.Errorf("Expected fileExists to return true for existing file '%s'", tmpFilePath)
	}

	// 测试不存在的文件
	nonExistentFile := tmpFilePath + ".nonexistent"
	if fileExists(nonExistentFile) {
		t.Errorf("Expected fileExists to return false for non-existent file '%s'", nonExistentFile)
	}

	// 测试目录
	tmpDir, err := os.MkdirTemp("", "file-exists-dir-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if fileExists(tmpDir) {
		t.Errorf("Expected fileExists to return false for directory '%s'", tmpDir)
	}
}
