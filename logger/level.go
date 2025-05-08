package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
)

// Level 定义日志级别
type Level = zapcore.Level

// 日志级别常量
const (
	DebugLevel  = zapcore.DebugLevel
	InfoLevel   = zapcore.InfoLevel
	WarnLevel   = zapcore.WarnLevel
	ErrorLevel  = zapcore.ErrorLevel
	DPanicLevel = zapcore.DPanicLevel
	PanicLevel  = zapcore.PanicLevel
	FatalLevel  = zapcore.FatalLevel
)

// 日志级别配置常量
const (
	// DefaultLevelFile 默认日志级别配置文件路径
	DefaultLevelFile = ".loglevel"
)

// ParseLevel 解析日志级别字符串
func ParseLevel(levelStr string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "dpanic":
		return DPanicLevel, nil
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	default:
		return InfoLevel, fmt.Errorf("未知的日志级别: %s", levelStr)
	}
}

// SaveLogLevel 保存日志级别到文件
func SaveLogLevel(level Level, path string) error {
	if path == "" {
		path = DefaultLevelFile
	}

	var levelStr string
	switch level {
	case DebugLevel:
		levelStr = "debug"
	case InfoLevel:
		levelStr = "info"
	case WarnLevel:
		levelStr = "warn"
	case ErrorLevel:
		levelStr = "error"
	case DPanicLevel:
		levelStr = "dpanic"
	case PanicLevel:
		levelStr = "panic"
	case FatalLevel:
		levelStr = "fatal"
	default:
		levelStr = "info"
	}

	// 将日志级别字符串写入文件
	return os.WriteFile(path, []byte(levelStr), 0644)
}

// SetGlobalLevel 设置全局日志级别
func SetGlobalLevel(level Level) {
	SetLevel(level)
}

// SetGlobalLevelFromFile 从文件加载并设置全局日志级别
func SetGlobalLevelFromFile(path string) error {
	level, err := LoadLogLevel(path)
	if err != nil {
		return err
	}

	SetGlobalLevel(level)
	return nil
}

// LoadLogLevel 从文件加载日志级别
func LoadLogLevel(path string) (Level, error) {
	if path == "" {
		path = DefaultLevelFile
	}

	// 读取文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		return InfoLevel, fmt.Errorf("读取日志级别文件失败: %w", err)
	}

	// 解析日志级别
	level, err := ParseLevel(string(data))
	if err != nil {
		return InfoLevel, fmt.Errorf("解析日志级别失败: %w", err)
	}

	return level, nil
}

// SaveGlobalLevel 保存当前全局日志级别到文件
func SaveGlobalLevel(path string) error {
	// 在实现之前需要确保全局日志记录器已初始化
	checkGlobalLogger()

	// 从全局记录器中获取当前级别(类型断言)
	if zl, ok := globalLogger.(*zapLogger); ok {
		return SaveLogLevel(zl.level, path)
	}

	// 如果不是zapLogger实例，则无法获取其级别
	return fmt.Errorf("全局日志记录器不是zapLogger实例，无法获取日志级别")
}
