// Package boot 提供应用程序启动和生命周期管理功能。
// 该包实现了一个类似Spring Boot的应用程序引导框架，支持自动配置、组件管理和生命周期控制。
//
// # 应用生命周期
//
// 应用程序生命周期分为以下几个阶段：
//
// 1. 创建阶段：创建应用实例和基础设施组件
//   - 加载配置文件
//   - 创建组件注册表
//   - 创建事件总线
//   - 创建自动配置引擎
//
// 2. 初始化阶段：应用配置和组件注册
//   - 注册自动配置器
//   - 注册用户定义的组件
//   - 配置所有组件
//   - 解析组件依赖关系
//
// 3. 启动阶段：按照依赖顺序启动组件
//   - 按类型优先级启动组件（基础设施 -> 数据源 -> 核心 -> Web）
//   - 发布ApplicationStartedEvent事件
//   - 启动健康检查
//
// 4. 运行阶段：应用正常运行
//   - 处理请求
//   - 定期执行健康检查
//   - 监听关闭信号
//
// 5. 关闭阶段：优雅停止应用
//   - 发布ApplicationStoppingEvent事件
//   - 反向顺序停止组件
//   - 释放资源
//   - 发布ApplicationStoppedEvent事件
//
// # 使用示例
//
//	package main
//
//	import (
//	    "github.com/guanzhenxing/go-snap/boot"
//	    "github.com/guanzhenxing/go-snap/web"
//	)
//
//	func main() {
//	    // 创建启动器
//	    bootApp := boot.NewBoot()
//
//	    // 配置启动器
//	    bootApp.SetConfigPath("./config")
//	    bootApp.AddComponent(&MyCustomComponent{})
//
//	    // 运行应用
//	    if err := bootApp.Run(); err != nil {
//	        panic(err)
//	    }
//	}
//
// # 扩展点
//
// 应用程序提供以下主要扩展点：
//
// 1. 组件 (Component)：实现特定功能的单元
// 2. 自动配置器 (AutoConfigurer)：负责自动配置特定类型的组件
// 3. 组件激活器 (ComponentActivator)：控制组件是否被激活
// 4. 插件 (Plugin)：提供一组相关功能的捆绑包
package boot

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Boot 应用启动器
// 是应用程序的主要入口点，负责配置和启动应用
// 提供流式API，允许链式调用配置方法
type Boot struct {
	// configPath 配置文件所在路径
	configPath string

	// components 用户定义的组件列表
	components []Component

	// plugins 用户定义的插件列表
	plugins []Plugin

	// configurers 自定义配置器列表
	configurers []AutoConfigurer

	// activators 组件激活器列表
	activators []ComponentActivator
}

// NewBoot 创建启动器
// 返回一个配置了默认值的Boot实例
//
// 默认配置：
//   - configPath: "configs"
//   - 空的组件、插件、配置器和激活器列表
//
// 示例：
//
//	bootApp := boot.NewBoot()
func NewBoot() *Boot {
	return &Boot{
		configPath:  "configs",
		components:  []Component{},
		plugins:     []Plugin{},
		configurers: []AutoConfigurer{},
		activators:  []ComponentActivator{},
	}
}

// SetConfigPath 设置配置路径
// 指定配置文件所在的目录路径
//
// 参数：
//   - path: 配置文件目录的路径
//
// 返回：
//   - *Boot: 启动器实例，用于链式调用
//
// 示例：
//
//	bootApp.SetConfigPath("./configs")
func (b *Boot) SetConfigPath(path string) *Boot {
	b.configPath = path
	return b
}

// AddComponent 添加自定义组件
// 将组件实例添加到应用中
//
// 参数：
//   - component: 实现Component接口的组件实例
//
// 返回：
//   - *Boot: 启动器实例，用于链式调用
//
// 示例：
//
//	bootApp.AddComponent(&MyDbComponent{})
//	      .AddComponent(&MyWebComponent{})
func (b *Boot) AddComponent(component Component) *Boot {
	b.components = append(b.components, component)
	return b
}

// AddPlugin 添加插件
// 将插件添加到应用中
//
// 参数：
//   - plugin: 实现Plugin接口的插件实例
//
// 返回：
//   - *Boot: 启动器实例，用于链式调用
//
// 插件与组件的区别：
//   - 组件是单一功能单元
//   - 插件可以注册多个组件和配置器
//
// 示例：
//
//	bootApp.AddPlugin(myplugin.New())
func (b *Boot) AddPlugin(plugin Plugin) *Boot {
	b.plugins = append(b.plugins, plugin)
	return b
}

// AddConfigurer 添加配置器
// 将自定义配置器添加到应用中
//
// 参数：
//   - configurer: 实现AutoConfigurer接口的配置器实例
//
// 返回：
//   - *Boot: 启动器实例，用于链式调用
//
// 示例：
//
//	bootApp.AddConfigurer(&MyCustomConfigurer{})
func (b *Boot) AddConfigurer(configurer AutoConfigurer) *Boot {
	b.configurers = append(b.configurers, configurer)
	return b
}

// AddActivator 添加激活器
// 将组件激活器添加到应用中
//
// 参数：
//   - activator: 实现ComponentActivator接口的激活器实例
//
// 返回：
//   - *Boot: 启动器实例，用于链式调用
//
// 激活器用于控制组件是否应该被激活，可以基于条件判断
//
// 示例：
//
//	// 只在开发环境激活某些组件
//	bootApp.AddActivator(&DevEnvActivator{})
func (b *Boot) AddActivator(activator ComponentActivator) *Boot {
	b.activators = append(b.activators, activator)
	return b
}

// Run 运行应用
// 创建并启动应用程序，会阻塞直到应用关闭
//
// 返回：
//   - error: 启动过程中遇到的错误，如果正常关闭则返回nil
//
// 执行流程：
//  1. 创建应用实例
//  2. 初始化所有组件
//  3. 启动所有组件
//  4. 等待关闭信号
//  5. 优雅关闭应用
//
// 示例：
//
//	if err := bootApp.Run(); err != nil {
//	    log.Fatalf("应用启动失败: %v", err)
//	}
func (b *Boot) Run() error {
	// 创建应用
	app, err := b.createApplication()
	if err != nil {
		return err
	}

	// 运行应用
	return app.Run()
}

// Initialize 初始化应用并返回应用实例（不启动）
// 与Run不同，此方法仅初始化应用但不启动组件
//
// 返回：
//   - *Application: 初始化后的应用实例
//   - error: 初始化过程中遇到的错误
//
// 使用场景：
//   - 需要在启动前执行自定义操作
//   - 单元测试中模拟应用环境
//
// 示例：
//
//	app, err := bootApp.Initialize()
//	if err != nil {
//	    log.Fatalf("应用初始化失败: %v", err)
//	}
//
//	// 执行自定义操作
//
//	if err := app.Run(); err != nil {
//	    log.Fatalf("应用运行失败: %v", err)
//	}
func (b *Boot) Initialize() (*Application, error) {
	// 创建应用
	app, err := b.createApplication()
	if err != nil {
		return nil, err
	}

	// 初始化应用
	if err := app.Initialize(); err != nil {
		return nil, err
	}

	return app, nil
}

// createApplication 创建应用实例
// 配置应用程序并注册标准和自定义组件
//
// 返回：
//   - *Application: 创建的应用实例
//   - error: 创建过程中遇到的错误
//
// 内部执行流程：
//  1. 创建Application实例
//  2. 添加标准配置器
//  3. 添加自定义配置器
//  4. 添加激活器
//  5. 注册自定义组件
//  6. 注册插件
func (b *Boot) createApplication() (*Application, error) {
	// 创建应用
	app, err := NewApplication(b.configPath)
	if err != nil {
		return nil, err
	}

	// 添加标准配置器
	app.AddConfigurer(&ConfigConfigurer{})
	app.AddConfigurer(&LoggerConfigurer{})
	app.AddConfigurer(&DBStoreConfigurer{})
	app.AddConfigurer(&CacheConfigurer{})
	app.AddConfigurer(&WebConfigurer{})

	// 添加自定义配置器
	for _, configurer := range b.configurers {
		app.AddConfigurer(configurer)
	}

	// 添加激活器
	for _, activator := range b.activators {
		app.AddActivator(activator)
	}

	// 注册自定义组件
	for _, component := range b.components {
		if err := app.RegisterComponent(component); err != nil {
			log.Printf("Warning: Failed to register component %s: %v", component.Name(), err)
		}
	}

	// 注册插件
	for _, plugin := range b.plugins {
		if err := plugin.Register(app); err != nil {
			log.Printf("Warning: Failed to register plugin %s: %v", plugin.Name(), err)
		}
	}

	return app, nil
}

// Application 应用类
// 表示一个运行中的应用实例，管理组件生命周期和应用状态
type Application struct {
	// name 应用名称
	name string

	// version 应用版本
	version string

	// registry 组件注册表，管理所有组件
	registry *ComponentRegistry

	// propSource 属性源，提供配置
	propSource PropertySource

	// autoConfig 自动配置引擎
	autoConfig *AutoConfig

	// eventBus 事件总线，用于组件间通信
	eventBus *EventBus

	// ctx 应用上下文，用于控制生命周期
	ctx context.Context

	// cancel 取消函数，用于停止应用
	cancel context.CancelFunc

	// state 当前应用状态
	state AppState

	// stateMu 保护状态访问的互斥锁
	stateMu sync.RWMutex

	// shutdownCh 关闭信号通道
	shutdownCh chan os.Signal

	// shutdownTimeout 关闭超时时间（秒）
	shutdownTimeout int

	// healthChecker 健康检查器
	healthChecker *ApplicationHealthChecker

	// metrics 应用指标
	metrics *ApplicationMetrics
}

// ApplicationHealthChecker 应用健康检查器
// 负责定期检查所有组件的健康状态
type ApplicationHealthChecker struct {
	// registry 组件注册表的引用
	registry *ComponentRegistry

	// checkInterval 检查间隔时间
	checkInterval time.Duration

	// stopCh 停止通道，用于停止健康检查
	stopCh chan struct{}

	// mutex 保护健康状态的互斥锁
	mutex sync.RWMutex

	// lastCheck 上次检查时间
	lastCheck time.Time

	// healthStatus 最近的健康检查结果
	healthStatus map[string]error
}

// ApplicationMetrics 应用指标
// 收集应用运行时的各种指标数据
type ApplicationMetrics struct {
	// StartTime 应用启动时间
	StartTime time.Time

	// ComponentCount 组件数量
	ComponentCount int

	// HealthCheckCount 健康检查次数
	HealthCheckCount int

	// ErrorCount 错误数量
	ErrorCount int

	// mutex 保护指标的互斥锁
	mutex sync.RWMutex
}

// NewApplication 创建应用
// 初始化应用程序的基础设施组件
//
// 参数：
//   - configPath: 配置文件路径
//
// 返回：
//   - *Application: 创建的应用实例
//   - error: 创建过程中遇到的错误
//
// 内部执行流程：
//  1. 创建上下文和取消函数
//  2. 加载属性源
//  3. 加载环境变量
//  4. 创建组件注册表
//  5. 创建事件总线
//  6. 创建自动配置引擎
//  7. 创建健康检查器
//  8. 初始化应用指标
func NewApplication(configPath string) (*Application, error) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建属性源
	propSource, err := NewFilePropertySource(configPath)
	if err != nil {
		cancel()
		return nil, err
	}

	// 加载环境变量
	LoadEnvironmentVariables(propSource)

	// 应用名称和版本
	name := propSource.GetString("app.name", "GoBootApp")
	version := propSource.GetString("app.version", "1.0.0")

	// 创建组件注册表
	registry := NewComponentRegistry(ctx, propSource)

	// 创建事件总线
	eventBus := NewEventBus()

	// 创建自动配置引擎
	autoConfig := NewAutoConfig()

	// 创建健康检查器
	healthChecker := &ApplicationHealthChecker{
		registry:      registry,
		checkInterval: time.Duration(propSource.GetInt("app.health_check_interval", 30)) * time.Second,
		stopCh:        make(chan struct{}),
		healthStatus:  make(map[string]error),
	}

	// 创建指标收集器
	metrics := &ApplicationMetrics{
		StartTime: time.Now(),
	}

	return &Application{
		name:            name,
		version:         version,
		registry:        registry,
		propSource:      propSource,
		autoConfig:      autoConfig,
		eventBus:        eventBus,
		ctx:             ctx,
		cancel:          cancel,
		state:           AppStateCreated,
		shutdownCh:      make(chan os.Signal, 1),
		shutdownTimeout: propSource.GetInt("app.shutdown_timeout", 30),
		healthChecker:   healthChecker,
		metrics:         metrics,
	}, nil
}

// GetName 获取应用名称
func (a *Application) GetName() string {
	return a.name
}

// GetVersion 获取应用版本
func (a *Application) GetVersion() string {
	return a.version
}

// GetComponent 获取组件
func (a *Application) GetComponent(name string) (Component, bool) {
	return a.registry.GetComponent(name)
}

// GetPropertySource 获取属性源
func (a *Application) GetPropertySource() PropertySource {
	return a.propSource
}

// GetEventBus 获取事件总线
func (a *Application) GetEventBus() *EventBus {
	return a.eventBus
}

// GetRegistry 获取组件注册表
func (a *Application) GetRegistry() *ComponentRegistry {
	return a.registry
}

// RegisterComponent 注册组件
func (a *Application) RegisterComponent(component Component) error {
	return a.registry.RegisterComponent(component)
}

// AddConfigurer 添加配置器
func (a *Application) AddConfigurer(configurer AutoConfigurer) {
	a.autoConfig.AddConfigurer(configurer)
}

// AddActivator 添加激活器
func (a *Application) AddActivator(activator ComponentActivator) {
	a.autoConfig.AddActivator(activator)
}

// Initialize 初始化应用
func (a *Application) Initialize() error {
	// 设置应用状态
	a.setState(AppStateInitializing)

	// 执行自动配置
	if err := a.autoConfig.Configure(a.registry, a.propSource); err != nil {
		a.setState(AppStateFailed)
		return NewConfigError("Application", "自动配置失败", err)
	}

	// 解析组件依赖
	if err := a.registry.ResolveDependencies(); err != nil {
		a.setState(AppStateFailed)
		return NewConfigError("Application", "依赖解析失败", err)
	}

	// 初始化组件
	if err := a.initializeComponents(); err != nil {
		a.setState(AppStateFailed)
		return err
	}

	// 设置应用状态
	a.setState(AppStateInitialized)

	// 发布初始化完成事件
	a.eventBus.Publish("application.initialized", a)

	return nil
}

// initializeComponents 初始化组件
func (a *Application) initializeComponents() error {
	components := a.registry.GetAllComponentsSorted()

	for _, component := range components {
		if err := component.Initialize(a.ctx); err != nil {
			return NewComponentError(
				component.Name(),
				"initialize",
				"组件初始化失败",
				err,
			)
		}
	}

	// 更新指标
	a.metrics.mutex.Lock()
	a.metrics.ComponentCount = len(components)
	a.metrics.mutex.Unlock()

	return nil
}

// Run 运行应用
func (a *Application) Run() error {
	// 初始化应用
	if a.GetState() < AppStateInitialized {
		if err := a.Initialize(); err != nil {
			return err
		}
	}

	// 设置应用状态
	a.setState(AppStateStarting)

	// 启动组件
	if err := a.startComponents(); err != nil {
		a.setState(AppStateFailed)
		return err
	}

	// 启动健康检查
	a.startHealthChecker()

	// 设置应用状态
	a.setState(AppStateRunning)

	// 发布应用启动事件
	a.eventBus.Publish("application.started", a)

	// 设置信号处理
	signal.Notify(a.shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("应用 %s (版本 %s) 已启动", a.name, a.version)

	// 等待信号
	<-a.shutdownCh

	log.Printf("收到停止信号，开始关闭应用...")

	// 关闭应用
	shutdownCtx := context.Background()
	if a.shutdownTimeout > 0 {
		var cancel context.CancelFunc
		shutdownCtx, cancel = context.WithTimeout(context.Background(), time.Duration(a.shutdownTimeout)*time.Second)
		defer cancel()
	}

	return a.Shutdown(shutdownCtx)
}

// startComponents 启动组件
func (a *Application) startComponents() error {
	components := a.registry.GetAllComponentsSorted()

	for _, component := range components {
		if err := component.Start(a.ctx); err != nil {
			return NewComponentError(
				component.Name(),
				"start",
				"组件启动失败",
				err,
			)
		}
	}

	return nil
}

// startHealthChecker 启动健康检查器
func (a *Application) startHealthChecker() {
	go func() {
		ticker := time.NewTicker(a.healthChecker.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				a.performHealthCheck()
			case <-a.healthChecker.stopCh:
				return
			case <-a.ctx.Done():
				return
			}
		}
	}()
}

// performHealthCheck 执行健康检查
func (a *Application) performHealthCheck() {
	a.healthChecker.mutex.Lock()
	defer a.healthChecker.mutex.Unlock()

	a.healthChecker.lastCheck = time.Now()
	healthResults := a.registry.HealthCheck()

	// 更新健康状态
	a.healthChecker.healthStatus = healthResults

	// 更新指标
	a.metrics.mutex.Lock()
	a.metrics.HealthCheckCount++
	if len(healthResults) > 0 {
		a.metrics.ErrorCount += len(healthResults)
	}
	a.metrics.mutex.Unlock()

	// 发布健康检查事件
	if len(healthResults) > 0 {
		a.eventBus.Publish("application.health_check.failed", healthResults)
	} else {
		a.eventBus.Publish("application.health_check.passed", nil)
	}
}

// GetHealthStatus 获取健康状态
func (a *Application) GetHealthStatus() map[string]error {
	a.healthChecker.mutex.RLock()
	defer a.healthChecker.mutex.RUnlock()

	result := make(map[string]error)
	for k, v := range a.healthChecker.healthStatus {
		result[k] = v
	}
	return result
}

// GetMetrics 获取应用指标
func (a *Application) GetMetrics() map[string]interface{} {
	a.metrics.mutex.RLock()
	defer a.metrics.mutex.RUnlock()

	registryMetrics := a.registry.GetMetrics()

	return map[string]interface{}{
		"app_name":           a.name,
		"app_version":        a.version,
		"app_state":          a.GetState().String(),
		"start_time":         a.metrics.StartTime,
		"uptime":             time.Since(a.metrics.StartTime).String(),
		"component_count":    a.metrics.ComponentCount,
		"health_check_count": a.metrics.HealthCheckCount,
		"error_count":        a.metrics.ErrorCount,
		"registry_metrics":   registryMetrics,
	}
}

// Shutdown 关闭应用
func (a *Application) Shutdown(ctx context.Context) error {
	// 设置应用状态
	a.setState(AppStateStopping)

	// 发布应用停止事件
	a.eventBus.PublishSync("application.stopping", a)

	// 停止健康检查器
	close(a.healthChecker.stopCh)

	// 停止组件
	if err := a.stopComponents(ctx); err != nil {
		log.Printf("停止组件时发生错误: %v", err)
	}

	// 取消上下文
	a.cancel()

	// 设置应用状态
	a.setState(AppStateStopped)

	// 发布应用已停止事件
	a.eventBus.PublishSync("application.stopped", a)

	log.Printf("应用 %s 已停止", a.name)

	return nil
}

// stopComponents 停止组件
func (a *Application) stopComponents(ctx context.Context) error {
	components := a.registry.GetAllComponentsForShutdown()

	var errors []error
	for _, component := range components {
		if err := component.Stop(ctx); err != nil {
			// 记录错误，继续停止其他组件
			componentErr := NewComponentError(
				component.Name(),
				"stop",
				"组件停止失败",
				err,
			)
			errors = append(errors, componentErr)

			a.eventBus.Publish("component.stop.error", map[string]interface{}{
				"component": component.Name(),
				"error":     componentErr,
			})
		}
	}

	// 如果有错误，返回第一个错误
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// GetState 获取应用状态
func (a *Application) GetState() AppState {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()
	return a.state
}

// setState 设置应用状态
func (a *Application) setState(state AppState) {
	a.stateMu.Lock()
	oldState := a.state
	a.state = state
	a.stateMu.Unlock()

	// 发布状态变更事件
	if oldState != state {
		a.eventBus.Publish("application.state.changed", map[string]interface{}{
			"oldState": oldState,
			"newState": state,
		})
	}
}
