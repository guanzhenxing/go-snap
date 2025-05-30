// Package errors 提供带有错误码的错误处理功能。
//
// # 错误码系统设计
//
// 错误码系统设计用于标准化错误处理，特别适合构建API和微服务架构。核心设计理念：
//
// 1. 唯一性：每个错误码都是唯一的整数值，便于识别和处理特定错误类型
// 2. 自描述：错误码包含HTTP状态码、消息和参考文档URL等元数据
// 3. 全局注册：错误码在全局注册表中注册，确保一致性和可发现性
// 4. 国际化支持：错误消息可以通过国际化系统进行本地化
// 5. 与HTTP集成：每个错误码都有关联的HTTP状态码，便于在API中使用
//
// # 错误码分配原则
//
// 建议按照以下原则分配错误码：
//
// - 保留错误码1用于UnknownError（框架已实现）
// - 错误码2-999保留给框架使用
// - 错误码1000-9999用于应用级错误
// - 可按功能模块或服务划分错误码范围，例如：
//   - 1000-1999：用户和认证相关错误
//   - 2000-2999：资源相关错误
//   - 3000-3999：业务流程错误
//
// # 最佳实践
//
// 1. 在应用初始化时注册所有错误码：
//
//	const (
//	    // 用户相关错误（1000-1099）
//	    UserNotFound = 1000 + iota
//	    InvalidCredentials
//	    AccountLocked
//
//	    // 资源相关错误（2000-2099）
//	    ResourceNotFound = 2000 + iota
//	    ResourceAlreadyExists
//	    ResourceLocked
//	)
//
//	func init() {
//	    errors.RegisterErrorCode(UserNotFound, http.StatusNotFound, "用户不存在", "")
//	    errors.RegisterErrorCode(InvalidCredentials, http.StatusUnauthorized, "无效的凭证", "")
//	    // ...其他错误码
//	}
//
// 2. 在API处理中使用错误码：
//
//	user, err := userService.GetUser(userID)
//	if err != nil {
//	    // 根据错误类型返回适当的错误码
//	    if errors.Is(err, sql.ErrNoRows) {
//	        return errors.WrapWithCode(err, UserNotFound, "用户 %s 不存在", userID)
//	    }
//	    return errors.Wrap(err, "获取用户失败")
//	}
//
// 3. 在API响应中使用错误码：
//
//	type ErrorResponse struct {
//	    Code    int    `json:"code"`
//	    Message string `json:"message"`
//	    Details string `json:"details,omitempty"`
//	}
//
//	func WriteErrorResponse(w http.ResponseWriter, err error) {
//	    code := errors.GetErrorCodeFromError(err)
//	    resp := ErrorResponse{
//	        Code:    code.Code(),
//	        Message: code.Message(),
//	        Details: err.Error(),
//	    }
//	    w.WriteHeader(code.HTTPStatus())
//	    json.NewEncoder(w).Encode(resp)
//	}
//
// 4. 在客户端处理错误码：
//
//	resp, err := http.Get(apiURL)
//	if err != nil {
//	    return err
//	}
//	if resp.StatusCode != http.StatusOK {
//	    var errResp ErrorResponse
//	    json.NewDecoder(resp.Body).Decode(&errResp)
//
//	    switch errResp.Code {
//	    case UserNotFound:
//	        // 处理用户不存在
//	    case InvalidCredentials:
//	        // 处理凭证无效
//	    default:
//	        // 处理其他错误
//	    }
//	}
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
// 该接口用于统一错误码的处理方式，提供错误码、HTTP状态码和错误消息等信息。
// 实现此接口的类型可以用于错误码注册、错误创建和错误响应生成。
type ErrorCode interface {
	// HTTPStatus 返回应与错误码关联的HTTP状态码。
	// 用于在Web API中将错误转换为正确的HTTP响应。
	// 返回值应是标准HTTP状态码，如200、400、404、500等。
	HTTPStatus() int

	// String 返回面向用户的错误文本。
	// 通常包含错误码和描述性消息。
	// 格式为"错误码: 错误消息"，如"404: 资源未找到"。
	String() string

	// Message 是String的别名，返回面向用户的错误文本。
	// 提供更语义化的访问方式。
	// 返回纯错误消息，不包含错误码。
	Message() string

	// Reference 返回错误的详细文档链接。
	// 用于提供更详细的错误解释和解决方案。
	// 例如指向API文档或问题解决指南的URL。
	Reference() string

	// Code 返回错误码的整数值。
	// 用于唯一标识错误类型。
	// 由应用定义，应确保在应用内唯一。
	Code() int
}

// standardErrorCode 实现了ErrorCodeInfo接口。
// 这是ErrorCode接口的标准实现，用于创建和管理错误码。
// 作为默认实现，它提供了所有必要的错误码元数据。
type standardErrorCode struct {
	// ErrorCode 指错误码的整数值。
	// 用于唯一标识错误类型。
	ErrorCode int

	// HTTPStatusCode 指与错误码关联的HTTP状态码。
	// 用于在Web API中返回正确的HTTP状态码。
	HTTPStatusCode int

	// ExternalMessage 指面向用户的错误文本。
	// 这是对外展示的错误消息，应该简洁明了。
	ExternalMessage string

	// ReferenceURL 指定参考文档的链接。
	// 提供详细的错误文档URL，可以为空。
	ReferenceURL string
}

// Code 返回错误码的整数值。
// 实现ErrorCode接口的方法。
func (sec standardErrorCode) Code() int {
	return sec.ErrorCode
}

// String 实现了stringer接口。返回面向用户的错误消息。
// 格式为"错误码: 错误消息"。
// 便于日志记录和调试输出。
func (sec standardErrorCode) String() string {
	return fmt.Sprintf("%d: %s", sec.ErrorCode, sec.ExternalMessage)
}

// Message 返回面向用户的错误消息，与String相同。
// 提供更语义化的访问方式。
// 不包含错误码前缀，只返回纯文本消息。
func (sec standardErrorCode) Message() string {
	return sec.ExternalMessage
}

// HTTPStatus 返回与错误码关联的HTTP状态码。
// 如果未设置，则返回500（内部服务器错误）。
// 用于在Web API中返回正确的HTTP状态码。
func (sec standardErrorCode) HTTPStatus() int {
	if sec.HTTPStatusCode == 0 {
		return 500
	}

	return sec.HTTPStatusCode
}

// Reference 返回参考文档的链接。
// 提供错误的详细文档URL。
// 如果未设置，则返回空字符串。
func (sec standardErrorCode) Reference() string {
	return sec.ReferenceURL
}

// 错误码注册表使用sync.Map，提高并发性能
// 用于全局存储和检索错误码信息
// 使用sync.Map而不是普通map是为了保证并发安全性
var errorCodeRegistry sync.Map

// UnknownError 是默认的错误码，用于没有特定错误码的错误。
// 预定义的错误码，用于表示未知或未分类的错误。
// 错误码1保留给未知错误，不应被应用重新定义。
var (
	UnknownError ErrorCode = NewErrorCode(1, http.StatusInternalServerError, "发生了内部服务器错误", "http://github.com/guanzhenxing/go-snap/errors/README.md")
)

// 初始化函数，在包加载时执行
// 注册默认的错误码
// 这确保了即使应用没有注册任何错误码，框架也能正常工作
func init() {
	// 使用未知错误初始化注册表
	Register(UnknownError)
}

//=====================================================
// 错误码创建函数
//=====================================================

// NewErrorCode 创建一个新的错误码信息对象
// 接收错误码、HTTP状态码、错误消息和参考文档URL
// 参数：
//   - code: 错误码整数值，应在应用内唯一
//   - httpStatus: HTTP状态码，如http.StatusNotFound (404)
//   - message: 面向用户的错误消息
//   - reference: 错误文档URL，可选
//
// 返回：
//   - ErrorCode: 创建的错误码对象
//
// 示例：
//
//	NotFound := errors.NewErrorCode(1001, http.StatusNotFound, "资源未找到", "")
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
// 这是非强制验证的注册方法，适用于不关心冲突的场景。
//
// 参数：
//   - ec: 要注册的错误码对象
//
// 返回：
//   - ErrorCode: 注册的错误码对象，用于链式调用
//
// 特性：
//   - 会保护错误码1（UnknownError），防止被覆盖
//   - 会自动更新全局错误码注册表
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
	// 保护错误码1，确保它始终表示UnknownError
	if ec.Code() == 1 && ec != UnknownError {
		panic("错误码 `1` 已被 `go-snap` 保留用作 UnknownError 错误码")
	}

	// 将错误码存储到全局注册表
	errorCodeRegistry.Store(ec.Code(), ec)
	return ec
}

// RegisterErrorCodes 批量注册错误码，提高初始化效率
// 这对于在应用启动时一次性注册大量错误码特别有用
//
// 参数：
//   - codes: 要注册的错误码对象数组
//
// 特性：
//   - 会保护错误码1（UnknownError），防止被覆盖
//   - 会按顺序注册所有错误码
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
// 用于确保错误码的唯一性和一致性。
//
// 参数：
//   - ec: 要注册的错误码对象
//
// 返回：
//   - ErrorCode: 注册的错误码对象，用于链式调用
//
// 特性：
//   - 会检查错误码是否已存在，存在则引发panic
//   - 会保护错误码1（UnknownError），防止被覆盖
//
// 示例:
//
//	func init() {
//	    // 如果错误码已存在，会引发panic
//	    errors.MustRegister(errors.NewErrorCode(InvalidParameter, http.StatusBadRequest, "无效参数", ""))
//	}
func MustRegister(ec ErrorCode) ErrorCode {
	// 特殊处理UnknownError
	if ec.Code() == 1 && ec != UnknownError {
		panic("错误码 '1' 已被 'go-snap' 保留用作 UnknownError 错误码")
	}

	// 检查错误码是否已存在
	if _, exists := errorCodeRegistry.Load(ec.Code()); exists {
		panic(fmt.Sprintf("错误码: %d 已存在", ec.Code()))
	}

	// 注册错误码
	errorCodeRegistry.Store(ec.Code(), ec)
	return ec
}

// MustRegisterErrorCodes 批量注册错误码，如果有任何错误码已存在则会引发panic
// 这对于确保应用程序启动时不会出现错误码冲突特别有用
//
// 参数：
//   - codes: 要注册的错误码对象数组
//
// 特性：
//   - 会检查所有错误码是否已存在，任一存在则引发panic
//   - 全部检查通过后才会进行注册
//
// 示例:
//
//	func init() {
//	    // 如果任一错误码已存在，会引发panic
//	    errors.MustRegisterErrorCodes([]ErrorCode{
//	        errors.NewErrorCode(InvalidParameter, http.StatusBadRequest, "无效参数", ""),
//	        errors.NewErrorCode(ResourceNotFound, http.StatusNotFound, "资源未找到", ""),
//	    })
//	}
func MustRegisterErrorCodes(codes []ErrorCode) {
	// 先检查所有错误码
	for _, code := range codes {
		if code.Code() == 1 && code != UnknownError {
			panic("错误码 '1' 已被 'go-snap' 保留用作 UnknownError 错误码")
		}

		if _, exists := errorCodeRegistry.Load(code.Code()); exists {
			panic(fmt.Sprintf("错误码: %d 已存在", code.Code()))
		}
	}

	// 然后注册所有错误码
	for _, code := range codes {
		errorCodeRegistry.Store(code.Code(), code)
	}
}

// RegisterErrorCode 创建并注册一个错误码。
// 这是一个便捷方法，结合了NewErrorCode和Register的功能。
//
// 参数：
//   - code: 错误码整数值
//   - httpStatus: HTTP状态码
//   - message: 错误消息
//   - reference: 参考文档URL
//
// 返回：
//   - ErrorCode: 创建并注册的错误码对象
//
// 示例:
//
//	NotFound := errors.RegisterErrorCode(1001, http.StatusNotFound, "资源未找到", "")
func RegisterErrorCode(code int, httpStatus int, message string, reference string) ErrorCode {
	ec := NewErrorCode(code, httpStatus, message, reference)
	return Register(ec)
}

// MustRegisterErrorCode 创建并注册一个错误码，如果错误码已存在则会引发panic。
// 这是一个便捷方法，结合了NewErrorCode和MustRegister的功能。
//
// 参数：
//   - code: 错误码整数值
//   - httpStatus: HTTP状态码
//   - message: 错误消息
//   - reference: 参考文档URL
//
// 返回：
//   - ErrorCode: 创建并注册的错误码对象
//
// 示例:
//
//	// 如果错误码1001已存在，会引发panic
//	NotFound := errors.MustRegisterErrorCode(1001, http.StatusNotFound, "资源未找到", "")
func MustRegisterErrorCode(code int, httpStatus int, message string, reference string) ErrorCode {
	ec := NewErrorCode(code, httpStatus, message, reference)
	return MustRegister(ec)
}

//=====================================================
// 错误码查询函数
//=====================================================

// GetErrorCodeFromError 从错误中提取错误码信息。
// 如果错误实现了Code()方法，则返回对应的注册错误码；
// 否则返回UnknownError。
//
// 参数：
//   - err: 要检查的错误
//
// 返回：
//   - ErrorCode: 错误关联的错误码对象，未找到则返回UnknownError
//
// 示例:
//
//	err := doSomething()
//	errorCode := errors.GetErrorCodeFromError(err)
//	fmt.Printf("错误码: %d, HTTP状态: %d\n", errorCode.Code(), errorCode.HTTPStatus())
func GetErrorCodeFromError(err error) ErrorCode {
	if err == nil {
		return nil
	}

	// 检查错误是否提供了Code方法
	if coder, ok := err.(interface{ Code() int }); ok {
		code := coder.Code()
		if code != 0 {
			// 尝试从注册表获取错误码信息
			if ec, ok := errorCodeRegistry.Load(code); ok {
				return ec.(ErrorCode)
			}
		}
	}

	// 默认返回未知错误
	return UnknownError
}

// IsErrorCode 检查错误是否与指定错误码相关联。
// 如果错误具有相同的错误码，则返回true。
//
// 参数：
//   - err: 要检查的错误
//   - code: 要比较的错误码整数值
//
// 返回：
//   - bool: 如果错误具有指定错误码，则为true
//
// 示例:
//
//	err := doSomething()
//	if errors.IsErrorCode(err, ResourceNotFound) {
//	    // 处理资源未找到错误
//	} else if errors.IsErrorCode(err, InvalidParameter) {
//	    // 处理无效参数错误
//	}
func IsErrorCode(err error, code int) bool {
	if err == nil {
		return false
	}

	// 检查错误是否提供了Code方法
	if coder, ok := err.(interface{ Code() int }); ok {
		return coder.Code() == code
	}

	// 错误没有关联错误码
	return false
}

// ListErrorCodes 返回所有已注册的错误码的列表。
// 便于查看和调试所有已注册的错误码。
//
// 返回：
//   - []int: 所有已注册错误码的整数值列表
//
// 示例:
//
//	allCodes := errors.ListErrorCodes()
//	for _, code := range allCodes {
//	    fmt.Printf("已注册错误码: %d\n", code)
//	}
func ListErrorCodes() []int {
	var codes []int
	errorCodeRegistry.Range(func(key, value interface{}) bool {
		if code, ok := key.(int); ok {
			codes = append(codes, code)
		}
		return true
	})
	return codes
}

// GetErrorCode 根据错误码整数值获取对应的错误码对象。
// 如果找不到对应的错误码，返回UnknownError。
//
// 参数：
//   - code: 错误码整数值
//
// 返回：
//   - ErrorCode: 对应的错误码对象，未找到则返回UnknownError
//
// 示例:
//
//	errorCode := errors.GetErrorCode(ResourceNotFound)
//	fmt.Printf("错误消息: %s\n", errorCode.Message())
func GetErrorCode(code int) ErrorCode {
	if v, ok := errorCodeRegistry.Load(code); ok {
		return v.(ErrorCode)
	}
	return UnknownError
}

// HasErrorCode 检查指定的错误码是否已注册。
// 用于验证错误码是否存在。
//
// 参数：
//   - code: 要检查的错误码整数值
//
// 返回：
//   - bool: 如果错误码已注册，则为true
//
// 示例:
//
//	if errors.HasErrorCode(ResourceNotFound) {
//	    // 错误码已注册
//	} else {
//	    // 错误码未注册
//	}
func HasErrorCode(code int) bool {
	_, ok := errorCodeRegistry.Load(code)
	return ok
}
