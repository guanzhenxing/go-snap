package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ConfigType 配置文件类型
type ConfigType string

const (
	// ConfigTypeYAML YAML格式配置
	ConfigTypeYAML ConfigType = "yaml"
	// ConfigTypeJSON JSON格式配置
	ConfigTypeJSON ConfigType = "json"
	// ConfigTypeTOML TOML格式配置
	ConfigTypeTOML ConfigType = "toml"
)

// 默认配置值
const (
	defaultConfigName = "config"
	defaultConfigType = "yaml"
	defaultConfigPath = "./configs"
	defaultEnvPrefix  = "APP"
)

// Options 配置选项
type Options struct {
	// ConfigName 配置文件名（不含扩展名）
	ConfigName string
	// ConfigType 配置文件类型（yaml, json, toml）
	ConfigType string
	// ConfigPaths 配置文件搜索路径
	ConfigPaths []string
	// EnvPrefix 环境变量前缀
	EnvPrefix string
	// AutomaticEnv 是否自动绑定环境变量
	AutomaticEnv bool
	// EnvKeyReplacer 环境变量键替换器
	EnvKeyReplacer *strings.Replacer
	// WatchConfigFile 是否监听配置文件变更
	WatchConfigFile bool
	// DefaultValues 默认配置值
	DefaultValues map[string]interface{}
	// ConfigValidators 配置验证器
	ConfigValidators []ConfigValidator
}

// ConfigValidator 配置验证器函数类型
type ConfigValidator func(p Provider) error

// DefaultOptions 返回默认配置选项
func DefaultOptions() Options {
	return Options{
		ConfigName:      defaultConfigName,
		ConfigType:      defaultConfigType,
		ConfigPaths:     []string{defaultConfigPath},
		EnvPrefix:       defaultEnvPrefix,
		AutomaticEnv:    true,
		EnvKeyReplacer:  strings.NewReplacer(".", "_"),
		WatchConfigFile: false,
		DefaultValues: map[string]interface{}{
			"app.env":   "development",
			"app.debug": true,
			"app.name":  "go-snap-app",
		},
		ConfigValidators: []ConfigValidator{},
	}
}

// Option 配置选项设置函数
type Option func(*Options)

// WithConfigName 设置配置文件名
func WithConfigName(name string) Option {
	return func(o *Options) {
		o.ConfigName = name
	}
}

// WithConfigType 设置配置文件类型
func WithConfigType(configType string) Option {
	return func(o *Options) {
		o.ConfigType = configType
	}
}

// WithConfigPath 设置单个配置路径
func WithConfigPath(path string) Option {
	return func(o *Options) {
		// 确保路径存在
		if _, err := os.Stat(path); os.IsNotExist(err) {
			_ = os.MkdirAll(path, 0755)
		}
		o.ConfigPaths = []string{path}
	}
}

// WithConfigPaths 设置多个配置路径
func WithConfigPaths(paths []string) Option {
	return func(o *Options) {
		o.ConfigPaths = paths
	}
}

// AddConfigPath 添加配置路径
func AddConfigPath(path string) Option {
	return func(o *Options) {
		o.ConfigPaths = append(o.ConfigPaths, path)
	}
}

// WithEnvPrefix 设置环境变量前缀
func WithEnvPrefix(prefix string) Option {
	return func(o *Options) {
		o.EnvPrefix = prefix
	}
}

// WithAutomaticEnv 设置是否自动绑定环境变量
func WithAutomaticEnv(automatic bool) Option {
	return func(o *Options) {
		o.AutomaticEnv = automatic
	}
}

// WithEnvKeyReplacer 设置环境变量键替换器
func WithEnvKeyReplacer(replacer *strings.Replacer) Option {
	return func(o *Options) {
		o.EnvKeyReplacer = replacer
	}
}

// WithWatchConfigFile 设置是否监听配置文件变更
func WithWatchConfigFile(watch bool) Option {
	return func(o *Options) {
		o.WatchConfigFile = watch
	}
}

// WithDefaultValue 设置单个默认值
func WithDefaultValue(key string, value interface{}) Option {
	return func(o *Options) {
		if o.DefaultValues == nil {
			o.DefaultValues = make(map[string]interface{})
		}
		o.DefaultValues[key] = value
	}
}

// WithDefaultValues 设置多个默认值
func WithDefaultValues(values map[string]interface{}) Option {
	return func(o *Options) {
		for k, v := range values {
			if o.DefaultValues == nil {
				o.DefaultValues = make(map[string]interface{})
			}
			o.DefaultValues[k] = v
		}
	}
}

// WithConfigValidator 添加配置验证器
func WithConfigValidator(validator ConfigValidator) Option {
	return func(o *Options) {
		o.ConfigValidators = append(o.ConfigValidators, validator)
	}
}

// FindConfigFile 根据配置选项查找配置文件
func FindConfigFile(opts Options) string {
	// 使用当前工作目录作为基础路径
	cwd, _ := os.Getwd()

	// 构建完整的文件名（包含扩展名）
	configFile := opts.ConfigName + "." + opts.ConfigType

	// 获取当前环境
	currentEnv := string(GetCurrentEnvironment())

	// 环境特定配置文件名
	envConfigFile := ""
	if currentEnv != "" {
		envConfigFile = opts.ConfigName + "." + currentEnv + "." + opts.ConfigType
	}

	// 首先检查指定的路径
	for _, path := range opts.ConfigPaths {
		// 先检查环境特定配置（如果有）
		if envConfigFile != "" {
			envFullPath := filepath.Join(path, envConfigFile)
			if fileExists(envFullPath) {
				return envFullPath
			}
		}

		// 然后检查标准配置
		fullPath := filepath.Join(path, configFile)
		if fileExists(fullPath) {
			return fullPath
		}
	}

	// 然后检查当前目录
	if envConfigFile != "" && fileExists(filepath.Join(cwd, envConfigFile)) {
		return filepath.Join(cwd, envConfigFile)
	}

	if fileExists(filepath.Join(cwd, configFile)) {
		return filepath.Join(cwd, configFile)
	}

	// 最后检查./configs目录
	if envConfigFile != "" && fileExists(filepath.Join(cwd, "configs", envConfigFile)) {
		return filepath.Join(cwd, "configs", envConfigFile)
	}

	defaultPath := filepath.Join(cwd, "configs", configFile)
	if fileExists(defaultPath) {
		return defaultPath
	}

	return ""
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
