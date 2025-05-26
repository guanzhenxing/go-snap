package snapcore

import (
	"context"
	"time"

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

// DbStoreComponent 数据库组件适配器
type DbStoreComponent struct {
	name                string
	store               *dbstore.Store
	repository          dbstore.Repository
	config              config.Provider
	logger              logger.Logger
	opts                []dbstore.Option
	healthCheckInterval time.Duration
	stopChan            chan struct{}
	modelRegistry       map[string]interface{} // 存储模型的映射表
}

// NewDbStoreComponent 创建数据库组件
func NewDbStoreComponent(opts ...dbstore.Option) *DbStoreComponent {
	return &DbStoreComponent{
		name:                "db",
		opts:                opts,
		healthCheckInterval: time.Minute * 5,
		stopChan:            make(chan struct{}),
		modelRegistry:       make(map[string]interface{}),
	}
}

// RegisterModel 注册数据模型，用于自动迁移
func (c *DbStoreComponent) RegisterModel(name string, model interface{}) *DbStoreComponent {
	c.modelRegistry[name] = model
	return c
}

// RegisterModels 批量注册数据模型
func (c *DbStoreComponent) RegisterModels(models map[string]interface{}) *DbStoreComponent {
	for name, model := range models {
		c.RegisterModel(name, model)
	}
	return c
}

// WithHealthCheckInterval 设置健康检查间隔
func (c *DbStoreComponent) WithHealthCheckInterval(interval time.Duration) *DbStoreComponent {
	c.healthCheckInterval = interval
	return c
}

// Initialize 初始化数据库组件
func (c *DbStoreComponent) Initialize(ctx context.Context) error {
	// 从配置获取数据库配置
	var dbCfg DbConfig
	if err := c.config.UnmarshalKey("database", &dbCfg); err != nil {
		return errors.Wrap(err, "解析数据库配置失败")
	}

	// 转换字符串时间配置为时间对象
	storeConfig := dbstore.DefaultConfig()
	storeConfig.Driver = dbCfg.Driver
	storeConfig.DSN = dbCfg.DSN
	storeConfig.MaxOpenConns = dbCfg.MaxOpenConns
	storeConfig.MaxIdleConns = dbCfg.MaxIdleConns
	storeConfig.TablePrefix = dbCfg.TablePrefix
	storeConfig.SingularTable = dbCfg.SingularTable
	storeConfig.Debug = dbCfg.Debug
	storeConfig.SkipDefaultTxn = dbCfg.SkipDefaultTxn
	storeConfig.PrepareStmt = dbCfg.PrepareStmt
	storeConfig.DisableNestedTxn = dbCfg.DisableNestedTxn

	// 解析时间相关配置
	if dbCfg.ConnMaxLifetime != "" {
		if duration, err := time.ParseDuration(dbCfg.ConnMaxLifetime); err == nil {
			storeConfig.ConnMaxLifetime = duration
		} else {
			c.logger.Warn("解析ConnMaxLifetime失败，使用默认值",
				logger.String("error", err.Error()),
				logger.String("value", dbCfg.ConnMaxLifetime),
			)
		}
	}

	if dbCfg.ConnMaxIdleTime != "" {
		if duration, err := time.ParseDuration(dbCfg.ConnMaxIdleTime); err == nil {
			storeConfig.ConnMaxIdleTime = duration
		} else {
			c.logger.Warn("解析ConnMaxIdleTime失败，使用默认值",
				logger.String("error", err.Error()),
				logger.String("value", dbCfg.ConnMaxIdleTime),
			)
		}
	}

	if dbCfg.SlowThreshold != "" {
		if duration, err := time.ParseDuration(dbCfg.SlowThreshold); err == nil {
			storeConfig.SlowThreshold = duration
		} else {
			c.logger.Warn("解析SlowThreshold失败，使用默认值",
				logger.String("error", err.Error()),
				logger.String("value", dbCfg.SlowThreshold),
			)
		}
	}

	// 合并选项
	var allOpts []dbstore.Option
	if c.logger != nil {
		allOpts = append(allOpts, dbstore.WithLogger(c.logger))
	}

	// 添加来自配置的选项
	if dbCfg.Debug {
		allOpts = append(allOpts, dbstore.WithDebug(true))
	}

	if dbCfg.TablePrefix != "" {
		allOpts = append(allOpts, dbstore.WithTablePrefix(dbCfg.TablePrefix))
	}

	if dbCfg.SingularTable {
		allOpts = append(allOpts, dbstore.WithSingularTable(true))
	}

	if dbCfg.PrepareStmt {
		allOpts = append(allOpts, dbstore.WithPreparedStatement(true))
	}

	if dbCfg.SkipDefaultTxn {
		allOpts = append(allOpts, dbstore.WithSkipDefaultTransaction(true))
	}

	if dbCfg.DisableNestedTxn {
		allOpts = append(allOpts, dbstore.WithDisableNestedTransaction(true))
	}

	// 添加用户提供的选项（它们将覆盖来自配置的选项）
	allOpts = append(allOpts, c.opts...)

	// 创建数据库连接
	c.logger.Info("正在初始化数据库连接",
		logger.String("driver", storeConfig.Driver),
		logger.Int("max_open_conns", storeConfig.MaxOpenConns),
		logger.Int("max_idle_conns", storeConfig.MaxIdleConns),
	)

	store, err := dbstore.New(storeConfig, allOpts...)
	if err != nil {
		return errors.Wrap(err, "创建数据库连接失败")
	}

	c.store = store
	c.repository = dbstore.NewRepository(store)

	// 处理自动迁移
	if dbCfg.AutoMigrate {
		if err := c.handleAutoMigration(ctx, dbCfg.Models); err != nil {
			return errors.Wrap(err, "数据库自动迁移失败")
		}
	}

	c.logger.Info("数据库组件初始化成功",
		logger.String("driver", storeConfig.Driver),
	)

	return nil
}

// handleAutoMigration 处理数据库自动迁移
func (c *DbStoreComponent) handleAutoMigration(ctx context.Context, modelNames []string) error {
	if len(modelNames) == 0 && len(c.modelRegistry) == 0 {
		c.logger.Warn("没有找到要迁移的模型")
		return nil
	}

	c.logger.Info("开始数据库自动迁移")

	// 如果指定了模型名称，只迁移这些模型
	if len(modelNames) > 0 {
		models := make([]interface{}, 0, len(modelNames))
		for _, name := range modelNames {
			if model, exists := c.modelRegistry[name]; exists {
				models = append(models, model)
				c.logger.Debug("添加模型到迁移列表", logger.String("model", name))
			} else {
				c.logger.Warn("找不到要迁移的模型", logger.String("model", name))
			}
		}

		if len(models) > 0 {
			if err := c.store.Migrate(models...); err != nil {
				return err
			}
		}
	} else {
		// 否则迁移所有注册的模型
		models := make([]interface{}, 0, len(c.modelRegistry))
		for name, model := range c.modelRegistry {
			models = append(models, model)
			c.logger.Debug("添加模型到迁移列表", logger.String("model", name))
		}

		if len(models) > 0 {
			if err := c.store.Migrate(models...); err != nil {
				return err
			}
		}
	}

	c.logger.Info("数据库自动迁移完成")
	return nil
}

// Start 启动数据库组件
func (c *DbStoreComponent) Start(ctx context.Context) error {
	c.logger.Info("启动数据库组件")

	// 确保连接正常
	if err := c.store.Ping(); err != nil {
		return errors.Wrap(err, "数据库连接测试失败")
	}

	// 启动健康检查
	if c.healthCheckInterval > 0 {
		go c.startHealthCheck()
	}

	return nil
}

// startHealthCheck 启动定期健康检查
func (c *DbStoreComponent) startHealthCheck() {
	ticker := time.NewTicker(c.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.store.Ping(); err != nil {
				c.logger.Error("数据库健康检查失败",
					logger.String("error", err.Error()),
				)
			} else {
				c.logger.Debug("数据库健康检查成功")
			}

			// 输出连接池状态
			stats := c.store.Stats()
			c.logger.Debug("数据库连接池状态",
				logger.Any("stats", stats),
			)

		case <-c.stopChan:
			return
		}
	}
}

// Stop 停止数据库组件
func (c *DbStoreComponent) Stop(ctx context.Context) error {
	c.logger.Info("停止数据库组件")

	// 停止健康检查
	close(c.stopChan)

	// 关闭数据库连接
	if c.store != nil {
		if err := c.store.Close(); err != nil {
			return errors.Wrap(err, "关闭数据库连接失败")
		}
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

// GetRepository 获取数据库仓储实例
func (c *DbStoreComponent) GetRepository() dbstore.Repository {
	return c.repository
}

// CreateRepository 创建新的仓储实例
func (c *DbStoreComponent) CreateRepository() dbstore.Repository {
	return dbstore.NewRepository(c.store)
}

// SetConfig 设置配置提供器
func (c *DbStoreComponent) SetConfig(config config.Provider) {
	c.config = config
}

// SetLogger 设置日志器
func (c *DbStoreComponent) SetLogger(logger logger.Logger) {
	c.logger = logger
}
