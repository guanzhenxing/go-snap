package main

import (
	stderrors "errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

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
	ErrQueryTimeout       = 2003
	ErrTransactionFailed  = 2004

	// API相关错误 (3000-3999)
	ErrInvalidInput      = 3001
	ErrUnauthorized      = 3002
	ErrRateLimitExceeded = 3003
	ErrResourceNotFound  = 3004
	ErrMethodNotAllowed  = 3005

	// 业务逻辑错误 (4000-4999)
	ErrInsufficientFunds  = 4001
	ErrPaymentFailed      = 4002
	ErrOrderProcessing    = 4003
	ErrInvalidOrderStatus = 4004
	ErrProductUnavailable = 4005

	// 第三方服务错误 (5000-5999)
	ErrThirdPartyTimeout  = 5001
	ErrThirdPartyError    = 5002
	ErrGatewayTimeout     = 5003
	ErrServiceUnavailable = 5004
)

// 初始化并注册所有错误码
func init() {
	// 注册用户相关错误码
	errors.RegisterErrorCode(ErrUserNotFound, http.StatusNotFound, "用户未找到", "https://docs.example.com/errors/user-not-found")
	errors.RegisterErrorCode(ErrInvalidPassword, http.StatusBadRequest, "密码无效", "https://docs.example.com/errors/invalid-password")

	// 注册数据库相关错误码
	errors.RegisterErrorCode(ErrDatabaseConnection, http.StatusInternalServerError, "数据库连接失败", "https://docs.example.com/errors/db-connection")
	errors.RegisterErrorCode(ErrDuplicateKey, http.StatusConflict, "记录已存在", "https://docs.example.com/errors/duplicate-key")
	errors.RegisterErrorCode(ErrQueryTimeout, http.StatusGatewayTimeout, "数据库查询超时", "https://docs.example.com/errors/query-timeout")
	errors.RegisterErrorCode(ErrTransactionFailed, http.StatusInternalServerError, "事务执行失败", "https://docs.example.com/errors/transaction-failed")

	// 注册API相关错误码
	errors.RegisterErrorCode(ErrInvalidInput, http.StatusBadRequest, "输入无效", "")
	errors.RegisterErrorCode(ErrUnauthorized, http.StatusUnauthorized, "未授权访问", "")
	errors.RegisterErrorCode(ErrRateLimitExceeded, http.StatusTooManyRequests, "请求速率超限", "")
	errors.RegisterErrorCode(ErrResourceNotFound, http.StatusNotFound, "资源未找到", "")
	errors.RegisterErrorCode(ErrMethodNotAllowed, http.StatusMethodNotAllowed, "方法不允许", "")

	// 注册业务逻辑错误码
	errors.RegisterErrorCode(ErrInsufficientFunds, http.StatusBadRequest, "余额不足", "")
	errors.RegisterErrorCode(ErrPaymentFailed, http.StatusBadRequest, "支付失败", "")
	errors.RegisterErrorCode(ErrOrderProcessing, http.StatusBadRequest, "订单处理中", "")
	errors.RegisterErrorCode(ErrInvalidOrderStatus, http.StatusBadRequest, "订单状态无效", "")
	errors.RegisterErrorCode(ErrProductUnavailable, http.StatusBadRequest, "商品不可用", "")

	// 注册第三方服务错误码
	errors.RegisterErrorCode(ErrThirdPartyTimeout, http.StatusGatewayTimeout, "第三方服务超时", "")
	errors.RegisterErrorCode(ErrThirdPartyError, http.StatusBadGateway, "第三方服务错误", "")
	errors.RegisterErrorCode(ErrGatewayTimeout, http.StatusGatewayTimeout, "网关超时", "")
	errors.RegisterErrorCode(ErrServiceUnavailable, http.StatusServiceUnavailable, "服务不可用", "")
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

// ===========================================================================
// 新增实际应用场景示例
// ===========================================================================

// 场景1: 分布式系统中的微服务调用
type ServiceClient struct {
	name     string
	endpoint string
}

func NewServiceClient(name, endpoint string) *ServiceClient {
	return &ServiceClient{
		name:     name,
		endpoint: endpoint,
	}
}

// 模拟调用远程服务
func (c *ServiceClient) Call(method string, params map[string]interface{}) (interface{}, error) {
	// 模拟请求ID和追踪ID
	requestID := "req-" + time.Now().Format("20060102150405")
	traceID := "trace-abc-" + requestID

	// 添加调用上下文
	callContext := map[string]interface{}{
		"request_id":      requestID,
		"trace_id":        traceID,
		"source_service":  "api_gateway",
		"target_service":  c.name,
		"target_endpoint": c.endpoint,
		"method":          method,
		"timestamp":       time.Now().Format(time.RFC3339),
	}

	// 模拟超时错误
	if method == "getUserProfile" && params["simulate"] == "timeout" {
		baseErr := errors.NewWithCode(ErrThirdPartyTimeout, "调用%s服务超时", c.name)
		return nil, errors.WithContextMap(baseErr, callContext)
	}

	// 模拟服务错误
	if method == "getUserProfile" && params["simulate"] == "error" {
		// 模拟底层错误
		innerErr := stderrors.New("500 Internal Server Error")
		// 添加一层包装
		serviceErr := errors.Wrap(innerErr, "远程服务返回错误状态码")
		// 添加错误码
		codedErr := errors.WrapWithCode(serviceErr, ErrThirdPartyError, "调用%s服务失败", c.name)
		// 添加上下文
		return nil, errors.WithContextMap(codedErr, callContext)
	}

	// 模拟服务不可用
	if method == "getUserProfile" && params["simulate"] == "unavailable" {
		baseErr := errors.NewWithCode(ErrServiceUnavailable, "%s服务不可用", c.name)
		return nil, errors.WithContextMap(baseErr, callContext)
	}

	// 正常返回
	return map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user_id":   params["user_id"],
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}, nil
}

// 场景2: 数据库操作封装
type DatabaseClient struct {
	connString string
}

func NewDatabaseClient(connString string) *DatabaseClient {
	return &DatabaseClient{
		connString: connString,
	}
}

// 模拟数据库查询
func (db *DatabaseClient) Query(query string, params map[string]interface{}) (interface{}, error) {
	// 添加查询上下文
	queryContext := map[string]interface{}{
		"query":     query,
		"params":    params,
		"timestamp": time.Now().Format(time.RFC3339),
		"db_host":   "db-server-01",
	}

	// 模拟连接错误
	if params["simulate"] == "connection_error" {
		baseErr := stderrors.New("无法连接到数据库服务器")
		return nil, errors.WithContextMap(
			errors.WrapWithCode(baseErr, ErrDatabaseConnection, "数据库连接失败"),
			queryContext,
		)
	}

	// 模拟查询超时
	if params["simulate"] == "timeout" {
		baseErr := stderrors.New("查询执行超过30秒限制")
		return nil, errors.WithContextMap(
			errors.WrapWithCode(baseErr, ErrQueryTimeout, "数据库查询超时"),
			queryContext,
		)
	}

	// 模拟事务错误
	if params["simulate"] == "transaction_error" {
		// 可能的底层错误链
		sqlErr := stderrors.New("foreign key constraint failed")
		txErr := errors.Wrap(sqlErr, "无法更新关联表")
		return nil, errors.WithContextMap(
			errors.WrapWithCode(txErr, ErrTransactionFailed, "事务执行失败"),
			queryContext,
		)
	}

	// 正常返回
	return map[string]interface{}{
		"rows_affected":  1,
		"last_insert_id": 1001,
	}, nil
}

// 场景3: Web API处理器
type APIHandler struct {
	userService   *UserService
	dbClient      *DatabaseClient
	serviceClient *ServiceClient
}

func NewAPIHandler() *APIHandler {
	return &APIHandler{
		userService:   &UserService{},
		dbClient:      NewDatabaseClient("postgres://user:pass@localhost:5432/mydb"),
		serviceClient: NewServiceClient("user-profile", "https://user-profile.example.com"),
	}
}

// 模拟处理用户订单API请求
func (h *APIHandler) HandleOrderCreation(orderData map[string]interface{}) (interface{}, error) {
	// 创建API上下文
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	apiContext := map[string]interface{}{
		"request_id": requestID,
		"endpoint":   "/api/orders",
		"method":     "POST",
		"client_ip":  "192.168.1.1",
		"user_agent": "Mozilla/5.0",
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	// 1. 验证请求数据
	if _, ok := orderData["user_id"]; !ok {
		return nil, errors.WithContextMap(
			errors.NewWithCode(ErrInvalidInput, "用户ID不能为空"),
			apiContext,
		)
	}

	if _, ok := orderData["product_id"]; !ok {
		return nil, errors.WithContextMap(
			errors.NewWithCode(ErrInvalidInput, "商品ID不能为空"),
			apiContext,
		)
	}

	userID := orderData["user_id"].(int)
	productID := orderData["product_id"].(int)

	// 2. 获取用户信息
	_, err := h.userService.FindByID(userID)
	if err != nil {
		return nil, errors.WithContextMap(
			errors.Wrap(err, "获取订单用户信息失败"),
			apiContext,
		)
	}

	// 3. 检查商品库存
	_, err = h.dbClient.Query("SELECT stock FROM products WHERE id = ?", map[string]interface{}{
		"id":       productID,
		"simulate": orderData["simulate_db"],
	})
	if err != nil {
		return nil, errors.WithContextMap(
			errors.Wrap(err, "检查商品库存失败"),
			apiContext,
		)
	}

	// 4. 调用支付服务进行预授权
	_, err = h.serviceClient.Call("preAuthorizePayment", map[string]interface{}{
		"user_id":  userID,
		"amount":   orderData["amount"],
		"simulate": orderData["simulate_service"],
	})
	if err != nil {
		return nil, errors.WithContextMap(
			errors.Wrap(err, "支付预授权失败"),
			apiContext,
		)
	}

	// 5. 创建订单记录
	_, err = h.dbClient.Query(
		"INSERT INTO orders (user_id, product_id, amount, status) VALUES (?, ?, ?, ?)",
		map[string]interface{}{
			"user_id":    userID,
			"product_id": productID,
			"amount":     orderData["amount"],
			"status":     "pending",
			"simulate":   orderData["simulate_db"],
		},
	)
	if err != nil {
		// 发生错误，还需要取消支付预授权
		cancelErr := h.rollbackPayment(userID, orderData["amount"])
		if cancelErr != nil {
			// 两个错误都需要报告
			return nil, errors.WithContextMap(
				errors.NewAggregate([]error{
					errors.Wrap(err, "创建订单失败"),
					errors.Wrap(cancelErr, "取消支付预授权失败"),
				}),
				apiContext,
			)
		}
		return nil, errors.WithContextMap(
			errors.Wrap(err, "创建订单失败"),
			apiContext,
		)
	}

	// 成功响应
	return map[string]interface{}{
		"order_id":   12345,
		"status":     "created",
		"request_id": requestID,
	}, nil
}

// 模拟回滚支付
func (h *APIHandler) rollbackPayment(userID int, amount interface{}) error {
	_, err := h.serviceClient.Call("cancelPayment", map[string]interface{}{
		"user_id": userID,
		"amount":  amount,
	})
	if err != nil {
		return errors.Wrap(err, "取消支付失败")
	}
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

	// ===========================================================================
	// 运行新增的实际应用场景示例
	// ===========================================================================

	// 示例9: 分布式系统中的微服务调用
	fmt.Println("\n[示例9: 分布式系统中的微服务调用]")
	serviceClient := NewServiceClient("user-profile", "https://user-profile.example.com")

	// 9.1 模拟微服务超时
	_, err9a := serviceClient.Call("getUserProfile", map[string]interface{}{
		"user_id":  123,
		"simulate": "timeout",
	})
	fmt.Println("微服务调用超时:")
	HandleError(err9a)

	// 9.2 模拟微服务错误
	_, err9b := serviceClient.Call("getUserProfile", map[string]interface{}{
		"user_id":  123,
		"simulate": "error",
	})
	fmt.Println("微服务调用错误:")
	HandleError(err9b)

	// 示例10: 数据库操作
	fmt.Println("\n[示例10: 数据库操作场景]")
	dbClient := NewDatabaseClient("postgres://user:pass@localhost:5432/mydb")

	// 10.1 模拟数据库连接错误
	_, err10a := dbClient.Query("SELECT * FROM users", map[string]interface{}{
		"simulate": "connection_error",
	})
	fmt.Println("数据库连接错误:")
	HandleError(err10a)

	// 10.2 模拟数据库查询超时
	_, err10b := dbClient.Query("SELECT * FROM large_table", map[string]interface{}{
		"simulate": "timeout",
	})
	fmt.Println("数据库查询超时:")
	HandleError(err10b)

	// 10.3 模拟事务执行失败
	_, err10c := dbClient.Query("UPDATE orders SET status = 'completed'", map[string]interface{}{
		"simulate": "transaction_error",
	})
	fmt.Println("事务执行失败:")
	HandleError(err10c)

	// 示例11: Web API处理复杂订单创建
	fmt.Println("\n[示例11: Web API处理复杂订单创建]")
	apiHandler := NewAPIHandler()

	// 11.1 模拟订单创建过程中的数据库错误
	_, err11a := apiHandler.HandleOrderCreation(map[string]interface{}{
		"user_id":          123,
		"product_id":       456,
		"amount":           99.99,
		"simulate_db":      "transaction_error",
		"simulate_service": "normal",
	})
	fmt.Println("订单创建 - 数据库错误:")
	HandleError(err11a)

	// 11.2 模拟订单创建过程中的微服务错误
	_, err11b := apiHandler.HandleOrderCreation(map[string]interface{}{
		"user_id":          123,
		"product_id":       456,
		"amount":           99.99,
		"simulate_db":      "normal",
		"simulate_service": "timeout",
	})
	fmt.Println("订单创建 - 微服务错误:")
	HandleError(err11b)

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
