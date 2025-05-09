// Package errors 提供带有错误码的错误处理功能。
package errors

import (
	"fmt"
	"net/http"
	"sync"
)

//=====================================================
// 基础定义
//=====================================================

// ErrorCode 定义了错误码详细信息的接口。
type ErrorCode interface {
	// HTTPStatus 返回应与错误码关联的HTTP状态码。
	HTTPStatus() int

	// String 返回面向用户的错误文本。
	String() string

	// Message 是String的别名，返回面向用户的错误文本。
	Message() string

	// Reference 返回错误的详细文档链接。
	Reference() string

	// Code 返回错误码的整数值。
	Code() int
}

// standardErrorCode 实现了ErrorCodeInfo接口。
type standardErrorCode struct {
	// ErrorCode 指错误码的整数值。
	ErrorCode int

	// HTTPStatusCode 指与错误码关联的HTTP状态码。
	HTTPStatusCode int

	// ExternalMessage 指面向用户的错误文本。
	ExternalMessage string

	// ReferenceURL 指定参考文档的链接。
	ReferenceURL string
}

// Code 返回错误码的整数值。
func (sec standardErrorCode) Code() int {
	return sec.ErrorCode
}

// String 实现了stringer接口。返回面向用户的错误消息。
func (sec standardErrorCode) String() string {
	return fmt.Sprintf("%d: %s", sec.ErrorCode, sec.ExternalMessage)
}

// Message 返回面向用户的错误消息，与String相同。
func (sec standardErrorCode) Message() string {
	return sec.ExternalMessage
}

// HTTPStatus 返回与错误码关联的HTTP状态码。
// 如果未设置，则返回500（内部服务器错误）。
func (sec standardErrorCode) HTTPStatus() int {
	if sec.HTTPStatusCode == 0 {
		return 500
	}

	return sec.HTTPStatusCode
}

// Reference 返回参考文档的链接。
func (sec standardErrorCode) Reference() string {
	return sec.ReferenceURL
}

// 错误码注册表使用sync.Map，提高并发性能
var errorCodeRegistry sync.Map

// UnknownError 是默认的错误码，用于没有特定错误码的错误。
var (
	UnknownError ErrorCode = NewErrorCode(1, http.StatusInternalServerError, "发生了内部服务器错误", "http://github.com/guanzhenxing/go-snap/errors/README.md")
)

func init() {
	// 使用未知错误初始化注册表
	Register(UnknownError)
}

//=====================================================
// 错误码创建函数
//=====================================================

// NewErrorCode 创建一个新的错误码信息对象
func NewErrorCode(code int, httpStatus int, message string, reference string) ErrorCode {
	return standardErrorCode{
		ErrorCode:       code,
		HTTPStatusCode:  httpStatus,
		ExternalMessage: message,
		ReferenceURL:    reference,
	}
}

//=====================================================
// 错误码注册函数
//=====================================================

// Register 注册用户定义的错误码。
// 如果错误码已存在，将覆盖现有错误码。
//
// 示例:
//
//	const (
//	    // InvalidParameter - 1001: 无效参数错误
//	    InvalidParameter int = iota + 1001
//	    // ResourceNotFound - 1002: 资源未找到
//	    ResourceNotFound
//	)
//
//	func init() {
//	    errors.NewErrorCode(InvalidParameter, http.StatusBadRequest, "无效参数", "")
//	    errors.NewErrorCode(ResourceNotFound, http.StatusNotFound, "资源未找到", "")
//	}
func Register(ec ErrorCode) ErrorCode {
	// 特殊处理UnknownError
	if ec.Code() == 1 && ec != UnknownError {
		panic("错误码 `1` 已被 `go-snap` 保留用作 UnknownError 错误码")
	}

	errorCodeRegistry.Store(ec.Code(), ec)
	return ec
}

// RegisterErrorCodes 批量注册错误码，提高初始化效率
// 这对于在应用启动时一次性注册大量错误码特别有用
//
// 示例:
//
//	func init() {
//	    errors.RegisterErrorCodes([]ErrorCode{
//	        errors.NewErrorCode(InvalidParameter, http.StatusBadRequest, "无效参数", ""),
//	        errors.NewErrorCode(ResourceNotFound, http.StatusNotFound, "资源未找到", ""),
//	        // 更多错误码...
//	    })
//	}
func RegisterErrorCodes(codes []ErrorCode) {
	for _, code := range codes {
		Register(code)
	}
}

// MustRegister 注册用户定义的错误码，如果错误码已存在则会引发panic。
// 这对于在包init函数中定义的错误码特别有用。
func MustRegister(ec ErrorCode) ErrorCode {
	// 特殊处理UnknownError
	if ec.Code() == 1 && ec != UnknownError {
		panic("错误码 '1' 已被 'go-snap' 保留用作 UnknownError 错误码")
	}

	if _, exists := errorCodeRegistry.Load(ec.Code()); exists {
		panic(fmt.Sprintf("错误码: %d 已存在", ec.Code()))
	}

	errorCodeRegistry.Store(ec.Code(), ec)
	return ec
}

// MustRegisterErrorCodes 批量注册错误码，如果有任何错误码已存在则会引发panic
// 这对于确保应用程序启动时不会出现错误码冲突特别有用
//
// 示例:
//
//	func init() {
//	    errors.MustRegisterErrorCodes([]ErrorCode{
//	        errors.NewErrorCode(InvalidParameter, http.StatusBadRequest, "无效参数", ""),
//	        errors.NewErrorCode(ResourceNotFound, http.StatusNotFound, "资源未找到", ""),
//	        // 更多错误码...
//	    })
//	}
func MustRegisterErrorCodes(codes []ErrorCode) {
	for _, code := range codes {
		MustRegister(code)
	}
}

// RegisterErrorCode 创建并注册一个错误码
func RegisterErrorCode(code int, httpStatus int, message string, reference string) ErrorCode {
	ec := NewErrorCode(code, httpStatus, message, reference)
	return Register(ec)
}

// MustRegisterErrorCode 创建并注册一个错误码，如果已存在则panic
func MustRegisterErrorCode(code int, httpStatus int, message string, reference string) ErrorCode {
	ec := NewErrorCode(code, httpStatus, message, reference)
	return MustRegister(ec)
}

//=====================================================
// 错误码查询函数
//=====================================================

// GetErrorCodeFromError 将任何错误解析为ErrorCodeInfo接口。
// nil错误将直接返回nil。
// 非错误码错误将被解析为UnknownError。
//
// 示例:
//
//	err := doSomething()
//	coder := GetErrorCodeFromError(err)
//	fmt.Printf("错误码: %d, HTTP状态: %d\n", coder.Code(), coder.HTTPStatus())
func GetErrorCodeFromError(err error) ErrorCode {
	if err == nil {
		return nil
	}

	if v, ok := err.(interface{ Code() int }); ok {
		if coder, ok := errorCodeRegistry.Load(v.Code()); ok {
			return coder.(ErrorCode)
		}
	}

	return UnknownError
}

// IsErrorCode 检查错误链中是否包含给定的错误码。
//
// 示例:
//
//	const ResourceNotFound = 1002
//
//	err := doSomething()
//	if errors.IsErrorCode(err, ResourceNotFound) {
//	    // 处理未找到错误
//	}
func IsErrorCode(err error, code int) bool {
	if v, ok := err.(interface{ Code() int }); ok {
		if v.Code() == code {
			return true
		}
	}

	if v, ok := err.(interface{ Unwrap() error }); ok && v.Unwrap() != nil {
		return IsErrorCode(v.Unwrap(), code)
	}

	return false
}

// ListErrorCodes 返回所有已注册错误码的列表。
func ListErrorCodes() []int {
	var codeList []int

	errorCodeRegistry.Range(func(key, value interface{}) bool {
		codeList = append(codeList, key.(int))
		return true
	})

	return codeList
}

// GetErrorCode 返回指定错误码的ErrorCodeInfo。
// 如果错误码未注册，则返回nil。
func GetErrorCode(code int) ErrorCode {
	if coder, ok := errorCodeRegistry.Load(code); ok {
		return coder.(ErrorCode)
	}
	return nil
}

// HasErrorCode 检查指定的错误码是否已注册。
func HasErrorCode(code int) bool {
	_, ok := errorCodeRegistry.Load(code)
	return ok
}
