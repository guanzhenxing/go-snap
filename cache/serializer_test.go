package cache

import (
	"encoding/gob"
	"reflect"
	"testing"
)

type testStruct struct {
	Name  string
	Value int
	Map   map[string]interface{}
	Array []int
}

func TestJSONSerializer(t *testing.T) {
	serializer := &JSONSerializer{}

	testCases := []struct {
		name  string
		value interface{}
	}{
		{"String", "test string"},
		{"Int", 42},
		{"Float", 3.14},
		{"Bool", true},
		{"Array", []int{1, 2, 3}},
		{"Map", map[string]string{"key": "value"}},
		{"Struct", testStruct{
			Name:  "test",
			Value: 123,
			Map:   map[string]interface{}{"foo": "bar"},
			Array: []int{4, 5, 6},
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 序列化
			data, err := serializer.Marshal(tc.value)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// 反序列化
			var result interface{}
			err = serializer.Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// 针对不同类型验证结果
			switch tc.name {
			case "String", "Bool":
				// JSON会保留字符串和布尔类型
				if result != tc.value {
					t.Fatalf("Expected %v, got %v", tc.value, result)
				}
			case "Int", "Float":
				// JSON会将整数转换为浮点数
				expectedFloat := float64(0)
				switch v := tc.value.(type) {
				case int:
					expectedFloat = float64(v)
				case float64:
					expectedFloat = v
				}
				numResult, ok := result.(float64)
				if !ok {
					t.Fatalf("Expected float64 type, got %T", result)
				}
				if numResult != expectedFloat {
					t.Fatalf("Expected %v, got %v", expectedFloat, numResult)
				}
			case "Array":
				// JSON数字数组会变成float64数组
				arr, ok := result.([]interface{})
				if !ok {
					t.Fatalf("Expected array type, got %T", result)
				}
				if len(arr) != 3 {
					t.Fatalf("Expected array length 3, got %d", len(arr))
				}
				// 验证值
				for i, v := range arr {
					expected := float64(i + 1)
					if v != expected {
						t.Fatalf("Expected %v at index %d, got %v", expected, i, v)
					}
				}
			case "Map", "Struct":
				// 只检查非空，因为JSON会改变值类型
				if result == nil {
					t.Fatalf("Expected non-nil result for %s", tc.name)
				}
			}
		})
	}

	// 测试反序列化到特定类型
	t.Run("UnmarshalToType", func(t *testing.T) {
		original := testStruct{
			Name:  "specific type",
			Value: 456,
			Map:   map[string]interface{}{"test": 123},
			Array: []int{7, 8, 9},
		}

		// 序列化
		data, err := serializer.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		// 反序列化到特定类型
		var result testStruct
		err = serializer.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// 验证字段
		if result.Name != original.Name {
			t.Fatalf("Expected Name %s, got %s", original.Name, result.Name)
		}
		if result.Value != original.Value {
			t.Fatalf("Expected Value %d, got %d", original.Value, result.Value)
		}
		// 验证Map和Array字段
		if len(result.Map) != len(original.Map) {
			t.Fatalf("Expected Map length %d, got %d", len(original.Map), len(result.Map))
		}
		if len(result.Array) != len(original.Array) {
			t.Fatalf("Expected Array length %d, got %d", len(original.Array), len(result.Array))
		}
	})
}

func TestGobSerializer(t *testing.T) {
	// 由于Gob序列化的特殊性，我们需要先注册测试类型
	gob.Register(testStruct{})
	gob.Register(map[string]string{})
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})

	serializer := &GobSerializer{}

	testCases := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"String", "test string", "test string"},
		{"Int", 42, int64(42)}, // Gob会将int转换为int64
		{"Float", 3.14, 3.14},
		{"Bool", true, true},
		{"Array", []int{1, 2, 3}, []int{1, 2, 3}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 序列化
			data, err := serializer.Marshal(tc.value)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// 为了正确反序列化，我们需要使用具体类型
			var result interface{}
			switch tc.name {
			case "String":
				var s string
				err = serializer.Unmarshal(data, &s)
				result = s
			case "Int":
				var i int64
				err = serializer.Unmarshal(data, &i)
				result = i
			case "Float":
				var f float64
				err = serializer.Unmarshal(data, &f)
				result = f
			case "Bool":
				var b bool
				err = serializer.Unmarshal(data, &b)
				result = b
			case "Array":
				var a []int
				err = serializer.Unmarshal(data, &a)
				result = a
			default:
				t.Skipf("Skipping %s test for Gob", tc.name)
				return
			}

			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// 验证结果
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Expected %v, got %v", tc.expected, result)
			}
		})
	}

	// 测试Map类型
	t.Run("Map", func(t *testing.T) {
		original := map[string]string{"key": "value"}
		data, err := serializer.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result map[string]string
		err = serializer.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if !reflect.DeepEqual(result, original) {
			t.Fatalf("Expected %v, got %v", original, result)
		}
	})

	// 测试结构体类型
	t.Run("Struct", func(t *testing.T) {
		original := testStruct{
			Name:  "test",
			Value: 123,
			Map:   map[string]interface{}{"foo": "bar"},
			Array: []int{4, 5, 6},
		}
		data, err := serializer.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result testStruct
		err = serializer.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result.Name != original.Name || result.Value != original.Value {
			t.Fatalf("Expected {%s, %d}, got {%s, %d}",
				original.Name, original.Value, result.Name, result.Value)
		}
	})
}

func TestDefaultSerializer(t *testing.T) {
	serializer := DefaultSerializer()

	// 验证默认是JSON序列化器
	if _, ok := serializer.(*JSONSerializer); !ok {
		t.Fatalf("Expected default serializer to be JSONSerializer, got %T", serializer)
	}

	// 简单测试序列化/反序列化
	value := "test default serializer"
	data, err := serializer.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result string
	err = serializer.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result != value {
		t.Fatalf("Expected %v, got %v", value, result)
	}
}
