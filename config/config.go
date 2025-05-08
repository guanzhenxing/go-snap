package config

import (
	"time"
)

// Provider 定义配置提供者接口
type Provider interface {
	// Get 获取指定键的配置值，如果不存在返回nil
	Get(key string) interface{}
	// GetString 获取字符串类型的配置
	GetString(key string) string
	// GetBool 获取布尔类型的配置
	GetBool(key string) bool
	// GetInt 获取整型的配置
	GetInt(key string) int
	// GetInt64 获取int64类型的配置
	GetInt64(key string) int64
	// GetFloat64 获取浮点类型的配置
	GetFloat64(key string) float64
	// GetTime 获取时间类型的配置
	GetTime(key string) time.Time
	// GetDuration 获取时间段类型的配置
	GetDuration(key string) time.Duration
	// GetStringSlice 获取字符串切片类型的配置
	GetStringSlice(key string) []string
	// GetStringMap 获取字符串映射类型的配置
	GetStringMap(key string) map[string]interface{}
	// GetStringMapString 获取字符串-字符串映射类型的配置
	GetStringMapString(key string) map[string]string
	// IsSet 判断配置项是否存在
	IsSet(key string) bool
	// Set 设置配置项
	Set(key string, value interface{})
	// UnmarshalKey 将指定键下的配置值解析到结构体中
	UnmarshalKey(key string, rawVal interface{}) error
	// Unmarshal 将所有配置解析到结构体中
	Unmarshal(rawVal interface{}) error
	// LoadConfig 加载配置文件
	LoadConfig() error
	// WatchConfig 启用配置热重载
	WatchConfig()
	// OnConfigChange 注册配置变更回调函数
	OnConfigChange(run func())
	// ValidateConfig 验证配置是否合法
	ValidateConfig() error
}

// Environment 环境类型
type Environment string

const (
	// EnvDevelopment 开发环境
	EnvDevelopment Environment = "development"
	// EnvTesting 测试环境
	EnvTesting Environment = "testing"
	// EnvStaging 预生产环境
	EnvStaging Environment = "staging"
	// EnvProduction 生产环境
	EnvProduction Environment = "production"
)

// Config 全局配置管理器实例
var Config Provider

// 初始化默认配置提供者
func init() {
	// 注意：在init中不应该加载配置，由应用决定何时初始化配置
	// Viper的初始化将在应用调用InitConfig时进行
}

// InitConfig 初始化配置管理器
func InitConfig(opts ...Option) error {
	Config = NewViperProvider(opts...)
	return Config.LoadConfig()
}

// GetCurrentEnvironment 获取当前环境
func GetCurrentEnvironment() Environment {
	if Config == nil {
		return EnvDevelopment // 默认为开发环境
	}

	env := Config.GetString("app.env")
	switch env {
	case string(EnvTesting):
		return EnvTesting
	case string(EnvStaging):
		return EnvStaging
	case string(EnvProduction):
		return EnvProduction
	default:
		return EnvDevelopment
	}
}

// IsDevelopment 检查是否为开发环境
func IsDevelopment() bool {
	return GetCurrentEnvironment() == EnvDevelopment
}

// IsTesting 检查是否为测试环境
func IsTesting() bool {
	return GetCurrentEnvironment() == EnvTesting
}

// IsStaging 检查是否为预生产环境
func IsStaging() bool {
	return GetCurrentEnvironment() == EnvStaging
}

// IsProduction 检查是否为生产环境
func IsProduction() bool {
	return GetCurrentEnvironment() == EnvProduction
}
