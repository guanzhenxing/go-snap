package errors

import (
	"fmt"
	"strings"
	"testing"
)

// 测试添加单个上下文值
func TestWithContext(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")
	contextErr := WithContext(baseErr, "request_id", "abc-123")

	// 检查错误消息
	if contextErr.Error() != baseErr.Error() {
		t.Errorf("错误消息不匹配: 期望 %q, 得到 %q", baseErr.Error(), contextErr.Error())
	}

	// 获取上下文值
	value, exists := GetContext(contextErr, "request_id")
	if !exists {
		t.Error("未能从错误中获取请求ID上下文")
	}

	strValue, ok := value.(string)
	if !ok {
		t.Errorf("上下文值类型错误: 期望 string, 得到 %T", value)
	}

	if strValue != "abc-123" {
		t.Errorf("上下文值不匹配: 期望 %q, 得到 %q", "abc-123", strValue)
	}

	// 检查不存在的键
	_, exists = GetContext(contextErr, "non_existent")
	if exists {
		t.Error("不应存在非存在的上下文键")
	}

	// nil错误处理
	nilResult := WithContext(nil, "key", "value")
	if nilResult != nil {
		t.Error("WithContext(nil, ...) 应该返回 nil")
	}
}

// 测试添加多个上下文值
func TestWithContextMap(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")
	contextMap := map[string]interface{}{
		"request_id": "abc-123",
		"user_id":    123,
		"timestamp":  "2023-04-05T12:34:56Z",
	}

	contextErr := WithContextMap(baseErr, contextMap)

	// 检查所有上下文值
	allContext := GetAllContext(contextErr)
	if len(allContext) != len(contextMap) {
		t.Errorf("上下文值数量不匹配: 期望 %d, 得到 %d", len(contextMap), len(allContext))
	}

	// 逐个检查值
	for k, expected := range contextMap {
		actual, exists := allContext[k]
		if !exists {
			t.Errorf("上下文中应包含键 %q", k)
			continue
		}

		if actual != expected {
			t.Errorf("键 %q 的值不匹配: 期望 %v, 得到 %v", k, expected, actual)
		}
	}

	// nil错误处理
	nilResult := WithContextMap(nil, contextMap)
	if nilResult != nil {
		t.Error("WithContextMap(nil, ...) 应该返回 nil")
	}

	// 空映射处理
	emptyResult := WithContextMap(baseErr, nil)
	if emptyResult != baseErr {
		t.Error("WithContextMap(err, nil) 应该返回原始错误")
	}
}

// 测试上下文错误链
func TestContextErrorChain(t *testing.T) {
	// 创建带上下文的错误链
	baseErr := fmt.Errorf("基础错误")
	firstContext := WithContext(baseErr, "level", "base")
	secondContext := WithContext(firstContext, "request_id", "abc-123")
	thirdContext := WithContext(secondContext, "user_id", 456)

	// 检查可以获取所有上下文值
	level, exists := GetContext(thirdContext, "level")
	if !exists || level != "base" {
		t.Errorf("应能从链中获取第一层上下文，得到: %v", level)
	}

	requestID, exists := GetContext(thirdContext, "request_id")
	if !exists || requestID != "abc-123" {
		t.Errorf("应能从链中获取第二层上下文，得到: %v", requestID)
	}

	userID, exists := GetContext(thirdContext, "user_id")
	if !exists || userID != 456 {
		t.Errorf("应能从链中获取第三层上下文，得到: %v", userID)
	}

	// 获取所有上下文
	allContext := GetAllContext(thirdContext)
	if len(allContext) != 3 {
		t.Errorf("应包含所有三个上下文键，但得到 %d 个", len(allContext))
	}
}

// 测试上下文与包装函数的交互
func TestContextWithWrapping(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")

	// 先添加上下文，再包装
	errWithContext := WithContext(baseErr, "request_id", "abc-123")
	wrappedErr := Wrap(errWithContext, "包装错误")

	// 应该能从包装错误中获取上下文
	requestID, exists := GetContext(wrappedErr, "request_id")
	if !exists || requestID != "abc-123" {
		t.Errorf("应能从包装错误中获取上下文，得到: %v", requestID)
	}

	// 先包装，再添加上下文
	wrappedFirst := Wrap(baseErr, "包装错误")
	contextAfterWrap := WithContext(wrappedFirst, "user_id", 789)

	// 应该能从上下文后的包装错误中获取上下文
	userID, exists := GetContext(contextAfterWrap, "user_id")
	if !exists || userID != 789 {
		t.Errorf("应能从上下文后的包装错误中获取上下文，得到: %v", userID)
	}
}

// 测试错误格式化输出中包含上下文
func TestFormatWithContext(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")

	// 添加多个上下文
	contextErr := WithContext(baseErr, "request_id", "abc-123")
	contextErr = WithContext(contextErr, "user_id", 456)

	// 测试详细格式化输出（带+标志）
	formatted := fmt.Sprintf("%+v", contextErr)

	// 检查输出中包含上下文信息
	if !strings.Contains(formatted, "request_id: abc-123") {
		t.Errorf("格式化输出应包含request_id上下文，但得到: %s", formatted)
	}

	if !strings.Contains(formatted, "user_id: 456") {
		t.Errorf("格式化输出应包含user_id上下文，但得到: %s", formatted)
	}
}

// 测试复杂数据类型的上下文
func TestComplexContextTypes(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")

	// 添加各种类型的上下文数据
	contextMap := map[string]interface{}{
		"string_value":  "字符串",
		"int_value":     42,
		"boolean_value": true,
		"float_value":   3.14,
		"nil_value":     nil,
		"slice_value":   []string{"a", "b", "c"},
		"map_value":     map[string]int{"one": 1, "two": 2},
		"struct_value":  struct{ Name string }{"测试"},
	}

	contextErr := WithContextMap(baseErr, contextMap)

	// 获取并验证所有上下文值
	allContext := GetAllContext(contextErr)

	for key, expected := range contextMap {
		actual, exists := allContext[key]
		if !exists {
			t.Errorf("上下文中应包含键 %q", key)
			continue
		}

		// 特殊处理slice、map和struct的比较
		switch key {
		case "slice_value", "map_value", "struct_value":
			// 仅检查类型，因为复杂类型的值相等比较较为复杂
			expectedType := fmt.Sprintf("%T", expected)
			actualType := fmt.Sprintf("%T", actual)
			if expectedType != actualType {
				t.Errorf("键 %q 的类型不匹配: 期望 %s, 得到 %s", key, expectedType, actualType)
			}
		default:
			// 基本类型可以直接比较
			if actual != expected {
				t.Errorf("键 %q 的值不匹配: 期望 %v, 得到 %v", key, expected, actual)
			}
		}
	}
}

// 以下是示例代码

// ExampleWithContext 演示如何使用错误上下文功能为错误添加额外信息
func ExampleWithContext() {
	// 模拟从数据库查询失败的错误
	baseErr := fmt.Errorf("数据库连接失败")

	// 添加请求ID作为上下文
	contextErr := WithContext(baseErr, "request_id", "req-123456")

	// 添加更多上下文信息
	contextErr = WithContext(contextErr, "user_id", 42)
	contextErr = WithContext(contextErr, "operation", "查询用户")

	// 包装错误并提供更多上下文
	finalErr := Wrap(contextErr, "处理用户请求失败")

	// 从错误中获取上下文信息
	requestID, _ := GetContext(finalErr, "request_id")
	fmt.Printf("请求ID: %s\n", requestID)

	userID, _ := GetContext(finalErr, "user_id")
	fmt.Printf("用户ID: %d\n", userID)

	// 获取所有上下文信息
	allContext := GetAllContext(finalErr)
	fmt.Printf("所有上下文: %v\n", allContext)

	// 使用格式化输出错误的详细信息（包括上下文）
	fmt.Printf("详细错误: %+v\n", finalErr)
}

// Example_errorContext_HTTP 演示如何使用错误上下文与HTTP处理程序结合
func Example_errorContext_HTTP() {
	// 模拟HTTP请求处理
	handleHTTPRequest()
}

// handleHTTPRequest 模拟HTTP请求处理函数
func handleHTTPRequest() {
	// 模拟请求信息
	requestID := "req-789012"
	userID := 123

	// 调用业务逻辑
	err := mockBusinessLogic(requestID, userID)
	if err != nil {
		// 在HTTP处理程序中，我们可以从错误中提取相关信息
		code := -1
		if c, ok := err.(interface{ Code() int }); ok {
			code = c.Code()
		}

		// 从错误中提取上下文信息
		reqID, _ := GetContext(err, "request_id")
		uid, _ := GetContext(err, "user_id")

		// 输出错误响应（在实际应用中，这会发送HTTP响应）
		fmt.Printf("错误响应: {\"code\":%d, \"message\":\"%s\", \"request_id\":\"%v\", \"user_id\":%v}\n",
			code, err.Error(), reqID, uid)
		return
	}

	fmt.Println("请求处理成功")
}

// mockBusinessLogic 模拟业务逻辑
func mockBusinessLogic(requestID string, userID int) error {
	// 模拟数据库操作错误
	dbErr := fmt.Errorf("用户数据未找到")
	_ = dbErr // 避免未使用警告

	// 创建带有错误码和上下文的错误
	err := NewWithCode(404, "用户不存在")

	// 添加请求上下文
	err = WithContextMap(err, map[string]interface{}{
		"request_id": requestID,
		"user_id":    userID,
		"timestamp":  "2023-04-05T12:34:56Z",
	})

	// 包装原始错误
	return Wrap(err, "获取用户数据失败")
}
