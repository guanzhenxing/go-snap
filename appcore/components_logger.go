package appcore

import (
	"context"
	"fmt"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
)

// LoggerComponent 日志组件适配器
type LoggerComponent struct {
	name    string
	logger  logger.Logger
	config  config.Provider
	closeFn func() error
}

// 确保LoggerComponent实现了LoggerProvider接口
var _ LoggerProvider = (*LoggerComponent)(nil)

// NewLoggerComponent 创建日志组件
func NewLoggerComponent() *LoggerComponent {
	return &LoggerComponent{
		name: "logger",
	}
}

// Initialize 初始化日志组件
func (c *LoggerComponent) Initialize(ctx context.Context) error {
	// 确保配置已设置
	if c.config == nil {
		return errors.New("日志组件需要配置")
	}

	// 准备日志选项
	var opts []logger.Option

	// 日志级别
	if c.config.IsSet("logger.level") {
		levelStr := c.config.GetString("logger.level")
		if level, err := logger.ParseLevel(levelStr); err == nil {
			opts = append(opts, logger.WithLevel(level))
		}
	}

	// 日志文件
	if c.config.IsSet("logger.file.path") {
		path := c.config.GetString("logger.file.path")
		if path != "" {
			opts = append(opts, logger.WithFilename(path))
		}
	}

	// JSON格式
	if c.config.IsSet("logger.json") && c.config.GetBool("logger.json") {
		opts = append(opts, logger.WithJSONConsole(true))
	}

	// 创建日志器
	c.logger = logger.New(opts...)

	// 添加服务信息
	fields := []logger.Field{}

	if c.config.IsSet("app.name") {
		fields = append(fields, logger.String("service", c.config.GetString("app.name")))
	}

	if c.config.IsSet("app.env") {
		fields = append(fields, logger.String("env", c.config.GetString("app.env")))
	}

	if c.config.IsSet("app.version") {
		fields = append(fields, logger.String("version", c.config.GetString("app.version")))
	}

	if len(fields) > 0 {
		c.logger = c.logger.With(fields...)
	}

	// 设置关闭函数
	c.closeFn = func() error {
		return c.logger.Sync()
	}

	// 输出日志器初始化信息
	c.logger.Debug("日志系统初始化成功")

	return nil
}

// Start 启动日志组件
func (c *LoggerComponent) Start(ctx context.Context) error {
	if c.logger == nil {
		return fmt.Errorf("logger component not initialized")
	}

	c.logger.Info("日志组件已启动")
	return nil
}

// Stop 停止日志组件
func (c *LoggerComponent) Stop(ctx context.Context) error {
	if c.logger != nil {
		c.logger.Info("日志组件正在停止")
	}

	if c.closeFn != nil {
		return c.closeFn()
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

// GetLogger 获取日志器
func (c *LoggerComponent) GetLogger() logger.Logger {
	return c.logger
}

// SetConfig 设置配置提供器
func (c *LoggerComponent) SetConfig(config config.Provider) {
	c.config = config
}

// NewLoggerProvider 提供日志组件
func NewLoggerProvider(config config.Provider) (LoggerProvider, logger.Logger, error) {
	component := NewLoggerComponent()
	component.SetConfig(config)
	if err := component.Initialize(context.Background()); err != nil {
		return nil, nil, err
	}
	return component, component.GetLogger(), nil
}
