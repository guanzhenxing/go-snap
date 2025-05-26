package snapcore

import (
	"context"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/dbstore"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
)

// DbStoreComponent 数据库组件适配器
type DbStoreComponent struct {
	name   string
	store  *dbstore.Store
	config config.Provider
	logger logger.Logger
	opts   []dbstore.Option
}

// NewDatabaseComponent 创建数据库组件
func NewDbStoreComponent(opts ...dbstore.Option) *DbStoreComponent {
	return &DbStoreComponent{
		name: "db",
		opts: opts,
	}
}

// Initialize 初始化数据库组件
func (c *DbStoreComponent) Initialize(ctx context.Context) error {
	// 从配置获取数据库配置
	var dbConfig dbstore.Config
	if err := c.config.UnmarshalKey("database", &dbConfig); err != nil {
		return errors.Wrap(err, "unmarshal database config failed")
	}

	// 合并选项
	var allOpts []dbstore.Option
	if c.logger != nil {
		allOpts = append(allOpts, dbstore.WithLogger(c.logger))
	}
	allOpts = append(allOpts, c.opts...)

	// 创建数据库连接
	store, err := dbstore.New(dbConfig, allOpts...)
	if err != nil {
		return errors.Wrap(err, "create database store failed")
	}

	c.store = store
	return nil
}

// Start 启动数据库组件
func (c *DbStoreComponent) Start(ctx context.Context) error {
	// 数据库连接已在初始化时建立，这里可以进行额外的启动逻辑
	return nil
}

// Stop 停止数据库组件
func (c *DbStoreComponent) Stop(ctx context.Context) error {
	if c.store != nil {
		return c.store.Close()
	}
	return nil
}

// Name 获取组件名称
func (c *DbStoreComponent) Name() string {
	return c.name
}

// Type 获取组件类型
func (c *DbStoreComponent) Type() ComponentType {
	return ComponentTypeDataSource
}

// Dependencies 获取组件依赖
func (c *DbStoreComponent) Dependencies() []string {
	return []string{"config", "logger"}
}

// GetStore 获取数据库存储实例
func (c *DbStoreComponent) GetStore() *dbstore.Store {
	return c.store
}

// SetConfig 设置配置提供器
func (c *DbStoreComponent) SetConfig(config config.Provider) {
	c.config = config
}

// SetLogger 设置日志器
func (c *DbStoreComponent) SetLogger(logger logger.Logger) {
	c.logger = logger
}
