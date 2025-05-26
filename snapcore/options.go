package snapcore

import (
	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/logger"
)

// Option 配置选项函数
type Option func(*App)

// WithConfigFile 设置配置文件路径
func WithConfigFile(path string) Option {
	return func(a *App) {
		a.configPath = path
	}
}

// WithConfig 设置配置提供器
func WithConfig(provider config.Provider) Option {
	return func(a *App) {
		a.config = provider
	}
}

// WithLogger 设置日志器
func WithLogger(logger logger.Logger) Option {
	return func(a *App) {
		a.logger = logger
	}
}

// WithComponent 注册组件
func WithComponent(name string, component Component, dependencies ...string) Option {
	return func(a *App) {
		// 组件将在App.Run()时添加到依赖图中
		a.pendingComponents = append(a.pendingComponents, ComponentRegistration{
			Name:         name,
			Component:    component,
			Dependencies: dependencies,
		})
	}
}

// WithHook 添加钩子
func WithHook(hookType HookType, hookFunc HookFunc) Option {
	return func(a *App) {
		a.hooks[hookType] = append(a.hooks[hookType], hookFunc)
	}
}

// WithShutdownTimeout 设置关闭超时时间
func WithShutdownTimeout(timeout int) Option {
	return func(a *App) {
		a.shutdownTimeout = timeout
	}
}

// WithGracefulShutdown 设置是否启用优雅关闭
func WithGracefulShutdown(enabled bool) Option {
	return func(a *App) {
		a.gracefulShutdown = enabled
	}
}

// WithStateMonitor 设置状态监控器
func WithStateMonitor(enabled bool) Option {
	return func(a *App) {
		a.monitorState = enabled
	}
}

// WithPlugin 添加插件
func WithPlugin(plugin Plugin) Option {
	return func(a *App) {
		a.plugins = append(a.plugins, plugin)
	}
}

// WithDecorator 添加组件装饰器
func WithDecorator(decorator ComponentDecorator) Option {
	return func(a *App) {
		a.decorators = append(a.decorators, decorator)
	}
}
