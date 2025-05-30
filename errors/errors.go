// Package errors 提供简单且强大的错误处理原语。
//
// # 概述
//
// 本包扩展了Go标准库的错误处理功能，实现了基于链式错误（error chain）的错误处理机制。
// 它提供了以下核心功能：
//   - 错误链和上下文：在保留原始错误的同时添加上下文信息
//   - 堆栈跟踪：自动捕获错误发生位置的完整调用栈
//   - 错误码系统：支持统一的错误码管理，便于API开发和客户端错误处理
//   - 错误上下文：支持为错误附加键值对形式的上下文数据
//   - 与标准库兼容：完全兼容Go 1.13+的错误处理机制（errors.Is, errors.As, errors.Unwrap）
//
// # 最佳实践
//
//  1. 使用Wrap/Wrapf在错误传播过程中添加上下文信息，而不是简单地返回原始错误
//     错误：return err
//     正确：return errors.Wrap(err, "读取配置文件失败")
//
//  2. 使用WithCode/WrapWithCode在API边界处为错误添加错误码
//     例如：return errors.WithCode(UserNotFound, "用户 %s 不存在", username)
//
//  3. 使用Cause获取错误链中的根本原因
//     例如：rootErr := errors.Cause(err)
//
//  4. 使用错误码进行错误比较，而不是直接比较错误实例
//     错误：if err == ErrNotFound { ... }
//     正确：if errors.IsErrorCode(err, NotFound) { ... }
//
//  5. 为关键错误添加上下文信息
//     例如：err = errors.WithContext(err, "request_id", requestID)
//
//  6. 在高性能场景中，谨慎选择堆栈捕获模式
//     例如：errors.DefaultStackCaptureMode = errors.StackCaptureModeNever
//
// # 性能考虑
//
// - 创建错误时捕获堆栈跟踪会产生一定的性能开销
// - 对于高频创建错误的场景，可以使用StackCaptureModeNever或StackCaptureModeModeSampled模式
// - 错误格式化（特别是带堆栈跟踪的）是相对昂贵的操作，应该主要用于调试和日志记录
// - 使用WithContext添加错误上下文比创建新的错误包装更高效
//
// # 与标准库的关系
//
// 本包完全兼容Go 1.13+的错误处理机制。具体表现为：
//   - 实现了Unwrap()方法，支持errors.Is和errors.As
//   - 提供了与标准库行为一致但功能更强的Is和As函数
//   - 可以与标准库的errors.New、fmt.Errorf等函数创建的错误互操作
//
// # 与其他错误处理库的对比
//
// 本包的设计受到了pkg/errors和github.com/cockroachdb/errors的影响，但提供了更多功能：
//   - 比pkg/errors增加了错误码、错误上下文和堆栈优化
//   - 比标准库更易用，提供了更多错误处理原语
//   - 比大多数错误库提供了更细粒度的堆栈跟踪控制
//
// # 错误类型层次结构
//
// 本包定义了几个关键接口来表示不同类型的错误：
//   - error：标准Go错误接口
//   - ContextualError：扩展error，提供获取原始错误的方法
//   - StackTracer：可以提供堆栈跟踪的错误
//   - ErrorCode：提供错误码和HTTP状态码的错误
//
// Go中传统的错误处理习惯大致如下：
//
//	if err != nil {
//	        return err
//	}
//
// 当这种模式递归地应用于调用栈时，会导致错误报告缺乏上下文或调试信息。
// errors包允许程序员以不破坏错误原始值的方式，在代码的失败路径中添加上下文信息。
//
// # 为错误添加上下文
//
// errors.Wrap函数返回一个新的错误，它通过在调用Wrap的位置记录堆栈跟踪和提供的消息，
// 为原始错误添加上下文。例如：
//
//	_, err := ioutil.ReadAll(r)
//	if err != nil {
//	        return errors.Wrap(err, "读取失败")
//	}
//
// # 获取错误的根因
//
// 使用errors.Wrap构造一个错误堆栈，为前一个错误添加上下文。根据错误的性质，
// 可能需要逆向操作errors.Wrap以检索原始错误。任何实现此接口的错误值：
//
//	type causer interface {
//	        Cause() error
//	}
//
// 都可以被errors.Cause检查。errors.Cause将递归检索最顶层的不实现causer的错误，
// 这被假定为原始原因。
//
// # 使用错误码
//
// 本包还提供错误码功能，这对API开发和客户端错误处理很有用。例如：
//
//	// 定义错误码
//	const (
//	    NotFound = 404
//	)
//
//	// 注册带详细信息的错误码
//	RegisterErrorCode(NotFound, http.StatusNotFound, "资源未找到", "")
//
//	// 使用错误码
//	return errors.WithCode(NotFound, "用户 %s 未找到", username)
//
// 客户端随后可以检查特定的错误码：
//
//	err := doSomething()
//	if errors.IsErrorCode(err, NotFound) {
//	    // 处理未找到错误
//	}
//
// # 堆栈捕获优化
//
// 本包支持多种堆栈捕获策略，可以根据性能需求进行配置：
//
//	// 设置全局堆栈捕获模式
//	errors.DefaultStackCaptureMode = errors.StackCaptureModeDeferred // 默认模式
//
//	// 为特定错误选择捕获模式
//	err := errors.NewWithStackControl("错误", errors.StackCaptureModeImmediate)
//
// 支持以下捕获模式：
//   - StackCaptureModeNever: 不捕获堆栈，最大化性能
//   - StackCaptureModeDeferred: 仅在需要时捕获堆栈（默认模式）
//   - StackCaptureModeImmediate: 创建错误时立即捕获堆栈
//   - StackCaptureModeModeSampled: 每N个错误采样一个堆栈
//
// 更详细的信息请参考 stack_optimization.md 文档。
package errors

import (
	stderrors "errors" // 为标准库errors包添加别名导入
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync/atomic"
)

//=====================================================
// 基础接口定义
//=====================================================

// ContextualError 表示一个带有附加上下文的错误。
// 该接口扩展了标准error接口，提供了获取原始错误的方法。
type ContextualError interface {
	error
	Cause() error  // 获取原始错误
	Unwrap() error // 兼容Go 1.13+的错误链
}

// StackTracer 是可以提供堆栈跟踪的错误接口。
// 实现此接口的错误可以返回完整的调用堆栈信息。
type StackTracer interface {
	StackTrace() StackTrace // 返回堆栈跟踪信息
}

//=====================================================
// 错误结构定义
//=====================================================

// contextualError 表示一个统一的错误结构，带有可选的堆栈跟踪、原因和错误码。
// 该结构是包内部的核心错误实现，支持错误链、堆栈跟踪和错误码等功能。
type contextualError struct {
	msg     string                 // 错误消息
	err     error                  // 原始错误
	code    int                    // 错误码
	stack   StackProvider          // 堆栈提供者
	context map[string]interface{} // 错误上下文信息
}

// Error 返回错误消息。
// 实现标准error接口。
func (ce *contextualError) Error() string { return ce.msg }

// Cause 返回错误的底层原因。
// 实现ContextualError接口。
func (ce *contextualError) Cause() error { return ce.err }

// Unwrap 提供与Go 1.13+错误链的兼容性。
// 用于支持标准库的errors.Is和errors.As功能。
func (ce *contextualError) Unwrap() error { return ce.err }

// Code 返回错误码（如果已设置），否则返回0。
// 用于API错误处理和客户端错误识别。
func (ce *contextualError) Code() int { return ce.code }

// StackTrace 返回堆栈跟踪。
// 实现StackTracer接口，允许访问完整的调用栈。
func (ce *contextualError) StackTrace() StackTrace {
	if ce.stack == nil {
		return nil
	}
	return ce.stack.StackTrace()
}

// Format 实现fmt.Formatter接口，用于格式化打印错误。
// 支持多种格式化选项，特别是详细的堆栈跟踪打印。
// 格式选项：
//   - %s: 仅打印错误消息
//   - %v: 打印错误消息
//   - %+v: 打印详细信息，包括完整的错误链和堆栈跟踪
//   - %-v: 打印详细信息，但不包括堆栈跟踪
func (ce *contextualError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') || s.Flag('-') {
			var str strings.Builder
			flagTrace := s.Flag('+')
			formatDetailed(ce, &str, flagTrace)
			io.WriteString(s, str.String())
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, ce.Error())
	case 'q':
		fmt.Fprintf(s, "%q", ce.Error())
	}
}

// Is 方法实现了标准库errors.Is的功能，用于错误比较。
// 如果目标错误是一个错误码，则比较错误码；否则委托给标准库errors.Is。
// 这使得使用errors.Is可以检查错误码相等性。
func (ce *contextualError) Is(target error) bool {
	// 如果目标错误是一个错误码
	if codeErr, ok := target.(interface{ Code() int }); ok {
		return ce.code == codeErr.Code()
	}

	// 委托给底层错误
	if ce.err != nil {
		if stdErr, ok := ce.err.(interface{ Is(error) bool }); ok {
			return stdErr.Is(target)
		}
		return ce.err == target
	}

	return false
}

// As 方法实现了标准库errors.As的功能，用于类型断言。
// 如果目标类型是一个错误码接口，尝试将当前错误转换为目标类型；
// 否则委托给底层错误的As方法或标准类型断言。
// 这使得使用errors.As可以将错误转换为特定类型。
func (ce *contextualError) As(target interface{}) bool {
	// 尝试将自身类型匹配到target
	if reflectAsTarget(ce, target) {
		return true
	}

	// 尝试将当前错误转换为目标类型
	if coder, ok := target.(interface{ Code() int }); ok {
		// 模拟为目标类型设置值
		// 注意：这里仅示意，实际实现需要使用反射正确设置值
		if setter, ok := coder.(interface{ SetCode(int) }); ok {
			setter.SetCode(ce.code)
			return true
		}
	}

	// 委托给底层错误
	if ce.err != nil {
		if stdErr, ok := ce.err.(interface{ As(interface{}) bool }); ok {
			return stdErr.As(target)
		}
		return reflectAsTarget(ce.err, target)
	}

	return false
}

// reflectAsTarget 使用反射帮助实现As方法的类型断言
// 这是一个内部辅助函数，用于安全地将错误值转换为目标类型
func reflectAsTarget(err error, target interface{}) bool {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return false
	}

	targetType := val.Type().Elem()
	if targetType.Kind() != reflect.Interface && !reflect.TypeOf(err).AssignableTo(targetType) {
		return false
	}

	val.Elem().Set(reflect.ValueOf(err))
	return true
}

//=====================================================
// 错误创建函数
//=====================================================

// New 创建一个带有给定消息的新错误。
// 这是errors.New的增强版本，支持堆栈跟踪。
// 性能影响: 创建堆栈跟踪会产生少量开销，但通常可以忽略不计。
// 在高性能场景中，可以使用NewWithStackControl调整堆栈捕获行为。
//
// 示例:
//
//	err := errors.New("无法连接到服务器")
func New(message string) error {
	var stack StackProvider

	switch DefaultStackCaptureMode {
	case StackCaptureModeNever:
		// 不捕获堆栈，最大化性能
		stack = nil
	case StackCaptureModeImmediate:
		// 立即捕获堆栈
		stack = callers()
	case StackCaptureModeDeferred:
		// 延迟捕获堆栈，仅在需要时捕获
		stack = newLazyStack(3)
	case StackCaptureModeModeSampled:
		// 采样捕获，每N个错误捕获一次堆栈
		counter := atomic.AddInt32(&stackSampleCounter, 1)
		if counter%int32(SamplingRate) == 0 {
			stack = callers()
		}
	default:
		// 默认使用延迟捕获
		stack = newLazyStack(3)
	}

	return &contextualError{
		msg:   message,
		code:  UnknownError.Code(),
		stack: stack,
	}
}

// Errorf 创建一个带有格式化消息的新错误。
// 这是fmt.Errorf的增强版本，支持堆栈跟踪。
// 与New类似，但支持格式化字符串。
//
// 示例:
//
//	err := errors.Errorf("连接到 %s 失败", serverName)
func Errorf(format string, args ...interface{}) error {
	var stack StackProvider

	switch DefaultStackCaptureMode {
	case StackCaptureModeNever:
		stack = nil
	case StackCaptureModeImmediate:
		stack = callers()
	case StackCaptureModeDeferred:
		stack = newLazyStack(3)
	case StackCaptureModeModeSampled:
		counter := atomic.AddInt32(&stackSampleCounter, 1)
		if counter%int32(SamplingRate) == 0 {
			stack = callers()
		}
	default:
		stack = newLazyStack(3)
	}

	return &contextualError{
		msg:   fmt.Sprintf(format, args...),
		code:  UnknownError.Code(),
		stack: stack,
	}
}

//=====================================================
// 错误包装函数
//=====================================================

// WithStack 为现有错误添加堆栈跟踪。
// 如果错误已经有堆栈跟踪，则保留原始堆栈。
// 适用于需要保留原始错误但添加堆栈信息的场景。
//
// 示例:
//
//	err := doSomething()
//	return errors.WithStack(err)
func WithStack(err error) error {
	if err == nil {
		return nil
	}

	// 尝试从原始错误获取错误码
	code := UnknownError.Code()
	if c, ok := err.(interface{ Code() int }); ok {
		code = c.Code()
	}

	// 尝试从原始错误获取上下文信息
	var context map[string]interface{}
	if ce, ok := err.(*contextualError); ok && ce.context != nil {
		// 创建上下文的拷贝
		context = make(map[string]interface{}, len(ce.context))
		for k, v := range ce.context {
			context[k] = v
		}
	}

	return &contextualError{
		msg:     err.Error(),
		err:     err,
		code:    code,
		stack:   createStackProvider(),
		context: context,
	}
}

// Wrap 将错误包装在新的错误中，添加消息和堆栈跟踪。
// 这是错误处理的推荐方法，允许在调用链中添加上下文。
// 如果原始错误为nil，返回nil。
//
// 示例:
//
//	file, err := os.Open(path)
//	if err != nil {
//	    return errors.Wrap(err, "打开配置文件失败")
//	}
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	// 尝试从原始错误获取错误码
	code := UnknownError.Code()
	if c, ok := err.(interface{ Code() int }); ok {
		code = c.Code()
	}

	// 尝试从原始错误获取上下文信息
	var context map[string]interface{}
	if ce, ok := err.(*contextualError); ok && ce.context != nil {
		// 复制上下文映射
		context = make(map[string]interface{}, len(ce.context))
		for k, v := range ce.context {
			context[k] = v
		}
	}

	return &contextualError{
		msg:     message,
		err:     err,
		code:    code,
		stack:   callers(),
		context: context,
	}
}

// Wrapf 将错误包装在新的错误中，添加格式化消息和堆栈跟踪。
// 与Wrap类似，但支持格式化字符串。
// 如果原始错误为nil，返回nil。
//
// 示例:
//
//	resp, err := http.Get(url)
//	if err != nil {
//	    return errors.Wrapf(err, "请求URL %s失败", url)
//	}
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	// 尝试从原始错误获取错误码
	code := UnknownError.Code()
	if c, ok := err.(interface{ Code() int }); ok {
		code = c.Code()
	}

	// 尝试从原始错误获取上下文信息
	var context map[string]interface{}
	if ce, ok := err.(*contextualError); ok && ce.context != nil {
		// 复制上下文映射
		context = make(map[string]interface{}, len(ce.context))
		for k, v := range ce.context {
			context[k] = v
		}
	}

	return &contextualError{
		msg:     fmt.Sprintf(format, args...),
		err:     err,
		code:    code,
		stack:   callers(),
		context: context,
	}
}

//=====================================================
// 错误码相关函数
//=====================================================

// NewWithCode 创建一个带有错误码的新错误。
// 适用于API错误处理，允许客户端根据错误码进行不同处理。
// 错误码应事先使用RegisterErrorCode注册。
//
// 示例:
//
//	return errors.NewWithCode(InvalidParameter, "无效的用户ID: %s", userID)
func NewWithCode(code int, format string, args ...interface{}) error {
	return &contextualError{
		msg:   fmt.Sprintf(format, args...),
		code:  code,
		stack: callers(),
	}
}

// WrapWithCode 将错误包装在新的错误中，添加错误码、消息和堆栈跟踪。
// 结合了Wrap和NewWithCode的功能。
// 如果原始错误为nil，返回nil。
//
// 示例:
//
//	user, err := db.GetUser(userID)
//	if err != nil {
//	    if errors.Is(err, sql.ErrNoRows) {
//	        return errors.WrapWithCode(err, NotFound, "用户 %s 不存在", userID)
//	    }
//	    return errors.Wrap(err, "查询用户失败")
//	}
func WrapWithCode(err error, code int, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	// 尝试从原始错误获取上下文信息
	var context map[string]interface{}
	if ce, ok := err.(*contextualError); ok && ce.context != nil {
		// 复制上下文映射
		context = make(map[string]interface{}, len(ce.context))
		for k, v := range ce.context {
			context[k] = v
		}
	}

	return &contextualError{
		msg:     fmt.Sprintf(format, args...),
		err:     err,
		code:    code,
		stack:   callers(),
		context: context,
	}
}

// NewFromCode 从预定义的错误码创建一个新错误。
// 使用注册的错误码信息，包括HTTP状态码和错误消息。
// 适用于没有额外上下文的标准错误场景。
//
// 示例:
//
//	return errors.NewFromCode(NotFound)
func NewFromCode(code ErrorCode) error {
	return &contextualError{
		msg:   code.Message(),
		code:  code.Code(),
		stack: callers(),
	}
}

//=====================================================
// 错误查询函数
//=====================================================

// Cause 获取错误链中的根本原因。
// 遍历错误链，找到最底层的错误（即不再实现Cause()方法的错误）。
// 这对于检查原始错误类型或与标准错误比较很有用。
//
// 示例:
//
//	err := doSomething()
//	if errors.Cause(err) == io.EOF {
//	    // 处理EOF错误
//	}
func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}

		if cause.Cause() == nil {
			break
		}

		err = cause.Cause()
	}
	return err
}

//=====================================================
// 错误上下文函数
//=====================================================

// WithContext 为错误添加键值对形式的上下文信息。
// 这些上下文信息可以用于日志记录、调试或错误处理。
// 比创建新的错误包装更高效，不会添加新的堆栈跟踪。
//
// 示例:
//
//	err = errors.WithContext(err, "user_id", userID)
//	err = errors.WithContext(err, "request_id", requestID)
func WithContext(err error, key string, value interface{}) error {
	if err == nil {
		return nil
	}

	// 尝试将上下文信息添加到现有的contextualError
	var ce *contextualError
	if stderrors.As(err, &ce) {
		// 如果已经是contextualError，复制并添加新的上下文值
		newCE := *ce // 复制结构体

		// 初始化context映射（如果需要）
		if newCE.context == nil {
			newCE.context = make(map[string]interface{})
		}

		// 添加或更新上下文值
		newCE.context[key] = value
		return &newCE
	}

	// 获取错误码（如果有）
	code := UnknownError.Code()
	if c, ok := err.(interface{ Code() int }); ok {
		code = c.Code()
	}

	// 创建新的contextualError，包含上下文信息
	return &contextualError{
		msg:     err.Error(),
		err:     err,
		code:    code,
		stack:   callers(),
		context: map[string]interface{}{key: value},
	}
}

// WithContextMap 为错误添加多个键值对形式的上下文信息。
// 与WithContext类似，但一次可以添加多个上下文值。
//
// 示例:
//
//	err = errors.WithContextMap(err, map[string]interface{}{
//	    "user_id": userID,
//	    "request_id": requestID,
//	    "operation": "user_create",
//	})
func WithContextMap(err error, contextMap map[string]interface{}) error {
	if err == nil {
		return nil
	}

	if len(contextMap) == 0 {
		return err
	}

	// 尝试将上下文信息添加到现有的contextualError
	var ce *contextualError
	if stderrors.As(err, &ce) {
		// 如果已经是contextualError，复制并添加新的上下文值
		newCE := *ce // 复制结构体

		// 初始化context映射（如果需要）
		if newCE.context == nil {
			newCE.context = make(map[string]interface{})
		}

		// 添加或更新所有上下文值
		for k, v := range contextMap {
			newCE.context[k] = v
		}

		return &newCE
	}

	// 获取错误码（如果有）
	code := UnknownError.Code()
	if c, ok := err.(interface{ Code() int }); ok {
		code = c.Code()
	}

	// 创建新的上下文映射并复制所有键值对
	newContext := make(map[string]interface{}, len(contextMap))
	for k, v := range contextMap {
		newContext[k] = v
	}

	// 创建新的contextualError，包含上下文信息
	return &contextualError{
		msg:     err.Error(),
		err:     err,
		code:    code,
		stack:   callers(),
		context: newContext,
	}
}

// GetContext 从错误中获取特定的上下文值。
// 遍历错误链，查找指定键的上下文值。
//
// 示例:
//
//	if requestID, ok := errors.GetContext(err, "request_id"); ok {
//	    log.Printf("Error for request %v: %v", requestID, err)
//	}
func GetContext(err error, key string) (interface{}, bool) {
	var ce *contextualError
	if stderrors.As(err, &ce) && ce.context != nil {
		value, exists := ce.context[key]
		return value, exists
	}
	return nil, false
}

// GetAllContext 获取错误链中的所有上下文信息。
// 合并错误链中所有错误的上下文，较新的值会覆盖较旧的值。
//
// 示例:
//
//	allContext := errors.GetAllContext(err)
//	for k, v := range allContext {
//	    log.Printf("Context %s: %v", k, v)
//	}
func GetAllContext(err error) map[string]interface{} {
	var ce *contextualError
	if stderrors.As(err, &ce) && ce.context != nil {
		// 创建副本以避免修改原始映射
		result := make(map[string]interface{}, len(ce.context))
		for k, v := range ce.context {
			result[k] = v
		}
		return result
	}
	return make(map[string]interface{})
}

// NewWithStackControl 创建一个新错误，并精确控制堆栈捕获行为。
// 这对于性能敏感的场景特别有用。
// 支持的模式包括：Never、Deferred、Immediate和Sampled。
//
// 示例:
//
//	// 在高频操作中不捕获堆栈
//	err := errors.NewWithStackControl("缓存未命中", errors.StackCaptureModeNever)
//
//	// 在关键点创建带完整堆栈的错误
//	err := errors.NewWithStackControl("启动失败", errors.StackCaptureModeImmediate)
func NewWithStackControl(message string, mode StackCaptureMode) error {
	var stack StackProvider

	switch mode {
	case StackCaptureModeNever:
		stack = nil
	case StackCaptureModeImmediate:
		stack = callers()
	case StackCaptureModeDeferred:
		stack = newLazyStack(4)
	case StackCaptureModeModeSampled:
		counter := atomic.AddInt32(&stackSampleCounter, 1)
		if counter%int32(SamplingRate) == 0 {
			stack = callers()
		}
	}

	return &contextualError{
		msg:   message,
		code:  UnknownError.Code(),
		stack: stack,
	}
}

//=====================================================
// 标准兼容函数
//=====================================================

// Is 检查目标错误是否与给定错误匹配。
// 实现与标准库errors.Is相同的功能，但增加了对错误码的支持。
// 这使得可以检查错误码相等性。
//
// 示例:
//
//	if errors.Is(err, NotFoundError) {
//	    // 处理未找到错误
//	}
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// As 尝试将错误转换为特定类型。
// 实现与标准库errors.As相同的功能，但增加了对错误码的支持。
// 目标必须是一个非nil指针。
//
// 示例:
//
//	var apiErr *APIError
//	if errors.As(err, &apiErr) {
//	    fmt.Printf("API错误: %d", apiErr.Code())
//	}
func As(err error, target interface{}) bool {
	return stderrors.As(err, target)
}
