package appcore

import (
	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/logger"
)

// Option 配置选项函数
type Option func(*App)

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

// WithComponentManager 设置组件管理器
func WithComponentManager(manager *ComponentManager) Option {
	return func(a *App) {
		a.componentManager = manager
	}
}
