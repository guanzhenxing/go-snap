package appcore

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
	name             string
	version          string
	componentManager *ComponentManager
	config           config.Provider
	logger           logger.Logger
	hooks            map[HookType][]HookFunc
	ctx              context.Context
	cancel           context.CancelFunc
	shutdownCh       chan os.Signal
	shutdownTimeout  int
	gracefulShutdown bool
	state            AppState
	stateMu          sync.RWMutex
	stateListeners   []StateChangeListener
}

// 确保App实现ApplicationContext接口
var _ ApplicationContext = (*App)(nil)

// New 创建一个新的应用实例
func New(name string, version string) *App {
	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		name:             name,
		version:          version,
		hooks:            make(map[HookType][]HookFunc),
		ctx:              ctx,
		cancel:           cancel,
		shutdownCh:       make(chan os.Signal, 1),
		shutdownTimeout:  30, // 默认30秒超时
		gracefulShutdown: true,
		state:            AppStateCreated,
		stateListeners:   []StateChangeListener{},
	}
}

// SetComponentManager 设置组件管理器
func (a *App) SetComponentManager(manager *ComponentManager) {
	a.componentManager = manager
}

// SetConfig 设置配置
func (a *App) SetConfig(config config.Provider) {
	a.config = config
}

// SetLogger 设置日志
func (a *App) SetLogger(logger logger.Logger) {
	a.logger = logger
}

// Run 运行应用
func (a *App) Run() error {
	// 执行初始化前钩子
	if err := a.runHooks(HookBeforeInitialize); err != nil {
		return err
	}

	// 设置应用状态
	a.setAppState(AppStateInitializing)

	// 注意：组件已通过Wire初始化，这里不再执行初始化

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

// 启动组件
func (a *App) startComponents() error {
	components := a.componentManager.GetAllComponentsSorted()

	for _, component := range components {
		a.logger.Info("Starting component",
			logger.String("component", component.Name()))

		if err := component.Start(a.ctx); err != nil {
			return errors.Wrapf(err, "failed to start component %s", component.Name())
		}
	}

	return nil
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

// 停止组件
func (a *App) stopComponents(ctx context.Context) error {
	components := a.componentManager.GetAllComponentsForShutdown()

	for _, component := range components {
		a.logger.Info("Stopping component",
			logger.String("component", component.Name()))

		if err := component.Stop(ctx); err != nil {
			a.logger.Error("Error stopping component",
				logger.String("component", component.Name()),
				logger.String("error", err.Error()),
			)
		}
	}

	return nil
}

// runHooks 运行钩子
func (a *App) runHooks(hookType HookType) error {
	return a.runHooksWithContext(a.ctx, hookType)
}

// runHooksWithContext 使用指定上下文运行钩子
func (a *App) runHooksWithContext(ctx context.Context, hookType HookType) error {
	hooks := a.hooks[hookType]
	for _, hook := range hooks {
		if err := hook(ctx); err != nil {
			return err
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

	// 通知状态监听器
	for _, listener := range a.stateListeners {
		listener(a.name, oldState, state)
	}
}

// GetAppState 获取应用状态
func (a *App) GetAppState() AppState {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()
	return a.state
}

// GetComponent 获取组件
func (a *App) GetComponent(name string) (Component, bool) {
	return a.componentManager.GetComponent(name)
}

// GetComponentByType 获取指定类型的组件
func (a *App) GetComponentByType(t ComponentType) (Component, bool) {
	return a.componentManager.GetComponentByType(t)
}

// GetComponentsByType 获取指定类型的所有组件
func (a *App) GetComponentsByType(t ComponentType) []Component {
	return a.componentManager.GetComponentsByType(t)
}

// RegisterStateChangeListener 注册状态变更监听器
func (a *App) RegisterStateChangeListener(listener StateChangeListener) {
	a.stateListeners = append(a.stateListeners, listener)
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
