package logger

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// AsyncOption 定义异步日志选项
type AsyncOption struct {
	Enabled       bool          // 是否启用异步
	QueueSize     int           // 异步队列大小
	FlushInterval time.Duration // 刷新间隔
	Workers       int           // 工作协程数量
	DropWhenFull  bool          // 队列满时丢弃日志
}

// SampleConfig 定义日志采样配置
type SampleConfig struct {
	Initial    int           // 初始全部记录的数量
	Thereafter int           // 之后每N条记录一条
	Tick       time.Duration // 时间窗口
}

// TimeWindowSampler 时间窗口采样器
type TimeWindowSampler struct {
	Window    time.Duration // 时间窗口
	Threshold int           // 窗口内阈值
}

// HookFunc 定义日志钩子函数
type HookFunc func(Level, string, ...Field) error

// RollConfig 定义日志滚动配置
type RollConfig struct {
	MaxSize     int    // 单个文件最大大小MB
	MaxBackups  int    // 最大备份数量
	MaxAge      int    // 最大保留天数
	Compress    bool   // 是否压缩
	RollOnDate  bool   // 是否按日期滚动
	DatePattern string // 日期格式
}

// Config 定义日志配置
type Config struct {
	level         Level
	console       bool
	consoleJSON   bool // 是否以JSON格式输出到控制台
	filename      string
	maxSize       int
	maxBackups    int
	maxAge        int
	compress      bool
	encoderConfig zapcore.EncoderConfig
	format        FormatOption
	sample        *SampleConfig
	async         AsyncOption
	timeWindow    *TimeWindowSampler
	hooks         []HookFunc
	rollConfig    *RollConfig // 日志滚动配置
	otelEnabled   bool        // 是否启用OpenTelemetry集成
	sensitiveKeys []string    // 需要自动脱敏的敏感字段
}

// defaultConfig 返回默认配置
func defaultConfig() *Config {
	return &Config{
		level:       InfoLevel,
		console:     true,
		consoleJSON: false,
		filename:    "",
		maxSize:     100, // MB
		maxBackups:  10,
		maxAge:      30, // days
		compress:    true,
		encoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
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
		},
		format: FormatOption{
			TimeFormat:        time.RFC3339,
			CallerSkip:        1,
			Stacktrace:        true,
			StacktraceLevel:   ErrorLevel,
			DisableCaller:     false,
			DisableStacktrace: false,
		},
		sample: &SampleConfig{
			Initial:    100,
			Thereafter: 100,
			Tick:       time.Second,
		},
		async: AsyncOption{
			Enabled:       false,
			QueueSize:     10000,
			FlushInterval: time.Second * 3,
			Workers:       1,
			DropWhenFull:  false,
		},
		hooks: make([]HookFunc, 0),
		rollConfig: &RollConfig{
			MaxSize:     100,
			MaxBackups:  10,
			MaxAge:      30,
			Compress:    true,
			RollOnDate:  false,
			DatePattern: "2006-01-02",
		},
		otelEnabled:   false,
		sensitiveKeys: []string{}, // 默认无敏感字段
	}
}

// Option 定义配置选项函数
type Option func(*Config)

// WithLevel 设置日志级别
func WithLevel(level Level) Option {
	return func(c *Config) {
		c.level = level
	}
}

// WithConsole 设置是否输出到控制台
func WithConsole(console bool) Option {
	return func(c *Config) {
		c.console = console
	}
}

// WithJSONConsole 设置是否以JSON格式输出到控制台
func WithJSONConsole(enabled bool) Option {
	return func(c *Config) {
		c.consoleJSON = enabled
	}
}

// WithFilename 设置日志文件名
func WithFilename(filename string) Option {
	return func(c *Config) {
		c.filename = filename
	}
}

// WithMaxSize 设置单个日志文件最大大小（MB）
func WithMaxSize(maxSize int) Option {
	return func(c *Config) {
		c.maxSize = maxSize
		c.rollConfig.MaxSize = maxSize
	}
}

// WithMaxBackups 设置最大备份文件数
func WithMaxBackups(maxBackups int) Option {
	return func(c *Config) {
		c.maxBackups = maxBackups
		c.rollConfig.MaxBackups = maxBackups
	}
}

// WithMaxAge 设置日志文件保留天数
func WithMaxAge(maxAge int) Option {
	return func(c *Config) {
		c.maxAge = maxAge
		c.rollConfig.MaxAge = maxAge
	}
}

// WithCompress 设置是否压缩备份文件
func WithCompress(compress bool) Option {
	return func(c *Config) {
		c.compress = compress
		c.rollConfig.Compress = compress
	}
}

// WithEncoderConfig 设置编码器配置
func WithEncoderConfig(encoderConfig zapcore.EncoderConfig) Option {
	return func(c *Config) {
		c.encoderConfig = encoderConfig
	}
}

// WithFormat 设置日志格式化选项
func WithFormat(format FormatOption) Option {
	return func(c *Config) {
		c.format = format
	}
}

// WithSample 设置日志采样配置
func WithSample(sample *SampleConfig) Option {
	return func(c *Config) {
		c.sample = sample
	}
}

// WithAsync 设置异步日志配置
func WithAsync(opt AsyncOption) Option {
	return func(c *Config) {
		opt.Enabled = true
		c.async = opt
	}
}

// WithTimeWindowSampling 设置时间窗口采样
func WithTimeWindowSampling(window time.Duration, threshold int) Option {
	return func(c *Config) {
		c.timeWindow = &TimeWindowSampler{
			Window:    window,
			Threshold: threshold,
		}
	}
}

// WithHook 添加日志钩子
func WithHook(hook HookFunc) Option {
	return func(c *Config) {
		c.hooks = append(c.hooks, hook)
	}
}

// WithRollConfig 设置日志滚动配置
func WithRollConfig(cfg RollConfig) Option {
	return func(c *Config) {
		c.rollConfig = &cfg
		// 同步到旧的配置
		c.maxSize = cfg.MaxSize
		c.maxBackups = cfg.MaxBackups
		c.maxAge = cfg.MaxAge
		c.compress = cfg.Compress
	}
}

// WithRollOnDate 设置是否按日期滚动日志
func WithRollOnDate(enabled bool, pattern string) Option {
	return func(c *Config) {
		c.rollConfig.RollOnDate = enabled
		if pattern != "" {
			c.rollConfig.DatePattern = pattern
		}
	}
}

// WithOpenTelemetry 启用OpenTelemetry集成
func WithOpenTelemetry(enabled bool) Option {
	return func(c *Config) {
		c.otelEnabled = enabled
	}
}

// WithSensitiveKeys 设置需要自动脱敏的敏感字段
func WithSensitiveKeys(keys ...string) Option {
	return func(c *Config) {
		c.sensitiveKeys = keys
	}
}
