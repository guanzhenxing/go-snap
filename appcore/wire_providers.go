package appcore

import (
	"github.com/google/wire"
	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/logger"
)

// ProvideComponentManager 提供组件管理器
func ProvideComponentManager(
	configProvider ConfigProvider,
	loggerProvider LoggerProvider,
	dbStoreProvider DBStoreProvider,
	cacheProvider CacheProvider,
	webProvider WebProvider,
) *ComponentManager {
	manager := NewComponentManager()

	// 注册所有组件
	manager.RegisterComponent(configProvider)
	manager.RegisterComponent(loggerProvider)
	manager.RegisterComponent(dbStoreProvider)
	manager.RegisterComponent(cacheProvider)
	manager.RegisterComponent(webProvider)

	return manager
}

// ProvideApp 提供应用实例
func ProvideApp(
	name string,
	version string,
	manager *ComponentManager,
	config config.Provider,
	logger logger.Logger,
) *App {
	app := New(name, version)
	app.SetComponentManager(manager)
	app.SetConfig(config)
	app.SetLogger(logger)

	return app
}

// 基础Provider集
var BaseSet = wire.NewSet(
	NewConfigProvider,
	NewLoggerProvider,
)

// 数据存储Provider集
var StorageSet = wire.NewSet(
	NewDBStoreProvider,
	NewCacheProvider,
)

// Web Provider集
var WebSet = wire.NewSet(
	NewWebProvider,
)

// 完整应用Provider集
var AppSet = wire.NewSet(
	BaseSet,
	StorageSet,
	WebSet,
	ProvideComponentManager,
	ProvideApp,
)
