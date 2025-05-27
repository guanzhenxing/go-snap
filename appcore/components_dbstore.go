package appcore

import (
	"context"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/dbstore"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
)

// DbConfig 数据库配置结构，便于从配置文件映射
type DbConfig struct {
	Driver           string                 `json:"driver" validate:"required,oneof=mysql postgres sqlite"`
	DSN              string                 `json:"dsn" validate:"required"`
	MaxOpenConns     int                    `json:"max_open_conns" validate:"min=1"`
	MaxIdleConns     int                    `json:"max_idle_conns" validate:"min=1"`
	ConnMaxLifetime  string                 `json:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime  string                 `json:"conn_max_idle_time" validate:"required"`
	TablePrefix      string                 `json:"table_prefix"`
	SingularTable    bool                   `json:"singular_table"`
	Debug            bool                   `json:"debug"`
	SlowThreshold    string                 `json:"slow_threshold" validate:"required"`
	SkipDefaultTxn   bool                   `json:"skip_default_txn"`
	PrepareStmt      bool                   `json:"prepare_stmt"`
	DisableNestedTxn bool                   `json:"disable_nested_txn"`
	AutoMigrate      bool                   `json:"auto_migrate"`
	Models           []string               `json:"models"`
	ModelRegistry    map[string]interface{} `json:"-"` // 模型注册表，不从配置读取
}

// DBStoreComponent 数据库组件适配器
type DBStoreComponent struct {
	name   string
	db     *dbstore.Store
	config config.Provider
	logger logger.Logger
}

// 确保DBStoreComponent实现了DBStoreProvider接口
var _ DBStoreProvider = (*DBStoreComponent)(nil)

// NewDBStoreComponent 创建数据库组件
func NewDBStoreComponent() *DBStoreComponent {
	return &DBStoreComponent{
		name: "dbstore",
	}
}

// Initialize 初始化数据库组件
func (c *DBStoreComponent) Initialize(ctx context.Context) error {
	// 确保配置已设置
	if c.config == nil {
		return errors.New("数据库组件需要配置")
	}

	// 从配置中获取数据库配置
	var dbConfig dbstore.Config
	if err := c.config.UnmarshalKey("database", &dbConfig); err != nil {
		return errors.Wrap(err, "解析数据库配置失败")
	}

	// 创建数据库实例
	var options []dbstore.Option
	if c.logger != nil {
		options = append(options, dbstore.WithLogger(c.logger))
	}

	db, err := dbstore.New(dbConfig, options...)
	if err != nil {
		return errors.Wrap(err, "创建数据库实例失败")
	}

	c.db = db

	if c.logger != nil {
		c.logger.Info("数据库组件初始化成功",
			logger.String("driver", dbConfig.Driver),
		)
	}

	return nil
}

// Start 启动数据库组件
func (c *DBStoreComponent) Start(ctx context.Context) error {
	if c.logger != nil {
		c.logger.Info("数据库组件已启动")
	}
	return nil
}

// Stop 停止数据库组件
func (c *DBStoreComponent) Stop(ctx context.Context) error {
	if c.logger != nil {
		c.logger.Info("数据库组件正在停止")
	}

	if c.db != nil {
		if err := c.db.Close(); err != nil {
			return errors.Wrap(err, "关闭数据库连接失败")
		}
	}
	return nil
}

// Name 获取组件名称
func (c *DBStoreComponent) Name() string {
	return c.name
}

// Type 获取组件类型
func (c *DBStoreComponent) Type() ComponentType {
	return ComponentTypeDataSource
}

// GetDBStore 获取数据库实例
func (c *DBStoreComponent) GetDBStore() *dbstore.Store {
	return c.db
}

// SetConfig 设置配置提供器
func (c *DBStoreComponent) SetConfig(config config.Provider) {
	c.config = config
}

// SetLogger 设置日志器
func (c *DBStoreComponent) SetLogger(logger logger.Logger) {
	c.logger = logger
}

// NewDBStoreProvider 提供数据库组件
func NewDBStoreProvider(config config.Provider, logger logger.Logger) (DBStoreProvider, error) {
	component := NewDBStoreComponent()
	component.SetConfig(config)
	component.SetLogger(logger)
	if err := component.Initialize(context.Background()); err != nil {
		return nil, err
	}
	return component, nil
}
