package errors

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// 测试带错误码的错误
func TestNewWithCode(t *testing.T) {
	const (
		code    = 404
		message = "资源未找到"
	)

	// 注册错误码
	RegisterErrorCode(code, http.StatusNotFound, "Resource not found", "")

	err := NewWithCode(code, message)

	if err == nil {
		t.Fatal("NewWithCode应该返回一个错误，但返回了nil")
	}

	// 检查错误码
	codeErr, ok := err.(interface{ Code() int })
	if !ok {
		t.Fatal("NewWithCode返回的错误应该有Code方法")
	}

	if codeErr.Code() != code {
		t.Errorf("错误码错误: 期望 %d, 得到 %d", code, codeErr.Code())
	}

	// 检查错误消息
	if err.Error() != message {
		t.Errorf("错误消息错误: 期望 %q, 得到 %q", message, err.Error())
	}
}

// 测试带错误码的包装错误
func TestWrapWithCode(t *testing.T) {
	const (
		baseErrorMsg = "基础错误"
		code         = 500
		wrapMsg      = "数据库错误"
	)

	baseErr := fmt.Errorf(baseErrorMsg)
	err := WrapWithCode(baseErr, code, wrapMsg)

	if err == nil {
		t.Fatal("WrapWithCode应该返回一个错误，但返回了nil")
	}

	// 检查错误码
	codeErr, ok := err.(interface{ Code() int })
	if !ok {
		t.Fatal("WrapWithCode返回的错误应该有Code方法")
	}

	if codeErr.Code() != code {
		t.Errorf("错误码错误: 期望 %d, 得到 %d", code, codeErr.Code())
	}

	// 检查错误消息
	if err.Error() != wrapMsg {
		t.Errorf("错误消息错误: 期望 %q, 得到 %q", wrapMsg, err.Error())
	}

	// 检查错误链
	cause := Cause(err)
	if cause != baseErr {
		t.Errorf("Cause错误: 期望 %q, 得到 %q", baseErr, cause)
	}

	// 检查nil错误处理
	if WrapWithCode(nil, code, wrapMsg) != nil {
		t.Error("WrapWithCode(nil, code, msg)应该返回nil")
	}
}

// 测试从ErrorCodeInfo创建错误
func TestNewFromCode(t *testing.T) {
	const code = 403

	// 注册错误码
	RegisterErrorCode(code, http.StatusForbidden, "禁止访问", "")

	// 获取错误码
	codeInfo := GetErrorCode(code)
	if codeInfo == nil {
		t.Fatal("应该能够获取注册的错误码")
	}

	// 从错误码创建错误
	err := NewFromCode(codeInfo)

	if err == nil {
		t.Fatal("NewFromCode应该返回一个错误，但返回了nil")
	}

	// 检查错误码
	codeErr, ok := err.(interface{ Code() int })
	if !ok {
		t.Fatal("NewFromCode返回的错误应该有Code方法")
	}

	if codeErr.Code() != code {
		t.Errorf("错误码错误: 期望 %d, 得到 %d", code, codeErr.Code())
	}

	// 检查错误消息
	if err.Error() != codeInfo.Message() {
		t.Errorf("错误消息错误: 期望 %q, 得到 %q", codeInfo.Message(), err.Error())
	}
}

// 测试带错误码的格式化
func TestCodeErrorFormat(t *testing.T) {
	const (
		baseErrorMsg = "基础错误"
		wrapMsg      = "包装错误"
		code         = 404
	)

	// 注册错误码
	RegisterErrorCode(code, http.StatusNotFound, "资源未找到", "")

	baseErr := fmt.Errorf(baseErrorMsg)
	err := WrapWithCode(baseErr, code, wrapMsg)

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
	if !strings.Contains(detailed, fmt.Sprintf("(%d)", code)) {
		t.Errorf("详细格式化应该包含错误码: %d", code)
	}
}

// 测试错误码注册和查询
func TestErrorCode(t *testing.T) {
	const (
		code       = 429
		httpStatus = http.StatusTooManyRequests
		message    = "请求过多"
		reference  = "rate-limit-docs"
	)

	// 注册错误码
	RegisterErrorCode(code, httpStatus, message, reference)

	// 检查错误码是否已注册
	if !HasErrorCode(code) {
		t.Errorf("错误码 %d 应该已注册", code)
	}

	// 获取错误码信息
	codeInfo := GetErrorCode(code)
	if codeInfo == nil {
		t.Fatalf("应该能够获取错误码 %d 的信息", code)
	}

	// 检查错误码信息是否正确
	if codeInfo.Code() != code {
		t.Errorf("错误码错误: 期望 %d, 得到 %d", code, codeInfo.Code())
	}

	if codeInfo.HTTPStatus() != httpStatus {
		t.Errorf("HTTP状态码错误: 期望 %d, 得到 %d", httpStatus, codeInfo.HTTPStatus())
	}

	if codeInfo.Message() != message {
		t.Errorf("错误消息错误: 期望 %q, 得到 %q", message, codeInfo.Message())
	}

	if codeInfo.Reference() != reference {
		t.Errorf("参考文档错误: 期望 %q, 得到 %q", reference, codeInfo.Reference())
	}

	// 测试列出所有错误码
	codes := ListErrorCodes()
	found := false
	for _, c := range codes {
		if c == code {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("ListErrorCodes应该包含已注册的错误码 %d", code)
	}
}

// 测试错误码默认行为
func TestErrorCodeDefaults(t *testing.T) {
	// 测试未设置HTTP状态码时的默认值
	ec := NewErrorCode(100, 0, "测试消息", "")
	if ec.HTTPStatus() != 500 {
		t.Errorf("未设置HTTP状态码时应默认为500，但得到 %d", ec.HTTPStatus())
	}

	// 测试String方法
	expected := "100: 测试消息"
	if ec.String() != expected {
		t.Errorf("String()返回错误: 期望 %q, 得到 %q", expected, ec.String())
	}
}

// 测试MustRegister方法
func TestMustRegister(t *testing.T) {
	const code = 599

	// 正常注册应该不会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustRegister不应该panic: %v", r)
		}
	}()

	ec := NewErrorCode(code, http.StatusInternalServerError, "测试强制注册", "")
	MustRegister(ec)

	// 验证注册成功
	if !HasErrorCode(code) {
		t.Error("MustRegister应该成功注册错误码")
	}
}

// 测试MustRegister对重复错误码的处理
func TestMustRegisterDuplicate(t *testing.T) {
	const code = 598

	// 首次注册
	ec1 := NewErrorCode(code, http.StatusInternalServerError, "第一个", "")
	Register(ec1)

	// 重复注册应该panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("重复使用MustRegister应该panic")
		}
	}()

	ec2 := NewErrorCode(code, http.StatusInternalServerError, "第二个", "")
	MustRegister(ec2)
}

// 测试检查错误中的错误码
func TestIsErrorCode(t *testing.T) {
	const code = 400

	// 注册错误码
	RegisterErrorCode(code, http.StatusBadRequest, "无效请求", "")

	// 创建带错误码的错误
	err := NewWithCode(code, "参数无效")

	// 检查错误码
	if !IsErrorCode(err, code) {
		t.Errorf("IsErrorCode应该为错误码 %d 返回true", code)
	}

	// 检查其他错误码
	if IsErrorCode(err, code+1) {
		t.Errorf("IsErrorCode应该为错误码 %d 返回false", code+1)
	}

	// 包装错误后检查错误码
	wrappedErr := Wrap(err, "包装的无效请求")
	if !IsErrorCode(wrappedErr, code) {
		t.Errorf("IsErrorCode应该在包装的错误中检测到错误码 %d", code)
	}

	// 检查nil错误
	if IsErrorCode(nil, code) {
		t.Error("IsErrorCode应该为nil错误返回false")
	}
}

// 测试GetErrorCodeFromError函数
func TestGetErrorCodeFromError(t *testing.T) {
	const code = 401

	// 注册错误码
	RegisterErrorCode(code, http.StatusUnauthorized, "未授权", "")

	// 创建带错误码的错误
	err := NewWithCode(code, "需要登录")

	// 获取错误码
	codeInfo := GetErrorCodeFromError(err)
	if codeInfo == nil {
		t.Fatal("应该能够从错误中获取错误码信息")
	}

	if codeInfo.Code() != code {
		t.Errorf("错误码错误: 期望 %d, 得到 %d", code, codeInfo.Code())
	}

	// 测试普通错误
	plainErr := fmt.Errorf("普通错误")
	plainCodeInfo := GetErrorCodeFromError(plainErr)
	if plainCodeInfo != UnknownError {
		t.Error("普通错误应该返回UnknownError")
	}

	// 测试nil错误
	nilCodeInfo := GetErrorCodeFromError(nil)
	if nilCodeInfo != nil {
		t.Error("nil错误应该返回nil错误码信息")
	}
}

// 测试MustRegisterErrorCode函数
func TestMustRegisterErrorCode(t *testing.T) {
	const code = 597

	// 正常注册应该不会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustRegisterErrorCode不应该panic: %v", r)
		}
	}()

	MustRegisterErrorCode(code, http.StatusInternalServerError, "测试强制注册", "")

	// 验证注册成功
	if !HasErrorCode(code) {
		t.Error("MustRegisterErrorCode应该成功注册错误码")
	}
}

// 测试禁止使用错误码0的情况
func TestReservedZeroCode(t *testing.T) {
	// 尝试使用保留的错误码0应该会panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("使用错误码0应该panic")
		}
	}()

	NewErrorCode(0, http.StatusInternalServerError, "尝试使用保留的错误码", "")
	MustRegister(NewErrorCode(0, http.StatusInternalServerError, "尝试使用保留的错误码", ""))
}

// 测试不存在的错误码
func TestNonExistentErrorCode(t *testing.T) {
	const nonExistentCode = 999998

	// 尝试获取不存在的错误码
	codeInfo := GetErrorCode(nonExistentCode)
	if codeInfo != nil {
		t.Errorf("获取不存在的错误码应该返回nil，但得到: %v", codeInfo)
	}
}

// 测试验证错误码不存在
func TestHasErrorCodeNegative(t *testing.T) {
	const nonExistentCode = 999997

	// 验证不存在的错误码
	if HasErrorCode(nonExistentCode) {
		t.Error("HasErrorCode应该为不存在的错误码返回false")
	}
}

// 测试使用Unwrap函数的错误码检测
func TestIsErrorCodeWithUnwrap(t *testing.T) {
	const code = 411
	RegisterErrorCode(code, http.StatusLengthRequired, "需要内容长度", "")

	// 创建一个带错误码的嵌套错误
	codeErr := NewWithCode(code, "错误消息")
	wrappingErr := Wrap(codeErr, "包装错误")

	// 测试是否可以检测到嵌套错误的错误码
	if !IsErrorCode(wrappingErr, code) {
		t.Errorf("IsErrorCode应该能识别嵌套的错误码 %d", code)
	}
}
