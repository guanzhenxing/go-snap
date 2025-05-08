package logger

import (
	"reflect"
	"testing"
	"time"
)

func TestWithLevel(t *testing.T) {
	config := &Config{}
	option := WithLevel(DebugLevel)
	option(config)

	if config.level != DebugLevel {
		t.Errorf("WithLevel failed, expected %v, got %v", DebugLevel, config.level)
	}
}

func TestWithConsole(t *testing.T) {
	config := &Config{}

	option := WithConsole(true)
	option(config)
	if !config.console {
		t.Errorf("WithConsole(true) failed, expected true")
	}

	option = WithConsole(false)
	option(config)
	if config.console {
		t.Errorf("WithConsole(false) failed, expected false")
	}
}

func TestWithJSONConsole(t *testing.T) {
	config := &Config{}

	option := WithJSONConsole(true)
	option(config)
	if !config.consoleJSON {
		t.Errorf("WithJSONConsole(true) failed, expected true")
	}

	option = WithJSONConsole(false)
	option(config)
	if config.consoleJSON {
		t.Errorf("WithJSONConsole(false) failed, expected false")
	}
}

func TestWithFilename(t *testing.T) {
	config := &Config{}
	filename := "/tmp/test.log"

	option := WithFilename(filename)
	option(config)

	if config.filename != filename {
		t.Errorf("WithFilename failed, expected %s, got %s", filename, config.filename)
	}
}

func TestWithMaxSize(t *testing.T) {
	config := defaultConfig()
	size := 100

	option := WithMaxSize(size)
	option(config)

	if config.maxSize != size {
		t.Errorf("WithMaxSize failed, expected %d, got %d", size, config.maxSize)
	}

	// 检查是否同步到rollConfig
	if config.rollConfig.MaxSize != size {
		t.Errorf("WithMaxSize failed to sync to rollConfig, expected %d, got %d", size, config.rollConfig.MaxSize)
	}
}

func TestWithMaxBackups(t *testing.T) {
	config := defaultConfig()
	backups := 5

	option := WithMaxBackups(backups)
	option(config)

	if config.maxBackups != backups {
		t.Errorf("WithMaxBackups failed, expected %d, got %d", backups, config.maxBackups)
	}

	// 检查是否同步到rollConfig
	if config.rollConfig.MaxBackups != backups {
		t.Errorf("WithMaxBackups failed to sync to rollConfig, expected %d, got %d", backups, config.rollConfig.MaxBackups)
	}
}

func TestWithMaxAge(t *testing.T) {
	config := defaultConfig()
	age := 7

	option := WithMaxAge(age)
	option(config)

	if config.maxAge != age {
		t.Errorf("WithMaxAge failed, expected %d, got %d", age, config.maxAge)
	}

	// 检查是否同步到rollConfig
	if config.rollConfig.MaxAge != age {
		t.Errorf("WithMaxAge failed to sync to rollConfig, expected %d, got %d", age, config.rollConfig.MaxAge)
	}
}

func TestWithCompress(t *testing.T) {
	config := defaultConfig()

	// 先设置为false
	config.compress = false
	if config.rollConfig != nil {
		config.rollConfig.Compress = false
	}

	option := WithCompress(true)
	option(config)

	if !config.compress {
		t.Errorf("WithCompress failed, expected true")
	}

	// 检查是否同步到rollConfig
	if !config.rollConfig.Compress {
		t.Errorf("WithCompress failed to sync to rollConfig, expected true")
	}
}

func TestWithEncoderConfig(t *testing.T) {
	config := defaultConfig() // 使用默认配置

	// 创建自定义encoderConfig
	oldTimeKey := config.encoderConfig.TimeKey
	newTimeKey := "custom_time"

	// 复制现有配置并修改
	encoderConfig := config.encoderConfig
	encoderConfig.TimeKey = newTimeKey

	option := WithEncoderConfig(encoderConfig)
	option(config)

	if config.encoderConfig.TimeKey != newTimeKey {
		t.Errorf("WithEncoderConfig failed, expected TimeKey %s, got %s",
			newTimeKey, config.encoderConfig.TimeKey)
	}

	if config.encoderConfig.TimeKey == oldTimeKey {
		t.Errorf("WithEncoderConfig failed, TimeKey not changed")
	}
}

func TestWithFormat(t *testing.T) {
	config := defaultConfig()
	format := FormatOption{
		TimeFormat:        time.RFC3339Nano,
		CallerSkip:        2,
		Stacktrace:        false,
		StacktraceLevel:   PanicLevel,
		DisableCaller:     true,
		DisableStacktrace: true,
	}

	option := WithFormat(format)
	option(config)

	if !reflect.DeepEqual(config.format, format) {
		t.Errorf("WithFormat failed, format not set correctly")
	}
}

func TestWithAsync(t *testing.T) {
	config := &Config{}

	// 基本异步配置
	asyncOpt := AsyncOption{
		QueueSize:     500,
		FlushInterval: 5 * time.Second,
		Workers:       2,
		DropWhenFull:  true,
	}

	option := WithAsync(asyncOpt)
	option(config)

	if !config.async.Enabled {
		t.Errorf("WithAsync failed, async not enabled")
	}

	if config.async.QueueSize != 500 {
		t.Errorf("WithAsync failed, expected QueueSize 500, got %d", config.async.QueueSize)
	}

	if config.async.Workers != 2 {
		t.Errorf("WithAsync failed, expected Workers 2, got %d", config.async.Workers)
	}

	if !config.async.DropWhenFull {
		t.Errorf("WithAsync failed, expected DropWhenFull true")
	}
}

func TestWithTimeWindowSampling(t *testing.T) {
	config := &Config{}
	window := 5 * time.Second
	threshold := 100

	option := WithTimeWindowSampling(window, threshold)
	option(config)

	if config.timeWindow == nil {
		t.Fatalf("WithTimeWindowSampling failed, timeWindow is nil")
	}

	if config.timeWindow.Window != window {
		t.Errorf("WithTimeWindowSampling failed for window, expected %v, got %v",
			window, config.timeWindow.Window)
	}

	if config.timeWindow.Threshold != threshold {
		t.Errorf("WithTimeWindowSampling failed for threshold, expected %d, got %d",
			threshold, config.timeWindow.Threshold)
	}
}

func TestWithHook(t *testing.T) {
	config := &Config{}

	// 创建一个简单的hook
	testHook := func(level Level, msg string, fields ...Field) error {
		// 仅用于测试
		return nil
	}

	option := WithHook(testHook)
	option(config)

	if len(config.hooks) != 1 {
		t.Errorf("WithHook failed, expected 1 hook, got %d", len(config.hooks))
	}
}

func TestWithRollConfig(t *testing.T) {
	config := &Config{}
	rollConfig := RollConfig{
		MaxSize:     200,
		MaxBackups:  5,
		MaxAge:      14,
		Compress:    true,
		RollOnDate:  true,
		DatePattern: "2006-01-02",
	}

	option := WithRollConfig(rollConfig)
	option(config)

	if config.rollConfig == nil {
		t.Fatalf("WithRollConfig failed, rollConfig is nil")
	}

	if config.rollConfig.MaxSize != rollConfig.MaxSize {
		t.Errorf("WithRollConfig failed for MaxSize")
	}

	if config.rollConfig.MaxBackups != rollConfig.MaxBackups {
		t.Errorf("WithRollConfig failed for MaxBackups")
	}

	if config.rollConfig.DatePattern != rollConfig.DatePattern {
		t.Errorf("WithRollConfig failed for DatePattern")
	}

	// 检查是否同步到主配置
	if config.maxSize != rollConfig.MaxSize {
		t.Errorf("WithRollConfig failed to sync MaxSize to main config")
	}

	if config.maxBackups != rollConfig.MaxBackups {
		t.Errorf("WithRollConfig failed to sync MaxBackups to main config")
	}
}

func TestWithRollOnDate(t *testing.T) {
	config := defaultConfig()
	enabled := true
	pattern := "2006-01-02-15"

	option := WithRollOnDate(enabled, pattern)
	option(config)

	if config.rollConfig == nil {
		t.Errorf("WithRollOnDate failed, RollConfig is nil")
		return
	}

	if !config.rollConfig.RollOnDate {
		t.Errorf("WithRollOnDate failed, RollOnDate not enabled")
	}

	if config.rollConfig.DatePattern != pattern {
		t.Errorf("WithRollOnDate failed, expected pattern %s, got %s",
			pattern, config.rollConfig.DatePattern)
	}
}

func TestWithOpenTelemetry(t *testing.T) {
	config := &Config{}

	option := WithOpenTelemetry(true)
	option(config)

	if !config.otelEnabled {
		t.Errorf("WithOpenTelemetry failed, expected true")
	}
}

func TestWithSensitiveKeys(t *testing.T) {
	config := &Config{}
	keys := []string{"password", "ssn", "credit_card"}

	option := WithSensitiveKeys(keys...)
	option(config)

	if len(config.sensitiveKeys) != len(keys) {
		t.Errorf("WithSensitiveKeys failed, expected %d keys, got %d",
			len(keys), len(config.sensitiveKeys))
		return
	}

	for i, key := range keys {
		if config.sensitiveKeys[i] != key {
			t.Errorf("WithSensitiveKeys failed, expected key %s, got %s",
				key, config.sensitiveKeys[i])
		}
	}
}
