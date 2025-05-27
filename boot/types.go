// Package boot 提供类似Spring Boot的自动配置框架
package boot

import "context"

// Component 组件接口
type Component interface {
	// Name 返回组件名称
	Name() string

	// Type 返回组件类型
	Type() ComponentType

	// Initialize 初始化组件
	Initialize(ctx context.Context) error

	// Start 启动组件
	Start(ctx context.Context) error

	// Stop 停止组件
	Stop(ctx context.Context) error
}

// ComponentType 组件类型
type ComponentType int

const (
	// ComponentTypeInfrastructure 基础设施组件
	ComponentTypeInfrastructure ComponentType = iota

	// ComponentTypeDataSource 数据源组件
	ComponentTypeDataSource

	// ComponentTypeCore 核心业务组件
	ComponentTypeCore

	// ComponentTypeWeb Web服务组件
	ComponentTypeWeb
)

// ComponentFactory 组件工厂接口
type ComponentFactory interface {
	// Create 创建组件实例
	Create(ctx context.Context, props PropertySource) (Component, error)

	// Dependencies 返回依赖的组件名称
	Dependencies() []string
}

// PropertySource 属性源接口
type PropertySource interface {
	// GetProperty 获取属性值
	GetProperty(key string) (interface{}, bool)

	// GetString 获取字符串属性
	GetString(key string, defaultValue string) string

	// GetBool 获取布尔属性
	GetBool(key string, defaultValue bool) bool

	// GetInt 获取整型属性
	GetInt(key string, defaultValue int) int

	// GetFloat 获取浮点属性
	GetFloat(key string, defaultValue float64) float64

	// HasProperty 判断属性是否存在
	HasProperty(key string) bool

	// SetProperty 设置属性值
	SetProperty(key string, value interface{})
}

// Condition 条件接口
type Condition interface {
	// Matches 判断条件是否匹配
	Matches(props PropertySource) bool
}

// Plugin 插件接口
type Plugin interface {
	// Name 返回插件名称
	Name() string

	// Register 注册插件
	Register(app *Application) error
}

// BeanProvider Bean提供者接口
type BeanProvider interface {
	// GetBean 获取Bean实例
	GetBean(name string, bean interface{}) error
}

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

// EventListener 事件监听器
type EventListener func(eventName string, eventData interface{})

// ConfigError 配置错误
type ConfigError struct {
	Message string
	Cause   error
}

// Error 实现error接口
func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Common errors
var (
	ErrComponentNotFound  = &ConfigError{Message: "组件未找到"}
	ErrComponentExists    = &ConfigError{Message: "组件已存在"}
	ErrDependencyCycle    = &ConfigError{Message: "组件依赖循环"}
	ErrNotBeanProvider    = &ConfigError{Message: "组件不是Bean提供者"}
	ErrInvalidConfig      = &ConfigError{Message: "无效的配置"}
	ErrComponentInitError = &ConfigError{Message: "组件初始化失败"}
)
