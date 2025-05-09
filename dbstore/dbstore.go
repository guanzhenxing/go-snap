// Package dbstore 提供了基于GORM的数据库操作封装，简化项目中的数据库操作
package dbstore

import (
	"context"
	"fmt"
	"time"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Config 数据库配置
type Config struct {
	Driver           string        `json:"driver" validate:"required,oneof=mysql postgres sqlite"`
	DSN              string        `json:"dsn" validate:"required"`
	MaxOpenConns     int           `json:"max_open_conns" validate:"min=1"`
	MaxIdleConns     int           `json:"max_idle_conns" validate:"min=1"`
	ConnMaxLifetime  time.Duration `json:"conn_max_lifetime" validate:"min=1"`
	ConnMaxIdleTime  time.Duration `json:"conn_max_idle_time" validate:"min=1"`
	TablePrefix      string        `json:"table_prefix"`
	SingularTable    bool          `json:"singular_table"`
	Debug            bool          `json:"debug"`
	SlowThreshold    time.Duration `json:"slow_threshold" validate:"min=1ms"`
	SkipDefaultTxn   bool          `json:"skip_default_txn"`
	PrepareStmt      bool          `json:"prepare_stmt"`
	DisableNestedTxn bool          `json:"disable_nested_txn"`
}

// DefaultConfig 返回默认数据库配置
func DefaultConfig() Config {
	return Config{
		Driver:           "mysql",
		MaxOpenConns:     100,
		MaxIdleConns:     10,
		ConnMaxLifetime:  time.Hour,
		ConnMaxIdleTime:  time.Minute * 10,
		TablePrefix:      "",
		SingularTable:    false,
		Debug:            false,
		SlowThreshold:    time.Millisecond * 200,
		SkipDefaultTxn:   false,
		PrepareStmt:      true,
		DisableNestedTxn: false,
	}
}

// LoadFromProvider 从配置提供器加载数据库配置
func (c *Config) LoadFromProvider(p config.Provider) error {
	if err := p.Unmarshal(c); err != nil {
		return errors.Wrap(err, "unmarshal database config failed")
	}
	return nil
}

// Store 数据库存储实例，封装了GORM的基本操作
type Store struct {
	db     *gorm.DB
	config Config
	log    logger.Logger
}

// New 创建数据库存储实例
func New(cfg Config, opts ...Option) (*Store, error) {
	s := &Store{
		config: cfg,
		log:    logger.New(),
	}

	// 应用选项
	for _, opt := range opts {
		opt(s)
	}

	// 初始化数据库连接
	if err := s.initialize(); err != nil {
		return nil, err
	}

	return s, nil
}

// NewFromProvider 从配置提供器创建数据库存储实例
func NewFromProvider(p config.Provider, configPath string, opts ...Option) (*Store, error) {
	cfg := DefaultConfig()

	// 获取数据库配置
	if configPath != "" {
		// 使用子配置路径
		if err := p.UnmarshalKey(configPath, &cfg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal database config")
		}
	} else {
		// 直接解析
		if err := p.Unmarshal(&cfg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal database config")
		}
	}

	return New(cfg, opts...)
}

// 初始化数据库连接
func (s *Store) initialize() error {
	var dialector gorm.Dialector

	// 根据驱动类型创建不同的方言
	switch s.config.Driver {
	case "mysql":
		dialector = mysql.Open(s.config.DSN)
	case "postgres":
		dialector = postgres.Open(s.config.DSN)
	case "sqlite":
		dialector = sqlite.Open(s.config.DSN)
	default:
		return errors.Errorf("unsupported database driver: %s", s.config.Driver)
	}

	// 创建GORM配置
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   s.config.TablePrefix,
			SingularTable: s.config.SingularTable,
		},
		Logger:                                   newLogger(s.log, s.config.SlowThreshold, s.config.Debug),
		SkipDefaultTransaction:                   s.config.SkipDefaultTxn,
		PrepareStmt:                              s.config.PrepareStmt,
		DisableNestedTransaction:                 s.config.DisableNestedTxn,
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 打开数据库连接
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get database connection")
	}

	sqlDB.SetMaxOpenConns(s.config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(s.config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(s.config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(s.config.ConnMaxIdleTime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return errors.Wrap(err, "failed to ping database")
	}

	s.db = db
	return nil
}

// DB 返回原始的GORM数据库实例
func (s *Store) DB() *gorm.DB {
	return s.db
}

// WithContext 返回带有上下文的数据库会话
func (s *Store) WithContext(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}

// Transaction 执行事务操作
func (s *Store) Transaction(fn func(tx *gorm.DB) error) error {
	return s.db.Transaction(fn)
}

// TransactionWithContext 执行带有上下文的事务操作
func (s *Store) TransactionWithContext(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return s.db.WithContext(ctx).Transaction(fn)
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get database connection")
	}
	return sqlDB.Close()
}

// Migrate 执行数据库迁移
func (s *Store) Migrate(models ...interface{}) error {
	return s.db.AutoMigrate(models...)
}

// Ping 检查数据库连接
func (s *Store) Ping() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get database connection")
	}
	return sqlDB.Ping()
}

// Stats 返回数据库连接池统计信息
func (s *Store) Stats() interface{} {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %v", err)
	}
	return sqlDB.Stats()
}
