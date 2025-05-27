package appcore

import (
	"context"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
	"github.com/guanzhenxing/go-snap/web"
)

// WebComponent Web组件适配器
type WebComponent struct {
	name    string
	server  *web.Server
	config  config.Provider
	logger  logger.Logger
	options []web.Option
}

// 确保WebComponent实现了WebProvider接口
var _ WebProvider = (*WebComponent)(nil)

// NewWebComponent 创建Web组件
func NewWebComponent(options ...web.Option) *WebComponent {
	return &WebComponent{
		name:    "web",
		options: options,
	}
}

// Initialize 初始化Web组件
func (c *WebComponent) Initialize(ctx context.Context) error {
	// 从配置获取Web配置
	var webConfig web.Config
	if err := c.config.UnmarshalKey("web", &webConfig); err != nil {
		return errors.Wrap(err, "unmarshal web config failed")
	}

	// 合并选项
	var allOptions []web.Option
	if c.logger != nil {
		allOptions = append(allOptions, web.WithLogger(c.logger))
	}
	allOptions = append(allOptions, c.options...)

	// 创建Web服务器
	c.server = web.New(webConfig, allOptions...)
	return nil
}

// Start 启动Web组件
func (c *WebComponent) Start(ctx context.Context) error {
	if c.server != nil {
		return c.server.Start()
	}
	return nil
}

// Stop 停止Web组件
func (c *WebComponent) Stop(ctx context.Context) error {
	if c.server != nil {
		return c.server.Stop(ctx)
	}
	return nil
}

// Name 获取组件名称
func (c *WebComponent) Name() string {
	return c.name
}

// Type 获取组件类型
func (c *WebComponent) Type() ComponentType {
	return ComponentTypeWeb
}

// GetServer 获取Web服务器实例
func (c *WebComponent) GetServer() *web.Server {
	return c.server
}

// SetConfig 设置配置提供器
func (c *WebComponent) SetConfig(config config.Provider) {
	c.config = config
}

// SetLogger 设置日志器
func (c *WebComponent) SetLogger(logger logger.Logger) {
	c.logger = logger
}

// NewWebProvider 提供Web组件
func NewWebProvider(config config.Provider, logger logger.Logger) (WebProvider, error) {
	component := NewWebComponent()
	component.SetConfig(config)
	component.SetLogger(logger)
	if err := component.Initialize(context.Background()); err != nil {
		return nil, err
	}
	return component, nil
}
