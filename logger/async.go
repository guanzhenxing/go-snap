package logger

import (
	"fmt"
	"sync/atomic"
	"time"
)

// 这里不重复定义asyncLogEntry结构体，在logger.go中已经定义

// asyncWorker 异步日志工作协程
func (l *zapLogger) asyncWorker() {
	defer l.asyncWg.Done()

	for {
		select {
		case entry, ok := <-l.asyncQueue:
			if !ok {
				return
			}
			// 实际写入日志
			start := time.Now()
			switch entry.level {
			case DebugLevel:
				entry.logger.Debug(entry.msg, entry.fields...)
			case InfoLevel:
				entry.logger.Info(entry.msg, entry.fields...)
			case WarnLevel:
				entry.logger.Warn(entry.msg, entry.fields...)
			case ErrorLevel:
				entry.logger.Error(entry.msg, entry.fields...)
			case DPanicLevel:
				entry.logger.DPanic(entry.msg, entry.fields...)
			case PanicLevel:
				entry.logger.Panic(entry.msg, entry.fields...)
			case FatalLevel:
				entry.logger.Fatal(entry.msg, entry.fields...)
			}

			l.updateMetrics(start)
			if entry.counter != nil {
				atomic.AddInt64(entry.counter, 1)
			}

		case <-l.asyncQuit:
			return
		}
	}
}

// periodicFlush 定期刷新日志
func (l *zapLogger) periodicFlush() {
	ticker := time.NewTicker(l.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.Sync()
		case <-l.asyncQuit:
			return
		}
	}
}

// asyncLog 异步写入日志
func (l *zapLogger) asyncLog(level Level, msg string, fields ...Field) {
	var counter *int64

	// 自动脱敏处理敏感字段
	if globalSensitiveKeys != nil && len(globalSensitiveKeys) > 0 && len(fields) > 0 {
		for i, field := range fields {
			if field.Key != "" && contains(globalSensitiveKeys, field.Key) {
				// 替换为脱敏后的字段
				fields[i] = AutoMask(field.Key, fmt.Sprintf("%v", field.Interface))
			}
		}
	}

	// 根据级别获取对应的计数器
	switch level {
	case DebugLevel:
		counter = &l.stats.DebugCount
	case InfoLevel:
		counter = &l.stats.InfoCount
	case WarnLevel:
		counter = &l.stats.WarnCount
	case ErrorLevel:
		counter = &l.stats.ErrorCount
	case DPanicLevel:
		counter = &l.stats.DPanicCount
	case PanicLevel:
		counter = &l.stats.PanicCount
	case FatalLevel:
		counter = &l.stats.FatalCount
	}

	// 复制字段以防被修改
	fieldsCopy := make([]Field, len(fields))
	copy(fieldsCopy, fields)

	entry := asyncLogEntry{
		level:   level,
		msg:     msg,
		fields:  fieldsCopy,
		logger:  l.zap,
		counter: counter,
	}

	// 更新队列长度指标
	queueLen := int64(len(l.asyncQueue))
	atomic.StoreInt64(&l.metrics.AsyncQueueLen, queueLen)

	// 发送到异步队列
	if l.dropWhenFull {
		// 非阻塞发送，队列满时丢弃
		select {
		case l.asyncQueue <- entry:
			// 成功发送
		default:
			// 队列满，丢弃日志
			atomic.AddInt64(&l.metrics.DroppedLogs, 1)
		}
	} else {
		// 阻塞发送，直到队列有空间
		l.asyncQueue <- entry
	}
}
