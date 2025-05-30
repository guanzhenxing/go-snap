// Package dbstore 提供了基于GORM的数据库操作封装，简化项目中的数据库操作
//
// # 数据库抽象层设计
//
// 本包设计了一个基于GORM的数据库抽象层，提供以下核心功能：
//
// 1. 连接管理：处理数据库连接的创建、配置和生命周期管理
// 2. 事务处理：简化事务操作，支持嵌套事务和上下文传递
// 3. 查询构建：提供流畅的API进行数据库查询和操作
// 4. 结构映射：在Go结构体和数据库表之间建立映射关系
// 5. 迁移工具：支持数据库结构迁移和版本管理
// 6. 性能监控：记录慢查询和连接池状态
//
// # 支持的数据库
//
// 目前支持以下数据库系统：
//
// - MySQL/MariaDB
// - PostgreSQL
// - SQLite
//
// # 连接池管理
//
// 本包提供了连接池的完整配置和管理：
//
// - MaxOpenConns：最大打开连接数，控制并发连接上限
// - MaxIdleConns：最大空闲连接数，优化资源使用
// - ConnMaxLifetime：连接最大生存时间，防止连接资源泄露
// - ConnMaxIdleTime：空闲连接最大存活时间，优化资源回收
//
// 连接池设计目标是在性能和资源使用之间取得平衡，避免频繁创建连接的开销，
// 同时防止过多连接占用数据库资源。
//
// # 事务处理
//
// 事务处理是数据库操作中的关键环节，本包提供了便捷的事务API：
//
//	err := store.Transaction(func(tx *gorm.DB) error {
//	    // 事务内部操作
//	    if err := tx.Create(&user).Error; err != nil {
//	        // 返回错误会自动回滚事务
//	        return err
//	    }
//
//	    // 更多数据库操作...
//
//	    // 事务成功完成，自动提交
//	    return nil
//	})
//
// 支持带上下文的事务处理，便于传递请求级信息和控制超时：
//
//	err := store.TransactionWithContext(ctx, func(tx *gorm.DB) error {
//	    // 事务内部操作
//	    return nil
//	})
//
// # 性能优化
//
// 1. 预处理语句：通过PrepareStmt选项启用语句预处理，减少解析开销
// 2. 连接池配置：根据应用负载特性优化连接池参数
// 3. 跳过默认事务：对于只读操作，可通过SkipDefaultTxn提高性能
// 4. 慢查询日志：自动记录超过阈值的慢查询，便于优化
// 5. 查询缓存：可与缓存系统集成，减少数据库负载
//
// # 使用示例
//
// 1. 创建数据库连接：
//
//	// 使用默认配置创建
//	cfg := dbstore.DefaultConfig()
//	cfg.Driver = "mysql"
//	cfg.DSN = "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
//
//	store, err := dbstore.New(cfg)
//	if err != nil {
//	    log.Fatalf("连接数据库失败: %v", err)
//	}
//	defer store.Close()
//
// 2. 基本CRUD操作：
//
//	// 创建记录
//	user := User{Name: "张三", Age: 25}
//	if err := store.DB().Create(&user).Error; err != nil {
//	    log.Printf("创建用户失败: %v", err)
//	}
//
//	// 查询记录
//	var result User
//	if err := store.DB().First(&result, "name = ?", "张三").Error; err != nil {
//	    log.Printf("查询用户失败: %v", err)
//	}
//
//	// 更新记录
//	if err := store.DB().Model(&result).Update("age", 26).Error; err != nil {
//	    log.Printf("更新用户失败: %v", err)
//	}
//
//	// 删除记录
//	if err := store.DB().Delete(&result).Error; err != nil {
//	    log.Printf("删除用户失败: %v", err)
//	}
//
// 3. 使用Repository模式：
//
//	// 定义用户仓库
//	type UserRepository struct {
//	    dbstore.Repository
//	}
//
//	// 创建新的用户仓库
//	userRepo := &UserRepository{
//	    Repository: dbstore.NewRepository(store.DB(), &User{}),
//	}
//
//	// 使用仓库方法
//	user, err := userRepo.FindByID(ctx, userID)
//
// # 最佳实践
//
// 1. 使用结构化的代码组织，如Repository模式分离数据访问逻辑
// 2. 适当使用事务确保数据一致性
// 3. 监控并优化慢查询
// 4. 为复杂查询编写单元测试
// 5. 使用迁移工具管理数据库结构变更
// 6. 合理配置连接池参数，避免连接耗尽
// 7. 在高并发环境中注意锁和死锁问题
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
// 定义数据库连接和行为的所有配置参数
// 可以通过代码直接设置或从配置文件加载
type Config struct {
	// Driver 数据库驱动类型
	// 支持的值: "mysql", "postgres", "sqlite"
	// 必填字段，决定使用哪种数据库系统
	Driver string `json:"driver" validate:"required,oneof=mysql postgres sqlite"`

	// DSN 数据源名称，即数据库连接字符串
	// 各驱动的格式不同:
	// - MySQL: "user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// - PostgreSQL: "host=localhost port=5432 user=user password=pass dbname=db sslmode=disable"
	// - SQLite: "file:db.sqlite?cache=shared"
	// 必填字段，包含连接数据库所需的所有信息
	DSN string `json:"dsn" validate:"required"`

	// MaxOpenConns 最大打开连接数
	// 控制到数据库的最大并发连接数
	// 推荐值: 根据数据库服务器能力和应用并发量设置，通常为10-100
	// 设置过大可能导致数据库压力过大，设置过小可能导致连接等待
	MaxOpenConns int `json:"max_open_conns" validate:"min=1"`

	// MaxIdleConns 最大空闲连接数
	// 连接池中保持的空闲连接数
	// 推荐值: 通常为MaxOpenConns的10%-30%
	// 设置合理的空闲连接数可以减少频繁创建连接的开销
	MaxIdleConns int `json:"max_idle_conns" validate:"min=1"`

	// ConnMaxLifetime 连接最大生存时间
	// 一个连接最多可以重用的时间，超过则关闭
	// 推荐值: 1小时到8小时，取决于数据库配置和网络环境
	// 防止连接因为网络问题或数据库设置而处于无效状态
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" validate:"min=1"`

	// ConnMaxIdleTime 连接最大空闲时间
	// 空闲连接保持多久后被关闭
	// 推荐值: 5分钟到30分钟
	// 释放长时间不用的连接，减少资源占用
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" validate:"min=1"`

	// TablePrefix 表名前缀
	// 所有表名自动添加的前缀，如"app_"会使user表变成app_user
	// 可选字段，适用于多应用共享数据库的场景
	TablePrefix string `json:"table_prefix"`

	// SingularTable 是否使用单数表名
	// true: 使用单数表名(user)，false: 使用复数表名(users)
	// 默认为false，使用复数表名
	SingularTable bool `json:"singular_table"`

	// Debug 是否启用调试模式
	// 启用后会打印所有SQL语句
	// 生产环境建议设为false，开发环境可以设为true
	Debug bool `json:"debug"`

	// SlowThreshold 慢查询阈值
	// 执行时间超过此阈值的查询会被记录为慢查询
	// 推荐值: 100ms-500ms，根据应用性能需求调整
	// 用于发现性能问题和优化目标
	SlowThreshold time.Duration `json:"slow_threshold" validate:"min=1ms"`

	// SkipDefaultTxn 是否跳过默认事务
	// true: 单个创建、更新、删除操作不启用事务
	// 对于只读查询或高性能要求的简单写操作有性能优势
	SkipDefaultTxn bool `json:"skip_default_txn"`

	// PrepareStmt 是否启用预处理语句
	// 启用后会缓存预处理语句，提高重复查询的性能
	// 推荐大多数场景下启用，除非有特殊需求
	PrepareStmt bool `json:"prepare_stmt"`

	// DisableNestedTxn 是否禁用嵌套事务
	// true: 禁止在事务中启动新事务
	// 在复杂业务逻辑中可以防止事务使用错误
	DisableNestedTxn bool `json:"disable_nested_txn"`
}

// DefaultConfig 返回默认数据库配置
// 返回：
//
//	Config: 包含合理默认值的配置实例
//
// 这些默认值适合大多数中小规模应用，大型应用可能需要根据负载特性调整
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
