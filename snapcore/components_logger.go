package snapcore

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
)

// LogConfig 日志配置结构
type LogConfig struct {
	// 基本配置
	Level           string   `json:"level"`            // 日志级别
	Format          string   `json:"format"`           // 日志格式: json或text
	Console         bool     `json:"console"`          // 是否输出到控制台
	ConsoleJSON     bool     `json:"console_json"`     // 控制台输出是否使用JSON格式
	Filename        string   `json:"filename"`         // 日志文件名
	TimeFormat      string   `json:"time_format"`      // 时间格式化
	StacktraceLevel string   `json:"stacktrace_level"` // 堆栈跟踪级别
	DisableCaller   bool     `json:"disable_caller"`   // 是否禁用调用者信息
	SensitiveKeys   []string `json:"sensitive_keys"`   // 敏感字段列表

	// 日志文件滚动配置
	MaxSize     int    `json:"max_size"`     // 单个日志文件最大大小(MB)
	MaxBackups  int    `json:"max_backups"`  // 最大备份数量
	MaxAge      int    `json:"max_age"`      // 日志文件最大保存天数
	Compress    bool   `json:"compress"`     // 是否压缩旧日志文件
	RollOnDate  bool   `json:"roll_on_date"` // 是否按日期滚动
	DatePattern string `json:"date_pattern"` // 日期格式(默认2006-01-02)

	// 异步日志配置
	Async          bool `json:"async"`            // 是否启用异步日志
	AsyncQueueSize int  `json:"async_queue_size"` // 异步队列大小
	AsyncWorkers   int  `json:"async_workers"`    // 异步工作者数量

	// 日志采样配置
	SampleEnabled    bool `json:"sample_enabled"`    // 是否启用采样
	SampleInitial    int  `json:"sample_initial"`    // 初始记录数量
	SampleThereafter int  `json:"sample_thereafter"` // 之后每N条记录一条

	// 动态日志级别
	DynamicLevel  bool   `json:"dynamic_level"`   // 是否启用动态日志级别
	LevelFilePath string `json:"level_file_path"` // 日志级别文件路径
}

// LoggerComponent 日志组件适配器
type LoggerComponent struct {
	name             string
	logger           logger.Logger
	config           config.Provider
	opts             []logger.Option
	logConfig        LogConfig
	levelCheckTicker *time.Ticker
	stopChan         chan struct{}
}

// NewLoggerComponent 创建日志组件
func NewLoggerComponent(opts ...logger.Option) *LoggerComponent {
	return &LoggerComponent{
		name:     "logger",
		opts:     opts,
		stopChan: make(chan struct{}),
	}
}

// Initialize 初始化日志组件
func (c *LoggerComponent) Initialize(ctx context.Context) error {
	// 设置默认配置
	c.logConfig = LogConfig{
		Level:            "info",
		Format:           "text",
		Console:          true,
		ConsoleJSON:      false,
		TimeFormat:       time.RFC3339,
		MaxSize:          100,
		MaxBackups:       10,
		MaxAge:           30,
		Compress:         true,
		RollOnDate:       true,
		DatePattern:      "2006-01-02",
		Async:            false,
		AsyncQueueSize:   10000,
		AsyncWorkers:     1,
		SampleEnabled:    false,
		SampleInitial:    100,
		SampleThereafter: 100,
		DynamicLevel:     false,
		LevelFilePath:    ".loglevel",
	}

	// 从配置中获取日志配置
	if err := c.config.UnmarshalKey("logger", &c.logConfig); err != nil {
		return errors.Wrap(err, "解析日志配置失败")
	}

	// 合并配置选项
	var allOpts []logger.Option

	// 解析日志级别
	level := logger.InfoLevel
	if c.logConfig.Level != "" {
		if parsedLevel, err := logger.ParseLevel(c.logConfig.Level); err == nil {
			level = parsedLevel
		}
	}
	allOpts = append(allOpts, logger.WithLevel(level))

	// 设置控制台输出
	allOpts = append(allOpts, logger.WithConsole(c.logConfig.Console))
	if c.logConfig.ConsoleJSON {
		allOpts = append(allOpts, logger.WithJSONConsole(true))
	}

	// 设置日志文件
	if c.logConfig.Filename != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(c.logConfig.Filename)
		if logDir != "." && logDir != ".." {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return errors.Wrap(err, "创建日志目录失败")
			}
		}

		allOpts = append(allOpts, logger.WithFilename(c.logConfig.Filename))

		// 设置日志文件滚动配置
		rollCfg := logger.RollConfig{
			MaxSize:     c.logConfig.MaxSize,
			MaxBackups:  c.logConfig.MaxBackups,
			MaxAge:      c.logConfig.MaxAge,
			Compress:    c.logConfig.Compress,
			RollOnDate:  c.logConfig.RollOnDate,
			DatePattern: c.logConfig.DatePattern,
		}
		allOpts = append(allOpts, logger.WithRollConfig(rollCfg))
	}

	// 设置格式化选项
	stacktraceLevel := logger.ErrorLevel
	if c.logConfig.StacktraceLevel != "" {
		if parsedLevel, err := logger.ParseLevel(c.logConfig.StacktraceLevel); err == nil {
			stacktraceLevel = parsedLevel
		}
	}

	formatOpt := logger.FormatOption{
		TimeFormat:        c.logConfig.TimeFormat,
		StacktraceLevel:   stacktraceLevel,
		DisableCaller:     c.logConfig.DisableCaller,
		DisableStacktrace: false,
	}
	allOpts = append(allOpts, logger.WithFormat(formatOpt))

	// 设置异步日志
	if c.logConfig.Async {
		asyncOpt := logger.AsyncOption{
			Enabled:       true,
			QueueSize:     c.logConfig.AsyncQueueSize,
			FlushInterval: time.Second * 3,
			Workers:       c.logConfig.AsyncWorkers,
			DropWhenFull:  false,
		}
		allOpts = append(allOpts, logger.WithAsync(asyncOpt))
	}

	// 设置日志采样
	if c.logConfig.SampleEnabled {
		sampleCfg := &logger.SampleConfig{
			Initial:    c.logConfig.SampleInitial,
			Thereafter: c.logConfig.SampleThereafter,
			Tick:       time.Second,
		}
		allOpts = append(allOpts, logger.WithSample(sampleCfg))
	}

	// 设置敏感字段
	if len(c.logConfig.SensitiveKeys) > 0 {
		allOpts = append(allOpts, logger.WithSensitiveKeys(c.logConfig.SensitiveKeys...))
	}

	// 添加用户提供的选项
	allOpts = append(allOpts, c.opts...)

	// 初始化日志器
	c.logger = logger.New(allOpts...)

	// 如果启用动态日志级别，初始化级别文件
	if c.logConfig.DynamicLevel {
		// 检查日志级别文件是否存在，不存在则创建
		if _, err := os.Stat(c.logConfig.LevelFilePath); os.IsNotExist(err) {
			if err := logger.SaveLogLevel(level, c.logConfig.LevelFilePath); err != nil {
				c.logger.Warn("创建日志级别文件失败",
					logger.String("error", err.Error()),
					logger.String("path", c.logConfig.LevelFilePath),
				)
			}
		} else {
			// 尝试从文件加载日志级别
			if fileLevel, err := logger.LoadLogLevel(c.logConfig.LevelFilePath); err == nil {
				c.logger.SetLevel(fileLevel)
				c.logger.Info("从文件加载日志级别",
					logger.String("level", fileLevel.String()),
					logger.String("path", c.logConfig.LevelFilePath),
				)
			}
		}
	}

	// 记录初始化成功
	c.logger.Info("日志组件初始化成功",
		logger.String("level", level.String()),
		logger.String("format", c.logConfig.Format),
		logger.Bool("async", c.logConfig.Async),
	)

	return nil
}

// Start 启动日志组件
func (c *LoggerComponent) Start(ctx context.Context) error {
	// 启动动态日志级别检查（如果启用）
	if c.logConfig.DynamicLevel {
		c.startLevelChecker()
		c.logger.Info("启动动态日志级别检查",
			logger.String("level_file", c.logConfig.LevelFilePath),
		)
	}

	c.logger.Info("日志组件启动成功")
	return nil
}

// startLevelChecker 启动日志级别检查器
func (c *LoggerComponent) startLevelChecker() {
	c.levelCheckTicker = time.NewTicker(time.Second * 10)

	go func() {
		for {
			select {
			case <-c.levelCheckTicker.C:
				// 检查日志级别文件是否有变化
				if level, err := logger.LoadLogLevel(c.logConfig.LevelFilePath); err == nil {
					// 获取当前级别（这里需要类型断言，实际可能需要修改）
					var currentLevel logger.Level
					if stats := c.logger.GetStats(); stats.DebugCount > 0 {
						// 这是一个不完美的方法，实际项目中可能需要从logger模块暴露更好的API
						currentLevel = logger.DebugLevel
					} else {
						currentLevel = logger.InfoLevel
					}

					// 如果级别有变化，更新日志级别
					if level != currentLevel {
						c.logger.SetLevel(level)
						c.logger.Info("更新日志级别",
							logger.String("level", level.String()),
							logger.String("path", c.logConfig.LevelFilePath),
						)
					}
				}
			case <-c.stopChan:
				return
			}
		}
	}()
}

// Stop 停止日志组件
func (c *LoggerComponent) Stop(ctx context.Context) error {
	// 停止日志级别检查
	if c.levelCheckTicker != nil {
		c.levelCheckTicker.Stop()
		close(c.stopChan)
	}

	// 保存当前日志级别（如果启用动态日志级别）
	if c.logConfig.DynamicLevel {
		// 由于无法直接获取当前级别，暂不实现保存功能
	}

	c.logger.Info("日志组件停止")

	// 关闭日志器
	if c.logger != nil {
		if err := c.logger.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "关闭日志器失败")
		}
	}

	return nil
}

// Name 获取组件名称
func (c *LoggerComponent) Name() string {
	return c.name
}

// Type 获取组件类型
func (c *LoggerComponent) Type() ComponentType {
	return ComponentTypeInfrastructure
}

// Dependencies 获取组件依赖
func (c *LoggerComponent) Dependencies() []string {
	return []string{"config"} // 日志组件依赖配置组件
}

// GetLogger 获取日志器实例
func (c *LoggerComponent) GetLogger() logger.Logger {
	return c.logger
}

// SetConfig 设置配置提供器
func (c *LoggerComponent) SetConfig(config config.Provider) {
	c.config = config
}

// GetLogConfig 获取当前日志配置
func (c *LoggerComponent) GetLogConfig() LogConfig {
	return c.logConfig
}

// WithSensitiveKeys 设置敏感字段
func (c *LoggerComponent) WithSensitiveKeys(keys ...string) *LoggerComponent {
	if c.logConfig.SensitiveKeys == nil {
		c.logConfig.SensitiveKeys = make([]string, 0)
	}
	c.logConfig.SensitiveKeys = append(c.logConfig.SensitiveKeys, keys...)
	return c
}

// WithLogLevel 设置日志级别
func (c *LoggerComponent) WithLogLevel(level string) *LoggerComponent {
	c.logConfig.Level = level
	return c
}

// WithLogFile 设置日志文件
func (c *LoggerComponent) WithLogFile(filename string) *LoggerComponent {
	c.logConfig.Filename = filename
	return c
}

// WithDynamicLevel 设置动态日志级别
func (c *LoggerComponent) WithDynamicLevel(enabled bool, levelFilePath string) *LoggerComponent {
	c.logConfig.DynamicLevel = enabled
	if levelFilePath != "" {
		c.logConfig.LevelFilePath = levelFilePath
	}
	return c
}
