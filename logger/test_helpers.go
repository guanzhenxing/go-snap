package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 测试辅助函数，捕获日志输出到buffer
// 将输出定向到提供的buffer，创建一个临时的logger实例
// 适用于所有logger包测试场景
func captureOutput(t *testing.T, f func()) *bytes.Buffer {
	var buf bytes.Buffer

	// 创建自定义 encoder 和输出
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 获取现有logger的级别
	var level zapcore.Level = zapcore.DebugLevel
	if zl, ok := globalLogger.(*zapLogger); ok && zl != nil {
		level = zl.level
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		level,
	)

	originalLogger := globalLogger

	// 创建临时字段池
	pool := &sync.Pool{
		New: func() interface{} {
			return make([]Field, 0, 16)
		},
	}

	// 创建新logger，保持原来logger的级别
	newLogger := &zapLogger{
		zap:    zap.New(core),
		level:  level,
		fields: make([]Field, 0),
		pool:   pool,
	}

	// 复制原始logger的过滤器
	if zl, ok := globalLogger.(*zapLogger); ok && zl != nil {
		newLogger.filters = zl.filters
	}

	// 重要：设置全局logger为新创建的logger
	globalLogger = newLogger

	defer func() {
		// 确保将buf的内容写入完毕
		_ = newLogger.Sync()
		globalLogger = originalLogger
	}()

	// 执行测试函数
	f()

	return &buf
}

// 从buffer解析日志条目为map[string]interface{}
// 用于测试日志输出内容的验证
func parseLogEntry(t *testing.T, buf *bytes.Buffer) map[string]interface{} {
	t.Helper()

	// 读取并处理日志输出
	logText := buf.String()
	if len(logText) == 0 {
		t.Fatalf("Log buffer is empty")
		return nil
	}

	// 仅保留JSON部分
	jsonPart := logText
	if idx := strings.IndexRune(logText, '{'); idx > 0 {
		jsonPart = logText[idx:]
	}

	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(jsonPart), &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v\nRaw text: %s\nJSON part: %s",
			err, logText, jsonPart)
		return nil
	}

	return entry
}

// 为兼容性保留，与captureOutput功能相同
// 推荐使用captureOutput替代此函数
// @deprecated 使用captureOutput替代
func captureTestOutput(t *testing.T, f func()) *bytes.Buffer {
	return captureOutput(t, f)
}

// 为兼容性保留，与parseLogEntry功能相同
// 推荐使用parseLogEntry替代此函数
// @deprecated 使用parseLogEntry替代
func parseTestLogEntry(t *testing.T, buf *bytes.Buffer) map[string]interface{} {
	// 复用parseLogEntry实现
	return parseLogEntry(t, buf)
}

// 创建标准测试上下文
// 创建一个包含trace_id和request_id的上下文
func createTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, TraceIDKey, "trace-123")
	ctx = context.WithValue(ctx, RequestIDKey, "req-456")
	return ctx
}

// 创建带缓冲区的测试logger
// 返回一个缓冲区和对应的zap.Logger实例
func createBufferedLogger(t *testing.T) (*bytes.Buffer, *zap.Logger) {
	var buf bytes.Buffer
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)

	return &buf, zap.New(core)
}
