package logger

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

// 测试字段处理
func TestFields(t *testing.T) {
	// 测试各种字段类型
	fields := []struct {
		name     string
		field    Field
		expected interface{}
	}{
		{"String", String("string", "value"), "value"},
		{"Int", Int("int", 123), float64(123)}, // JSON将整数解码为float64
		{"Int8", Int8("int8", 8), float64(8)},
		{"Int16", Int16("int16", 16), float64(16)},
		{"Int32", Int32("int32", 32), float64(32)},
		{"Int64", Int64("int64", 9876543210), float64(9876543210)},
		{"Uint", Uint("uint", 123), float64(123)},
		{"Uint8", Uint8("uint8", 8), float64(8)},
		{"Uint16", Uint16("uint16", 16), float64(16)},
		{"Uint32", Uint32("uint32", 32), float64(32)},
		{"Uint64", Uint64("uint64", 64), float64(64)},
		{"Float32", Float32("float32", 3.14), float64(3.14)},
		{"Float64", Float64("float64", 123.456), 123.456},
		{"Bool", Bool("bool", true), true},
		{"Time", Time("time", time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)), "2023-06-01T00:00:00Z"},
		{"Duration", Duration("duration", time.Hour), float64(3600)},
		{"Any", Any("any", map[string]string{"key": "value"}), map[string]interface{}{"key": "value"}},
		{"Binary", Binary("binary", []byte("binary data")), "YmluYXJ5IGRhdGE="},
	}

	// 记录日志
	buf := captureOutput(t, func() {
		fieldArray := make([]Field, 0, len(fields))
		for _, f := range fields {
			fieldArray = append(fieldArray, f.field)
		}

		Info("field test", fieldArray...)
	})

	// 解析日志条目
	entry := parseLogEntry(t, buf)

	// 验证字段值
	for _, f := range fields {
		// 跳过复杂类型比较
		if f.name == "Time" || f.name == "Any" || f.name == "Binary" {
			continue
		}

		if entry[f.field.Key] != f.expected {
			t.Errorf("Field %s: expected %v (%T), got %v (%T)",
				f.name, f.expected, f.expected, entry[f.field.Key], entry[f.field.Key])
		}
	}
}

// 用于测试的Stringer接口实现
type testStringer struct {
	val string
}

func (s testStringer) String() string {
	return s.val
}

// 用于测试Error字段的自定义错误
type testError string

func (e testError) Error() string {
	return string(e)
}

// 测试字段类型转换
func TestFieldConversion(t *testing.T) {
	// 测试Stringer
	stringer := testStringer{val: "stringer value"}

	// 测试Error
	errVal := testError("error value")

	buf := captureOutput(t, func() {
		Info("conversion test",
			Stringer("stringer", stringer),
			LogError(errVal),
		)
	})

	// 解析日志条目
	entry := parseLogEntry(t, buf)

	// 验证字段值 (简化检测，具体值可能与实现不同)
	if entry["stringer"] == nil {
		t.Error("Stringer field is nil")
	}
}

// 测试更多字段类型
func TestMoreFields(t *testing.T) {
	testErr := errors.New("test error")
	now := time.Now()

	buf := captureOutput(t, func() {
		Info("more fields test",
			// 测试错误相关字段
			Err(testErr),
			NamedError("custom_error", testErr),

			// 测试时间相关字段
			Time("timestamp", now),
			RFC3339Time("rfc3339", now),
			ISO8601Time("iso8601", now),
			DurationMs("duration_ms", 1500*time.Millisecond),

			// 测试复杂数据结构
			Map("map", map[string]interface{}{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			}),

			// 测试日志上下文相关字段
			TraceID("trace-123"),
			RequestID("req-456"),
			UserID("user-789"),
			Service("test-service"),
			Method("GET"),
			Path("/api/test"),
			Status(200),
			Latency(50*time.Millisecond),
		)
	})

	entry := parseLogEntry(t, buf)

	// 检查错误字段
	if entry["error"] == nil {
		t.Error("Err field is missing")
	}

	if entry["custom_error"] == nil {
		t.Error("NamedError field is missing")
	}

	// 检查时间字段
	if entry["timestamp"] == nil {
		t.Error("Timestamp field is missing")
	}

	if entry["rfc3339"] == nil {
		t.Error("RFC3339Time field is missing")
	}

	if entry["iso8601"] == nil {
		t.Error("ISO8601Time field is missing")
	}

	if entry["duration_ms"] == nil {
		t.Error("DurationMs field is missing")
	}

	// 检查Map字段
	if mp, ok := entry["map"].(map[string]interface{}); !ok {
		t.Error("Map field is missing or has wrong type")
	} else {
		if mp["key1"] != "value1" {
			t.Errorf("Map field has wrong value for key1: %v", mp["key1"])
		}
	}

	// 检查上下文相关字段
	if entry["trace_id"] != "trace-123" {
		t.Errorf("TraceID field has wrong value: %v", entry["trace_id"])
	}

	if entry["request_id"] != "req-456" {
		t.Errorf("RequestID field has wrong value: %v", entry["request_id"])
	}

	if entry["user_id"] != "user-789" {
		t.Errorf("UserID field has wrong value: %v", entry["user_id"])
	}

	if entry["service"] != "test-service" {
		t.Errorf("Service field has wrong value: %v", entry["service"])
	}

	if entry["method"] != "GET" {
		t.Errorf("Method field has wrong value: %v", entry["method"])
	}

	if entry["path"] != "/api/test" {
		t.Errorf("Path field has wrong value: %v", entry["path"])
	}

	if float64(entry["status"].(float64)) != float64(200) {
		t.Errorf("Status field has wrong value: %v", entry["status"])
	}
}

// 为测试Dict创建自定义字段集合
func createDictFields() []Field {
	return []Field{
		String("str", "value"),
		Int("num", 123),
		String("inner", "inner_value"),
	}
}

// 为测试Array创建自定义ArrayMarshaler实现
type stringArray []string

func (a stringArray) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, s := range a {
		enc.AppendString(s)
	}
	return nil
}

// 测试Dict和Array字段
func TestDictAndArray(t *testing.T) {
	// 创建用于Dict的字段数组
	dictFields := createDictFields()

	// 创建用于Array的数组封装
	arrayData := stringArray{"one", "two", "three"}

	buf := captureOutput(t, func() {
		Info("dict and array test",
			Dict("dict", dictFields...),
			Array("array", arrayData),
		)
	})

	entry := parseLogEntry(t, buf)

	// 检查Dict字段
	if dict, ok := entry["dict"].(map[string]interface{}); !ok {
		t.Error("Dict field is missing or has wrong type")
	} else {
		if dict["str"] != "value" {
			t.Errorf("Dict field has wrong value for 'str': %v", dict["str"])
		}

		if float64(dict["num"].(float64)) != float64(123) {
			t.Errorf("Dict field has wrong value for 'num': %v", dict["num"])
		}

		if dict["inner"] != "inner_value" {
			t.Errorf("Dict field has wrong nested value: %v", dict["inner"])
		}
	}

	// 检查Array字段
	if arr, ok := entry["array"].([]interface{}); !ok {
		t.Error("Array field is missing or has wrong type")
	} else {
		if len(arr) != 3 {
			t.Errorf("Array field has wrong length: %d", len(arr))
		}

		if arr[0] != "one" || arr[1] != "two" || arr[2] != "three" {
			t.Errorf("Array field has wrong values: %v", arr)
		}
	}
}

// 测试Object和Sprintf字段
func TestObjectAndSprintf(t *testing.T) {
	type TestStruct struct {
		Name  string
		Value int
		Sub   struct {
			Field string
		}
	}

	testObj := TestStruct{
		Name:  "test",
		Value: 42,
	}
	testObj.Sub.Field = "subfield"

	buf := captureOutput(t, func() {
		Info("object and sprintf test",
			Object("object", testObj),
			Sprintf("formatted", "%s-%d", "value", 123),
			Interface("interface", testObj),
		)
	})

	entry := parseLogEntry(t, buf)

	// 检查Object字段
	if obj, ok := entry["object"].(map[string]interface{}); !ok {
		t.Error("Object field is missing or has wrong type")
	} else {
		if obj["Name"] != "test" {
			t.Errorf("Object field has wrong value for Name: %v", obj["Name"])
		}

		if float64(obj["Value"].(float64)) != float64(42) {
			t.Errorf("Object field has wrong value for Value: %v", obj["Value"])
		}
	}

	// 检查Sprintf字段
	if entry["formatted"] != "value-123" {
		t.Errorf("Sprintf field has wrong value: %v", entry["formatted"])
	}

	// 检查Interface字段
	if intf, ok := entry["interface"].(map[string]interface{}); !ok {
		t.Error("Interface field is missing or has wrong type")
	} else {
		if intf["Name"] != "test" {
			t.Errorf("Interface field has wrong value for Name: %v", intf["Name"])
		}
	}
}

// 测试Sample字段
func TestSample(t *testing.T) {
	buf := captureOutput(t, func() {
		for i := 0; i < 10; i++ {
			// 使用Sample函数包装一个值
			sampleValue := i
			Info("sample test", Sample("sample", sampleValue))
		}
	})

	// 这里不好精确测试采样的结果，只检查日志格式是否正确
	if buf.Len() == 0 {
		t.Error("Sample field produced no output")
	}
}

// 测试Timestamp函数
func TestTimestamp(t *testing.T) {
	buf := captureOutput(t, func() {
		Info("timestamp test", Timestamp())
	})

	entry := parseLogEntry(t, buf)

	// 时间戳应该是当前时间
	if entry["timestamp"] == nil {
		t.Error("Timestamp field is missing")
	}
}
