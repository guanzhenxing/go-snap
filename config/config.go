// Package config 提供应用配置管理功能
// 基于Viper实现，支持多种配置源、配置热重载和配置验证
//
// # 配置系统架构
//
// 本包设计了一个灵活的配置抽象层，允许从多种来源加载和管理配置。主要组件：
//
// 1. Provider接口：定义配置提供者的抽象接口，允许不同的底层实现
// 2. ViperProvider：基于Viper库的默认实现，提供强大的配置功能
// 3. 环境管理：支持不同运行环境（开发、测试、预生产、生产）的配置隔离
// 4. 配置验证：支持对配置进行结构化验证，确保配置的正确性
// 5. 配置监听：支持配置热重载，动态响应配置变更
//
// # 支持的配置格式
//
// 配置系统支持多种常见的配置格式：
//
// - YAML (.yaml, .yml)
// - JSON (.json)
// - TOML (.toml)
// - HCL (.hcl)
// - INI (.ini)
// - 环境变量
// - 命令行参数
//
// # 配置加载顺序
//
// 配置系统按以下顺序加载配置，后加载的会覆盖先加载的：
//
// 1. 默认值
// 2. 配置文件
// 3. 环境变量 (可选，格式为APP_SECTION_KEY)
// 4. 命令行参数 (可选)
//
// # 配置热重载
//
// 配置系统支持监听配置文件变更并自动重新加载：
//
//	config.InitConfig(config.WithWatchConfig(true))
//	config.Config.OnConfigChange(func() {
//	    log.Println("配置已更新")
//	})
//
// # 配置验证
//
// 支持使用结构体标签验证配置的正确性：
//
//	type AppConfig struct {
//	    Port    int    `validate:"required,min=1024,max=65535"`
//	    LogPath string `validate:"required"`
//	}
//
//	var appConfig AppConfig
//	if err := config.Config.UnmarshalKey("app", &appConfig); err != nil {
//	    log.Fatalf("配置验证失败: %v", err)
//	}
//
// # 使用示例
//
// 1. 基本使用：
//
//	// 初始化配置
//	err := config.InitConfig(
//	    config.WithConfigPath("./configs"),
//	    config.WithConfigName("app"),
//	    config.WithConfigType("yaml"),
//	)
//	if err != nil {
//	    log.Fatalf("初始化配置失败: %v", err)
//	}
//
//	// 获取配置值
//	serverPort := config.Config.GetInt("server.port")
//	dbURL := config.Config.GetString("database.url")
//
// 2. 结构化配置：
//
//	type DatabaseConfig struct {
//	    Driver string `json:"driver" validate:"required,oneof=mysql postgres sqlite"`
//	    URL    string `json:"url" validate:"required"`
//	}
//
//	var dbConfig DatabaseConfig
//	if err := config.Config.UnmarshalKey("database", &dbConfig); err != nil {
//	    log.Fatalf("解析数据库配置失败: %v", err)
//	}
//
// 3. 环境特定配置：
//
//	if config.IsProduction() {
//	    // 生产环境特定逻辑
//	} else if config.IsDevelopment() {
//	    // 开发环境特定逻辑
//	}
//
// # 最佳实践
//
// 1. 配置分层：将配置按功能领域分组（app, server, database, cache等）
// 2. 敏感信息：敏感信息（如密码、密钥）使用环境变量或专用的密钥管理系统
// 3. 默认值：为所有配置项提供合理的默认值
// 4. 验证：对关键配置项进行验证
// 5. 文档化：为配置项添加详细注释说明用途和取值范围
package config

import (
	"time"
)

// Provider 定义配置提供者接口
// 抽象配置操作，允许不同的底层实现（如Viper、环境变量等）
type Provider interface {
	// Get 获取指定键的配置值，如果不存在返回nil
	// 参数：
	//   key: 配置键名，通常使用点号分隔的路径格式（如app.name）
	// 返回：
	//   配置值，类型为interface{}
	Get(key string) interface{}

	// GetString 获取字符串类型的配置
	// 参数：
	//   key: 配置键名
	// 返回：
	//   字符串类型的配置值，如果键不存在则返回空字符串
	GetString(key string) string

	// GetBool 获取布尔类型的配置
	// 参数：
	//   key: 配置键名
	// 返回：
	//   布尔类型的配置值，如果键不存在则返回false
	GetBool(key string) bool

	// GetInt 获取整型的配置
	// 参数：
	//   key: 配置键名
	// 返回：
	//   整型的配置值，如果键不存在则返回0
	GetInt(key string) int

	// GetInt64 获取int64类型的配置
	// 参数：
	//   key: 配置键名
	// 返回：
	//   int64类型的配置值，如果键不存在则返回0
	GetInt64(key string) int64

	// GetFloat64 获取浮点类型的配置
	// 参数：
	//   key: 配置键名
	// 返回：
	//   float64类型的配置值，如果键不存在则返回0.0
	GetFloat64(key string) float64

	// GetTime 获取时间类型的配置
	// 参数：
	//   key: 配置键名
	// 返回：
	//   time.Time类型的配置值，如果键不存在则返回零值时间
	GetTime(key string) time.Time

	// GetDuration 获取时间段类型的配置
	// 参数：
	//   key: 配置键名
	// 返回：
	//   time.Duration类型的配置值，如果键不存在则返回0
	GetDuration(key string) time.Duration

	// GetStringSlice 获取字符串切片类型的配置
	// 参数：
	//   key: 配置键名
	// 返回：
	//   字符串切片，如果键不存在则返回空切片
	GetStringSlice(key string) []string

	// GetStringMap 获取字符串映射类型的配置
	// 参数：
	//   key: 配置键名
	// 返回：
	//   字符串到任意值的映射，如果键不存在则返回空映射
	GetStringMap(key string) map[string]interface{}

	// GetStringMapString 获取字符串-字符串映射类型的配置
	// 参数：
	//   key: 配置键名
	// 返回：
	//   字符串到字符串的映射，如果键不存在则返回空映射
	GetStringMapString(key string) map[string]string

	// IsSet 判断配置项是否存在
	// 参数：
	//   key: 配置键名
	// 返回：
	//   如果配置项存在则为true，否则为false
	IsSet(key string) bool

	// Set 设置配置项
	// 参数：
	//   key: 配置键名
	//   value: 要设置的配置值
	Set(key string, value interface{})

	// UnmarshalKey 将指定键下的配置值解析到结构体中
	// 参数：
	//   key: 配置键名
	//   rawVal: 目标结构体指针
	// 返回：
	//   解析过程中遇到的错误，如果解析成功则返回nil
	UnmarshalKey(key string, rawVal interface{}) error

	// Unmarshal 将所有配置解析到结构体中
	// 参数：
	//   rawVal: 目标结构体指针
	// 返回：
	//   解析过程中遇到的错误，如果解析成功则返回nil
	Unmarshal(rawVal interface{}) error

	// LoadConfig 加载配置文件
	// 返回：
	//   加载过程中遇到的错误，如果加载成功则返回nil
	LoadConfig() error

	// WatchConfig 启用配置热重载
	// 当配置文件变更时自动重新加载
	WatchConfig()

	// OnConfigChange 注册配置变更回调函数
	// 参数：
	//   run: 配置变更时要执行的回调函数
	OnConfigChange(run func())

	// ValidateConfig 验证配置是否合法
	// 返回：
	//   验证过程中遇到的错误，如果验证通过则返回nil
	ValidateConfig() error
}

// Environment 环境类型，表示应用运行的不同环境
type Environment string

const (
	// EnvDevelopment 开发环境，用于本地开发和调试
	EnvDevelopment Environment = "development"
	// EnvTesting 测试环境，用于自动化测试和QA测试
	EnvTesting Environment = "testing"
	// EnvStaging 预生产环境，用于模拟生产环境的最终测试
	EnvStaging Environment = "staging"
	// EnvProduction 生产环境，用于正式部署给最终用户
	EnvProduction Environment = "production"
)

// Config 全局配置管理器实例
// 应用中可以直接使用此变量访问配置
var Config Provider

// 初始化默认配置提供者
func init() {
	// 注意：在init中不应该加载配置，由应用决定何时初始化配置
	// Viper的初始化将在应用调用InitConfig时进行
}

// InitConfig 初始化配置管理器
// 参数：
//
//	opts: 配置选项，用于自定义配置行为
//
// 返回：
//
//	初始化过程中遇到的错误，如果初始化成功则返回nil
//
// 示例：
//
//	err := config.InitConfig(
//	  config.WithConfigPath("./configs"),
//	  config.WithConfigName("app"),
//	  config.WithConfigType("yaml"),
//	)
func InitConfig(opts ...Option) error {
	Config = NewViperProvider(opts...)
	return Config.LoadConfig()
}

// GetCurrentEnvironment 获取当前应用环境
// 返回：
//
//	当前环境类型，基于app.env配置项决定
//
// 如果未设置app.env或Config未初始化，则默认返回开发环境
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
// 返回：
//
//	如果当前是开发环境则为true，否则为false
func IsDevelopment() bool {
	return GetCurrentEnvironment() == EnvDevelopment
}

// IsTesting 检查是否为测试环境
// 返回：
//
//	如果当前是测试环境则为true，否则为false
func IsTesting() bool {
	return GetCurrentEnvironment() == EnvTesting
}

// IsStaging 检查是否为预生产环境
// 返回：
//
//	如果当前是预生产环境则为true，否则为false
func IsStaging() bool {
	return GetCurrentEnvironment() == EnvStaging
}

// IsProduction 检查是否为生产环境
// 返回：
//
//	如果当前是生产环境则为true，否则为false
func IsProduction() bool {
	return GetCurrentEnvironment() == EnvProduction
}
