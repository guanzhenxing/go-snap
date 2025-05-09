package dbstore

import (
	"github.com/guanzhenxing/go-snap/logger"
)

// Option 定义Store配置选项
type Option func(*Store)

// WithLogger 设置自定义日志器
func WithLogger(log logger.Logger) Option {
	return func(s *Store) {
		s.log = log
	}
}

// WithDebug 设置调试模式
func WithDebug(debug bool) Option {
	return func(s *Store) {
		s.config.Debug = debug
	}
}

// WithTablePrefix 设置表前缀
func WithTablePrefix(prefix string) Option {
	return func(s *Store) {
		s.config.TablePrefix = prefix
	}
}

// WithSingularTable 设置单数表名
func WithSingularTable(singular bool) Option {
	return func(s *Store) {
		s.config.SingularTable = singular
	}
}

// WithPreparedStatement 设置预编译语句
func WithPreparedStatement(enabled bool) Option {
	return func(s *Store) {
		s.config.PrepareStmt = enabled
	}
}

// WithSkipDefaultTransaction 设置跳过默认事务
func WithSkipDefaultTransaction(skip bool) Option {
	return func(s *Store) {
		s.config.SkipDefaultTxn = skip
	}
}

// WithDisableNestedTransaction 设置禁用嵌套事务
func WithDisableNestedTransaction(disable bool) Option {
	return func(s *Store) {
		s.config.DisableNestedTxn = disable
	}
}
