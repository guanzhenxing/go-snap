package errors

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// 测试错误格式化 %s
func TestErrorFormatString(t *testing.T) {
	const errorMessage = "测试错误"
	err := New(errorMessage)

	formatted := fmt.Sprintf("%s", err)
	if formatted != errorMessage {
		t.Errorf("使用%%s格式化错误应该返回错误消息，期望 %q, 得到 %q", errorMessage, formatted)
	}
}

// 测试错误格式化 %q
func TestErrorFormatQuoted(t *testing.T) {
	const errorMessage = "测试错误"
	err := New(errorMessage)

	formatted := fmt.Sprintf("%q", err)
	expected := strconv.Quote(errorMessage)
	if formatted != expected {
		t.Errorf("使用%%q格式化错误应该返回引用的错误消息，期望 %q, 得到 %q", expected, formatted)
	}
}

// 测试错误格式化 %v
func TestErrorFormatValue(t *testing.T) {
	const errorMessage = "测试错误"
	err := New(errorMessage)

	formatted := fmt.Sprintf("%v", err)
	if formatted != errorMessage {
		t.Errorf("使用%%v格式化错误应该返回错误消息，期望 %q, 得到 %q", errorMessage, formatted)
	}
}

// 测试错误格式化 %+v
func TestErrorFormatDetailedValue(t *testing.T) {
	// 保存原始配置
	originalMode := DefaultStackCaptureMode
	DefaultStackCaptureMode = StackCaptureModeImmediate // 确保使用立即模式进行测试
	defer func() {
		DefaultStackCaptureMode = originalMode
	}()

	const errorMessage = "测试错误"
	err := New(errorMessage)

	formatted := fmt.Sprintf("%+v", err)
	if !strings.Contains(formatted, errorMessage) {
		t.Errorf("使用%%+v格式化错误应该包含错误消息，但得到: %q", formatted)
	}

	// 检查是否包含调用位置
	if !strings.Contains(formatted, "/errors/format_test.go:") {
		t.Errorf("使用%%+v格式化错误应该包含堆栈信息，但得到: %q", formatted)
	}
}

// 测试错误格式化 %-v
func TestErrorFormatMinusV(t *testing.T) {
	const errorMessage = "测试错误"
	err := New(errorMessage)

	formatted := fmt.Sprintf("%-v", err)
	if !strings.Contains(formatted, errorMessage) {
		t.Errorf("使用%%-v格式化错误应该包含错误消息，但得到: %q", formatted)
	}
}

// 测试带嵌套错误的格式化
func TestErrorFormatNested(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")
	wrappedErr := Wrap(baseErr, "包装错误")

	// 标准格式应该只显示最外层错误
	simple := fmt.Sprintf("%s", wrappedErr)
	if simple != "包装错误" {
		t.Errorf("嵌套错误的简单格式化错误，期望 %q, 得到 %q", "包装错误", simple)
	}

	// 详细格式应该包含错误链
	detailed := fmt.Sprintf("%+v", wrappedErr)
	if !strings.Contains(detailed, "包装错误") || !strings.Contains(detailed, "基础错误") {
		t.Errorf("嵌套错误的详细格式化应该包含整个错误链，但得到: %q", detailed)
	}
}

// 测试带错误码的格式化
func TestErrorFormatWithCode(t *testing.T) {
	const (
		code    = 404
		message = "资源未找到"
	)

	// 注册错误码
	RegisterErrorCode(code, 404, "Resource not found", "")
	err := NewWithCode(code, message)

	// 详细格式应该包含错误码
	detailed := fmt.Sprintf("%+v", err)
	if !strings.Contains(detailed, fmt.Sprintf("(%d)", code)) {
		t.Errorf("带错误码的详细格式化应该包含错误码，但得到: %q", detailed)
	}
}

// 测试formatDetailed无堆栈模式
func TestFormatDetailedNoTrace(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")
	err1 := Wrap(baseErr, "错误1")
	err2 := Wrap(err1, "错误2")

	var str strings.Builder
	formatDetailed(err2, &str, false)

	result := str.String()

	// 不显示堆栈时应该只有第一个错误
	if strings.Contains(result, "错误1") || strings.Contains(result, "基础错误") {
		t.Errorf("禁用堆栈的formatDetailed不应包含完整错误链，但得到: %s", result)
	}

	if !strings.Contains(result, "错误2") {
		t.Errorf("formatDetailed应该包含最外层错误消息，但得到: %s", result)
	}
}

// 测试空错误的格式化
func TestEmptyErrorFormat(t *testing.T) {
	var str strings.Builder
	formatDetailed(nil, &str, true)

	result := str.String()
	if result != "" {
		t.Errorf("格式化nil错误应该返回空字符串，但得到: %q", result)
	}
}

// 测试一个非常深的错误链格式化
func TestDeepErrorChainFormat(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")
	err := baseErr

	// 创建一个深度为10的错误链
	for i := 1; i <= 10; i++ {
		err = Wrap(err, fmt.Sprintf("错误层级%d", i))
	}

	// 详细格式应该包含所有错误
	detailed := fmt.Sprintf("%+v", err)

	// 检查最外层和最内层错误
	if !strings.Contains(detailed, "错误层级10") || !strings.Contains(detailed, "基础错误") {
		t.Errorf("深层错误链的详细格式化应该包含内外层错误，但得到: %q", detailed)
	}

	// 检查分隔符数量
	separatorCount := strings.Count(detailed, ";")
	if separatorCount < 9 { // 10个错误应该有至少9个分隔符
		t.Errorf("深层错误链的详细格式化应该包含多个分隔符，但只找到%d个", separatorCount)
	}
}

// 测试extractErrorFormatInfo函数的自定义错误码支持
func TestExtractErrorFormatInfoWithCustomCode(t *testing.T) {
	const (
		code    = 502
		message = "网关错误"
	)

	// 注册错误码
	RegisterErrorCode(code, 502, message, "")

	// 创建带错误码的错误
	err := NewWithCode(code, "自定义错误")

	// 提取格式化信息
	info := extractErrorFormatInfo(err)

	// 验证错误码和消息
	if info.code != code {
		t.Errorf("提取的错误码错误: 期望 %d, 得到 %d", code, info.code)
	}

	// 注意：message实际上是从错误码注册表中获取的完整消息，它包含错误码前缀
	expected := fmt.Sprintf("%d: %s", code, message)
	if info.message != expected {
		t.Errorf("提取的错误消息错误, 期望 %q, 得到 %q", expected, info.message)
	}
}

// 测试未注册错误码的情况
func TestUnregisteredCodeFormat(t *testing.T) {
	// 使用一个肯定未注册的错误码
	const unregisteredCode = 999999

	err := &contextualError{
		msg:  "未注册的错误码",
		code: unregisteredCode,
	}

	// 提取格式化信息
	info := extractErrorFormatInfo(err)

	// 应该使用UnknownError的信息
	if info.code != unregisteredCode {
		t.Errorf("未注册错误码应该保留原始值: 期望 %d, 得到 %d", unregisteredCode, info.code)
	}
}
