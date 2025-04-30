// Package errors 提供简单且强大的错误处理原语。
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
package errors

import (
	"fmt"
	"io"
	"strings"
)

//=====================================================
// 基础接口定义
//=====================================================

// ContextualError 表示一个带有附加上下文的错误。
type ContextualError interface {
	error
	Cause() error
	Unwrap() error
}

// StackTracer 是可以提供堆栈跟踪的错误接口。
type StackTracer interface {
	StackTrace() StackTrace
}

// contextualError 表示一个统一的错误结构，带有可选的堆栈跟踪、原因和错误码。
type contextualError struct {
	msg   string
	err   error
	code  int
	stack *stack
}

// Error 返回错误消息。
func (ce *contextualError) Error() string { return ce.msg }

// Cause 返回错误的底层原因。
func (ce *contextualError) Cause() error { return ce.err }

// Unwrap 提供与Go 1.13+错误链的兼容性。
func (ce *contextualError) Unwrap() error { return ce.err }

// Code 返回错误码（如果已设置），否则返回0。
func (ce *contextualError) Code() int { return ce.code }

// StackTrace 返回堆栈跟踪。
func (ce *contextualError) StackTrace() StackTrace {
	if ce.stack == nil {
		return nil
	}
	return ce.stack.StackTrace()
}

// Format 实现fmt.Formatter接口，用于格式化打印错误。
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

//=====================================================
// 错误创建函数
//=====================================================

// New 返回带有提供消息和堆栈跟踪的错误。
// 它是标准库errors.New的直接替代品，额外记录堆栈跟踪。
//
// 示例:
//
//	err := errors.New("连接被拒绝")
func New(message string) error {
	return &contextualError{
		msg:   message,
		stack: callers(),
	}
}

// Errorf 根据格式说明符格式化并返回满足error的字符串值，同时记录堆栈跟踪。
// 它是fmt.Errorf的直接替代品，额外记录堆栈跟踪。
//
// 示例:
//
//	err := errors.Errorf("连接到%s被拒绝", hostname)
func Errorf(format string, args ...interface{}) error {
	return &contextualError{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(),
	}
}

//=====================================================
// 错误包装函数
//=====================================================

// WithStack 为err添加调用WithStack时的堆栈跟踪。
// 如果err为nil，WithStack返回nil。
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

	return &contextualError{
		msg:   err.Error(),
		err:   err,
		code:  code,
		stack: callers(),
	}
}

// Wrap 返回一个错误，它在调用Wrap的位置为err添加堆栈跟踪和提供的消息。
// 如果err为nil，Wrap返回nil。
//
// 示例:
//
//	err := db.Query()
//	if err != nil {
//	    return errors.Wrap(err, "查询数据库失败")
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

	return &contextualError{
		msg:   message,
		err:   err,
		code:  code,
		stack: callers(),
	}
}

// Wrapf 返回一个错误，它在调用Wrapf的位置为err添加堆栈跟踪和格式说明符。
// 如果err为nil，Wrapf返回nil。
//
// 示例:
//
//	err := db.Query()
//	if err != nil {
//	    return errors.Wrapf(err, "查询用户%s失败", username)
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

	return &contextualError{
		msg:   fmt.Sprintf(format, args...),
		err:   err,
		code:  code,
		stack: callers(),
	}
}

//=====================================================
// 错误码相关函数
//=====================================================

// NewWithCode 返回带有指定错误码和格式化消息的错误。
// 这对于API开发提供一致的错误码非常有用。
//
// 示例:
//
//	const NotFound = 404
//	return errors.NewWithCode(NotFound, "用户%s未找到", username)
func NewWithCode(code int, format string, args ...interface{}) error {
	return &contextualError{
		msg:   fmt.Sprintf(format, args...),
		code:  code,
		stack: callers(),
	}
}

// WrapWithCode 返回一个错误，它为err添加指定的错误码和格式化消息。
// 如果err为nil，WrapWithCode返回nil。
//
// 示例:
//
//	const DBError = 500
//	err := db.Query()
//	if err != nil {
//	    return errors.WrapWithCode(err, DBError, "数据库故障")
//	}
func WrapWithCode(err error, code int, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return &contextualError{
		msg:   fmt.Sprintf(format, args...),
		err:   err,
		code:  code,
		stack: callers(),
	}
}

// NewFromCode 从已有的ErrorCode创建一个带有该错误码的新错误。
// 这对于使用已注册的错误码创建标准错误特别有用。
//
// 示例:
//
//	notFoundCode := errors.GetErrorCode(404)
//	return errors.NewFromCodeInfo(notFoundCode)
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

// Cause 返回错误的底层原因（如果可能）。
// 如果错误实现了causer接口，则它有一个原因。
// 如果错误未实现Cause，将返回原始错误。
// 如果错误为nil，将不做进一步调查而返回nil。
//
// 示例:
//
//	err := errors.Wrap(originalErr, "额外上下文")
//	originalErr == errors.Cause(err) // true
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
