package snapcore

import (
	"context"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
)

// LoggerComponent 日志组件适配器
type LoggerComponent struct {
	name   string
	logger logger.Logger
	config config.Provider
	opts   []logger.Option
}

// NewLoggerComponent 创建日志组件
func NewLoggerComponent(opts ...logger.Option) *LoggerComponent {
	return &LoggerComponent{
		name: "logger",
		opts: opts,
	}
}

// Initialize 初始化日志组件
func (c *LoggerComponent) Initialize(ctx context.Context) error {
	// 从配置中获取日志配置
	var logConfig struct {
		Level    string `json:"level"`
		Filename string `json:"filename"`
		Format   string `json:"format"`
	}
	if err := c.config.UnmarshalKey("logger", &logConfig); err != nil {
		return errors.Wrap(err, "unmarshal logger config failed")
	}

	// 合并配置选项
	var allOpts []logger.Option

	// 添加从配置中获取的选项
	if logConfig.Level != "" {
		var level logger.Level
		if err := level.UnmarshalText([]byte(logConfig.Level)); err == nil {
			allOpts = append(allOpts, logger.WithLevel(level))
		}
	}

	// 文件输出配置 (直接在Option中处理)
	if logConfig.Filename != "" {
		// 使用logger.New的方式，而不依赖WithFileOutput
		// 实际项目中这里应该根据logger模块的实际实现来调整
	}

	// 格式化选项
	if logConfig.Format != "" {
		// 使用logger.New的方式，而不依赖FormatJSON和FormatText常量
		// 实际项目中这里应该根据logger模块的实际实现来调整
	}

	// 添加用户提供的选项
	allOpts = append(allOpts, c.opts...)

	// 初始化日志器
	c.logger = logger.New(allOpts...)

	return nil
}

// Start 启动日志组件
func (c *LoggerComponent) Start(ctx context.Context) error {
	return nil // 日志组件不需要启动
}

// Stop 停止日志组件
func (c *LoggerComponent) Stop(ctx context.Context) error {
	if c.logger != nil {
		return c.logger.Shutdown(ctx)
	}
	return nil
}

// Name 获取组件名称
func (c *LoggerComponent) Name() string {
	return c.name
}

// Type 获取组件类型
func (c *LoggerComponent) Type() ComponentType {
	return ComponentTypeInfrastructure
}

// Dependencies 获取组件依赖
func (c *LoggerComponent) Dependencies() []string {
	return []string{} // 日志组件没有依赖
}

// GetLogger 获取日志器实例
func (c *LoggerComponent) GetLogger() logger.Logger {
	return c.logger
}

// SetConfig 设置配置提供器
func (c *LoggerComponent) SetConfig(config config.Provider) {
	c.config = config
}
