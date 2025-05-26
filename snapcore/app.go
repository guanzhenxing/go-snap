package snapcore

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
)

// App 是SnapCore的核心类，负责管理整个应用生命周期
type App struct {
	name              string
	version           string
	components        map[string]Component
	dependencies      *DependencyGraph
	config            config.Provider
	logger            logger.Logger
	hooks             map[HookType][]HookFunc
	ctx               context.Context
	cancel            context.CancelFunc
	shutdownCh        chan os.Signal
	shutdownTimeout   int
	gracefulShutdown  bool
	monitorState      bool
	state             AppState
	stateMu           sync.RWMutex
	componentStates   map[string]ComponentState
	componentMu       sync.RWMutex
	stateListeners    []StateChangeListener
	configPath        string
	pendingComponents []ComponentRegistration
	plugins           []Plugin
	decorators        []ComponentDecorator
}

// 确保App实现ApplicationContext接口
var _ ApplicationContext = (*App)(nil)

// New 创建一个新的应用实例
func New(name string, version string) *App {
	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		name:              name,
		version:           version,
		components:        make(map[string]Component),
		dependencies:      NewDependencyGraph(),
		hooks:             make(map[HookType][]HookFunc),
		ctx:               ctx,
		cancel:            cancel,
		shutdownCh:        make(chan os.Signal, 1),
		shutdownTimeout:   30, // 默认30秒超时
		gracefulShutdown:  true,
		monitorState:      true,
		state:             AppStateCreated,
		componentStates:   make(map[string]ComponentState),
		pendingComponents: []ComponentRegistration{},
		plugins:           []Plugin{},
		decorators:        []ComponentDecorator{},
	}
}

// Run 运行应用
func (a *App) Run() error {
	// 初始化配置
	if err := a.initConfig(); err != nil {
		return err
	}

	// 初始化日志
	if err := a.initLogger(); err != nil {
		return err
	}

	// 注册组件
	if err := a.registerComponents(); err != nil {
		return err
	}

	// 执行初始化前钩子
	if err := a.runHooks(HookBeforeInitialize); err != nil {
		return err
	}

	// 设置应用状态
	a.setAppState(AppStateInitializing)

	// 初始化组件
	if err := a.initializeComponents(); err != nil {
		a.setAppState(AppStateFailed)
		return err
	}

	// 执行初始化后钩子
	if err := a.runHooks(HookAfterInitialize); err != nil {
		return err
	}

	// 设置应用状态
	a.setAppState(AppStateInitialized)
	a.setAppState(AppStateStarting)

	// 执行启动前钩子
	if err := a.runHooks(HookBeforeStart); err != nil {
		return err
	}

	// 启动组件
	if err := a.startComponents(); err != nil {
		a.setAppState(AppStateFailed)
		return err
	}

	// 执行启动后钩子
	if err := a.runHooks(HookAfterStart); err != nil {
		return err
	}

	// 设置应用状态
	a.setAppState(AppStateRunning)

	// 打印启动信息
	a.logger.Info("Application started",
		logger.String("name", a.name),
		logger.String("version", a.version),
	)

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

// Shutdown 停止应用
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down application",
		logger.String("name", a.name),
	)

	// 设置应用状态
	a.setAppState(AppStateStopping)

	// 执行关闭前钩子
	if err := a.runHooksWithContext(ctx, HookBeforeShutdown); err != nil {
		a.logger.Error("Error running before shutdown hooks",
			logger.String("error", err.Error()),
		)
	}

	// 关闭组件
	if err := a.stopComponents(ctx); err != nil {
		a.logger.Error("Error stopping components",
			logger.String("error", err.Error()),
		)
	}

	// 执行关闭后钩子
	if err := a.runHooksWithContext(ctx, HookAfterShutdown); err != nil {
		a.logger.Error("Error running after shutdown hooks",
			logger.String("error", err.Error()),
		)
	}

	// 取消上下文
	a.cancel()

	// 设置应用状态
	a.setAppState(AppStateStopped)

	a.logger.Info("Application stopped",
		logger.String("name", a.name),
	)

	return nil
}

// 初始化配置
func (a *App) initConfig() error {
	// 如果已经有配置提供器，则直接使用
	if a.config != nil {
		return nil
	}

	// 创建配置组件
	configComponent := NewConfigComponent(a.configPath)

	// 初始化配置组件
	if err := configComponent.Initialize(a.ctx); err != nil {
		return errors.Wrap(err, "initialize config component failed")
	}

	// 获取配置提供器
	a.config = configComponent.GetConfig()

	// 添加到组件列表
	a.components["config"] = configComponent

	return nil
}

// 初始化日志
func (a *App) initLogger() error {
	// 如果已经有日志器，则直接使用
	if a.logger != nil {
		return nil
	}

	// 创建日志组件
	logComponent := NewLoggerComponent()

	// 设置配置
	logComponent.SetConfig(a.config)

	// 初始化日志组件
	if err := logComponent.Initialize(a.ctx); err != nil {
		return errors.Wrap(err, "initialize logger component failed")
	}

	// 获取日志器
	a.logger = logComponent.GetLogger()

	// 添加到组件列表
	a.components["logger"] = logComponent

	return nil
}

// 注册组件
func (a *App) registerComponents() error {
	// 注册插件
	for _, plugin := range a.plugins {
		if err := plugin.Register(a); err != nil {
			return errors.Wrapf(err, "register plugin %s failed", plugin.Name())
		}
	}

	// 注册待处理的组件
	for _, reg := range a.pendingComponents {
		// 应用装饰器
		component := reg.Component
		for _, decorator := range a.decorators {
			component = decorator.Decorate(component)
		}

		// 设置日志器和配置（如果实现了相应接口）
		if loggerSetter, ok := component.(interface{ SetLogger(logger.Logger) }); ok && a.logger != nil {
			loggerSetter.SetLogger(a.logger)
		}

		if configSetter, ok := component.(interface{ SetConfig(config.Provider) }); ok && a.config != nil {
			configSetter.SetConfig(a.config)
		}

		// 添加到依赖图
		if err := a.dependencies.AddComponent(component, reg.Dependencies...); err != nil {
			return err
		}

		// 添加到组件映射
		a.components[component.Name()] = component

		// 初始化组件状态
		a.setComponentState(component.Name(), ComponentStateCreated)
	}

	return nil
}

// 初始化组件
func (a *App) initializeComponents() error {
	// 获取初始化顺序
	order, err := a.dependencies.SortComponentsForInitialization()
	if err != nil {
		return errors.Wrap(err, "sort components for initialization failed")
	}

	// 按顺序初始化组件
	for _, name := range order {
		component := a.components[name]

		// 设置组件状态
		a.setComponentState(name, ComponentStateInitializing)

		// 记录日志
		a.logger.Debug("Initializing component",
			logger.String("component", name),
		)

		// 初始化组件
		if err := component.Initialize(a.ctx); err != nil {
			a.logger.Error("Failed to initialize component",
				logger.String("component", name),
				logger.String("error", err.Error()),
			)
			return errors.Wrapf(err, "initialize component %s failed", name)
		}

		// 设置组件状态
		a.setComponentState(name, ComponentStateInitialized)
		a.dependencies.SetInitialized(name, true)

		a.logger.Debug("Component initialized",
			logger.String("component", name),
		)
	}

	return nil
}

// 启动组件
func (a *App) startComponents() error {
	// 获取初始化顺序（同样适用于启动）
	order, err := a.dependencies.SortComponentsForInitialization()
	if err != nil {
		return errors.Wrap(err, "sort components for starting failed")
	}

	// 按顺序启动组件
	for _, name := range order {
		component := a.components[name]

		// 设置组件状态
		a.setComponentState(name, ComponentStateStarting)

		// 记录日志
		a.logger.Debug("Starting component",
			logger.String("component", name),
		)

		// 启动组件
		if err := component.Start(a.ctx); err != nil {
			a.logger.Error("Failed to start component",
				logger.String("component", name),
				logger.String("error", err.Error()),
			)
			return errors.Wrapf(err, "start component %s failed", name)
		}

		// 设置组件状态
		a.setComponentState(name, ComponentStateRunning)
		a.dependencies.SetStarted(name, true)

		a.logger.Debug("Component started",
			logger.String("component", name),
		)
	}

	return nil
}

// 停止组件
func (a *App) stopComponents(ctx context.Context) error {
	// 获取关闭顺序
	order, err := a.dependencies.SortComponentsForShutdown()
	if err != nil {
		a.logger.Error("Failed to sort components for shutdown",
			logger.String("error", err.Error()),
		)
		// 继续使用初始化的逆序关闭
		initOrder, _ := a.dependencies.SortComponentsForInitialization()
		// 反转顺序
		order = make([]string, len(initOrder))
		for i, name := range initOrder {
			order[len(initOrder)-1-i] = name
		}
	}

	var lastErr error

	// 按顺序停止组件
	for _, name := range order {
		component, exists := a.components[name]
		if !exists || !a.dependencies.IsStarted(name) {
			continue
		}

		// 设置组件状态
		a.setComponentState(name, ComponentStateStopping)

		// 记录日志
		a.logger.Debug("Stopping component",
			logger.String("component", name),
		)

		// 停止组件
		if err := component.Stop(ctx); err != nil {
			a.logger.Error("Failed to stop component",
				logger.String("component", name),
				logger.String("error", err.Error()),
			)
			lastErr = err
			a.setComponentState(name, ComponentStateFailed)
		} else {
			// 设置组件状态
			a.setComponentState(name, ComponentStateStopped)
			a.dependencies.SetStarted(name, false)

			a.logger.Debug("Component stopped",
				logger.String("component", name),
			)
		}
	}

	return lastErr
}

// 执行钩子
func (a *App) runHooks(hookType HookType) error {
	return a.runHooksWithContext(a.ctx, hookType)
}

// 执行带有上下文的钩子
func (a *App) runHooksWithContext(ctx context.Context, hookType HookType) error {
	hooks := a.hooks[hookType]
	for i, hook := range hooks {
		if err := hook(ctx); err != nil {
			return errors.Wrapf(err, "run hook %d of type %d failed", i, hookType)
		}
	}
	return nil
}

// 设置应用状态
func (a *App) setAppState(state AppState) {
	a.stateMu.Lock()
	oldState := a.state
	a.state = state
	a.stateMu.Unlock()

	// 通知状态变更
	if a.monitorState {
		for _, listener := range a.stateListeners {
			listener(a.name, oldState, state)
		}
	}
}

// 设置组件状态
func (a *App) setComponentState(name string, state ComponentState) {
	a.componentMu.Lock()
	oldState, exists := a.componentStates[name]
	a.componentStates[name] = state
	a.componentMu.Unlock()

	// 通知状态变更
	if a.monitorState {
		for _, listener := range a.stateListeners {
			if exists {
				listener(name, oldState, state)
			} else {
				listener(name, ComponentStateCreated, state)
			}
		}
	}
}

// GetAppState 获取应用状态
func (a *App) GetAppState() AppState {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()
	return a.state
}

// GetComponentState 获取组件状态
func (a *App) GetComponentState(name string) ComponentState {
	a.componentMu.RLock()
	defer a.componentMu.RUnlock()
	state, exists := a.componentStates[name]
	if !exists {
		return ComponentStateCreated
	}
	return state
}

// RegisterStateChangeListener 注册状态变更监听器
func (a *App) RegisterStateChangeListener(listener StateChangeListener) {
	a.stateListeners = append(a.stateListeners, listener)
}

// GetComponent 获取指定名称的组件
func (a *App) GetComponent(name string) (Component, bool) {
	component, exists := a.components[name]
	return component, exists
}

// GetComponentByType 获取指定类型的第一个组件
func (a *App) GetComponentByType(t ComponentType) (Component, bool) {
	for _, component := range a.components {
		if component.Type() == t {
			return component, true
		}
	}
	return nil, false
}

// GetComponentsByType 获取指定类型的所有组件
func (a *App) GetComponentsByType(t ComponentType) []Component {
	return a.dependencies.GetComponentsByType(t)
}

// GetAllComponents 获取所有组件
func (a *App) GetAllComponents() map[string]Component {
	return a.dependencies.GetAllComponents()
}

// WithConfigFile 设置配置文件路径
func (a *App) WithConfigFile(path string) *App {
	a.configPath = path
	return a
}

// WithConfig 设置配置提供器
func (a *App) WithConfig(provider config.Provider) *App {
	a.config = provider
	return a
}

// WithLogger 设置日志器
func (a *App) WithLogger(logger logger.Logger) *App {
	a.logger = logger
	return a
}

// WithComponent 注册组件
func (a *App) WithComponent(name string, component Component, dependencies ...string) *App {
	// 组件将在App.Run()时添加到依赖图中
	a.pendingComponents = append(a.pendingComponents, ComponentRegistration{
		Name:         name,
		Component:    component,
		Dependencies: dependencies,
	})
	return a
}

// WithHook 添加钩子
func (a *App) WithHook(hookType HookType, hookFunc HookFunc) *App {
	a.hooks[hookType] = append(a.hooks[hookType], hookFunc)
	return a
}

// WithShutdownTimeout 设置关闭超时时间
func (a *App) WithShutdownTimeout(timeout int) *App {
	a.shutdownTimeout = timeout
	return a
}

// WithGracefulShutdown 设置是否启用优雅关闭
func (a *App) WithGracefulShutdown(enabled bool) *App {
	a.gracefulShutdown = enabled
	return a
}

// WithStateMonitor 设置状态监控器
func (a *App) WithStateMonitor(enabled bool) *App {
	a.monitorState = enabled
	return a
}

// WithPlugin 添加插件
func (a *App) WithPlugin(plugin Plugin) *App {
	a.plugins = append(a.plugins, plugin)
	return a
}

// WithDecorator 添加组件装饰器
func (a *App) WithDecorator(decorator ComponentDecorator) *App {
	a.decorators = append(a.decorators, decorator)
	return a
}
