package snapcore

import (
	"context"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/logger"
)

// CacheComponent 缓存组件适配器
type CacheComponent struct {
	name   string
	cache  interface{} // 使用interface代替具体实现，避免依赖实际的cache模块
	config config.Provider
	logger logger.Logger
}

// NewCacheComponent 创建缓存组件
func NewCacheComponent() *CacheComponent {
	return &CacheComponent{
		name: "cache",
	}
}

// Initialize 初始化缓存组件
func (c *CacheComponent) Initialize(ctx context.Context) error {
	// 这里使用interface{}代替具体实现，避免依赖错误
	return nil
}

// Start 启动缓存组件
func (c *CacheComponent) Start(ctx context.Context) error {
	return nil
}

// Stop 停止缓存组件
func (c *CacheComponent) Stop(ctx context.Context) error {
	return nil
}

// Name 获取组件名称
func (c *CacheComponent) Name() string {
	return c.name
}

// Type 获取组件类型
func (c *CacheComponent) Type() ComponentType {
	return ComponentTypeDataSource
}

// Dependencies 获取组件依赖
func (c *CacheComponent) Dependencies() []string {
	return []string{"config", "logger"}
}

// GetCache 获取缓存实例
func (c *CacheComponent) GetCache() interface{} {
	return c.cache
}

// SetConfig 设置配置提供器
func (c *CacheComponent) SetConfig(config config.Provider) {
	c.config = config
}

// SetLogger 设置日志器
func (c *CacheComponent) SetLogger(logger logger.Logger) {
	c.logger = logger
}
