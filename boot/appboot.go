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
type Boot struct {
	configPath  string
	components  []Component
	plugins     []Plugin
	configurers []AutoConfigurer
	activators  []ComponentActivator
}

// NewBoot 创建启动器
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
func (b *Boot) SetConfigPath(path string) *Boot {
	b.configPath = path
	return b
}

// AddComponent 添加自定义组件
func (b *Boot) AddComponent(component Component) *Boot {
	b.components = append(b.components, component)
	return b
}

// AddPlugin 添加插件
func (b *Boot) AddPlugin(plugin Plugin) *Boot {
	b.plugins = append(b.plugins, plugin)
	return b
}

// AddConfigurer 添加配置器
func (b *Boot) AddConfigurer(configurer AutoConfigurer) *Boot {
	b.configurers = append(b.configurers, configurer)
	return b
}

// AddActivator 添加激活器
func (b *Boot) AddActivator(activator ComponentActivator) *Boot {
	b.activators = append(b.activators, activator)
	return b
}

// Run 运行应用
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
type Application struct {
	name            string
	version         string
	registry        *ComponentRegistry
	propSource      PropertySource
	autoConfig      *AutoConfig
	eventBus        *EventBus
	ctx             context.Context
	cancel          context.CancelFunc
	state           AppState
	stateMu         sync.RWMutex
	shutdownCh      chan os.Signal
	shutdownTimeout int
}

// NewApplication 创建应用
func NewApplication(configPath string) (*Application, error) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 确保在所有错误路径上调用cancel

	// 创建属性源
	propSource, err := NewFilePropertySource(configPath)
	if err != nil {
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

	// 创建应用实例，并保留cancel函数，避免被defer调用
	cancelFunc := cancel
	cancel = func() {} // 防止defer调用原始cancel

	return &Application{
		name:            name,
		version:         version,
		registry:        registry,
		propSource:      propSource,
		autoConfig:      autoConfig,
		eventBus:        eventBus,
		ctx:             ctx,
		cancel:          cancelFunc,
		state:           AppStateCreated,
		shutdownCh:      make(chan os.Signal, 1),
		shutdownTimeout: propSource.GetInt("app.shutdown_timeout", 30),
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
		return err
	}

	// 解析组件依赖
	if err := a.registry.ResolveDependencies(); err != nil {
		a.setState(AppStateFailed)
		return err
	}

	// 初始化组件
	if err := a.initializeComponents(); err != nil {
		a.setState(AppStateFailed)
		return err
	}

	// 设置应用状态
	a.setState(AppStateInitialized)

	return nil
}

// 初始化组件
func (a *Application) initializeComponents() error {
	components := a.registry.GetAllComponentsSorted()

	for _, component := range components {
		if err := component.Initialize(a.ctx); err != nil {
			return &ConfigError{
				Message: "初始化组件失败: " + component.Name(),
				Cause:   err,
			}
		}
	}

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

	// 设置应用状态
	a.setState(AppStateRunning)

	// 发布应用启动事件
	a.eventBus.Publish("application.started", a)

	// 设置信号处理
	signal.Notify(a.shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-a.shutdownCh

	// 关闭应用
	shutdownCtx := context.Background()
	if a.shutdownTimeout > 0 {
		var cancel context.CancelFunc
		shutdownCtx, cancel = context.WithTimeout(context.Background(), time.Duration(a.shutdownTimeout)*time.Second)
		defer cancel()
	}

	return a.Shutdown(shutdownCtx)
}

// 启动组件
func (a *Application) startComponents() error {
	components := a.registry.GetAllComponentsSorted()

	for _, component := range components {
		if err := component.Start(a.ctx); err != nil {
			return &ConfigError{
				Message: "启动组件失败: " + component.Name(),
				Cause:   err,
			}
		}
	}

	return nil
}

// Shutdown 关闭应用
func (a *Application) Shutdown(ctx context.Context) error {
	// 设置应用状态
	a.setState(AppStateStopping)

	// 发布应用停止事件
	a.eventBus.PublishSync("application.stopping", a)

	// 停止组件
	if err := a.stopComponents(ctx); err != nil {
		return err
	}

	// 取消上下文
	a.cancel()

	// 设置应用状态
	a.setState(AppStateStopped)

	// 发布应用已停止事件
	a.eventBus.PublishSync("application.stopped", a)

	return nil
}

// 停止组件
func (a *Application) stopComponents(ctx context.Context) error {
	components := a.registry.GetAllComponentsForShutdown()

	for _, component := range components {
		if err := component.Stop(ctx); err != nil {
			// 仅记录错误，继续停止其他组件
			a.eventBus.Publish("component.stop.error", map[string]interface{}{
				"component": component.Name(),
				"error":     err,
			})
		}
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
