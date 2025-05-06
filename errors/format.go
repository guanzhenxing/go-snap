// Package errors 提供错误格式化功能。
package errors

import (
	"fmt"
	"sort"
	"strings"
)

// errorFormatInfo 包含用于格式化的基本错误信息。
type errorFormatInfo struct {
	code    int                    // 错误码
	message string                 // 格式化消息
	err     string                 // 原始错误消息
	stack   StackProvider          // 堆栈跟踪（如果有）
	context map[string]interface{} // 上下文信息
}

// formatDetailed 使用附加详细信息（包括堆栈跟踪）格式化错误。
// 当使用fmt.Printf("%+v", err)或fmt.Printf("%-v", err)时带有'+'或'-'标志，
// 用于提供丰富的错误信息。
//
// 参数:
//   - err: 要格式化的错误
//   - str: 用于写入格式化输出的字符串构建器
//   - showTrace: 如果为true，包含完整的错误链/跟踪
func formatDetailed(err error, str *strings.Builder, showTrace bool) {
	errs := unwrapErrorChain(err)

	for i, err := range errs {
		if i > 0 {
			str.WriteString("; ")
		}

		info := extractErrorFormatInfo(err)

		// 添加详细信息
		if info.stack != nil && info.stack.StackTrace() != nil && len(info.stack.StackTrace()) > 0 {
			f := info.stack.StackTrace()[0]
			fmt.Fprintf(str, "%s - [%s:%d (%s)] (%d)",
				info.err,
				f.file(),
				f.line(),
				f.name(),
				info.code,
			)
		} else {
			fmt.Fprintf(str, "%s (%d)", info.err, info.code)
		}

		// 添加上下文信息（如果有）
		if len(info.context) > 0 {
			str.WriteString(" {")

			// 获取排序后的键，使输出顺序一致
			keys := make([]string, 0, len(info.context))
			for k := range info.context {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for j, k := range keys {
				if j > 0 {
					str.WriteString(", ")
				}
				fmt.Fprintf(str, "%s: %v", k, info.context[k])
			}

			str.WriteString("}")
		}

		// 除非我们想要跟踪，否则在第一个错误后停止
		if !showTrace {
			break
		}
	}
}

// unwrapErrorChain 将错误堆栈转换为简单数组。
// 它解包嵌套错误以创建错误链中所有错误的平面列表。
func unwrapErrorChain(err error) []error {
	ret := []error{}

	if err != nil {
		if w, ok := err.(interface{ Unwrap() error }); ok {
			ret = append(ret, err)
			ret = append(ret, unwrapErrorChain(w.Unwrap())...)
		} else {
			ret = append(ret, err)
		}
	}

	return ret
}

// extractErrorFormatInfo 从错误中提取格式化信息。
// 它处理不同的错误类型并提取代码、消息和堆栈信息。
func extractErrorFormatInfo(err error) *errorFormatInfo {
	var info *errorFormatInfo

	switch e := err.(type) {
	case *contextualError:
		code := e.code
		var coder ErrorCode = UnknownError

		// 使用sync.Map的Load方法而不是map索引
		if coderVal, ok := errorCodeRegistry.Load(code); ok {
			coder = coderVal.(ErrorCode)
		}

		extMsg := coder.String()
		if extMsg == "" {
			extMsg = e.msg
		}

		info = &errorFormatInfo{
			code:    code,
			message: extMsg,
			err:     e.msg,
			stack:   e.stack,
			context: e.context, // 添加上下文信息
		}
	default:
		info = &errorFormatInfo{
			code:    UnknownError.Code(),
			message: err.Error(),
			err:     err.Error(),
		}
	}

	return info
}
