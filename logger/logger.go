// Package logger 提供一个高性能、可配置的结构化日志系统
// 支持异步日志、日志级别控制、采样、字段脱敏等功能
package logger

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// FormatOption 定义日志格式化选项
type FormatOption struct {
	TimeFormat        string
	CallerSkip        int
	Stacktrace        bool
	StacktraceLevel   Level
	DisableCaller     bool
	DisableStacktrace bool
}

// LogEntry 定义日志条目
type LogEntry struct {
	Level   Level
	Message string
	Fields  []Field
	Time    time.Time
}

// Metrics 定义日志监控指标
type Metrics struct {
	WriteLatency  time.Duration
	BufferSize    int
	DroppedLogs   int64
	FilteredLogs  int64
	SampledLogs   int64
	TotalLogs     int64
	AsyncQueueLen int64 // 异步队列长度
}

// Logger 定义日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	DPanic(msg string, fields ...Field)
	Panic(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
	WithLogContext(lctx *ContextLogger) Logger
	SetLevel(level Level)
	AddFilter(f FilterFunc)
	GetStats() Stats
	GetMetrics() Metrics
	Sync() error
	Shutdown(ctx context.Context) error
}

// FilterFunc 定义日志过滤函数
type FilterFunc func(Level, string, ...Field) bool

// Stats 定义日志统计信息
type Stats struct {
	DebugCount  int64
	InfoCount   int64
	WarnCount   int64
	ErrorCount  int64
	DPanicCount int64
	PanicCount  int64
	FatalCount  int64
	LastError   error
	LastErrorAt time.Time
}

// asyncLogEntry 异步日志条目
type asyncLogEntry struct {
	level   Level
	msg     string
	fields  []Field
	logger  *zap.Logger
	counter *int64
}

// zapLogger zap日志实现
type zapLogger struct {
	zap     *zap.Logger
	level   Level
	fields  []Field
	filters []FilterFunc
	stats   Stats
	metrics Metrics
	mu      sync.RWMutex
	format  FormatOption
	sample  *SampleConfig
	pool    *sync.Pool

	// 异步日志相关
	async         bool
	asyncQueue    chan asyncLogEntry
	asyncWorkers  int
	asyncWg       sync.WaitGroup
	asyncQuit     chan struct{}
	dropWhenFull  bool
	flushInterval time.Duration

	// 时间窗口采样
	timeWindow    *TimeWindowSampler
	samplingStats map[Level]*samplingState

	// 钩子
	hooks []HookFunc
}

// samplingState 采样状态
type samplingState struct {
	counter   int
	lastReset time.Time
	mu        sync.Mutex
}

var (
	globalLogger Logger
	once         sync.Once
	fieldPool    = sync.Pool{
		New: func() interface{} {
			return make([]Field, 0, 16)
		},
	}

	// 初始化全局敏感字段列表
	globalSensitiveKeys []string
)

// checkGlobalLogger 检查全局日志实例是否初始化
func checkGlobalLogger() {
	if globalLogger == nil {
		panic("logger not initialized, please call logger.Init() first")
	}
}

// Init 初始化全局日志实例
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
