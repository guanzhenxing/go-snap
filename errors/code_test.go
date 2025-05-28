package errors

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// ==================================================
// 错误码基本功能测试
// ==================================================

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
		t.Errorf("在列出的错误码中应该包含 %d", code)
	}
}

// 测试错误码默认值
func TestErrorCodeDefaults(t *testing.T) {
	const (
		code = 418 // I'm a teapot
	)

	// 注册错误码，但不提供HTTP状态和引用
	RegisterErrorCode(code, 0, "我是茶壶", "")

	codeInfo := GetErrorCode(code)
	if codeInfo == nil {
		t.Fatalf("应该能够获取错误码 %d 的信息", code)
	}

	// 检查默认HTTP状态码是否设置为500
	if codeInfo.HTTPStatus() != http.StatusInternalServerError {
		t.Errorf("默认HTTP状态码错误: 期望 %d, 得到 %d", http.StatusInternalServerError, codeInfo.HTTPStatus())
	}
}

// 测试错误码必须注册函数
func TestMustRegister(t *testing.T) {
	const (
		code       = 451
		httpStatus = http.StatusUnavailableForLegalReasons
		message    = "由于法律原因不可用"
	)

	// 正常注册应该不会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustRegisterErrorCode不应该panic: %v", r)
		}
	}()

	MustRegisterErrorCode(code, httpStatus, message, "")
}

// 测试重复注册导致panic
func TestMustRegisterDuplicate(t *testing.T) {
	const (
		code    = 452
		message = "重复测试"
	)

	// 先正常注册
	RegisterErrorCode(code, http.StatusBadRequest, message, "")

	// 重复注册应该panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustRegisterErrorCode应该在重复注册时panic")
		}
	}()

	MustRegisterErrorCode(code, http.StatusBadRequest, "另一个消息", "")
}

// 测试错误码判断
func TestIsErrorCode(t *testing.T) {
	const (
		code1 = 460
		code2 = 461
	)

	// 注册错误码
	RegisterErrorCode(code1, http.StatusBadRequest, "错误1", "")
	RegisterErrorCode(code2, http.StatusBadRequest, "错误2", "")

	// 创建带错误码的错误
	err1 := NewWithCode(code1, "第一个错误")

	// 直接判断
	if !IsErrorCode(err1, code1) {
		t.Errorf("IsErrorCode错误: 期望返回true，但得到false")
	}

	if IsErrorCode(err1, code2) {
		t.Errorf("IsErrorCode错误: 期望返回false，但得到true")
	}

	// 测试包装错误
	wrappedErr := Wrap(err1, "包装的第一个错误")
	if !IsErrorCode(wrappedErr, code1) {
		t.Errorf("IsErrorCode错误: 对于包装错误，期望返回true，但得到false")
	}

	// 测试没有错误码的错误
	stdErr := fmt.Errorf("标准错误")
	if IsErrorCode(stdErr, code1) {
		t.Errorf("IsErrorCode错误: 对于没有错误码的错误，期望返回false，但得到true")
	}

	// 测试nil错误
	if IsErrorCode(nil, code1) {
		t.Errorf("IsErrorCode错误: 对于nil，期望返回false，但得到true")
	}
}

// 测试从错误中获取错误码
func TestGetErrorCodeFromError(t *testing.T) {
	const (
		code       = 470
		httpStatus = http.StatusBadRequest
		message    = "测试错误"
		reference  = "test-ref"
	)

	// 注册错误码
	RegisterErrorCode(code, httpStatus, message, reference)

	// 创建带错误码的错误
	err := NewWithCode(code, "实际错误消息")

	// 获取错误码
	codeInfo := GetErrorCodeFromError(err)
	if codeInfo == nil {
		t.Fatal("应该能够从错误中获取错误码信息")
	}

	// 验证错误码信息
	if codeInfo.Code() != code {
		t.Errorf("错误码错误: 期望 %d, 得到 %d", code, codeInfo.Code())
	}

	if codeInfo.HTTPStatus() != httpStatus {
		t.Errorf("HTTP状态码错误: 期望 %d, 得到 %d", httpStatus, codeInfo.HTTPStatus())
	}

	if codeInfo.Message() != message {
		t.Errorf("注册的错误消息错误: 期望 %q, 得到 %q", message, codeInfo.Message())
	}

	// 测试包装错误
	wrappedErr := Wrap(err, "包装错误")
	wrappedCodeInfo := GetErrorCodeFromError(wrappedErr)
	if wrappedCodeInfo == nil {
		t.Fatal("应该能够从包装错误中获取错误码信息")
	}

	if wrappedCodeInfo.Code() != code {
		t.Errorf("包装错误的错误码错误: 期望 %d, 得到 %d", code, wrappedCodeInfo.Code())
	}

	// 测试没有错误码的错误
	stdErr := fmt.Errorf("标准错误")
	stdCodeInfo := GetErrorCodeFromError(stdErr)
	if stdCodeInfo == nil {
		t.Fatal("对于标准错误，应该返回默认的UnknownError")
	}

	if stdCodeInfo.Code() != UnknownError.Code() {
		t.Errorf("标准错误的错误码错误: 期望 %d, 得到 %d", UnknownError.Code(), stdCodeInfo.Code())
	}
}

// 测试创建错误码
func TestNewErrorCode(t *testing.T) {
	const (
		code       = 480
		httpStatus = http.StatusBadRequest
		message    = "创建的错误码"
		reference  = "create-ref"
	)

	errorCode := NewErrorCode(code, httpStatus, message, reference)

	if errorCode.Code() != code {
		t.Errorf("错误码错误: 期望 %d, 得到 %d", code, errorCode.Code())
	}

	if errorCode.HTTPStatus() != httpStatus {
		t.Errorf("HTTP状态码错误: 期望 %d, 得到 %d", httpStatus, errorCode.HTTPStatus())
	}

	if errorCode.Message() != message {
		t.Errorf("错误消息错误: 期望 %q, 得到 %q", message, errorCode.Message())
	}

	if errorCode.Reference() != reference {
		t.Errorf("参考文档错误: 期望 %q, 得到 %q", reference, errorCode.Reference())
	}
}

// 测试0错误码现在可以正常注册
func TestZeroCodeAllowed(t *testing.T) {
	const (
		code    = 0
		message = "零错误码"
	)

	// 注册0错误码应该成功，不会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("注册错误码0不应该panic: %v", r)
		}
	}()

	// 注册错误码0
	RegisterErrorCode(code, http.StatusOK, message, "")

	// 验证注册成功
	if !HasErrorCode(code) {
		t.Error("错误码0应该已成功注册")
	}

	codeInfo := GetErrorCode(code)
	if codeInfo == nil {
		t.Fatal("应该能够获取错误码0的信息")
	}

	if codeInfo.Code() != code {
		t.Errorf("错误码错误: 期望 %d, 得到 %d", code, codeInfo.Code())
	}

	if codeInfo.Message() != message {
		t.Errorf("错误消息错误: 期望 %q, 得到 %q", message, codeInfo.Message())
	}
}

// 测试不存在的错误码
func TestNonExistentErrorCode(t *testing.T) {
	const nonExistentCode = 99999

	// 获取不存在的错误码应该返回nil
	codeInfo := GetErrorCode(nonExistentCode)
	if codeInfo != nil {
		t.Errorf("GetErrorCode(%d)应该返回nil，但返回 %v", nonExistentCode, codeInfo)
	}
}

// 测试HasErrorCode的负数处理
func TestHasErrorCodeNegative(t *testing.T) {
	const negativeCode = -1

	// 默认情况下不应该有负数错误码
	if HasErrorCode(negativeCode) {
		t.Errorf("错误码 %d 不应该存在", negativeCode)
	}

	// 注册负数错误码（虽然不推荐）
	RegisterErrorCode(negativeCode, http.StatusBadRequest, "负数错误码", "")

	// 现在应该可以找到
	if !HasErrorCode(negativeCode) {
		t.Errorf("错误码 %d 应该已注册", negativeCode)
	}
}

// ==================================================
// 错误码高级功能与性能测试
// ==================================================

// 测试批量注册错误码
func TestBatchRegisterErrorCodes(t *testing.T) {
	// 准备测试数据
	codes := []ErrorCode{
		NewErrorCode(2001, http.StatusBadRequest, "测试错误1", ""),
		NewErrorCode(2002, http.StatusBadRequest, "测试错误2", ""),
		NewErrorCode(2003, http.StatusBadRequest, "测试错误3", ""),
	}

	// 批量注册
	RegisterErrorCodes(codes)

	// 验证是否注册成功
	for _, code := range codes {
		if !HasErrorCode(code.Code()) {
			t.Errorf("错误码 %d 应该已注册", code.Code())
		}

		registeredCode := GetErrorCode(code.Code())
		if registeredCode == nil {
			t.Errorf("应该能够获取错误码 %d", code.Code())
			continue
		}

		if registeredCode.Message() != code.Message() {
			t.Errorf("错误消息不匹配: 期望 %q, 得到 %q", code.Message(), registeredCode.Message())
		}
	}
}

// 测试批量MustRegister错误码
func TestBatchMustRegisterErrorCodes(t *testing.T) {
	// 准备测试数据
	codes := []ErrorCode{
		NewErrorCode(3001, http.StatusBadRequest, "必须注册错误1", ""),
		NewErrorCode(3002, http.StatusBadRequest, "必须注册错误2", ""),
		NewErrorCode(3003, http.StatusBadRequest, "必须注册错误3", ""),
	}

	// 测试成功情况
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustRegisterErrorCodes不应该panic: %v", r)
		}
	}()

	MustRegisterErrorCodes(codes)

	// 验证是否注册成功
	for _, code := range codes {
		if !HasErrorCode(code.Code()) {
			t.Errorf("错误码 %d 应该已注册", code.Code())
		}
	}
}

// 测试批量MustRegister错误码的失败情况
func TestBatchMustRegisterErrorCodesDuplicate(t *testing.T) {
	// 准备测试数据 - 包含一个重复的错误码
	code1 := NewErrorCode(4001, http.StatusBadRequest, "唯一错误", "")
	code2 := NewErrorCode(4001, http.StatusBadRequest, "重复错误", "") // 故意使用相同的错误码

	// 注册第一个错误码
	Register(code1)

	// 测试重复注册应该panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("对于重复错误码，MustRegisterErrorCodes应该panic")
		}
	}()

	MustRegisterErrorCodes([]ErrorCode{code2})
}

// 并发注册错误码的性能测试
func BenchmarkConcurrentRegister(b *testing.B) {
	// 重置计时器
	b.ResetTimer()

	// 创建一个WaitGroup来同步goroutine
	var wg sync.WaitGroup

	// 启动b.N个goroutine同时注册错误码
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			code := 5000 + i
			Register(NewErrorCode(code, http.StatusBadRequest, "并发测试错误", ""))
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
}

// 批量注册错误码的性能测试
func BenchmarkBatchRegisterErrorCodes(b *testing.B) {
	// 准备测试数据
	const batchSize = 100
	for i := 0; i < b.N; i++ {
		// 每次迭代创建一批错误码
		codes := make([]ErrorCode, batchSize)
		baseCode := 6000 + i*batchSize

		for j := 0; j < batchSize; j++ {
			codes[j] = NewErrorCode(baseCode+j, http.StatusBadRequest, "批量测试错误", "")
		}

		// 重置计时器（仅测量注册时间，不包括准备数据的时间）
		b.ResetTimer()

		// 批量注册
		RegisterErrorCodes(codes)

		// 停止计时器
		b.StopTimer()
	}
}
