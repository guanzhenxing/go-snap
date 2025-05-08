package main

import (
	"time"

	"github.com/guanzhenxing/go-snap/logger"
)

func mainAsyncLogger() {
	// 创建异步日志选项
	asyncOpt := logger.AsyncOption{
		QueueSize:     10000,           // 队列大小
		FlushInterval: 3 * time.Second, // 刷新间隔
		Workers:       2,               // 工作线程数
		DropWhenFull:  false,           // 队列满时是否丢弃日志
	}

	// 初始化日志，启用异步模式
	logger.Init(
		logger.WithLevel(logger.DebugLevel),
		logger.WithConsole(true),
		logger.WithAsync(asyncOpt),
	)

	// 记录一条普通日志
	logger.Info("异步日志初始化完成")

	// 记录大量日志 - 在高并发场景下，异步日志可以提高性能
	for i := 0; i < 1000; i++ {
		logger.Debug("异步日志测试",
			logger.Int("index", i),
			logger.Time("timestamp", time.Now()),
		)
	}

	// 重要信息日志
	logger.Info("重要操作完成",
		logger.String("operation", "data_backup"),
		logger.Bool("success", true),
	)

	// 警告日志 - 即使是异步模式，高级别日志依然很重要
	logger.Warn("系统资源不足",
		logger.Int("memory_usage_percent", 85),
		logger.Int("cpu_usage_percent", 90),
	)

	// 错误日志
	logger.Error("服务调用失败",
		logger.String("service", "payment-api"),
		logger.String("error", "timeout after 30s"),
		logger.Int("status_code", 504),
	)

	// 在应用程序退出前，确保所有日志都已写入
	// 这对于异步日志特别重要，否则可能会丢失最后的一些日志
	logger.Sync()

	// 实际应用中应该优雅地关闭日志
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// logger.Shutdown(ctx)

	// 注意：在真实应用中，可能需要等待一段时间确保异步日志刷新
	// time.Sleep(3 * time.Second)
}
