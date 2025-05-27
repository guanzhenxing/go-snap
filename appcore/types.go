// Package snapcore 提供了应用生命周期管理、组件协调和配置分发功能
package appcore

import (
	"context"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/dbstore"
	"github.com/guanzhenxing/go-snap/logger"
	"github.com/guanzhenxing/go-snap/web"
)

// ComponentType 组件类型枚举，用于控制初始化和关闭顺序
type ComponentType int

const (
	// ComponentTypeInfrastructure 基础设施组件（如日志、监控）
	ComponentTypeInfrastructure ComponentType = iota
	// ComponentTypeDataSource 数据源组件（如数据库、缓存）
	ComponentTypeDataSource
	// ComponentTypeCore 核心业务组件
	ComponentTypeCore
	// ComponentTypeWeb API/Web 服务组件
	ComponentTypeWeb
)

// Component 所有可被SnapCore管理的组件都应实现此接口
type Component interface {
	// Initialize 初始化组件
	Initialize(ctx context.Context) error

	// Start 启动组件
	Start(ctx context.Context) error

	// Stop 停止组件
	Stop(ctx context.Context) error

	// Name 获取组件名称
	Name() string

	// Type 获取组件类型
	Type() ComponentType
}

// 组件功能接口定义

// ConfigProvider 配置组件接口
type ConfigProvider interface {
	Component
	GetConfig() config.Provider
}

// LoggerProvider 日志组件接口
type LoggerProvider interface {
	Component
	GetLogger() logger.Logger
}

// DBStoreProvider 数据库组件接口
type DBStoreProvider interface {
	Component
	GetDBStore() *dbstore.Store
}

// CacheProvider 缓存组件接口
type CacheProvider interface {
	Component
	GetCache() interface{} // 根据实际缓存接口调整
}

// WebProvider Web服务组件接口
type WebProvider interface {
	Component
	GetServer() *web.Server
}

// HookType 定义生命周期钩子类型
type HookType int

const (
	// HookBeforeInitialize 应用初始化前
	HookBeforeInitialize HookType = iota

	// HookAfterInitialize 应用初始化后
	HookAfterInitialize

	// HookBeforeStart 组件启动前
	HookBeforeStart

	// HookAfterStart 组件启动后
	HookAfterStart

	// HookBeforeShutdown 应用关闭前
	HookBeforeShutdown

	// HookAfterShutdown 应用关闭后
	HookAfterShutdown
)

// HookFunc 钩子函数签名
type HookFunc func(ctx context.Context) error

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

// StateChangeListener 状态变更监听器
type StateChangeListener func(name string, oldState, newState interface{})

// ApplicationContext 应用上下文接口
type ApplicationContext interface {
	// GetComponent 获取组件
	GetComponent(name string) (Component, bool)

	// GetComponentByType 获取指定类型的组件
	GetComponentByType(t ComponentType) (Component, bool)

	// GetComponentsByType 获取指定类型的所有组件
	GetComponentsByType(t ComponentType) []Component

	// GetAppState 获取应用状态
	GetAppState() AppState

	// RegisterStateChangeListener 注册状态变更监听器
	RegisterStateChangeListener(listener StateChangeListener)
}
