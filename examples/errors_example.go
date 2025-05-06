package main

import (
	stderrors "errors"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/guanzhenxing/go-snap/errors"
)

// 定义应用错误码
const (
	// 用户相关错误 (1000-1999)
	ErrUserNotFound    = 1001
	ErrInvalidPassword = 1002

	// 数据库相关错误 (2000-2999)
	ErrDatabaseConnection = 2001
	ErrDuplicateKey       = 2002

	// API相关错误 (3000-3999)
	ErrInvalidInput      = 3001
	ErrUnauthorized      = 3002
	ErrRateLimitExceeded = 3003
)

// 初始化并注册所有错误码
func init() {
	// 注册用户相关错误码
	errors.RegisterErrorCode(ErrUserNotFound, http.StatusNotFound, "用户未找到", "https://docs.example.com/errors/user-not-found")
	errors.RegisterErrorCode(ErrInvalidPassword, http.StatusBadRequest, "密码无效", "https://docs.example.com/errors/invalid-password")

	// 注册数据库相关错误码
	errors.RegisterErrorCode(ErrDatabaseConnection, http.StatusInternalServerError, "数据库连接失败", "https://docs.example.com/errors/db-connection")
	errors.RegisterErrorCode(ErrDuplicateKey, http.StatusConflict, "记录已存在", "https://docs.example.com/errors/duplicate-key")

	// 注册API相关错误码
	errors.RegisterErrorCode(ErrInvalidInput, http.StatusBadRequest, "输入无效", "")
	errors.RegisterErrorCode(ErrUnauthorized, http.StatusUnauthorized, "未授权访问", "")
	errors.RegisterErrorCode(ErrRateLimitExceeded, http.StatusTooManyRequests, "请求速率超限", "")
}

// 模拟用户服务
type UserService struct{}

// 模拟根据ID查找用户
func (s *UserService) FindByID(id int) (string, error) {
	// 模拟用户不存在的情况
	if id <= 0 {
		// 使用错误码创建错误
		return "", errors.NewWithCode(ErrUserNotFound, "用户ID %d 不存在", id)
	}

	// 模拟数据库错误
	if id == 999 {
		// 创建基础错误
		dbErr := stderrors.New("连接超时")
		// 包装错误并添加错误码
		return "", errors.WrapWithCode(dbErr, ErrDatabaseConnection, "获取用户数据失败")
	}

	return fmt.Sprintf("用户-%d", id), nil
}

// 模拟验证用户密码
func (s *UserService) ValidatePassword(id int, password string) error {
	// 错误密码
	if password == "wrong" {
		// 使用基本错误创建
		baseErr := errors.New("密码校验失败")
		// 添加错误上下文
		return errors.WithContext(baseErr, "user_id", id)
	}

	// 空密码
	if password == "" {
		// 使用格式化错误
		return errors.Errorf("用户 %d 提供了空密码", id)
	}

	return nil
}

// 模拟API处理函数
func HandleUserRequest(userID int, requestID string) error {
	userService := &UserService{}

	// 查找用户
	username, err := userService.FindByID(userID)
	if err != nil {
		// 为错误添加请求上下文
		return errors.WithContextMap(err, map[string]interface{}{
			"request_id": requestID,
			"timestamp":  "2023-05-06T12:34:56Z",
		})
	}

	// 模拟其他操作可能产生的错误
	err1 := userService.ValidatePassword(userID, "")
	err2 := userService.ValidatePassword(userID, "wrong")

	// 如果有多个错误，使用错误聚合
	if err1 != nil && err2 != nil {
		// 使用Aggregate聚合多个错误
		return errors.NewAggregate([]error{
			errors.Wrap(err1, "密码验证失败"),
			errors.Wrap(err2, "身份验证失败"),
		})
	}

	fmt.Printf("成功处理用户 %s 的请求\n", username)
	return nil
}

// 处理API错误并生成HTTP响应
func HandleError(err error) {
	// 获取并输出错误的详细信息
	fmt.Printf("错误详情: %+v\n\n", err)

	// 检查是否为特定错误码
	if errors.IsErrorCode(err, ErrUserNotFound) {
		fmt.Println("检测到用户未找到错误")
	}

	// 获取错误码信息
	errCode := errors.GetErrorCodeFromError(err)
	if errCode != nil {
		fmt.Printf("错误码: %d\n", errCode.Code())
		fmt.Printf("HTTP状态: %d\n", errCode.HTTPStatus())
		fmt.Printf("错误消息: %s\n", errCode.Message())
		if ref := errCode.Reference(); ref != "" {
			fmt.Printf("文档引用: %s\n", ref)
		}
	}

	// 获取错误上下文
	if reqID, ok := errors.GetContext(err, "request_id"); ok {
		fmt.Printf("请求ID: %v\n", reqID)
	}

	// 获取所有上下文
	allContext := errors.GetAllContext(err)
	if len(allContext) > 0 {
		fmt.Println("所有上下文:")
		for k, v := range allContext {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}

	// 获取错误原因
	cause := errors.Cause(err)
	if cause != nil && !reflect.DeepEqual(cause, err) {
		fmt.Printf("根本原因: %v\n", cause)
	}

	// 检查是否为聚合错误
	if agg, ok := err.(interface{ Errors() []error }); ok {
		fmt.Println("聚合错误包含以下错误:")
		for i, e := range agg.Errors() {
			fmt.Printf("  %d: %v\n", i+1, e)
		}
	}

	fmt.Println("----------------------------------------")
}

func main() {
	fmt.Println("=== errors模块使用示例 ===")

	// 示例1: 基本错误
	fmt.Println("\n[示例1: 基本错误]")
	err1 := errors.New("这是一个基本错误")
	HandleError(err1)

	// 示例2: 带错误码的错误
	fmt.Println("\n[示例2: 带错误码的错误]")
	err2 := errors.NewWithCode(ErrInvalidInput, "输入参数格式错误")
	HandleError(err2)

	// 示例3: 错误包装
	fmt.Println("\n[示例3: 错误包装]")
	baseErr := stderrors.New("基础错误")
	wrappedErr := errors.Wrap(baseErr, "包装的错误")
	HandleError(wrappedErr)

	// 示例4: 模拟API请求处理 - 用户不存在
	fmt.Println("\n[示例4: 用户不存在]")
	err4 := HandleUserRequest(-1, "req-001")
	HandleError(err4)

	// 示例5: 模拟API请求处理 - 数据库错误
	fmt.Println("\n[示例5: 数据库错误]")
	err5 := HandleUserRequest(999, "req-002")
	HandleError(err5)

	// 示例6: 聚合错误
	fmt.Println("\n[示例6: 聚合错误]")
	err6a := errors.New("第一个错误")
	err6b := errors.NewWithCode(ErrRateLimitExceeded, "请求次数过多")
	aggErr := errors.NewAggregate([]error{err6a, err6b})
	HandleError(aggErr)

	// 示例7: 与标准库兼容性
	fmt.Println("\n[示例7: 与标准库兼容性]")
	stdErr := fmt.Errorf("标准错误")
	ourErr := errors.NewWithCode(ErrUnauthorized, "未授权访问")
	wrappedOurErr := fmt.Errorf("包装的错误: %w", ourErr)

	// 使用标准库的errors.Is和errors.As
	fmt.Printf("标准库errors.Is检测: %v\n", stderrors.Is(wrappedOurErr, ourErr))
	fmt.Printf("标准库errors.Is检测标准错误: %v\n", stderrors.Is(wrappedOurErr, stdErr))

	var codeErr interface{ Code() int }
	if stderrors.As(wrappedOurErr, &codeErr) {
		fmt.Printf("标准库errors.As获取错误码: %d\n", codeErr.Code())
	}

	HandleError(wrappedOurErr)

	// 示例8: 堆栈捕获优化
	fmt.Println("\n[示例8: 堆栈捕获优化]")

	// 设置全局堆栈捕获模式为延迟捕获
	fmt.Println("默认堆栈捕获模式:", describeStackCaptureMode(errors.DefaultStackCaptureMode))

	// 创建不同堆栈捕获模式的错误
	err8a := errors.NewWithStackControl("无堆栈错误", errors.StackCaptureModeNever)
	err8b := errors.NewWithStackControl("立即堆栈错误", errors.StackCaptureModeImmediate)
	err8c := errors.NewWithStackControl("延迟堆栈错误", errors.StackCaptureModeDeferred)

	// 处理并展示不同模式的错误
	fmt.Println("StackCaptureModeNever错误:")
	HandleError(err8a)

	fmt.Println("StackCaptureModeImmediate错误:")
	HandleError(err8b)

	fmt.Println("StackCaptureModeDeferred错误:")
	HandleError(err8c)

	log.Println("示例运行完成")
}

// 描述堆栈捕获模式
func describeStackCaptureMode(mode errors.StackCaptureMode) string {
	switch mode {
	case errors.StackCaptureModeNever:
		return "Never (不捕获堆栈)"
	case errors.StackCaptureModeImmediate:
		return "Immediate (立即捕获堆栈)"
	case errors.StackCaptureModeDeferred:
		return "Deferred (延迟捕获堆栈)"
	case errors.StackCaptureModeModeSampled:
		return "Sampled (采样捕获堆栈)"
	default:
		return fmt.Sprintf("未知模式(%d)", mode)
	}
}
