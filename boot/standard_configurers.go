package boot

import (
	"context"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/logger"
)

// 基础组件配置器

// LoggerConfigurer 日志配置器
type LoggerConfigurer struct{}

// Configure 配置日志组件
func (c *LoggerConfigurer) Configure(registry *ComponentRegistry, props PropertySource) error {
	// 检查是否启用日志
	enabled := props.GetBool("logger.enabled", true)
	if !enabled {
		return nil
	}

	// 创建日志组件工厂
	registry.RegisterFactory("logger", &LoggerComponentFactory{})

	return nil
}

// Order 配置顺序
func (c *LoggerConfigurer) Order() int {
	return 100 // 日志应该最先配置
}

// ConfigConfigurer 配置配置器
type ConfigConfigurer struct{}

// Configure 配置配置组件
func (c *ConfigConfigurer) Configure(registry *ComponentRegistry, props PropertySource) error {
	// 配置组件总是启用
	registry.RegisterFactory("config", &ConfigComponentFactory{})

	return nil
}

// Order 配置顺序
func (c *ConfigConfigurer) Order() int {
	return 50 // 配置比日志更优先
}

// 数据源组件配置器

// DBStoreConfigurer 数据库配置器
type DBStoreConfigurer struct{}

// Configure 配置数据库组件
func (c *DBStoreConfigurer) Configure(registry *ComponentRegistry, props PropertySource) error {
	// 检查是否启用数据库
	enabled := props.GetBool("database.enabled", false)
	if !enabled {
		return nil
	}

	// 创建数据库组件工厂
	registry.RegisterFactory("dbstore", &DBStoreComponentFactory{})

	return nil
}

// Order 配置顺序
func (c *DBStoreConfigurer) Order() int {
	return 200
}

// CacheConfigurer 缓存配置器
type CacheConfigurer struct{}

// Configure 配置缓存组件
func (c *CacheConfigurer) Configure(registry *ComponentRegistry, props PropertySource) error {
	// 检查是否启用缓存
	enabled := props.GetBool("cache.enabled", true)
	if !enabled {
		return nil
	}

	// 创建缓存组件工厂
	registry.RegisterFactory("cache", &CacheComponentFactory{})

	return nil
}

// Order 配置顺序
func (c *CacheConfigurer) Order() int {
	return 300
}

// Web组件配置器

// WebConfigurer Web配置器
type WebConfigurer struct{}

// Configure 配置Web组件
func (c *WebConfigurer) Configure(registry *ComponentRegistry, props PropertySource) error {
	// 检查是否启用Web
	enabled := props.GetBool("web.enabled", false)
	if !enabled {
		return nil
	}

	// 创建Web组件工厂
	registry.RegisterFactory("web", &WebComponentFactory{})

	return nil
}

// Order 配置顺序
func (c *WebConfigurer) Order() int {
	return 400
}

// 组件工厂实现

// LoggerComponentFactory 日志组件工厂
type LoggerComponentFactory struct{}

// Create 创建日志组件
func (f *LoggerComponentFactory) Create(ctx context.Context, props PropertySource) (Component, error) {
	level := props.GetString("logger.level", "info")

	var opts []logger.Option

	// 日志级别
	if logLevel, err := logger.ParseLevel(level); err == nil {
		opts = append(opts, logger.WithLevel(logLevel))
	}

	// 日志文件
	if props.HasProperty("logger.file.path") {
		path := props.GetString("logger.file.path", "")
		if path != "" {
			opts = append(opts, logger.WithFilename(path))
		}
	}

	// JSON格式
	if props.HasProperty("logger.json") {
		json := props.GetBool("logger.json", false)
		if json {
			opts = append(opts, logger.WithJSONConsole(true))
		}
	}

	// 创建日志器
	loggerInstance := logger.New(opts...)

	// 构建组件
	component := &LoggerComponent{
		name:   "logger",
		logger: loggerInstance,
	}

	return component, nil
}

// Dependencies 依赖
func (f *LoggerComponentFactory) Dependencies() []string {
	return []string{}
}

// ConfigComponentFactory 配置组件工厂
type ConfigComponentFactory struct{}

// Create 创建配置组件
func (f *ConfigComponentFactory) Create(ctx context.Context, props PropertySource) (Component, error) {
	// 获取配置提供者
	configProvider, ok := props.(interface {
		PropertySource
		GetConfigProvider() config.Provider
	})

	if !ok {
		// 如果属性源不提供配置提供者，则使用全局配置
		configProvider := config.Config

		// 构建组件
		component := &ConfigComponent{
			name:   "config",
			config: configProvider,
		}

		return component, nil
	}

	// 构建组件
	component := &ConfigComponent{
		name:   "config",
		config: configProvider.GetConfigProvider(),
	}

	return component, nil
}

// Dependencies 依赖
func (f *ConfigComponentFactory) Dependencies() []string {
	return []string{}
}

// 组件实现

// LoggerComponent 日志组件
type LoggerComponent struct {
	name   string
	logger logger.Logger
}

// Name 组件名称
func (c *LoggerComponent) Name() string {
	return c.name
}

// Type 组件类型
func (c *LoggerComponent) Type() ComponentType {
	return ComponentTypeInfrastructure
}

// Initialize 初始化组件
func (c *LoggerComponent) Initialize(ctx context.Context) error {
	c.logger.Debug("Logger component initialized")
	return nil
}

// Start 启动组件
func (c *LoggerComponent) Start(ctx context.Context) error {
	c.logger.Info("Logger component started")
	return nil
}

// Stop 停止组件
func (c *LoggerComponent) Stop(ctx context.Context) error {
	c.logger.Info("Logger component stopping")
	return c.logger.Sync()
}

// GetLogger 获取日志器
func (c *LoggerComponent) GetLogger() logger.Logger {
	return c.logger
}

// ConfigComponent 配置组件
type ConfigComponent struct {
	name   string
	config config.Provider
}

// Name 组件名称
func (c *ConfigComponent) Name() string {
	return c.name
}

// Type 组件类型
func (c *ConfigComponent) Type() ComponentType {
	return ComponentTypeInfrastructure
}

// Initialize 初始化组件
func (c *ConfigComponent) Initialize(ctx context.Context) error {
	return nil
}

// Start 启动组件
func (c *ConfigComponent) Start(ctx context.Context) error {
	return nil
}

// Stop 停止组件
func (c *ConfigComponent) Stop(ctx context.Context) error {
	return nil
}

// GetConfig 获取配置提供者
func (c *ConfigComponent) GetConfig() config.Provider {
	return c.config
}
