package main

import (
	"context"
	"time"

	"github.com/guanzhenxing/go-snap/logger"
)

func main() {
	// 运行不同的示例
	basicExample()
	contextExample()
	mainAsyncLogger() // 调用async_logger_example.go中的函数
	mainFileLogger()  // 调用file_logger_example.go中的函数
	mainMaskData()    // 调用mask_data_example.go中的函数
}

// 1. 基本日志示例
func basicExample() {
	println("\n===== 基本日志示例 =====")

	// 初始化日志
	logger.Init(
		logger.WithLevel(logger.DebugLevel), // 设置日志级别为Debug
		logger.WithConsole(true),            // 输出到控制台
		logger.WithJSONConsole(true),        // 使用JSON格式输出
	)

	// 基础日志示例
	logger.Debug("这是一条调试日志")
	logger.Info("这是一条信息日志")
	logger.Warn("这是一条警告日志")
	logger.Error("这是一条错误日志")

	// 带字段的日志
	logger.Info("带字段的信息日志",
		logger.String("name", "张三"),
		logger.Int("age", 30),
		logger.Bool("active", true),
	)

	// 更多字段类型
	logger.Info("更多字段类型示例",
		logger.Duration("duration", 500*time.Millisecond),
		logger.Time("current_time", time.Now()),
		logger.Any("data", map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}),
	)

	// 使用With方法预设字段
	log := logger.With(
		logger.String("component", "user-service"),
		logger.String("server_id", "srv-01"),
	)

	// 使用预设字段的日志记录器
	log.Info("用户登录成功",
		logger.String("user_id", "user_123"),
		logger.Int("login_count", 5),
	)

	// 同步刷新缓冲区，确保所有日志写入
	logger.Sync()
}

// 2. 上下文日志示例
func contextExample() {
	println("\n===== 上下文日志示例 =====")

	// 初始化日志
	logger.Init(
		logger.WithLevel(logger.DebugLevel),
		logger.WithConsole(true),
		logger.WithJSONConsole(true),
	)

	// 方法1：使用ContextLogger
	contextLogger := logger.NewContextLogger().
		WithTraceID("trace-123").
		WithRequestID("req-456").
		WithUserID("user-789").
		WithClientIP("192.168.1.100").
		WithPath("/api/users").
		WithMethod("GET")

	// 从ContextLogger创建Logger
	log := logger.NewLogContextLoggerWithCtx(contextLogger)
	log.Info("处理API请求")

	// 方法2：使用context.Context
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TraceIDKey, "trace-abc")
	ctx = context.WithValue(ctx, logger.RequestIDKey, "req-xyz")
	ctx = context.WithValue(ctx, "user_id", "user-100")

	// 使用带有上下文的Logger
	logger.WithContext(ctx).Info("用户请求处理完成",
		logger.Int("status", 200),
		logger.Duration("latency", 150),
	)

	// 确保日志写入
	logger.Sync()
}

// 注：原始的asyncLoggerExample, fileLoggerExample, maskDataExample函数
// 已经移动到各自的文件中并重命名为mainAsyncLogger, mainFileLogger, mainMaskData
// 这里不再重复这些示例函数的代码
