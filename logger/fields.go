package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Field 封装zapcore.Field
type Field = zapcore.Field

// 字段创建函数
var (
	Skip        = zap.Skip
	Binary      = zap.Binary
	Bool        = zap.Bool
	Boolp       = zap.Boolp
	ByteString  = zap.ByteString
	Complex128  = zap.Complex128
	Complex128p = zap.Complex128p
	Complex64   = zap.Complex64
	Complex64p  = zap.Complex64p
	Float64     = zap.Float64
	Float64p    = zap.Float64p
	Float32     = zap.Float32
	Float32p    = zap.Float32p
	Int         = zap.Int
	Intp        = zap.Intp
	Int64       = zap.Int64
	Int64p      = zap.Int64p
	Int32       = zap.Int32
	Int32p      = zap.Int32p
	Int16       = zap.Int16
	Int16p      = zap.Int16p
	Int8        = zap.Int8
	Int8p       = zap.Int8p
	String      = zap.String
	Stringp     = zap.Stringp
	Uint        = zap.Uint
	Uintp       = zap.Uintp
	Uint64      = zap.Uint64
	Uint64p     = zap.Uint64p
	Uint32      = zap.Uint32
	Uint32p     = zap.Uint32p
	Uint16      = zap.Uint16
	Uint16p     = zap.Uint16p
	Uint8       = zap.Uint8
	Uint8p      = zap.Uint8p
	Uintptr     = zap.Uintptr
	Uintptrp    = zap.Uintptrp
	Reflect     = zap.Reflect
	Namespace   = zap.Namespace
	Stringer    = zap.Stringer
	Time        = zap.Time
	Timep       = zap.Timep
	Stack       = zap.Stack
	StackSkip   = zap.StackSkip
	Duration    = zap.Duration
	Durationp   = zap.Durationp
	Any         = zap.Any
)

// LogError 记录错误，包含错误信息和堆栈
func LogError(err error) Field {
	if err == nil {
		return Skip()
	}
	return zap.Error(err)
}

// Dict 创建嵌套字段
func Dict(key string, fields ...Field) Field {
	return zap.Dict(key, fields...)
}

// Array 创建数组字段
func Array(key string, arr zapcore.ArrayMarshaler) Field {
	return zap.Array(key, arr)
}

// Err 创建错误字段（兼容别名）
func Err(err error) Field {
	return LogError(err)
}

// errors 别名
var (
	ErrorField = zap.Error // 重命名以避免冲突
)

// NamedError 创建带名称的错误字段
func NamedError(key string, err error) Field {
	if err == nil {
		return Skip()
	}
	return zap.NamedError(key, err)
}

// Timestamp 记录当前时间戳
func Timestamp() Field {
	return Time("timestamp", time.Now())
}

// RFC3339Time 以RFC3339格式记录时间
func RFC3339Time(key string, t time.Time) Field {
	return String(key, t.Format(time.RFC3339))
}

// ISO8601Time 以ISO8601格式记录时间
func ISO8601Time(key string, t time.Time) Field {
	return String(key, t.Format("2006-01-02T15:04:05.000Z07:00"))
}

// Human readable duration formats
func DurationMs(key string, d time.Duration) Field {
	return Int64(key, d.Milliseconds())
}

// Map 将map转换为字段
func Map(key string, val map[string]interface{}) Field {
	return zap.Reflect(key, val)
}

// Interface 接口别名，与Any相同
func Interface(key string, val interface{}) Field {
	return Any(key, val)
}

// Object 记录对象，与Any相同但名称更清晰
func Object(key string, val interface{}) Field {
	return Any(key, val)
}

// Sprintf 格式化字符串
func Sprintf(key, format string, args ...interface{}) Field {
	return String(key, fmt.Sprintf(format, args...))
}

// 采样相关字段
func Sample(key string, val interface{}) Field {
	return zap.Any(key, val)
}

// 追踪相关字段
func TraceID(id string) Field {
	return String("trace_id", id)
}

func RequestID(id string) Field {
	return String("request_id", id)
}

// 业务相关字段
func UserID(id string) Field {
	return String("user_id", id)
}

func Service(name string) Field {
	return String("service", name)
}

func Method(name string) Field {
	return String("method", name)
}

func Path(path string) Field {
	return String("path", path)
}

func Status(code int) Field {
	return Int("status", code)
}

func Latency(duration time.Duration) Field {
	return Duration("latency", duration)
}
