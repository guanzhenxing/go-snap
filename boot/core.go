package boot

import (
	"fmt"
	"time"
)

// AppState 应用状态枚举类型，表示应用在其生命周期中的不同阶段
type AppState int

const (
	// AppStateCreated 应用已创建，表示应用实例已创建但尚未初始化
	AppStateCreated AppState = iota

	// AppStateInitializing 应用正在初始化，表示应用正在进行初始化操作
	AppStateInitializing

	// AppStateInitialized 应用已初始化，表示应用已完成初始化但尚未启动
	AppStateInitialized

	// AppStateStarting 应用正在启动，表示应用正在启动其组件
	AppStateStarting

	// AppStateRunning 应用正在运行，表示应用已启动并正常运行中
	AppStateRunning

	// AppStateStopping 应用正在停止，表示应用正在停止其组件
	AppStateStopping

	// AppStateStopped 应用已停止，表示应用已完全停止
	AppStateStopped

	// AppStateFailed 应用运行失败，表示应用在初始化或运行过程中遇到错误
	AppStateFailed
)

// String 返回应用状态的字符串表示
// 返回：
//
//	表示应用状态的可读字符串
func (s AppState) String() string {
	switch s {
	case AppStateCreated:
		return "Created"
	case AppStateInitializing:
		return "Initializing"
	case AppStateInitialized:
		return "Initialized"
	case AppStateStarting:
		return "Starting"
	case AppStateRunning:
		return "Running"
	case AppStateStopping:
		return "Stopping"
	case AppStateStopped:
		return "Stopped"
	case AppStateFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// EventListener 事件监听器函数类型，用于订阅和处理应用事件
// 参数：
//
//	eventName: 事件名称
//	eventData: 事件数据
type EventListener func(eventName string, eventData interface{})

// Plugin 插件接口，用于扩展应用功能
type Plugin interface {
	// Name 返回插件名称
	// 返回：
	//   插件的唯一名称
	Name() string

	// Register 注册插件到应用
	// 参数：
	//   app: 应用实例
	// 返回：
	//   注册过程中遇到的错误，如果注册成功则返回nil
	Register(app *Application) error

	// Version 返回插件版本
	// 返回：
	//   插件的版本号
	Version() string

	// Dependencies 返回插件依赖的其他插件列表
	// 返回：
	//   插件依赖的其他插件名称列表
	Dependencies() []string
}

// ConfigError 配置错误类型，表示配置过程中遇到的错误
type ConfigError struct {
	// Message 错误消息
	Message string
	// Cause 导致错误的原因
	Cause error
	// Component 发生错误的组件
	Component string
	// Timestamp 错误发生的时间
	Timestamp time.Time
}

// Error 实现error接口，返回格式化的错误消息
// 返回：
//
//	格式化的错误消息字符串
func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s", e.Component, e.Message, e.Cause.Error())
	}
	return fmt.Sprintf("[%s] %s", e.Component, e.Message)
}

// Unwrap 支持错误链，返回错误的原因
// 返回：
//
//	导致当前错误的原始错误
func (e *ConfigError) Unwrap() error {
	return e.Cause
}

// NewConfigError 创建新的配置错误实例
// 参数：
//
//	component: 发生错误的组件名称
//	message: 错误消息
//	cause: 导致错误的原因
//
// 返回：
//
//	新的ConfigError实例
func NewConfigError(component, message string, cause error) *ConfigError {
	return &ConfigError{
		Message:   message,
		Cause:     cause,
		Component: component,
		Timestamp: time.Now(),
	}
}

// ComponentError 组件错误类型，表示组件操作过程中遇到的错误
type ComponentError struct {
	*ConfigError
	// ComponentName 发生错误的组件名称
	ComponentName string
	// Operation 发生错误的操作名称
	Operation string
}

// NewComponentError 创建新的组件错误实例
// 参数：
//
//	componentName: 组件名称
//	operation: 操作名称
//	message: 错误消息
//	cause: 导致错误的原因
//
// 返回：
//
//	新的ComponentError实例
func NewComponentError(componentName, operation, message string, cause error) *ComponentError {
	return &ComponentError{
		ConfigError: &ConfigError{
			Message:   message,
			Cause:     cause,
			Component: componentName,
			Timestamp: time.Now(),
		},
		ComponentName: componentName,
		Operation:     operation,
	}
}

// DependencyError 依赖错误类型，表示组件依赖解析过程中遇到的错误
type DependencyError struct {
	*ConfigError
	// DependencyChain 依赖链，包含导致错误的依赖路径
	DependencyChain []string
}

// NewDependencyError 创建新的依赖错误实例
// 参数：
//
//	message: 错误消息
//	chain: 依赖链
//	cause: 导致错误的原因
//
// 返回：
//
//	新的DependencyError实例
func NewDependencyError(message string, chain []string, cause error) *DependencyError {
	return &DependencyError{
		ConfigError: &ConfigError{
			Message:   message,
			Cause:     cause,
			Component: "DependencyResolver",
			Timestamp: time.Now(),
		},
		DependencyChain: chain,
	}
}

// Common errors 常见错误预定义
var (
	// ErrComponentNotFound 组件未找到错误
	ErrComponentNotFound = &ConfigError{Message: "组件未找到", Component: "Registry", Timestamp: time.Now()}
	// ErrComponentExists 组件已存在错误
	ErrComponentExists = &ConfigError{Message: "组件已存在", Component: "Registry", Timestamp: time.Now()}
	// ErrDependencyCycle 组件依赖循环错误
	ErrDependencyCycle = &ConfigError{Message: "组件依赖循环", Component: "DependencyResolver", Timestamp: time.Now()}
	// ErrNotBeanProvider 组件不是Bean提供者错误
	ErrNotBeanProvider = &ConfigError{Message: "组件不是Bean提供者", Component: "Registry", Timestamp: time.Now()}
	// ErrInvalidConfig 无效的配置错误
	ErrInvalidConfig = &ConfigError{Message: "无效的配置", Component: "Config", Timestamp: time.Now()}
	// ErrComponentInitError 组件初始化失败错误
	ErrComponentInitError = &ConfigError{Message: "组件初始化失败", Component: "Component", Timestamp: time.Now()}
	// ErrComponentStartError 组件启动失败错误
	ErrComponentStartError = &ConfigError{Message: "组件启动失败", Component: "Component", Timestamp: time.Now()}
	// ErrComponentStopError 组件停止失败错误
	ErrComponentStopError = &ConfigError{Message: "组件停止失败", Component: "Component", Timestamp: time.Now()}
	// ErrHealthCheckFailed 健康检查失败错误
	ErrHealthCheckFailed = &ConfigError{Message: "健康检查失败", Component: "HealthChecker", Timestamp: time.Now()}
)
