package logger

import (
	"context"
)

// 上下文键常量
const (
	TraceIDKey   = "trace_id"
	RequestIDKey = "request_id"
)

// 性能度量键常量
const (
	LatencyKey = "latency_ms"
)

// 日志上下文键
const (
	// 基本上下文键
	LogUserIDKey     = "user_id"
	LogSessionIDKey  = "session_id"
	LogClientIPKey   = "client_ip"
	LogDeviceIDKey   = "device_id"
	LogAppVersionKey = "app_version"

	// 性能度量键
	LogLatencyKey = "latency_ms"
	LogSizeKey    = "size_bytes"
	LogCountKey   = "count"
)

// ContextLogger 日志上下文
type ContextLogger struct {
	TraceID   string
	RequestID string
	UserID    string
	SessionID string
	ClientIP  string
	Path      string
	Method    string
	Extra     map[string]interface{}
}

// NewContextLogger 创建新的日志上下文
func NewContextLogger() *ContextLogger {
	return &ContextLogger{
		Extra: make(map[string]interface{}),
	}
}

// WithTraceID 设置跟踪ID
func (lc *ContextLogger) WithTraceID(traceID string) *ContextLogger {
	lc.TraceID = traceID
	return lc
}

// WithRequestID 设置请求ID
func (lc *ContextLogger) WithRequestID(requestID string) *ContextLogger {
	lc.RequestID = requestID
	return lc
}

// WithUserID 设置用户ID
func (lc *ContextLogger) WithUserID(userID string) *ContextLogger {
	lc.UserID = userID
	return lc
}

// WithSessionID 设置会话ID
func (lc *ContextLogger) WithSessionID(sessionID string) *ContextLogger {
	lc.SessionID = sessionID
	return lc
}

// WithClientIP 设置客户端IP
func (lc *ContextLogger) WithClientIP(clientIP string) *ContextLogger {
	lc.ClientIP = clientIP
	return lc
}

// WithPath 设置请求路径
func (lc *ContextLogger) WithPath(path string) *ContextLogger {
	lc.Path = path
	return lc
}

// WithMethod 设置请求方法
func (lc *ContextLogger) WithMethod(method string) *ContextLogger {
	lc.Method = method
	return lc
}

// WithExtra 添加额外字段
func (lc *ContextLogger) WithExtra(key string, value interface{}) *ContextLogger {
	lc.Extra[key] = value
	return lc
}

// ToContext 将LogContext转换为context.Context
func (lc *ContextLogger) ToContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if lc.TraceID != "" {
		ctx = context.WithValue(ctx, TraceIDKey, lc.TraceID)
	}

	if lc.RequestID != "" {
		ctx = context.WithValue(ctx, RequestIDKey, lc.RequestID)
	}

	if lc.UserID != "" {
		ctx = context.WithValue(ctx, LogUserIDKey, lc.UserID)
	}

	if lc.SessionID != "" {
		ctx = context.WithValue(ctx, LogSessionIDKey, lc.SessionID)
	}

	return ctx
}

// ToFields 将LogContext转换为日志字段
func (lc *ContextLogger) ToFields() []Field {
	fields := make([]Field, 0, 8+len(lc.Extra))

	if lc.TraceID != "" {
		fields = append(fields, String(TraceIDKey, lc.TraceID))
	}

	if lc.RequestID != "" {
		fields = append(fields, String(RequestIDKey, lc.RequestID))
	}

	if lc.UserID != "" {
		fields = append(fields, String(LogUserIDKey, lc.UserID))
	}

	if lc.SessionID != "" {
		fields = append(fields, String(LogSessionIDKey, lc.SessionID))
	}

	if lc.ClientIP != "" {
		fields = append(fields, String(LogClientIPKey, lc.ClientIP))
	}

	if lc.Path != "" {
		fields = append(fields, String("path", lc.Path))
	}

	if lc.Method != "" {
		fields = append(fields, String("method", lc.Method))
	}

	// 添加额外字段
	for k, v := range lc.Extra {
		fields = append(fields, Any(k, v))
	}

	return fields
}

// NewLogContextLoggerWithCtx 创建一个带LogContext的日志记录器（全局方法）
func NewLogContextLoggerWithCtx(lctx *ContextLogger) Logger {
	checkGlobalLogger()

	if zl, ok := globalLogger.(*zapLogger); ok {
		return zl.WithLogContext(lctx)
	}

	return globalLogger.With(lctx.ToFields()...)
}

// LoggerFromContext 从context.Context提取LogContext
func LoggerFromContext(ctx context.Context) *ContextLogger {
	if ctx == nil {
		return NewContextLogger()
	}

	lc := NewContextLogger()

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		lc.TraceID = traceID
	}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		lc.RequestID = requestID
	}

	if userID, ok := ctx.Value(LogUserIDKey).(string); ok {
		lc.UserID = userID
	}

	if sessionID, ok := ctx.Value(LogSessionIDKey).(string); ok {
		lc.SessionID = sessionID
	}

	return lc
}
