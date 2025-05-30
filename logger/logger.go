// Package logger 提供一个高性能、可配置的结构化日志系统
// 支持异步日志、日志级别控制、采样、字段脱敏等功能
// 基于zap日志库实现，并增加了额外的功能扩展
//
// # 日志系统架构
//
// 本包提供了一个灵活而强大的日志系统，主要组件包括：
//
// 1. Logger接口：定义了日志记录和管理的核心方法
// 2. Field类型：用于结构化日志字段
// 3. Level类型：定义日志级别和过滤策略
// 4. 上下文集成：支持从context中提取日志信息
// 5. 采样和过滤：支持高级日志过滤和采样策略
// 6. 异步处理：支持异步日志记录以提高性能
// 7. 指标监控：收集日志系统的性能指标
//
// # 日志级别
//
// 系统支持以下日志级别（从低到高）：
//   - Debug: 调试信息，仅在开发环境启用
//   - Info: 常规信息，记录正常操作
//   - Warn: 警告信息，潜在问题但不影响运行
//   - Error: 错误信息，操作失败但程序可继续
//   - DPanic: 开发模式下触发panic的严重错误
//   - Panic: 导致panic的严重错误
//   - Fatal: 致命错误，记录后会终止程序
//
// # 性能优化
//
// 本日志系统针对高性能场景进行了多项优化：
//
// 1. 结构化日志：避免了字符串拼接和解析开销
// 2. 级别过滤：在源头进行过滤，避免不必要的处理
// 3. 异步处理：支持非阻塞日志记录
// 4. 对象池：减少内存分配和GC压力
// 5. 日志采样：在高流量下自动采样
// 6. 字段复用：结构化字段支持复用以减少分配
//
// # 使用示例
//
// 1. 基本用法：
//
//	// 创建日志器
//	log := logger.New()
//
//	// 记录不同级别的日志
//	log.Debug("调试信息")
//	log.Info("系统启动")
//	log.Warn("配置过期", logger.String("config", "database"))
//	log.Error("连接失败", logger.Int("retries", 3), logger.String("service", "database"))
//
// 2. 结构化字段：
//
//	// 使用各种类型的字段
//	log.Info("用户登录",
//	    logger.String("user_id", "12345"),
//	    logger.Int("login_count", 5),
//	    logger.Bool("admin", false),
//	    logger.Duration("session_time", time.Hour),
//	)
//
// 3. 创建子日志器：
//
//	// 创建带有固定字段的日志器
//	requestLog := log.With(
//	    logger.String("request_id", requestID),
//	    logger.String("user_agent", userAgent),
//	)
//
//	// 在请求处理过程中使用
//	requestLog.Info("处理请求开始")
//	// ... 处理请求 ...
//	requestLog.Info("处理请求完成", logger.Duration("duration", duration))
//
// 4. 上下文集成：
//
//	// 从上下文创建日志器
//	ctx = logger.WithContext(ctx, log)
//
//	// 在其他函数中获取日志器
//	func processRequest(ctx context.Context) {
//	    log := logger.FromContext(ctx)
//	    log.Info("处理请求")
//	}
//
// 5. 异步日志：
//
//	// 创建异步日志器
//	log := logger.New(
//	    logger.WithAsync(true),
//	    logger.WithAsyncQueueSize(1000),
//	    logger.WithWorkers(2),
//	)
//
//	// 使用方式与同步日志器相同
//	log.Info("这是异步记录的")
//
//	// 应用退出前刷新
//	log.Sync()
//
// # 安全和隐私
//
// 日志系统提供了保护敏感数据的机制：
//
// 1. 字段脱敏：支持对敏感字段进行自动脱敏
//
//	log.Info("用户数据",
//	    logger.Sensitive("password", "secret123", logger.MaskAll),
//	    logger.Sensitive("credit_card", "1234-5678-9012-3456", logger.MaskLast4),
//	)
//
// 2. 日志清理：支持在输出前清理敏感数据
//
//	logger.New(logger.WithScrubber(myScrubberFunc))
//
// # 最佳实践
//
// 1. 使用结构化日志而非字符串拼接
// 2. 为每个请求创建带上下文的日志器
// 3. 在高流量系统中使用异步日志
// 4. 使用适当的日志级别
// 5. 在生产环境禁用Debug级别
// 6. 为敏感数据使用脱敏功能
// 7. 定期检查日志指标以发现性能问题
// 8. 在应用退出前调用Sync()确保日志刷新
package logger

import (
	"context"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/guanzhenxing/go-snap/errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// FormatOption 定义日志格式化选项，控制日志输出的格式和行为
// 用于配置日志的格式化行为和特性
type FormatOption struct {
	// TimeFormat 时间格式，遵循Go标准时间格式
	// 例如："2006-01-02 15:04:05.000" 或 time.RFC3339
	// 默认为ISO8601格式（类似于RFC3339）
	TimeFormat string

	// CallerSkip 调用栈跳过的层数，用于正确显示日志调用位置
	// 在包装日志库时，需要增加此值以显示正确的调用位置
	// 默认值为0，直接显示调用Logger方法的位置
	CallerSkip int

	// Stacktrace 是否启用堆栈跟踪
	// 如果为true，则在指定级别的日志中包含堆栈跟踪
	// 默认为false，除非是错误日志
	Stacktrace bool

	// StacktraceLevel 堆栈跟踪的最低级别，只有达到或超过此级别才会记录堆栈
	// 例如：设置为Error，则Error及以上级别会记录堆栈
	// 默认为Error级别
	StacktraceLevel Level

	// DisableCaller 是否禁用调用者信息
	// 如果为true，则日志中不会包含文件名和行号
	// 默认为false，包含调用位置信息
	DisableCaller bool

	// DisableStacktrace 是否禁用堆栈跟踪
	// 如果为true，则即使错误级别也不会记录堆栈
	// 此设置优先级高于Stacktrace
	DisableStacktrace bool
}

// LogEntry 定义日志条目，表示单条日志的所有信息
// 用于在日志处理链中传递日志信息，如过滤器和钩子
type LogEntry struct {
	// Level 日志级别
	// 指示此日志条目的严重程度
	Level Level

	// Message 日志消息
	// 日志的主要文本内容
	Message string

	// Fields 结构化字段
	// 包含与日志相关的额外结构化数据
	Fields []Field

	// Time 日志时间
	// 日志创建的时间戳
	Time time.Time
}

// Metrics 定义日志监控指标，用于监控日志系统性能和状态
// 这些指标对于理解日志系统行为和诊断性能问题很有价值
type Metrics struct {
	// WriteLatency 日志写入延迟
	// 记录日志的平均写入时间
	WriteLatency time.Duration

	// BufferSize 缓冲区大小
	// 当前日志缓冲区的大小（字节）
	BufferSize int

	// DroppedLogs 丢弃的日志数量
	// 由于缓冲区满或其他原因丢弃的日志计数
	DroppedLogs int64

	// FilteredLogs 被过滤的日志数量
	// 被过滤器拦截的日志计数
	FilteredLogs int64

	// SampledLogs 被采样的日志数量
	// 通过采样决定不记录的日志计数
	SampledLogs int64

	// TotalLogs 总日志数量
	// 尝试记录的总日志数
	TotalLogs int64

	// AsyncQueueLen 异步队列长度，仅在异步模式下有效
	// 当前等待处理的日志条目数
	AsyncQueueLen int64
}

// Logger 定义日志接口，提供各种日志记录方法和管理功能
// 所有日志操作应通过此接口进行，确保一致的日志行为
// 此接口设计支持链式调用和上下文感知
type Logger interface {
	// Debug 记录调试级别日志
	// 参数：
	//   msg: 日志消息
	//   fields: 可选的结构化字段
	// 说明：
	//   调试信息，通常仅在开发或调试环境中启用
	//   包含详细信息，帮助开发人员理解系统行为
	//   在生产环境中通常被禁用以提高性能
	Debug(msg string, fields ...Field)

	// Info 记录信息级别日志
	// 参数：
	//   msg: 日志消息
	//   fields: 可选的结构化字段
	// 说明：
	//   常规操作信息，表示正常的系统活动
	//   例如：服务启动、关闭、配置加载等
	//   适合记录业务流程的重要节点
	Info(msg string, fields ...Field)

	// Warn 记录警告级别日志
	// 参数：
	//   msg: 日志消息
	//   fields: 可选的结构化字段
	// 说明：
	//   警告信息，表示潜在问题但不影响当前操作
	//   例如：使用已废弃的API、配置次优、性能下降等
	//   通常需要运维或开发人员关注，但不需要立即响应
	Warn(msg string, fields ...Field)

	// Error 记录错误级别日志
	// 参数：
	//   msg: 日志消息
	//   fields: 可选的结构化字段
	// 说明：
	//   错误信息，表示操作失败，但不会导致系统崩溃
	//   例如：请求处理失败、数据库连接失败等
	//   通常会记录详细的错误信息和上下文，包括堆栈跟踪
	//   错误日志通常需要调查和处理
	Error(msg string, fields ...Field)

	// DPanic 记录开发环境下会触发panic的日志
	// 参数：
	//   msg: 日志消息
	//   fields: 可选的结构化字段
	// 说明：
	//   开发环境下的严重错误，会触发panic
	//   在生产环境中只记录日志而不触发panic
	//   用于在开发时及早发现严重问题
	DPanic(msg string, fields ...Field)

	// Panic 记录日志后触发panic
	// 参数：
	//   msg: 日志消息
	//   fields: 可选的结构化字段
	// 说明：
	//   严重错误，记录日志后会触发panic
	//   通常用于无法继续执行的严重错误
	//   会导致当前goroutine崩溃，但应用可能继续运行
	Panic(msg string, fields ...Field)

	// Fatal 记录日志后调用os.Exit(1)
	// 参数：
	//   msg: 日志消息
	//   fields: 可选的结构化字段
	// 说明：
	//   致命错误，记录日志后会终止程序
	//   通常用于应用无法继续运行的情况
	//   会导致整个应用程序立即退出
	//   应谨慎使用，通常只在初始化失败时使用
	Fatal(msg string, fields ...Field)

	// With 创建带有额外字段的Logger实例
	// 参数：
	//   fields: 要添加的结构化字段
	// 返回：
	//   新的Logger实例，包含所有添加的字段
	// 说明：
	//   创建一个新的日志器，继承所有配置，并添加固定字段
	//   这些字段将出现在通过新日志器创建的所有日志中
	//   适合为一组相关日志添加共同的上下文
	//   例如：为请求处理添加请求ID和用户ID
	With(fields ...Field) Logger

	// WithContext 创建带有上下文的Logger实例
	// 参数：
	//   ctx: 上下文，可能包含日志相关信息
	// 返回：
	//   新的Logger实例，集成了上下文中的日志字段
	// 说明：
	//   从上下文中提取日志相关信息，创建新的日志器
	//   例如：从请求上下文中提取请求ID、用户ID等
	//   如果上下文中没有日志信息，则返回原始日志器的副本
	WithContext(ctx context.Context) Logger

	// WithLogContext 创建带有日志上下文的Logger实例
	// 参数：
	//   lctx: 日志上下文，包含预定义的字段
	// 返回：
	//   新的Logger实例，包含日志上下文中的字段
	// 说明：
	//   日志上下文是一种专门为日志设计的上下文容器
	//   提供更高效和类型安全的方式来传递日志字段
	WithLogContext(lctx *ContextLogger) Logger

	// SetLevel 设置日志级别
	// 参数：
	//   level: 新的日志级别
	// 说明：
	//   动态调整日志记录的级别阈值
	//   只有大于等于此级别的日志才会被记录
	//   可用于在运行时调整日志详细程度，无需重启应用
	SetLevel(level Level)

	// AddFilter 添加日志过滤器
	// 参数：
	//   f: 过滤函数，返回true表示记录日志，false表示过滤掉
	// 说明：
	//   过滤器可以基于日志级别、消息内容和字段做细粒度过滤
	//   多个过滤器按添加顺序执行，任一返回false则不记录日志
	//   可用于实现复杂的日志过滤逻辑，如敏感信息过滤
	AddFilter(f FilterFunc)

	// GetStats 获取日志统计信息
	// 返回：
	//   日志统计数据
	// 说明：
	//   返回当前日志器的统计信息，包括各级别日志数量和最近错误
	//   可用于监控和调试日志系统
	GetStats() Stats

	// GetMetrics 获取日志指标数据
	// 返回：
	//   日志性能和状态指标
	// 说明：
	//   返回详细的性能指标，如延迟、丢弃日志数等
	//   适合集成到监控系统中，监控日志系统的健康状态
	GetMetrics() Metrics

	// Sync 同步日志，确保所有日志都被写入
	// 返回：
	//   同步过程中遇到的错误
	// 说明：
	//   特别是在异步模式下，此方法确保所有日志都被处理
	//   应用程序退出前应调用此方法
	//   即使在同步模式下，也可能需要刷新底层写入器的缓冲区
	Sync() error

	// Shutdown 优雅关闭日志系统
	// 参数：
	//   ctx: 用于控制关闭超时的上下文
	// 返回：
	//   关闭过程中遇到的错误
	// 说明：
	//   优雅关闭日志系统，包括：
	//   - 处理所有剩余的日志条目
	//   - 关闭异步工作线程
	//   - 刷新和关闭所有输出
	//   应用程序退出前应调用此方法
	Shutdown(ctx context.Context) error
}

// FilterFunc 定义日志过滤函数类型
// 参数：
//
//	Level: 日志级别
//	string: 日志消息
//	...Field: 日志字段
//
// 返回：
//
//	bool: true表示记录日志，false表示过滤掉
//
// 过滤器可以基于日志级别、消息内容和字段进行过滤
// 可以用于实现复杂的日志过滤逻辑，如敏感信息过滤或采样
type FilterFunc func(Level, string, ...Field) bool

// Stats 定义日志统计信息，记录各级别日志的计数和最近错误
// 提供日志系统运行状态的概览，便于监控和诊断问题
type Stats struct {
	// DebugCount 调试级别日志计数
	// 记录自创建以来的调试日志总数
	DebugCount int64

	// InfoCount 信息级别日志计数
	// 记录自创建以来的信息日志总数
	InfoCount int64

	// WarnCount 警告级别日志计数
	// 记录自创建以来的警告日志总数
	WarnCount int64

	// ErrorCount 错误级别日志计数
	// 记录自创建以来的错误日志总数
	ErrorCount int64

	// DPanicCount 开发级别panic日志计数
	// 记录自创建以来的DPanic日志总数
	DPanicCount int64

	// PanicCount panic级别日志计数
	// 记录自创建以来的Panic日志总数
	PanicCount int64

	// FatalCount 致命级别日志计数
	// 记录自创建以来的Fatal日志总数
	FatalCount int64

	// LastError 最近记录的错误
	// 保存最后一个错误级别或更高级别的日志错误
	LastError error

	// LastErrorAt 最近错误的时间戳
	// 最后一个错误记录的时间
	LastErrorAt time.Time
}

// asyncLogEntry 异步日志条目，用于异步日志处理
type asyncLogEntry struct {
	// level 日志级别
	level Level
	// msg 日志消息
	msg string
	// fields 日志字段
	fields []Field
	// logger 底层zap日志器
	logger *zap.Logger
	// counter 计数器引用，用于统计
	counter *int64
}

// zapLogger zap日志实现，实现了Logger接口
type zapLogger struct {
	// zap 底层zap日志器
	zap *zap.Logger
	// level 当前日志级别
	level Level
	// fields 基础字段，所有日志都会包含
	fields []Field
	// filters 日志过滤器列表
	filters []FilterFunc
	// stats 日志统计信息
	stats Stats
	// metrics 日志指标数据
	metrics Metrics
	// mu 互斥锁，保护并发访问
	mu sync.RWMutex
	// format 格式化选项
	format FormatOption
	// sample 采样配置
	sample *SampleConfig
	// pool 对象池，用于字段复用
	pool *sync.Pool

	// 异步日志相关
	// async 是否启用异步日志
	async bool
	// asyncQueue 异步日志队列
	asyncQueue chan asyncLogEntry
	// asyncWorkers 异步工作线程数量
	asyncWorkers int
	// asyncWg 工作线程等待组
	asyncWg sync.WaitGroup
	// asyncQuit 退出信号通道
	asyncQuit chan struct{}
	// dropWhenFull 队列满时是否丢弃日志
	dropWhenFull bool
	// flushInterval 定期刷新间隔
	flushInterval time.Duration

	// timeWindow 时间窗口采样器
	timeWindow *TimeWindowSampler
	// samplingStats 采样统计，按级别记录
	samplingStats map[Level]*samplingState

	// hooks 日志钩子函数列表
	hooks []HookFunc
}

// samplingState 采样状态，用于实现基于时间窗口的采样
type samplingState struct {
	// counter 计数器
	counter int
	// lastReset 上次重置时间
	lastReset time.Time
	// mu 互斥锁，保护并发访问
	mu sync.Mutex
}

var (
	// globalLogger 全局日志实例，通过Init初始化
	globalLogger Logger
	// once 确保全局日志只初始化一次
	once sync.Once
	// fieldPool 字段对象池，减少内存分配
	fieldPool = sync.Pool{
		New: func() interface{} {
			return make([]Field, 0, 16)
		},
	}

	// globalSensitiveKeys 全局敏感字段列表，这些字段的值会被脱敏
	globalSensitiveKeys []string
)

// checkGlobalLogger 检查全局日志实例是否初始化
// 如果未初始化则触发panic
func checkGlobalLogger() {
	if globalLogger == nil {
		panic("logger not initialized, please call logger.Init() first")
	}
}

// Init 初始化全局日志实例
// 参数：
//
//	opts: 日志配置选项
//
// 注意：
//
//	此函数保证只执行一次，重复调用不会重新初始化
func Init(opts ...Option) {
	once.Do(func() {
		config := defaultConfig()
		for _, opt := range opts {
			opt(config)
		}

		// 保存全局敏感字段列表
		if len(config.sensitiveKeys) > 0 {
			globalSensitiveKeys = make([]string, len(config.sensitiveKeys))
			copy(globalSensitiveKeys, config.sensitiveKeys)
		}

		globalLogger = New(opts...)
	})
}

// New 创建新的日志实例
// 参数：
//
//	opts: 日志配置选项
//
// 返回：
//
//	配置好的Logger实例
func New(opts ...Option) Logger {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// 创建核心
	var cores []zapcore.Core

	// 控制台输出
	if config.console {
		var consoleEncoder zapcore.Encoder
		if config.consoleJSON {
			consoleEncoder = zapcore.NewJSONEncoder(config.encoderConfig)
		} else {
			consoleEncoder = zapcore.NewConsoleEncoder(config.encoderConfig)
		}

		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			config.level,
		)
		cores = append(cores, consoleCore)
	}

	// 文件输出
	if config.filename != "" {
		var fileWriter zapcore.WriteSyncer

		if config.rollConfig.RollOnDate {
			// 使用基于日期的文件名
			today := time.Now().Format(config.rollConfig.DatePattern)
			filename := strings.Replace(config.filename, ".log", "-"+today+".log", 1)

			fileWriter = zapcore.AddSync(&lumberjack.Logger{
				Filename:   filename,
				MaxSize:    config.rollConfig.MaxSize,
				MaxBackups: config.rollConfig.MaxBackups,
				MaxAge:     config.rollConfig.MaxAge,
				Compress:   config.rollConfig.Compress,
			})
		} else {
			fileWriter = zapcore.AddSync(&lumberjack.Logger{
				Filename:   config.filename,
				MaxSize:    config.maxSize,
				MaxBackups: config.maxBackups,
				MaxAge:     config.maxAge,
				Compress:   config.compress,
			})
		}

		fileEncoder := zapcore.NewJSONEncoder(config.encoderConfig)
		fileCore := zapcore.NewCore(
			fileEncoder,
			fileWriter,
			config.level,
		)
		cores = append(cores, fileCore)
	}

	// 创建zap logger
	core := zapcore.NewTee(cores...)
	z := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(config.format.CallerSkip),
		zap.AddStacktrace(config.format.StacktraceLevel),
	)

	logger := &zapLogger{
		zap:           z,
		level:         config.level,
		fields:        make([]Field, 0),
		format:        config.format,
		sample:        config.sample,
		pool:          &fieldPool,
		metrics:       Metrics{},
		async:         config.async.Enabled,
		asyncWorkers:  config.async.Workers,
		dropWhenFull:  config.async.DropWhenFull,
		flushInterval: config.async.FlushInterval,
		timeWindow:    config.timeWindow,
		samplingStats: make(map[Level]*samplingState),
		hooks:         config.hooks,
	}

	// 初始化采样状态
	for _, level := range []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, PanicLevel, FatalLevel} {
		logger.samplingStats[level] = &samplingState{
			counter:   0,
			lastReset: time.Now(),
		}
	}

	// 初始化异步写入
	if logger.async {
		logger.asyncQueue = make(chan asyncLogEntry, config.async.QueueSize)
		logger.asyncQuit = make(chan struct{})

		// 启动工作协程
		for i := 0; i < logger.asyncWorkers; i++ {
			logger.asyncWg.Add(1)
			go logger.asyncWorker()
		}

		// 启动定时刷新协程
		if logger.flushInterval > 0 {
			go logger.periodicFlush()
		}
	}

	return logger
}

// 实现Logger接口的方法
func (l *zapLogger) Debug(msg string, fields ...Field) {
	if l.shouldLog(DebugLevel, msg, fields...) {
		l.log(DebugLevel, msg, fields...)
		atomic.AddInt64(&l.stats.DebugCount, 1)
	}
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	if l.shouldLog(InfoLevel, msg, fields...) {
		l.log(InfoLevel, msg, fields...)
		atomic.AddInt64(&l.stats.InfoCount, 1)
	}
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	if l.shouldLog(WarnLevel, msg, fields...) {
		l.log(WarnLevel, msg, fields...)
		atomic.AddInt64(&l.stats.WarnCount, 1)
	}
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	if l.shouldLog(ErrorLevel, msg, fields...) {
		l.log(ErrorLevel, msg, fields...)
		atomic.AddInt64(&l.stats.ErrorCount, 1)
		l.mu.Lock()
		l.stats.LastError = errors.New(msg)
		l.stats.LastErrorAt = time.Now()
		l.mu.Unlock()
	}
}

func (l *zapLogger) DPanic(msg string, fields ...Field) {
	if l.shouldLog(DPanicLevel, msg, fields...) {
		l.log(DPanicLevel, msg, fields...)
		atomic.AddInt64(&l.stats.DPanicCount, 1)
	}
}

func (l *zapLogger) Panic(msg string, fields ...Field) {
	if l.shouldLog(PanicLevel, msg, fields...) {
		l.log(PanicLevel, msg, fields...)
		atomic.AddInt64(&l.stats.PanicCount, 1)
	}
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	if l.shouldLog(FatalLevel, msg, fields...) {
		l.log(FatalLevel, msg, fields...)
		atomic.AddInt64(&l.stats.FatalCount, 1)
	}
}

// log 统一日志记录函数
func (l *zapLogger) log(level Level, msg string, fields ...Field) {
	// 执行钩子
	for _, hook := range l.hooks {
		if err := hook(level, msg, fields...); err != nil {
			// 如果钩子执行失败，记录错误但继续执行
			l.zap.Error("hook execution failed", zap.Error(err))
		}
	}

	// 异步还是同步写入
	if l.async {
		l.asyncLog(level, msg, fields...)
	} else {
		start := time.Now()
		switch level {
		case DebugLevel:
			l.zap.Debug(msg, fields...)
		case InfoLevel:
			l.zap.Info(msg, fields...)
		case WarnLevel:
			l.zap.Warn(msg, fields...)
		case ErrorLevel:
			l.zap.Error(msg, fields...)
		case DPanicLevel:
			l.zap.DPanic(msg, fields...)
		case PanicLevel:
			l.zap.Panic(msg, fields...)
		case FatalLevel:
			l.zap.Fatal(msg, fields...)
		}
		l.updateMetrics(start)
	}
}

func (l *zapLogger) With(fields ...Field) Logger {
	newLogger := &zapLogger{
		zap:           l.zap.With(fields...),
		level:         l.level,
		fields:        append(l.fields, fields...),
		filters:       l.filters,
		stats:         l.stats,
		format:        l.format,
		sample:        l.sample,
		pool:          l.pool,
		metrics:       l.metrics,
		async:         l.async,
		asyncQueue:    l.asyncQueue,
		asyncWorkers:  l.asyncWorkers,
		asyncQuit:     l.asyncQuit,
		dropWhenFull:  l.dropWhenFull,
		flushInterval: l.flushInterval,
		timeWindow:    l.timeWindow,
		samplingStats: l.samplingStats,
		hooks:         l.hooks,
	}
	return newLogger
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}
	fields := l.pool.Get().([]Field)
	defer l.pool.Put(fields)

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		fields = append(fields, String(TraceIDKey, traceID))
	}
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		fields = append(fields, String(RequestIDKey, requestID))
	}
	if len(fields) > 0 {
		return l.With(fields...)
	}
	return l
}

func (l *zapLogger) WithLogContext(lctx *ContextLogger) Logger {
	if lctx == nil {
		return l
	}

	return l.With(lctx.ToFields()...)
}

func (l *zapLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *zapLogger) AddFilter(f FilterFunc) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.filters = append(l.filters, f)
}

func (l *zapLogger) GetStats() Stats {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.stats
}

func (l *zapLogger) GetMetrics() Metrics {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 更新异步队列长度
	if l.async {
		queueLen := int64(len(l.asyncQueue))
		atomic.StoreInt64(&l.metrics.AsyncQueueLen, queueLen)
	}

	return l.metrics
}

func (l *zapLogger) Sync() error {
	err := l.zap.Sync()

	// 忽略控制台sync的ioctl错误
	if err != nil && (strings.Contains(err.Error(), "inappropriate ioctl for device") ||
		strings.Contains(err.Error(), "invalid argument")) {
		return nil
	}

	return err
}

func (l *zapLogger) Shutdown(ctx context.Context) error {
	if l.async {
		// 优雅关闭异步队列
		close(l.asyncQuit)

		// 等待队列处理完毕
		done := make(chan struct{})
		go func() {
			l.asyncWg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// 关闭已完成
		case <-ctx.Done():
			return ctx.Err()
		}

		// 关闭队列
		close(l.asyncQueue)
	}

	// 同步缓存
	done := make(chan error, 1)
	go func() {
		done <- l.Sync()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// shouldLog 检查是否应该记录日志
func (l *zapLogger) shouldLog(level Level, msg string, fields ...Field) bool {
	if level < l.level {
		return false
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	// 检查过滤器
	for _, filter := range l.filters {
		if !filter(level, msg, fields...) {
			atomic.AddInt64(&l.metrics.FilteredLogs, 1)
			return false
		}
	}

	// 检查基本采样
	if l.sample != nil {
		if !l.shouldBasicSample() {
			atomic.AddInt64(&l.metrics.SampledLogs, 1)
			return false
		}
	}

	// 检查时间窗口采样
	if l.timeWindow != nil {
		if !l.shouldTimeWindowSample(level) {
			atomic.AddInt64(&l.metrics.SampledLogs, 1)
			return false
		}
	}

	atomic.AddInt64(&l.metrics.TotalLogs, 1)
	return true
}

// shouldBasicSample 基本采样逻辑
func (l *zapLogger) shouldBasicSample() bool {
	if l.sample == nil {
		return true
	}

	// 实现采样逻辑
	count := atomic.AddInt64(&l.metrics.SampledLogs, 1)
	if count <= int64(l.sample.Initial) {
		return true
	}
	return count%int64(l.sample.Thereafter) == 0
}

// shouldTimeWindowSample 时间窗口采样
func (l *zapLogger) shouldTimeWindowSample(level Level) bool {
	if l.timeWindow == nil {
		return true
	}

	state := l.samplingStats[level]
	state.mu.Lock()
	defer state.mu.Unlock()

	// 检查是否需要重置窗口
	now := time.Now()
	if now.Sub(state.lastReset) > l.timeWindow.Window {
		state.counter = 0
		state.lastReset = now
	}

	// 检查是否超过阈值
	state.counter++
	return state.counter <= l.timeWindow.Threshold
}

// updateMetrics 更新监控指标
func (l *zapLogger) updateMetrics(start time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.metrics.WriteLatency = time.Since(start)
}

// 全局日志方法
func Debug(msg string, fields ...Field) {
	checkGlobalLogger()
	globalLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	checkGlobalLogger()
	globalLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	checkGlobalLogger()
	globalLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	checkGlobalLogger()
	globalLogger.Error(msg, fields...)
}

func DPanic(msg string, fields ...Field) {
	checkGlobalLogger()
	globalLogger.DPanic(msg, fields...)
}

func Panic(msg string, fields ...Field) {
	checkGlobalLogger()
	globalLogger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	checkGlobalLogger()
	globalLogger.Fatal(msg, fields...)
}

func With(fields ...Field) Logger {
	checkGlobalLogger()
	return globalLogger.With(fields...)
}

func WithContext(ctx context.Context) Logger {
	checkGlobalLogger()
	return globalLogger.WithContext(ctx)
}

func SetLevel(level Level) {
	checkGlobalLogger()
	globalLogger.SetLevel(level)
}

func AddFilter(f FilterFunc) {
	checkGlobalLogger()
	globalLogger.AddFilter(f)
}

func GetStats() Stats {
	checkGlobalLogger()
	return globalLogger.GetStats()
}

func GetMetrics() Metrics {
	checkGlobalLogger()
	return globalLogger.GetMetrics()
}

func Sync() error {
	checkGlobalLogger()
	return globalLogger.Sync()
}

func Shutdown(ctx context.Context) error {
	checkGlobalLogger()
	return globalLogger.Shutdown(ctx)
}

// AddHook 添加日志钩子（全局方法）
func AddHook(hook HookFunc) {
	checkGlobalLogger()

	// 将钩子添加到全局logger
	if zap, ok := globalLogger.(*zapLogger); ok {
		zap.mu.Lock()
		defer zap.mu.Unlock()
		zap.hooks = append(zap.hooks, hook)
	}
}

// contains 判断字符串是否在字符串切片中
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
