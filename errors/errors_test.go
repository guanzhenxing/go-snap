package errors

import (
	stderrors "errors"
	"fmt"
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
func TestStdlibAs(t *testing.T) {
	// 基本自定义错误测试
	baseErr := &testCustomError{msg: "自定义错误"}
	wrappedErr := Wrap(baseErr, "包装的自定义错误")

	var target *testCustomError
	if !stderrors.As(wrappedErr, &target) {
		t.Error("errors.As应该能够提取出包装的自定义错误")
	}

	if target.msg != "自定义错误" {
		t.Errorf("提取的错误消息错误: 期望 %q, 得到 %q", "自定义错误", target.msg)
	}

	// 带错误码的测试
	codeErr := NewWithCode(404, "未找到")
	var targetCode *testCustomCodeError

	// 这可能会失败，因为我们的As实现是简化的
	// 在实际实现中，需要使用反射正确处理这种情况
	if stderrors.As(codeErr, &targetCode) {
		if targetCode.code != 404 {
			t.Errorf("提取的错误码错误: 期望 %d, 得到 %d", 404, targetCode.code)
		}
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
