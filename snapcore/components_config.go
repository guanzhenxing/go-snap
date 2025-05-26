package snapcore

import (
	"context"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/errors"
)

// ConfigComponent 配置组件适配器
type ConfigComponent struct {
	name       string
	config     config.Provider
	configPath string
	opts       []config.Option
}

// NewConfigComponent 创建配置组件
func NewConfigComponent(configPath string, opts ...config.Option) *ConfigComponent {
	return &ConfigComponent{
		name:       "config",
		configPath: configPath,
		opts:       opts,
	}
}

// Initialize 初始化配置组件
func (c *ConfigComponent) Initialize(ctx context.Context) error {

	err := config.InitConfig(c.opts...)
	if err != nil {
		return errors.Wrap(err, "init config failed")
	}
	c.config = config.Config
	return nil
}

// Start 启动配置组件
func (c *ConfigComponent) Start(ctx context.Context) error {
	// 启动配置热重载
	c.config.WatchConfig()
	return nil
}

// Stop 停止配置组件
func (c *ConfigComponent) Stop(ctx context.Context) error {
	return nil // 配置组件不需要特别的停止逻辑
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
	return []string{} // 配置组件没有依赖
}

// GetConfig 获取配置提供器
func (c *ConfigComponent) GetConfig() config.Provider {
	return c.config
}
