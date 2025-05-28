package boot

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

// Plugin 插件接口
type Plugin interface {
	// Name 返回插件名称
	Name() string

	// Register 注册插件
	Register(app *Application) error
}

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
