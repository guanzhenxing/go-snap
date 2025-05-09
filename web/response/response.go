// Package response 提供Web响应相关的工具和结构
package response

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guanzhenxing/go-snap/errors"
)

// 预定义的状态码
const (
	CodeSuccess      = 0 // 成功
	CodeUnknownError = 1 // 未知错误
	CodeParamError   = 2 // 参数错误
	CodeServerError  = 3 // 服务器错误
	CodeUnauthorized = 4 // 未授权
	CodeForbidden    = 5 // 禁止访问
	CodeNotFound     = 6 // 资源不存在
	CodeTimeout      = 7 // 请求超时
	CodeTooManyReqs  = 8 // 请求过于频繁
)

// StatusMessages 状态码与消息的映射
var StatusMessages = map[int]string{
	CodeSuccess:      "成功",
	CodeUnknownError: "未知错误",
	CodeParamError:   "参数错误",
	CodeServerError:  "服务器错误",
	CodeUnauthorized: "未授权",
	CodeForbidden:    "禁止访问",
	CodeNotFound:     "资源不存在",
	CodeTimeout:      "请求超时",
	CodeTooManyReqs:  "请求过于频繁",
}

// Response 统一的响应结构
type Response struct {
	Code      int         `json:"code"`                 // 业务状态码
	Message   string      `json:"message"`              // 状态消息
	Data      interface{} `json:"data,omitempty"`       // 响应数据
	RequestID string      `json:"request_id,omitempty"` // 请求ID
	Timestamp int64       `json:"timestamp"`            // 时间戳
}

// Context 自定义的请求上下文，封装Gin的Context
type Context struct {
	ginCtx      *gin.Context
	status      int
	body        *bytes.Buffer // 用于记录响应体
	isStreaming bool          // 是否为流式响应
}

// Init 初始化上下文对象
func (c *Context) Init(ginCtx *gin.Context) {
	c.ginCtx = ginCtx
	c.status = http.StatusOK
	c.body = &bytes.Buffer{}
	c.isStreaming = false
}

// GetStatus 获取HTTP状态码
func (c *Context) GetStatus() int {
	return c.status
}

// GetBodyString 获取响应体字符串
func (c *Context) GetBodyString() string {
	return c.body.String()
}

// IsStreaming 是否为流式响应
func (c *Context) IsStreaming() bool {
	return c.isStreaming
}

// GinContext 获取原始的Gin上下文
func (c *Context) GinContext() *gin.Context {
	return c.ginCtx
}

// RequestID 获取请求ID
func (c *Context) RequestID() string {
	requestID, _ := c.ginCtx.Get("X-Request-ID")
	if requestID == nil {
		return ""
	}
	return requestID.(string)
}

// Success 返回成功响应
func (c *Context) Success(data interface{}) {
	c.JSON(http.StatusOK, &Response{
		Code:      CodeSuccess,
		Message:   StatusMessages[CodeSuccess],
		Data:      data,
		RequestID: c.RequestID(),
		Timestamp: time.Now().Unix(),
	})
}

// Error 返回错误响应
func (c *Context) Error(code int, err error) {
	if err == nil {
		// 如果没有错误对象，使用提供的错误码
		c.JSON(getHTTPStatusFromCode(code), &Response{
			Code:      code,
			Message:   StatusMessages[code],
			RequestID: c.RequestID(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 从错误中获取错误码
	var errorCode int
	var errorMsg string
	var httpStatus int

	// 尝试从错误中获取错误码
	errorCodeInfo := errors.GetErrorCodeFromError(err)
	if errorCodeInfo != nil && errorCodeInfo != errors.UnknownError {
		// 使用错误中的信息
		errorCode = errorCodeInfo.Code()
		errorMsg = errorCodeInfo.Message()
		httpStatus = errorCodeInfo.HTTPStatus()
	} else {
		// 使用传入的错误码
		errorCode = code
		errorMsg = err.Error()
		httpStatus = getHTTPStatusFromCode(code)
	}

	// 如果提供了明确的错误码，优先使用它
	if code != CodeSuccess && code != CodeUnknownError {
		errorCode = code
		httpStatus = getHTTPStatusFromCode(code)
	}

	c.JSON(httpStatus, &Response{
		Code:      errorCode,
		Message:   errorMsg,
		RequestID: c.RequestID(),
		Timestamp: time.Now().Unix(),
	})
}

// Param 获取URL参数
func (c *Context) Param(key string) string {
	return c.ginCtx.Param(key)
}

// Query 获取查询参数
func (c *Context) Query(key string) string {
	return c.ginCtx.Query(key)
}

// DefaultQuery 获取查询参数，如果不存在则返回默认值
func (c *Context) DefaultQuery(key, defaultValue string) string {
	return c.ginCtx.DefaultQuery(key, defaultValue)
}

// Bind 绑定请求体，支持JSON、XML、Form等
func (c *Context) Bind(obj interface{}) error {
	if err := c.ginCtx.ShouldBind(obj); err != nil {
		return errors.Wrap(err, "failed to bind request body")
	}
	return nil
}

// BindJSON 绑定JSON请求体
func (c *Context) BindJSON(obj interface{}) error {
	if err := c.ginCtx.ShouldBindJSON(obj); err != nil {
		return errors.Wrap(err, "failed to bind JSON request body")
	}
	return nil
}

// BindQuery 绑定查询参数
func (c *Context) BindQuery(obj interface{}) error {
	if err := c.ginCtx.ShouldBindQuery(obj); err != nil {
		return errors.Wrap(err, "failed to bind query parameters")
	}
	return nil
}

// JSON 返回JSON响应
func (c *Context) JSON(httpStatus int, obj interface{}) {
	c.status = httpStatus

	// 记录响应体
	if obj != nil {
		data, _ := json.Marshal(obj)
		c.body.Write(data)
	}

	c.ginCtx.JSON(httpStatus, obj)
}

// String 返回字符串响应
func (c *Context) String(httpStatus int, format string, values ...interface{}) {
	c.status = httpStatus
	c.ginCtx.String(httpStatus, format, values...)
}

// HTML 返回HTML响应
func (c *Context) HTML(httpStatus int, name string, obj interface{}) {
	c.status = httpStatus
	c.isStreaming = true
	c.ginCtx.HTML(httpStatus, name, obj)
}

// Stream 返回流式响应
func (c *Context) Stream(httpStatus int, contentType string, reader func(w http.ResponseWriter) bool) {
	c.status = httpStatus
	c.isStreaming = true
	c.ginCtx.Header("Content-Type", contentType)
	c.ginCtx.Status(httpStatus)
	if reader(c.ginCtx.Writer) {
		c.ginCtx.Writer.Flush()
	}
}

// File 返回文件响应
func (c *Context) File(filepath string) {
	c.isStreaming = true
	c.ginCtx.File(filepath)
}

// BadRequest 返回400错误
func (c *Context) BadRequest(err error) {
	c.Error(CodeParamError, err)
}

// Unauthorized 返回401错误
func (c *Context) Unauthorized(err error) {
	c.Error(CodeUnauthorized, err)
}

// Forbidden 返回403错误
func (c *Context) Forbidden(err error) {
	c.Error(CodeForbidden, err)
}

// NotFound 返回404错误
func (c *Context) NotFound(err error) {
	c.Error(CodeNotFound, err)
}

// ServerError 返回500错误
func (c *Context) ServerError(err error) {
	c.Error(CodeServerError, err)
}

// TooManyRequests 返回429错误
func (c *Context) TooManyRequests(err error) {
	c.Error(CodeTooManyReqs, err)
}

// getHTTPStatusFromCode 根据业务状态码获取对应的HTTP状态码
func getHTTPStatusFromCode(code int) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeParamError:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeTimeout:
		return http.StatusRequestTimeout
	case CodeTooManyReqs:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// GetContext 从Gin上下文获取自定义上下文
func GetContext(ginCtx *gin.Context) *Context {
	ctx, exists := ginCtx.Get("ctx")
	if !exists {
		return nil
	}
	return ctx.(*Context)
}

// NewErrorResponse 创建新的错误响应
func NewErrorResponse(code int, message string, requestID string) *Response {
	if message == "" {
		message = StatusMessages[code]
	}
	return &Response{
		Code:      code,
		Message:   message,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	}
}

// NewSuccessResponse 创建新的成功响应
func NewSuccessResponse(data interface{}, requestID string) *Response {
	return &Response{
		Code:      CodeSuccess,
		Message:   StatusMessages[CodeSuccess],
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	}
}
