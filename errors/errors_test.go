package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// =================================================
// 基本错误功能测试
// =================================================

// 测试创建基本错误
func TestNew(t *testing.T) {
	const errorMessage = "测试错误消息"

	err := New(errorMessage)

	if err == nil {
		t.Fatal("New应该返回一个错误，但返回了nil")
	}

	if err.Error() != errorMessage {
		t.Errorf("错误消息错误: 期望 %q, 得到 %q", errorMessage, err.Error())
	}

	// 检查堆栈跟踪
	if st, ok := err.(StackTracer); !ok || st.StackTrace() == nil {
		t.Error("New返回的错误应该包含堆栈跟踪")
	}
}

// 测试使用格式创建错误
func TestErrorf(t *testing.T) {
	const (
		format = "错误 %d: %s"
		code   = 123
		detail = "详细信息"
	)

	expectedMessage := fmt.Sprintf(format, code, detail)
	err := Errorf(format, code, detail)

	if err == nil {
		t.Fatal("Errorf应该返回一个错误，但返回了nil")
	}

	if err.Error() != expectedMessage {
		t.Errorf("错误消息错误: 期望 %q, 得到 %q", expectedMessage, err.Error())
	}

	// 检查堆栈跟踪
	if st, ok := err.(StackTracer); !ok || st.StackTrace() == nil {
		t.Error("Errorf返回的错误应该包含堆栈跟踪")
	}
}

// 测试添加堆栈跟踪
func TestWithStack(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")
	err := WithStack(baseErr)

	if err == nil {
		t.Fatal("WithStack应该返回一个错误，但返回了nil")
	}

	if err.Error() != baseErr.Error() {
		t.Errorf("错误消息错误: 期望 %q, 得到 %q", baseErr.Error(), err.Error())
	}

	// 检查堆栈跟踪
	if st, ok := err.(StackTracer); !ok || st.StackTrace() == nil {
		t.Error("WithStack返回的错误应该包含堆栈跟踪")
	}

	// 检查nil错误处理
	if WithStack(nil) != nil {
		t.Error("WithStack(nil)应该返回nil")
	}
}

// 测试包装错误
func TestWrap(t *testing.T) {
	const (
		baseErrorMsg = "基础错误"
		wrapMsg      = "包装消息"
	)

	baseErr := fmt.Errorf(baseErrorMsg)
	err := Wrap(baseErr, wrapMsg)

	if err == nil {
		t.Fatal("Wrap应该返回一个错误，但返回了nil")
	}

	if err.Error() != wrapMsg {
		t.Errorf("错误消息错误: 期望 %q, 得到 %q", wrapMsg, err.Error())
	}

	// 检查错误链
	cause := Cause(err)
	if cause != baseErr {
		t.Errorf("Cause错误: 期望 %q, 得到 %q", baseErr, cause)
	}

	// 检查堆栈跟踪
	if st, ok := err.(StackTracer); !ok || st.StackTrace() == nil {
		t.Error("Wrap返回的错误应该包含堆栈跟踪")
	}

	// 检查nil错误处理
	if Wrap(nil, wrapMsg) != nil {
		t.Error("Wrap(nil, msg)应该返回nil")
	}
}

// 测试格式化包装错误
func TestWrapf(t *testing.T) {
	const (
		baseErrorMsg = "基础错误"
		wrapFormat   = "包装错误: %s"
		detail       = "详细信息"
	)

	expectedMessage := fmt.Sprintf(wrapFormat, detail)
	baseErr := fmt.Errorf(baseErrorMsg)
	err := Wrapf(baseErr, wrapFormat, detail)

	if err == nil {
		t.Fatal("Wrapf应该返回一个错误，但返回了nil")
	}

	if err.Error() != expectedMessage {
		t.Errorf("错误消息错误: 期望 %q, 得到 %q", expectedMessage, err.Error())
	}

	// 检查错误链
	cause := Cause(err)
	if cause != baseErr {
		t.Errorf("Cause错误: 期望 %q, 得到 %q", baseErr, cause)
	}

	// 检查nil错误处理
	if Wrapf(nil, wrapFormat, detail) != nil {
		t.Error("Wrapf(nil, format, args)应该返回nil")
	}
}

// 测试获取错误根因
func TestCause(t *testing.T) {
	const (
		baseErrorMsg = "根本原因"
		wrapMsg1     = "第一层包装"
		wrapMsg2     = "第二层包装"
	)

	baseErr := fmt.Errorf(baseErrorMsg)
	err1 := Wrap(baseErr, wrapMsg1)
	err2 := Wrap(err1, wrapMsg2)

	cause := Cause(err2)

	if cause != baseErr {
		t.Errorf("Cause错误: 期望 %q, 得到 %q", baseErr, cause)
	}

	// 测试直接错误的Cause
	directCause := Cause(baseErr)
	if directCause != baseErr {
		t.Errorf("直接错误的Cause错误: 期望 %q, 得到 %q", baseErr, directCause)
	}

	// 测试nil的Cause
	nilCause := Cause(nil)
	if nilCause != nil {
		t.Errorf("nil的Cause错误: 期望 nil, 得到 %v", nilCause)
	}

	// 测试Cause返回nil的情况
	errWithNilCause := &contextualError{
		msg: "有nil原因的错误",
		err: nil,
	}
	result := Cause(errWithNilCause)
	if result != errWithNilCause {
		t.Errorf("当Cause()返回nil时，应该返回原始错误")
	}
}

// 测试ContextualError接口
func TestContextualError(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")
	err := Wrap(baseErr, "包装消息")

	// 检查是否实现了ContextualError接口
	ce, ok := err.(ContextualError)
	if !ok {
		t.Fatal("包装的错误应该实现ContextualError接口")
	}

	// 检查Cause方法
	if ce.Cause() != baseErr {
		t.Errorf("ContextualError.Cause错误: 期望 %q, 得到 %q", baseErr, ce.Cause())
	}

	// 检查Unwrap方法
	if ce.Unwrap() != baseErr {
		t.Errorf("ContextualError.Unwrap错误: 期望 %q, 得到 %q", baseErr, ce.Unwrap())
	}
}

// 测试嵌套多层的错误链
func TestNestedErrors(t *testing.T) {
	err1 := fmt.Errorf("错误1")
	err2 := Wrap(err1, "错误2")
	err3 := Wrap(err2, "错误3")
	err4 := Wrap(err3, "错误4")

	// 检查最底层原因
	cause := Cause(err4)
	if cause != err1 {
		t.Errorf("嵌套错误的Cause错误: 期望 %q, 得到 %q", err1, cause)
	}
}

// 测试错误格式化
func TestErrorFormatting(t *testing.T) {
	const (
		baseErrorMsg = "基础错误"
		wrapMsg      = "包装错误"
	)

	baseErr := fmt.Errorf(baseErrorMsg)
	err := Wrap(baseErr, wrapMsg)

	// 测试简单格式化
	simple := fmt.Sprintf("%s", err)
	if simple != wrapMsg {
		t.Errorf("简单格式化错误: 期望 %q, 得到 %q", wrapMsg, simple)
	}

	// 测试详细格式化
	detailed := fmt.Sprintf("%+v", err)
	if !strings.Contains(detailed, wrapMsg) {
		t.Errorf("详细格式化应该包含错误消息: %q", wrapMsg)
	}

	// 测试其他格式化
	quoted := fmt.Sprintf("%q", err)
	expectedQuoted := fmt.Sprintf("%q", wrapMsg)
	if quoted != expectedQuoted {
		t.Errorf("引号格式化错误: 期望 %q, 得到 %q", expectedQuoted, quoted)
	}
}

// =================================================
// 标准库兼容性测试
// =================================================

// 测试Is方法 - 单个错误
func TestStdlibIs(t *testing.T) {
	const (
		code    = 404
		message = "资源未找到"
	)

	// 注册错误码
	RegisterErrorCode(code, 404, "资源未找到", "")

	// 创建错误
	err1 := NewWithCode(code, message)
	err2 := NewWithCode(code, "另一个消息但相同错误码")
	err3 := NewWithCode(code+1, "不同错误码")

	// 标准errors.Is使用我们的Is方法
	if !stderrors.Is(err1, err2) {
		t.Error("具有相同错误码的错误应该被errors.Is识别为相等")
	}

	if stderrors.Is(err1, err3) {
		t.Error("具有不同错误码的错误不应该被errors.Is识别为相等")
	}

	// 测试与标准错误比较
	stdErr := fmt.Errorf("标准错误")
	wrappedStdErr := Wrap(stdErr, "包装的标准错误")

	if !stderrors.Is(wrappedStdErr, stdErr) {
		t.Error("包装的标准错误应该与原始错误匹配")
	}
}

// 测试Is方法 - 错误链
func TestStdlibIsErrorChain(t *testing.T) {
	// 创建错误链
	baseErr := fmt.Errorf("基础错误")
	wrappedErr := Wrap(baseErr, "包装错误")
	doubleWrappedErr := Wrap(wrappedErr, "二次包装错误")

	// 使用标准errors.Is检查错误链
	if !stderrors.Is(doubleWrappedErr, baseErr) {
		t.Error("errors.Is应该能够在错误链中识别基础错误")
	}

	if !stderrors.Is(doubleWrappedErr, wrappedErr) {
		t.Error("errors.Is应该能够在错误链中识别中间错误")
	}

	// 测试不在链中的错误
	otherErr := fmt.Errorf("其他错误")
	if stderrors.Is(doubleWrappedErr, otherErr) {
		t.Error("errors.Is不应该将错误链与不相关的错误匹配")
	}
}

// 定义一个自定义错误类型用于As测试
type testCustomError struct {
	msg string
}

func (e *testCustomError) Error() string {
	return e.msg
}

// 定义一个带错误码的自定义错误类型
type testCustomCodeError struct {
	code int
	msg  string
}

func (e *testCustomCodeError) Error() string {
	return e.msg
}

func (e *testCustomCodeError) Code() int {
	return e.code
}

func (e *testCustomCodeError) SetCode(code int) {
	e.code = code
}

// 测试As方法
func TestAsWithCustomTypes(t *testing.T) {
	// 测试标准错误使用As
	originalErr := &testCustomError{msg: "原始错误"}
	wrappedErr := Wrap(originalErr, "包装错误")

	// 尝试提取自定义错误
	var customErr *testCustomError
	if !stderrors.As(wrappedErr, &customErr) {
		t.Error("标准库As应该能提取自定义错误类型")
	} else if customErr.msg != "原始错误" {
		t.Errorf("提取的自定义错误消息错误: %s", customErr.msg)
	}

	// 测试错误码使用As
	codeErr := &testCustomCodeError{code: 404, msg: "资源未找到"}
	wrappedCodeErr := WrapWithCode(codeErr, 500, "服务器错误")

	// 尝试提取带错误码的自定义错误
	var extractedCodeErr *testCustomCodeError
	if !stderrors.As(wrappedCodeErr, &extractedCodeErr) {
		t.Error("标准库As应该能提取带错误码的自定义错误类型")
	} else if extractedCodeErr.code != 404 {
		t.Errorf("提取的错误码错误: %d", extractedCodeErr.code)
	}

	// 测试非指针类型（会编译错误，因此跳过）
	// var s string
	// if stderrors.As(wrappedErr, s) {
	//	t.Error("As不应该接受非指针目标")
	// }

	// 测试nil错误
	var nilErr *testCustomError
	if stderrors.As(nil, &nilErr) {
		t.Error("As对nil错误应返回false")
	}
}

// 测试聚合错误的Is方法
func TestStdlibAggregateIs(t *testing.T) {
	err1 := fmt.Errorf("错误1")
	err2 := fmt.Errorf("错误2")
	agg := NewAggregate([]error{err1, err2})

	// 测试是否包含其中一个错误
	if !stderrors.Is(agg, err1) {
		t.Error("聚合错误应该与其包含的错误之一匹配")
	}

	if !stderrors.Is(agg, err2) {
		t.Error("聚合错误应该与其包含的错误之一匹配")
	}

	// 测试不包含的错误
	otherErr := fmt.Errorf("其他错误")
	if stderrors.Is(agg, otherErr) {
		t.Error("聚合错误不应该与不包含的错误匹配")
	}

	// 测试聚合错误与聚合错误的比较
	agg2 := NewAggregate([]error{err1, err2})
	if !stderrors.Is(agg, agg2) {
		t.Error("包含相同错误的聚合错误应该被识别为相等")
	}

	// 测试不同的聚合错误
	agg3 := NewAggregate([]error{err1, otherErr})
	if stderrors.Is(agg, agg3) {
		t.Error("包含不同错误的聚合错误不应该被识别为相等")
	}
}

// 测试多种堆栈捕获模式下的New函数
func TestNewWithAllStackCaptureModes(t *testing.T) {
	// 保存原始配置
	originalMode := DefaultStackCaptureMode
	defer func() {
		DefaultStackCaptureMode = originalMode
	}()

	// 测试所有模式
	modes := []StackCaptureMode{
		StackCaptureModeNever,
		StackCaptureModeImmediate,
		StackCaptureModeDeferred,
		StackCaptureModeModeSampled,
	}

	for _, mode := range modes {
		DefaultStackCaptureMode = mode
		err := New("测试错误")

		// 验证错误消息始终正确
		if err.Error() != "测试错误" {
			t.Errorf("模式 %v: New返回的错误消息不正确", mode)
		}

		// 验证堆栈跟踪行为符合预期
		hasStack := false
		if st, ok := err.(StackTracer); ok && st.StackTrace() != nil {
			hasStack = true
		}

		switch mode {
		case StackCaptureModeNever:
			if hasStack {
				t.Errorf("模式 %v: 不应有堆栈跟踪", mode)
			}
		case StackCaptureModeImmediate, StackCaptureModeDeferred:
			if !hasStack {
				t.Errorf("模式 %v: 应有堆栈跟踪", mode)
			}
			// 采样模式不做具体断言
		}
	}
}

// 测试多种堆栈捕获模式下的Errorf函数
func TestErrorfWithDifferentStackModes(t *testing.T) {
	// 保存原始配置
	originalMode := DefaultStackCaptureMode
	defer func() {
		DefaultStackCaptureMode = originalMode
	}()

	// 测试基本功能
	err := Errorf("格式化 %s 错误", "测试")
	if err.Error() != "格式化 测试 错误" {
		t.Errorf("Errorf消息格式化错误: %v", err)
	}

	// 测试带堆栈和不带堆栈的情况
	modes := []struct {
		mode            StackCaptureMode
		shouldHaveStack bool
	}{
		{StackCaptureModeNever, false},
		{StackCaptureModeImmediate, true},
		{StackCaptureModeDeferred, true},
	}

	for _, tc := range modes {
		DefaultStackCaptureMode = tc.mode
		err := Errorf("测试")

		hasStack := false
		if st, ok := err.(StackTracer); ok && st.StackTrace() != nil {
			hasStack = true
		}

		if hasStack != tc.shouldHaveStack {
			t.Errorf("模式 %v: 堆栈跟踪存在状态 %v, 期望 %v",
				tc.mode, hasStack, tc.shouldHaveStack)
		}
	}
}

// 测试WithContextMap函数增强版
func TestWithContextMapExtended(t *testing.T) {
	originalErr := New("原始错误")
	contextMap := map[string]interface{}{
		"requestID": "req123",
		"userID":    123,
	}

	// 添加上下文
	wrappedErr := WithContextMap(originalErr, contextMap)

	// 测试获取上下文
	if reqID, ok := GetContext(wrappedErr, "requestID"); !ok {
		t.Error("应能获取requestID上下文")
	} else if reqID != "req123" {
		t.Errorf("requestID值错误: %v", reqID)
	}

	// 测试获取所有上下文
	allContext := GetAllContext(wrappedErr)
	if len(allContext) != 2 {
		t.Errorf("上下文映射长度错误: %d", len(allContext))
	}

	// 测试nil错误
	nilContextErr := WithContextMap(nil, contextMap)
	if nilContextErr != nil {
		t.Error("对nil错误使用WithContextMap应返回nil")
	}

	// 测试使用snake_case和camelCase混合的键
	baseErr := fmt.Errorf("基础错误")
	mixedMap := map[string]interface{}{
		"request_id": "abc-123",
		"userID":     456,
	}

	mixedErr := WithContextMap(baseErr, mixedMap)

	// 验证可以获取两种命名风格的键
	if reqID, ok := GetContext(mixedErr, "request_id"); !ok || reqID != "abc-123" {
		t.Errorf("应能获取snake_case风格的请求ID: %v", reqID)
	}

	if userID, ok := GetContext(mixedErr, "userID"); !ok || userID != 456 {
		t.Errorf("应能获取camelCase风格的用户ID: %v", userID)
	}
}

// =================================================
// 错误上下文测试
// =================================================

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

	// 测试使用requestID/userID格式的上下文键（合并TestWithContextMapExtended）
	originalErr := New("原始错误")
	extendedMap := map[string]interface{}{
		"requestID": "req123",
		"userID":    123,
	}

	// 添加上下文
	wrappedErr := WithContextMap(originalErr, extendedMap)

	// 测试获取上下文
	if reqID, ok := GetContext(wrappedErr, "requestID"); !ok {
		t.Error("应能获取requestID上下文")
	} else if reqID != "req123" {
		t.Errorf("requestID值错误: %v", reqID)
	}

	// 测试获取所有上下文
	extendedContext := GetAllContext(wrappedErr)
	if len(extendedContext) != 2 {
		t.Errorf("上下文映射长度错误: %d", len(extendedContext))
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

// =================================================
// 错误上下文使用示例
// =================================================

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

// 定义一个实现Code和SetCode接口的错误类型
type testCodedError struct {
	code int
	msg  string
}

func (e *testCodedError) Error() string {
	return e.msg
}

func (e *testCodedError) Code() int {
	return e.code
}

func (e *testCodedError) SetCode(code int) {
	e.code = code
}

// 定义一个实现多层嵌套错误的类型
type testNestedError struct {
	err error
}

func (e *testNestedError) Error() string {
	if e.err != nil {
		return "嵌套: " + e.err.Error()
	}
	return "嵌套错误"
}

func (e *testNestedError) Unwrap() error {
	return e.err
}

// 定义一个同时实现多个接口的错误类型
type testMultiInterfaceError struct {
	code    int
	msg     string
	context map[string]interface{}
}

func (e *testMultiInterfaceError) Error() string {
	return e.msg
}

func (e *testMultiInterfaceError) Code() int {
	return e.code
}

func (e *testMultiInterfaceError) SetCode(code int) {
	e.code = code
}

func (e *testMultiInterfaceError) Context() map[string]interface{} {
	return e.context
}

// 测试As方法增强版
func TestAsEnhanced(t *testing.T) {
	// 1. 测试contextualError实现的As方法
	baseErr := &testCustomError{msg: "基础错误"}
	contextErr := &contextualError{
		msg:   "上下文错误",
		err:   baseErr,
		code:  404,
		stack: callers(),
		context: map[string]interface{}{
			"key": "value",
		},
	}

	// 1.1 测试直接匹配contextualError类型
	var ce *contextualError
	if !stderrors.As(contextErr, &ce) {
		t.Error("应该能够将contextualError匹配为它自己的类型")
	} else if ce.code != 404 {
		t.Errorf("匹配的contextualError的错误码错误: %d", ce.code)
	}

	// 1.2 测试匹配其内部嵌套的错误
	var customErr *testCustomError
	if !stderrors.As(contextErr, &customErr) {
		t.Error("应该能从contextualError中提取内部嵌套的testCustomError")
	} else if customErr.msg != "基础错误" {
		t.Errorf("内部嵌套错误的消息错误: %s", customErr.msg)
	}

	// 2. 测试匹配Code接口
	codedErr := &testCodedError{code: 999, msg: "编码错误"}
	wrappedCodedErr := Wrap(codedErr, "包装的编码错误")

	// 创建一个可以接收Code接口的目标
	type codeGetter interface {
		Code() int
	}
	var cg codeGetter
	if !stderrors.As(wrappedCodedErr, &cg) {
		t.Error("应该能够将wrappedCodedErr匹配为Code接口")
	} else if cg.Code() != 999 {
		t.Errorf("通过接口获取的错误码错误: %d", cg.Code())
	}

	// 3. 测试多层嵌套错误
	innerErr := &testCustomError{msg: "内层错误"}
	middleErr := &testNestedError{err: innerErr}
	outerErr := &testNestedError{err: middleErr}
	wrappedNestedErr := Wrap(outerErr, "包装的多层嵌套错误")

	var inner *testCustomError
	if !stderrors.As(wrappedNestedErr, &inner) {
		t.Error("应该能从多层嵌套错误中提取最内层错误")
	} else if inner.msg != "内层错误" {
		t.Errorf("内层错误消息错误: %s", inner.msg)
	}

	// 4. 测试实现多接口的错误
	multiErr := &testMultiInterfaceError{
		code: 501,
		msg:  "多接口错误",
		context: map[string]interface{}{
			"multi": true,
		},
	}
	wrappedMultiErr := Wrap(multiErr, "包装的多接口错误")

	// 4.1 测试匹配为Code接口
	if !stderrors.As(wrappedMultiErr, &cg) {
		t.Error("应该能够将包装的多接口错误匹配为Code接口")
	} else if cg.Code() != 501 {
		t.Errorf("通过接口获取的多接口错误码错误: %d", cg.Code())
	}

	// 4.2 测试匹配为完整的多接口错误类型
	var multi *testMultiInterfaceError
	if !stderrors.As(wrappedMultiErr, &multi) {
		t.Error("应该能够将包装的多接口错误匹配为原始类型")
	} else {
		if multi.code != 501 {
			t.Errorf("多接口错误码错误: %d", multi.code)
		}
		if !multi.context["multi"].(bool) {
			t.Error("多接口错误上下文错误")
		}
	}

	// 5. 测试直接类型断言失败的情况
	strErr := fmt.Errorf("字符串错误")
	var numErr *testCustomCodeError
	if stderrors.As(strErr, &numErr) {
		t.Error("不应该能够将字符串错误匹配为testCustomCodeError")
	}

	// 6. 测试nil错误和nil目标
	if stderrors.As(nil, &customErr) {
		t.Error("nil错误不应该匹配任何目标")
	}

	// 7. 测试非指针目标
	// 注意：这里通过反射模拟，因为直接传非指针会导致编译错误
	invalidTarget := "非指针目标"
	result := contextErr.As(invalidTarget) // 直接调用As方法而不是通过stderrors.As
	if result {
		t.Error("非指针目标应该返回匹配失败")
	}
}

// 测试As方法的额外场景
func TestAsMethodAdditionalScenarios(t *testing.T) {
	// 测试直接匹配nil错误
	var customErr *testCustomError
	if stderrors.As(nil, &customErr) {
		t.Error("对nil错误调用As不应匹配任何类型")
	}

	// 错误链中有nil的情况
	nilWrappedErr := &contextualError{
		msg: "包装nil",
		err: nil,
	}
	if stderrors.As(nilWrappedErr, &customErr) {
		t.Error("包含nil原因的错误不应匹配非自身类型")
	}

	// 目标值为nil
	baseErr := &testCustomError{msg: "基础错误"}
	wrappedErr := Wrap(baseErr, "包装错误")

	var nilTarget *testCustomError = nil
	if !stderrors.As(wrappedErr, &nilTarget) {
		t.Error("目标值为nil时，As应该能够正确设置目标")
	}
	if nilTarget == nil || nilTarget.msg != "基础错误" {
		t.Errorf("As应该正确设置目标值，got: %v", nilTarget)
	}

	// 嵌套多层的复杂情况
	customCodeErr := &testCustomCodeError{code: 123, msg: "编码错误"}
	nestedErr := Wrap(customCodeErr, "一级嵌套")
	doubleNestedErr := Wrap(nestedErr, "二级嵌套")
	tripleNestedErr := Wrap(doubleNestedErr, "三级嵌套")

	var extractedCodeErr *testCustomCodeError
	if !stderrors.As(tripleNestedErr, &extractedCodeErr) {
		t.Error("As应该能穿透多层嵌套查找特定类型")
	} else if extractedCodeErr.code != 123 {
		t.Errorf("提取的错误码错误: %d", extractedCodeErr.code)
	}
}

// 测试Errorf函数的边缘情况
func TestErrorfEdgeCases(t *testing.T) {
	// 测试空格式字符串
	err1 := Errorf("")
	if err1.Error() != "" {
		t.Errorf("空格式字符串应该产生空错误消息，得到: %q", err1.Error())
	}

	// 测试格式字符串中包含%字符但没有参数
	err2 := Errorf("格式字符串包含%%字符但没有参数")
	if !strings.Contains(err2.Error(), "格式字符串包含%字符但没有参数") {
		t.Errorf("含有%%的格式字符串没有正确处理: %q", err2.Error())
	}

	// 测试格式字符串中有多个参数
	err3 := Errorf("多个参数: %d, %s, %t", 123, "字符串", true)
	expected := "多个参数: 123, 字符串, true"
	if err3.Error() != expected {
		t.Errorf("多参数格式错误: 期望 %q, 得到 %q", expected, err3.Error())
	}

	// 测试在堆栈捕获模式为Never时的行为
	originalMode := DefaultStackCaptureMode
	DefaultStackCaptureMode = StackCaptureModeNever
	defer func() {
		DefaultStackCaptureMode = originalMode
	}()

	err4 := Errorf("无堆栈错误")
	if st, ok := err4.(StackTracer); ok && st.StackTrace() != nil {
		t.Error("StackCaptureModeNever模式下Errorf不应产生堆栈")
	}
}

// 测试WithStack的更多场景
func TestWithStackMoreScenarios(t *testing.T) {
	// 测试为已有堆栈的错误添加堆栈
	baseErr := New("带堆栈的错误")
	doubleStackErr := WithStack(baseErr)

	// 检查是否仍然保留了原始错误消息
	if doubleStackErr.Error() != baseErr.Error() {
		t.Errorf("添加堆栈后消息变化: 期望 %q, 得到 %q",
			baseErr.Error(), doubleStackErr.Error())
	}

	// 测试堆栈有效性
	st, ok := doubleStackErr.(StackTracer)
	if !ok || st.StackTrace() == nil {
		t.Error("WithStack后应该有堆栈")
	}

	// 测试不同堆栈捕获模式
	originalMode := DefaultStackCaptureMode
	defer func() {
		DefaultStackCaptureMode = originalMode
	}()

	// 测试延迟捕获模式
	DefaultStackCaptureMode = StackCaptureModeDeferred
	deferredErr := WithStack(fmt.Errorf("基础错误"))
	if _, ok := deferredErr.(StackTracer); !ok {
		t.Error("延迟模式应该返回可提供堆栈的错误")
	}

	// 测试立即捕获模式
	DefaultStackCaptureMode = StackCaptureModeImmediate
	immediateErr := WithStack(fmt.Errorf("基础错误"))
	if _, ok := immediateErr.(StackTracer); !ok {
		t.Error("立即模式应该返回可提供堆栈的错误")
	}

	// 测试采样模式
	DefaultStackCaptureMode = StackCaptureModeModeSampled
	_ = WithStack(fmt.Errorf("基础错误"))
	// 这里不做断言因为采样模式的行为是不确定的
}

// 测试Wrapf函数的更多场景
func TestWrapfMoreScenarios(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")

	// 测试基本用法补充
	err1 := Wrapf(baseErr, "简单包装 %s", "消息")
	if err1.Error() != "简单包装 消息" {
		t.Errorf("错误消息格式化错误: %q", err1.Error())
	}

	// 测试多个格式参数
	err2 := Wrapf(baseErr, "包装于 %s，代码: %d，状态: %t", "测试函数", 404, true)
	expected := "包装于 测试函数，代码: 404，状态: true"
	if err2.Error() != expected {
		t.Errorf("多参数格式错误: 期望 %q, 得到 %q", expected, err2.Error())
	}

	// 测试获取错误原因
	cause := Cause(err1)
	if cause != baseErr {
		t.Errorf("Wrapf(baseErr, ...)的Cause应为baseErr，得到: %v", cause)
	}

	// 错误为nil时应返回nil
	if Wrapf(nil, "这不应生成错误") != nil {
		t.Error("Wrapf(nil, msg)应该返回nil")
	}
}

// 测试WrapWithCode的更多场景
func TestWrapWithCodeMoreScenarios(t *testing.T) {
	baseErr := fmt.Errorf("基础错误")
	const code = 503

	// 注册错误码
	RegisterErrorCode(code, http.StatusServiceUnavailable, "服务不可用", "")

	// 测试基本用法补充
	err1 := WrapWithCode(baseErr, code, "简单包装错误")

	// 检查错误码
	codeErr, ok := err1.(interface{ Code() int })
	if !ok || codeErr.Code() != code {
		t.Errorf("错误码设置错误: 期望 %d, 得到 %v", code, codeErr)
	}

	// 测试格式化包装
	err2 := WrapWithCode(baseErr, code, "格式化包装错误: %d", 42)
	expected := "格式化包装错误: 42"
	if err2.Error() != expected {
		t.Errorf("格式化包装错误: 期望 %q, 得到 %q", expected, err2.Error())
	}

	// 测试错误原因传递
	cause := Cause(err1)
	if cause != baseErr {
		t.Errorf("WrapWithCode应保留原始错误: 期望 %v, 得到 %v", baseErr, cause)
	}

	// 测试错误码属性保留
	codedErr := GetErrorCodeFromError(err2)
	if codedErr == nil || codedErr.Code() != code {
		t.Errorf("WrapWithCode错误码属性错误: 期望 %d, 得到 %v", code, codedErr)
	}

	// nil错误处理
	if WrapWithCode(nil, code, "包装nil") != nil {
		t.Error("WrapWithCode(nil, code, msg)应该返回nil")
	}
}

// 测试NewWithStackControl的更多场景
func TestNewWithStackControlMoreScenarios(t *testing.T) {
	// 测试所有堆栈捕获模式
	modes := []struct {
		mode            StackCaptureMode
		shouldHaveStack bool
		name            string
	}{
		{StackCaptureModeNever, false, "Never"},
		{StackCaptureModeImmediate, true, "Immediate"},
		{StackCaptureModeDeferred, true, "Deferred"},
		// 注意：采样模式不做确定性断言
	}

	for _, tc := range modes {
		t.Run(tc.name, func(t *testing.T) {
			err := NewWithStackControl("测试错误", tc.mode)

			// 检查错误消息
			if err.Error() != "测试错误" {
				t.Errorf("错误消息错误: 期望 %q, 得到 %q", "测试错误", err.Error())
			}

			// 检查堆栈跟踪
			hasStack := false
			if st, ok := err.(StackTracer); ok && st.StackTrace() != nil {
				hasStack = true
			}

			if hasStack != tc.shouldHaveStack {
				t.Errorf("模式 %v: 堆栈存在 %v, 期望 %v", tc.mode, hasStack, tc.shouldHaveStack)
			}
		})
	}

	// 测试特殊情况：采样模式下的行为
	// 注意：这只是测试函数调用而不断言结果
	_ = NewWithStackControl("采样测试", StackCaptureModeModeSampled)

	// 测试空消息
	emptyErr := NewWithStackControl("", StackCaptureModeImmediate)
	if emptyErr.Error() != "" {
		t.Error("空消息应保持为空")
	}
}
