package snapcore

import (
	"context"
	"os"
	"path/filepath"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
)

// ConfigComponent 配置组件适配器
type ConfigComponent struct {
	name       string
	config     config.Provider
	configPath string
	opts       []config.Option
	watcher    *config.Watcher
	logger     logger.Logger
}

// NewConfigComponent 创建配置组件
func NewConfigComponent(configPath string, opts ...config.Option) *ConfigComponent {
	// 处理配置路径
	finalPath := configPath
	if configPath == "" {
		// 如果未指定配置路径，使用默认路径
		cwd, _ := os.Getwd()
		finalPath = filepath.Join(cwd, "configs")
	}

	// 创建配置组件
	return &ConfigComponent{
		name:       "config",
		configPath: finalPath,
		opts:       opts,
	}
}

// Initialize 初始化配置组件
func (c *ConfigComponent) Initialize(ctx context.Context) error {
	// 合并所有配置选项
	var allOpts []config.Option

	// 首先添加配置路径
	allOpts = append(allOpts, config.WithConfigPath(c.configPath))

	// 添加用户提供的选项
	allOpts = append(allOpts, c.opts...)

	// 添加配置文件监听选项
	allOpts = append(allOpts, config.WithWatchConfigFile(true))

	// 初始化配置系统
	err := config.InitConfig(allOpts...)
	if err != nil {
		return errors.Wrap(err, "初始化配置失败")
	}

	// 获取全局配置实例
	c.config = config.Config

	// 验证配置
	if err := c.config.ValidateConfig(); err != nil {
		return errors.Wrap(err, "配置验证失败")
	}

	// 如果有日志记录器，输出配置信息
	if c.logger != nil {
		c.logger.Info("配置系统初始化成功",
			logger.String("config_path", c.configPath),
			logger.String("environment", string(config.GetCurrentEnvironment())),
		)
	}

	return nil
}

// Start 启动配置组件
func (c *ConfigComponent) Start(ctx context.Context) error {
	// 确保配置已初始化
	if c.config == nil {
		return errors.New("配置未初始化")
	}

	// 创建配置监听器
	c.watcher = config.NewWatcher(c.config)

	// 监听配置变更
	c.config.WatchConfig()

	// 注册一些关键配置的变更监听
	c.watcher.WatchFunc("app.debug", func(event config.ConfigChangeEvent) {
		if c.logger != nil {
			c.logger.Info("调试模式已更改",
				logger.Bool("new_value", event.NewValue.(bool)),
			)
		}
	})

	c.watcher.WatchFunc("app.env", func(event config.ConfigChangeEvent) {
		if c.logger != nil {
			c.logger.Info("应用环境已更改",
				logger.String("new_value", event.NewValue.(string)),
			)
		}
	})

	if c.logger != nil {
		c.logger.Info("配置热重载已启用")
	}

	return nil
}

// Stop 停止配置组件
func (c *ConfigComponent) Stop(ctx context.Context) error {
	// 清理配置监听器
	if c.watcher != nil {
		c.watcher.UnwatchAll()
	}

	if c.logger != nil {
		c.logger.Info("配置组件已停止")
	}

	return nil
}

// Name 获取组件名称
func (c *ConfigComponent) Name() string {
	return c.name
}

// Type 获取组件类型
func (c *ConfigComponent) Type() ComponentType {
	return ComponentTypeInfrastructure
}

// Dependencies 获取组件依赖
func (c *ConfigComponent) Dependencies() []string {
	return []string{} // 配置组件通常是基础组件，没有依赖
}

// GetConfig 获取配置提供器
func (c *ConfigComponent) GetConfig() config.Provider {
	return c.config
}

// SetLogger 设置日志器
func (c *ConfigComponent) SetLogger(logger logger.Logger) {
	c.logger = logger
}
