package dbstore

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/guanzhenxing/go-snap/logger"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// gormLogger 实现了GORM的日志接口，将GORM日志适配到项目的日志系统
type gormLogger struct {
	logger        logger.Logger
	slowThreshold time.Duration
	debug         bool
}

// 创建新的GORM日志适配器
func newLogger(log logger.Logger, slowThreshold time.Duration, debug bool) gormlogger.Interface {
	return &gormLogger{
		logger:        log,
		slowThreshold: slowThreshold,
		debug:         debug,
	}
}

// LogMode 设置日志模式
func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

// Info 记录信息日志
func (l *gormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

// Warn 记录警告日志
func (l *gormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(msg, args...))
}

// Error 记录错误日志
func (l *gormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(msg, args...))
}

// Trace 记录SQL执行跟踪日志
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !l.debug {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 组合日志消息
	logMsg := fmt.Sprintf("[%.3fms] [rows:%d] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)

	// 根据错误级别和执行时间记录不同级别的日志
	switch {
	case err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound):
		l.logger.Error(fmt.Sprintf("%s [ERROR: %v]", logMsg, err))
	case elapsed > l.slowThreshold:
		l.logger.Warn(fmt.Sprintf("%s [SLOW]", logMsg))
	default:
		l.logger.Debug(logMsg)
	}
}
