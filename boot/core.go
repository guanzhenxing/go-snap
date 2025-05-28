package boot

import (
	"fmt"
	"time"
)

// AppState 应用状态
type AppState int

const (
	// AppStateCreated 应用已创建
	AppStateCreated AppState = iota

	// AppStateInitializing 应用正在初始化
	AppStateInitializing

	// AppStateInitialized 应用已初始化
	AppStateInitialized

	// AppStateStarting 应用正在启动
	AppStateStarting

	// AppStateRunning 应用正在运行
	AppStateRunning

	// AppStateStopping 应用正在停止
	AppStateStopping

	// AppStateStopped 应用已停止
	AppStateStopped

	// AppStateFailed 应用运行失败
	AppStateFailed
)

// String 返回状态字符串
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

// EventListener 事件监听器
type EventListener func(eventName string, eventData interface{})

// Plugin 插件接口
type Plugin interface {
	// Name 返回插件名称
	Name() string

	// Register 注册插件
	Register(app *Application) error

	// Version 返回插件版本
	Version() string

	// Dependencies 返回插件依赖
	Dependencies() []string
}

// ConfigError 配置错误
type ConfigError struct {
	Message   string
	Cause     error
	Component string
	Timestamp time.Time
}

// Error 实现error接口
func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s", e.Component, e.Message, e.Cause.Error())
	}
	return fmt.Sprintf("[%s] %s", e.Component, e.Message)
}

// Unwrap 支持错误链
func (e *ConfigError) Unwrap() error {
	return e.Cause
}

// NewConfigError 创建配置错误
func NewConfigError(component, message string, cause error) *ConfigError {
	return &ConfigError{
		Message:   message,
		Cause:     cause,
		Component: component,
		Timestamp: time.Now(),
	}
}

// ComponentError 组件错误
type ComponentError struct {
	*ConfigError
	ComponentName string
	Operation     string
}

// NewComponentError 创建组件错误
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

// DependencyError 依赖错误
type DependencyError struct {
	*ConfigError
	DependencyChain []string
}

// NewDependencyError 创建依赖错误
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

// Common errors
var (
	ErrComponentNotFound   = &ConfigError{Message: "组件未找到", Component: "Registry", Timestamp: time.Now()}
	ErrComponentExists     = &ConfigError{Message: "组件已存在", Component: "Registry", Timestamp: time.Now()}
	ErrDependencyCycle     = &ConfigError{Message: "组件依赖循环", Component: "DependencyResolver", Timestamp: time.Now()}
	ErrNotBeanProvider     = &ConfigError{Message: "组件不是Bean提供者", Component: "Registry", Timestamp: time.Now()}
	ErrInvalidConfig       = &ConfigError{Message: "无效的配置", Component: "Config", Timestamp: time.Now()}
	ErrComponentInitError  = &ConfigError{Message: "组件初始化失败", Component: "Component", Timestamp: time.Now()}
	ErrComponentStartError = &ConfigError{Message: "组件启动失败", Component: "Component", Timestamp: time.Now()}
	ErrComponentStopError  = &ConfigError{Message: "组件停止失败", Component: "Component", Timestamp: time.Now()}
	ErrHealthCheckFailed   = &ConfigError{Message: "健康检查失败", Component: "HealthChecker", Timestamp: time.Now()}
)
