package main

import (
	"path/filepath"
	"time"

	"github.com/guanzhenxing/go-snap/logger"
)

func mainFileLogger() {
	// 设置日志文件路径
	logPath := filepath.Join("logs", "app.log")

	// 初始化日志，配置文件输出
	logger.Init(
		logger.WithLevel(logger.InfoLevel),
		logger.WithConsole(true),     // 同时输出到控制台
		logger.WithFilename(logPath), // 设置日志文件路径
		logger.WithMaxSize(10),       // 单个日志文件最大大小，单位MB
		logger.WithMaxBackups(5),     // 最多保留5个备份
		logger.WithMaxAge(30),        // 最多保留30天
		logger.WithCompress(true),    // 压缩旧日志文件
	)

	// 记录一些日志
	logger.Info("应用程序启动")

	// 带有结构化字段的日志
	logger.Info("系统配置加载完成",
		logger.Int("worker_count", 8),
		logger.String("mode", "production"),
		logger.Bool("debug", false),
	)

	// 模拟一些操作
	for i := 0; i < 3; i++ {
		logger.Info("处理任务",
			logger.Int("task_id", i),
			logger.String("status", "processing"),
		)
		// 模拟处理耗时
		time.Sleep(100 * time.Millisecond)

		logger.Info("任务完成",
			logger.Int("task_id", i),
			logger.String("status", "completed"),
			logger.Duration("duration", 100*time.Millisecond),
		)
	}

	// 记录一个错误
	logger.Error("连接数据库失败",
		logger.String("db_host", "localhost:5432"),
		logger.String("error", "connection timeout"),
	)

	// 在退出前同步刷新日志缓冲区
	logger.Sync()

	logger.Info("应用程序正常退出")

	// 在实际应用中，可能还会需要停止日志系统
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// logger.Shutdown(ctx)
}
