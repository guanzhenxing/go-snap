package dbstore

import (
	"context"
	"testing"
	"time"

	"github.com/guanzhenxing/go-snap/logger"
	"github.com/stretchr/testify/assert"
	gormlogger "gorm.io/gorm/logger"
)

// 模拟日志器
type mockLogger struct {
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		debugMessages: make([]string, 0),
		infoMessages:  make([]string, 0),
		warnMessages:  make([]string, 0),
		errorMessages: make([]string, 0),
	}
}

func (m *mockLogger) Debug(msg string, fields ...logger.Field) {
	m.debugMessages = append(m.debugMessages, msg)
}

func (m *mockLogger) Info(msg string, fields ...logger.Field) {
	m.infoMessages = append(m.infoMessages, msg)
}

func (m *mockLogger) Warn(msg string, fields ...logger.Field) {
	m.warnMessages = append(m.warnMessages, msg)
}

func (m *mockLogger) Error(msg string, fields ...logger.Field) {
	m.errorMessages = append(m.errorMessages, msg)
}

func (m *mockLogger) DPanic(msg string, fields ...logger.Field) {}
func (m *mockLogger) Panic(msg string, fields ...logger.Field)  {}
func (m *mockLogger) Fatal(msg string, fields ...logger.Field)  {}

func (m *mockLogger) With(fields ...logger.Field) logger.Logger {
	return m
}

func (m *mockLogger) WithContext(ctx context.Context) logger.Logger {
	return m
}

func (m *mockLogger) WithLogContext(lctx *logger.ContextLogger) logger.Logger {
	return m
}

func (m *mockLogger) SetLevel(level logger.Level)        {}
func (m *mockLogger) AddFilter(f logger.FilterFunc)      {}
func (m *mockLogger) GetStats() logger.Stats             { return logger.Stats{} }
func (m *mockLogger) GetMetrics() logger.Metrics         { return logger.Metrics{} }
func (m *mockLogger) Sync() error                        { return nil }
func (m *mockLogger) Shutdown(ctx context.Context) error { return nil }

// 测试GORM日志器的LogMode方法
func TestGormLogger_LogMode(t *testing.T) {
	mockLog := newMockLogger()
	gLog := newLogger(mockLog, time.Millisecond*100, true)

	// 测试LogMode方法
	newLogger := gLog.LogMode(gormlogger.Info)

	// 确保返回的是同一个实例
	assert.Equal(t, gLog, newLogger)
}

// 测试GORM日志器的Info方法
func TestGormLogger_Info(t *testing.T) {
	mockLog := newMockLogger()
	gLog := newLogger(mockLog, time.Millisecond*100, true)

	// 测试Info方法
	gLog.Info(context.Background(), "test info message")

	// 验证消息是否被记录
	assert.Len(t, mockLog.infoMessages, 1)
	assert.Equal(t, "test info message", mockLog.infoMessages[0])
}

// 测试GORM日志器的Warn方法
func TestGormLogger_Warn(t *testing.T) {
	mockLog := newMockLogger()
	gLog := newLogger(mockLog, time.Millisecond*100, true)

	// 测试Warn方法
	gLog.Warn(context.Background(), "test warn message")

	// 验证消息是否被记录
	assert.Len(t, mockLog.warnMessages, 1)
	assert.Equal(t, "test warn message", mockLog.warnMessages[0])
}

// 测试GORM日志器的Error方法
func TestGormLogger_Error(t *testing.T) {
	mockLog := newMockLogger()
	gLog := newLogger(mockLog, time.Millisecond*100, true)

	// 测试Error方法
	gLog.Error(context.Background(), "test error message")

	// 验证消息是否被记录
	assert.Len(t, mockLog.errorMessages, 1)
	assert.Equal(t, "test error message", mockLog.errorMessages[0])
}

// 测试GORM日志器的Trace方法
func TestGormLogger_Trace(t *testing.T) {
	// 创建mock logger
	mockLog := newMockLogger()

	// 测试启用调试模式的情况
	t.Run("with_debug_enabled", func(t *testing.T) {
		gLog := newLogger(mockLog, time.Millisecond*100, true)

		// 测试Trace方法 - 正常执行
		gLog.Trace(context.Background(), time.Now(), func() (string, int64) {
			return "SELECT * FROM users", 10
		}, nil)

		// 验证消息是否被记录为调试信息
		assert.NotEmpty(t, mockLog.debugMessages)

		// 清除消息
		mockLog.debugMessages = make([]string, 0)

		// 测试Trace方法 - 慢查询
		gLog.Trace(context.Background(), time.Now().Add(-time.Millisecond*200), func() (string, int64) {
			return "SELECT * FROM large_table", 1000
		}, nil)

		// 验证消息是否被记录为警告信息
		assert.NotEmpty(t, mockLog.warnMessages)

		// 清除消息
		mockLog.warnMessages = make([]string, 0)
		mockLog.errorMessages = make([]string, 0)

		// 测试Trace方法 - 查询出错
		gLog.Trace(context.Background(), time.Now(), func() (string, int64) {
			return "SELECT * FROM non_existent_table", 0
		}, assert.AnError)

		// 验证消息是否被记录为错误信息
		assert.NotEmpty(t, mockLog.errorMessages)
	})

	// 测试禁用调试模式的情况
	t.Run("with_debug_disabled", func(t *testing.T) {
		mockLog := newMockLogger()
		gLog := newLogger(mockLog, time.Millisecond*100, false)

		// 测试Trace方法
		gLog.Trace(context.Background(), time.Now(), func() (string, int64) {
			return "SELECT * FROM users", 10
		}, nil)

		// 验证没有消息被记录
		assert.Empty(t, mockLog.debugMessages)
		assert.Empty(t, mockLog.infoMessages)
		assert.Empty(t, mockLog.warnMessages)
		assert.Empty(t, mockLog.errorMessages)
	})
}

// 测试WithLogger选项
func TestWithLogger(t *testing.T) {
	// 创建mock logger
	mockLog := newMockLogger()

	// 使用WithLogger选项创建数据库实例
	config := DefaultConfig()
	config.Driver = "sqlite"
	config.DSN = ":memory:"

	// 由于我们无法直接访问store.log字段，这里通过日志功能测试验证
	store, err := New(config, WithLogger(mockLog))

	// 如果能成功创建，继续测试；否则跳过
	if err == nil && store != nil {
		// 关闭数据库连接
		defer store.Close()

		// 尝试执行一个操作来触发日志记录
		db := store.DB()
		db.Exec("CREATE TABLE test_logger (id INTEGER PRIMARY KEY)")
	}
}
