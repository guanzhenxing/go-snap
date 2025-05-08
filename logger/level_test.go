package logger

import (
	"os"
	"path/filepath"
	"testing"
)

// 测试日志级别字符串表示
func TestLevelStrings(t *testing.T) {
	levels := []struct {
		level Level
		str   string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{DPanicLevel, "dpanic"},
		{PanicLevel, "panic"},
		{FatalLevel, "fatal"},
	}

	for _, l := range levels {
		if l.level.String() != l.str {
			t.Errorf("Level %d string expected %s, got %s", l.level, l.str, l.level.String())
		}
	}
}

// 测试解析有效日志级别字符串
func TestParseLevelValid(t *testing.T) {
	tests := []struct {
		str   string
		level Level
	}{
		{"debug", DebugLevel},
		{"info", InfoLevel},
		{"warn", WarnLevel},
		{"error", ErrorLevel},
		{"dpanic", DPanicLevel},
		{"panic", PanicLevel},
		{"fatal", FatalLevel},
		// 大写和混合大小写
		{"DEBUG", DebugLevel},
		{"Info", InfoLevel},
		{"Warn", WarnLevel},
	}

	for _, tc := range tests {
		level, err := ParseLevel(tc.str)
		if err != nil {
			t.Errorf("Failed to parse valid level %s: %v", tc.str, err)
		}
		if level != tc.level {
			t.Errorf("Parsed level expected %v, got %v", tc.level, level)
		}
	}
}

// 测试解析无效日志级别字符串
func TestParseLevelInvalid(t *testing.T) {
	invalid := []string{"", "trace", "UNKNOWN", "critical"}

	for _, str := range invalid {
		level, err := ParseLevel(str)
		if err == nil {
			t.Errorf("Expected error for invalid level %s, got none", str)
		}
		if level != InfoLevel { // 默认值应该是InfoLevel
			t.Errorf("Invalid level expected default InfoLevel, got %v", level)
		}
	}
}

// 测试保存和加载日志级别到文件
func TestSaveAndLoadLogLevel(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "level_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 测试文件路径
	filename := filepath.Join(tmpDir, "level.txt")

	// 保存日志级别
	if err := SaveLogLevel(DebugLevel, filename); err != nil {
		t.Fatalf("Failed to save log level: %v", err)
	}

	// 加载日志级别
	level, err := LoadLogLevel(filename)
	if err != nil {
		t.Fatalf("Failed to load log level: %v", err)
	}

	if level != DebugLevel {
		t.Errorf("Loaded level expected DebugLevel, got %v", level)
	}

	// 测试读取不存在的文件
	nonExistFile := filepath.Join(tmpDir, "non_exist.txt")
	level, err = LoadLogLevel(nonExistFile)
	if err == nil {
		t.Error("Expected error when loading from non-existing file, got none")
	}
	if level != InfoLevel { // 默认应返回InfoLevel
		t.Errorf("Level from non-existing file expected InfoLevel, got %v", level)
	}
}

// 测试设置全局日志级别
func TestSetGlobalLevel(t *testing.T) {
	// 保存原始全局logger
	originalLogger := globalLogger
	defer func() {
		globalLogger = originalLogger
	}()

	// 初始化全局logger
	globalLogger = New(WithLevel(InfoLevel))

	// 设置全局日志级别
	SetGlobalLevel(DebugLevel)

	// 检查是否设置成功
	zapLogger, ok := globalLogger.(*zapLogger)
	if !ok {
		t.Fatal("Global logger is not a zapLogger")
	}

	if zapLogger.level != DebugLevel {
		t.Errorf("Global level expected DebugLevel, got %v", zapLogger.level)
	}
}

// 测试从文件设置全局日志级别
func TestSetGlobalLevelFromFile(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "level_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 测试文件路径
	filename := filepath.Join(tmpDir, "level.txt")

	// 保存日志级别
	if err := SaveLogLevel(WarnLevel, filename); err != nil {
		t.Fatalf("Failed to save log level: %v", err)
	}

	// 保存原始全局logger
	originalLogger := globalLogger
	defer func() {
		globalLogger = originalLogger
	}()

	// 初始化全局logger
	globalLogger = New(WithLevel(InfoLevel))

	// 从文件加载设置全局日志级别
	err = SetGlobalLevelFromFile(filename)
	if err != nil {
		t.Fatalf("Failed to set global level from file: %v", err)
	}

	// 检查是否设置成功
	zapLogger, ok := globalLogger.(*zapLogger)
	if !ok {
		t.Fatal("Global logger is not a zapLogger")
	}

	if zapLogger.level != WarnLevel {
		t.Errorf("Global level expected WarnLevel, got %v", zapLogger.level)
	}

	// 测试使用不存在的文件
	err = SetGlobalLevelFromFile(filepath.Join(tmpDir, "non_exist.txt"))
	if err == nil {
		t.Error("Expected error when setting from non-existing file, got none")
	}
}

// 测试保存全局级别到文件
func TestSaveGlobalLevel(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "level_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 测试文件路径
	filename := filepath.Join(tmpDir, "global_level.txt")

	// 保存原始全局logger
	originalLogger := globalLogger
	defer func() {
		globalLogger = originalLogger
	}()

	// 初始化全局logger
	globalLogger = New(WithLevel(ErrorLevel))

	// 保存全局日志级别
	err = SaveGlobalLevel(filename)
	if err != nil {
		t.Fatalf("Failed to save global level: %v", err)
	}

	// 加载保存的日志级别
	level, err := LoadLogLevel(filename)
	if err != nil {
		t.Fatalf("Failed to load saved global level: %v", err)
	}

	if level != ErrorLevel {
		t.Errorf("Saved global level expected ErrorLevel, got %v", level)
	}
}
