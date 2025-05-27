//go:build wireinject
// +build wireinject

// Wire是Google开发的依赖注入框架，通过代码生成实现依赖注入
// 本文件定义了应用的初始化函数，这些函数会被Wire处理并生成实际的实现代码
// 使用`wire`命令生成代码：wire gen ./appcore
package appcore

import "github.com/google/wire"

// InitializeApp 初始化完整应用实例
// 参数:
//   - name: 应用名称
//   - version: 应用版本
//   - configPath: 配置文件路径
//
// 返回:
//   - *App: 应用实例
//   - error: 初始化错误
func InitializeApp(name, version, configPath string) (*App, error) {
	wire.Build(AppSet)
	return nil, nil
}

// InitializeBaseApp 初始化基础应用实例，只包含配置和日志组件
// 适用于简单应用或工具程序
func InitializeBaseApp(name, version, configPath string) (*App, error) {
	wire.Build(
		BaseSet,
		NewComponentManager,
		ProvideComponentManager,
		ProvideApp,
	)
	return nil, nil
}

// InitializeWebApp 初始化Web应用实例，不包含数据库组件
// 适用于无状态API服务
func InitializeWebApp(name, version, configPath string) (*App, error) {
	wire.Build(
		BaseSet,
		WebSet,
		NewComponentManager,
		wire.Struct(new(ComponentManager), "*"),
		ProvideApp,
	)
	return nil, nil
}

// InitializeApiApp 初始化API应用实例，包含Web和缓存组件
// 适用于需要缓存的API服务
func InitializeApiApp(name, version, configPath string) (*App, error) {
	wire.Build(
		BaseSet,          // 配置和日志组件
		WebSet,           // Web服务组件
		NewCacheProvider, // 缓存组件
		NewComponentManager,
		wire.FieldsOf(new(*ComponentManager), "components", "componentsByType"),
		ProvideApp,
	)
	return nil, nil
}

// InitializeDataApp 初始化数据处理应用实例，包含数据库组件但不包含Web组件
// 适用于后台数据处理、定时任务等场景
func InitializeDataApp(name, version, configPath string) (*App, error) {
	wire.Build(
		BaseSet,            // 配置和日志组件
		NewDBStoreProvider, // 数据库组件
		NewComponentManager,
		wire.FieldsOf(new(*ComponentManager), "components", "componentsByType"),
		ProvideApp,
	)
	return nil, nil
}

// InitializeCacheApp 初始化缓存应用实例，包含缓存组件但不包含Web和数据库组件
// 适用于缓存服务、数据缓存等场景
func InitializeCacheApp(name, version, configPath string) (*App, error) {
	wire.Build(
		BaseSet,          // 配置和日志组件
		NewCacheProvider, // 缓存组件
		NewComponentManager,
		wire.FieldsOf(new(*ComponentManager), "components", "componentsByType"),
		ProvideApp,
	)
	return nil, nil
}

// InitializeStorage 初始化存储应用实例，包含数据库和缓存组件
// 适用于存储服务、数据访问层等场景
func InitializeStorage(name, version, configPath string) (*App, error) {
	wire.Build(
		BaseSet,    // 配置和日志组件
		StorageSet, // 数据库和缓存组件
		NewComponentManager,
		wire.FieldsOf(new(*ComponentManager), "components", "componentsByType"),
		ProvideApp,
	)
	return nil, nil
}
