package boot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/guanzhenxing/go-snap/cache"
	"github.com/guanzhenxing/go-snap/config"
	"github.com/guanzhenxing/go-snap/logger"
)

// BaseComponent 基础组件实现
type BaseComponent struct {
	name      string
	compType  ComponentType
	status    ComponentStatus
	statusMu  sync.RWMutex
	metrics   map[string]interface{}
	metricsMu sync.RWMutex
	startTime time.Time
}

// NewBaseComponent 创建基础组件
func NewBaseComponent(name string, compType ComponentType) *BaseComponent {
	return &BaseComponent{
		name:     name,
		compType: compType,
		status:   ComponentStatusCreated,
		metrics:  make(map[string]interface{}),
	}
}

// Name 返回组件名称
func (c *BaseComponent) Name() string {
	return c.name
}

// Type 返回组件类型
func (c *BaseComponent) Type() ComponentType {
	return c.compType
}

// GetStatus 获取组件状态
func (c *BaseComponent) GetStatus() ComponentStatus {
	c.statusMu.RLock()
	defer c.statusMu.RUnlock()
	return c.status
}

// SetStatus 设置组件状态
func (c *BaseComponent) SetStatus(status ComponentStatus) {
	c.statusMu.Lock()
	defer c.statusMu.Unlock()
	c.status = status
}

// GetMetrics 获取组件指标
func (c *BaseComponent) GetMetrics() map[string]interface{} {
	c.metricsMu.RLock()
	defer c.metricsMu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range c.metrics {
		result[k] = v
	}

	// 添加基础指标
	result["status"] = c.GetStatus().String()
	result["name"] = c.name
	result["type"] = c.compType
	if !c.startTime.IsZero() {
		result["start_time"] = c.startTime
		result["uptime"] = time.Since(c.startTime)
	}

	return result
}

// SetMetric 设置指标
func (c *BaseComponent) SetMetric(key string, value interface{}) {
	c.metricsMu.Lock()
	defer c.metricsMu.Unlock()
	c.metrics[key] = value
}

// Initialize 初始化组件
func (c *BaseComponent) Initialize(ctx context.Context) error {
	c.SetStatus(ComponentStatusInitialized)
	c.SetMetric("initialized_at", time.Now())
	return nil
}

// Start 启动组件
func (c *BaseComponent) Start(ctx context.Context) error {
	c.SetStatus(ComponentStatusStarted)
	c.startTime = time.Now()
	c.SetMetric("started_at", c.startTime)
	return nil
}

// Stop 停止组件
func (c *BaseComponent) Stop(ctx context.Context) error {
	c.SetStatus(ComponentStatusStopped)
	c.SetMetric("stopped_at", time.Now())
	return nil
}

// HealthCheck 健康检查
func (c *BaseComponent) HealthCheck() error {
	status := c.GetStatus()
	if status == ComponentStatusFailed {
		return fmt.Errorf("组件 %s 处于失败状态", c.name)
	}
	if status != ComponentStatusStarted {
		return fmt.Errorf("组件 %s 未启动，当前状态: %s", c.name, status.String())
	}
	return nil
}

// -------------------- 组件配置器 --------------------

// LoggerConfigurer 日志配置器
type LoggerConfigurer struct{}

// Configure 配置日志组件
func (c *LoggerConfigurer) Configure(registry *ComponentRegistry, props PropertySource) error {
	// 检查是否启用日志
	enabled := props.GetBool("logger.enabled", true)
	if !enabled {
		return nil
	}

	// 创建日志组件工厂
	return registry.RegisterFactory("logger", &LoggerComponentFactory{})
}

// Order 配置顺序
func (c *LoggerConfigurer) Order() int {
	return 100 // 日志应该最先配置
}

// GetName 获取配置器名称
func (c *LoggerConfigurer) GetName() string {
	return "LoggerConfigurer"
}

// ConfigConfigurer 配置配置器
type ConfigConfigurer struct{}

// Configure 配置配置组件
func (c *ConfigConfigurer) Configure(registry *ComponentRegistry, props PropertySource) error {
	// 配置组件总是启用
	return registry.RegisterFactory("config", &ConfigComponentFactory{})
}

// Order 配置顺序
func (c *ConfigConfigurer) Order() int {
	return 50 // 配置比日志更优先
}

// GetName 获取配置器名称
func (c *ConfigConfigurer) GetName() string {
	return "ConfigConfigurer"
}

// DBStoreConfigurer 数据库配置器
type DBStoreConfigurer struct{}

// Configure 配置数据库组件
func (c *DBStoreConfigurer) Configure(registry *ComponentRegistry, props PropertySource) error {
	// 检查是否启用数据库
	enabled := props.GetBool("database.enabled", false)
	if !enabled {
		return nil
	}

	// 创建数据库组件工厂
	return registry.RegisterFactory("dbstore", &DBStoreComponentFactory{})
}

// Order 配置顺序
func (c *DBStoreConfigurer) Order() int {
	return 200
}

// GetName 获取配置器名称
func (c *DBStoreConfigurer) GetName() string {
	return "DBStoreConfigurer"
}

// CacheConfigurer 缓存配置器
type CacheConfigurer struct{}

// Configure 配置缓存组件
func (c *CacheConfigurer) Configure(registry *ComponentRegistry, props PropertySource) error {
	// 检查是否启用缓存
	enabled := props.GetBool("cache.enabled", true)
	if !enabled {
		return nil
	}

	// 创建缓存组件工厂
	return registry.RegisterFactory("cache", &CacheComponentFactory{})
}

// Order 配置顺序
func (c *CacheConfigurer) Order() int {
	return 300
}

// GetName 获取配置器名称
func (c *CacheConfigurer) GetName() string {
	return "CacheConfigurer"
}

// WebConfigurer Web配置器
type WebConfigurer struct{}

// Configure 配置Web组件
func (c *WebConfigurer) Configure(registry *ComponentRegistry, props PropertySource) error {
	// 检查是否启用Web
	enabled := props.GetBool("web.enabled", false)
	if !enabled {
		return nil
	}

	// 创建Web组件工厂
	return registry.RegisterFactory("web", &WebComponentFactory{})
}

// Order 配置顺序
func (c *WebConfigurer) Order() int {
	return 400
}

// GetName 获取配置器名称
func (c *WebConfigurer) GetName() string {
	return "WebConfigurer"
}

// -------------------- 组件工厂 --------------------

// LoggerComponentFactory 日志组件工厂
type LoggerComponentFactory struct{}

// Create 创建日志组件
func (f *LoggerComponentFactory) Create(ctx context.Context, props PropertySource) (Component, error) {
	level := props.GetString("logger.level", "info")

	var opts []logger.Option

	// 日志级别
	if logLevel, err := logger.ParseLevel(level); err == nil {
		opts = append(opts, logger.WithLevel(logLevel))
	}

	// 日志文件
	if props.HasProperty("logger.file.path") {
		path := props.GetString("logger.file.path", "")
		if path != "" {
			opts = append(opts, logger.WithFilename(path))
		}
	}

	// JSON格式
	if props.HasProperty("logger.json") {
		json := props.GetBool("logger.json", false)
		if json {
			opts = append(opts, logger.WithJSONConsole(true))
		}
	}

	// 创建日志器
	loggerInstance := logger.New(opts...)

	// 构建组件
	component := &LoggerComponent{
		BaseComponent: NewBaseComponent("logger", ComponentTypeInfrastructure),
		logger:        loggerInstance,
	}

	return component, nil
}

// Dependencies 依赖
func (f *LoggerComponentFactory) Dependencies() []string {
	return []string{}
}

// ValidateConfig 验证配置
func (f *LoggerComponentFactory) ValidateConfig(props PropertySource) error {
	// 验证日志级别
	if props.HasProperty("logger.level") {
		level := props.GetString("logger.level", "")
		if _, err := logger.ParseLevel(level); err != nil {
			return NewConfigError("logger", "无效的日志级别: "+level, err)
		}
	}
	return nil
}

// GetConfigSchema 获取配置模式
func (f *LoggerComponentFactory) GetConfigSchema() ConfigSchema {
	return ConfigSchema{
		RequiredProperties: []string{},
		Properties: map[string]PropertySchema{
			"logger.enabled": {
				Type:         "bool",
				DefaultValue: true,
				Description:  "是否启用日志",
				Required:     false,
			},
			"logger.level": {
				Type:         "string",
				DefaultValue: "info",
				Description:  "日志级别",
				Required:     false,
			},
			"logger.json": {
				Type:         "bool",
				DefaultValue: false,
				Description:  "是否使用JSON格式",
				Required:     false,
			},
			"logger.file.path": {
				Type:         "string",
				DefaultValue: "",
				Description:  "日志文件路径",
				Required:     false,
			},
		},
		Dependencies: []string{},
	}
}

// ConfigComponentFactory 配置组件工厂
type ConfigComponentFactory struct{}

// Create 创建配置组件
func (f *ConfigComponentFactory) Create(ctx context.Context, props PropertySource) (Component, error) {
	// 获取配置提供者
	configProvider, ok := props.(interface {
		PropertySource
		GetConfigProvider() config.Provider
	})

	var configInstance config.Provider
	if !ok {
		// 如果属性源不提供配置提供者，则使用全局配置
		configInstance = config.Config
	} else {
		configInstance = configProvider.GetConfigProvider()
	}

	// 构建组件
	component := &ConfigComponent{
		BaseComponent: NewBaseComponent("config", ComponentTypeInfrastructure),
		config:        configInstance,
	}

	return component, nil
}

// Dependencies 依赖
func (f *ConfigComponentFactory) Dependencies() []string {
	return []string{}
}

// ValidateConfig 验证配置
func (f *ConfigComponentFactory) ValidateConfig(props PropertySource) error {
	// 配置组件不需要特殊验证
	return nil
}

// GetConfigSchema 获取配置模式
func (f *ConfigComponentFactory) GetConfigSchema() ConfigSchema {
	return ConfigSchema{
		RequiredProperties: []string{},
		Properties:         map[string]PropertySchema{},
		Dependencies:       []string{},
	}
}

// DBStoreComponentFactory 数据库组件工厂
type DBStoreComponentFactory struct{}

// Create 创建数据库组件
func (f *DBStoreComponentFactory) Create(ctx context.Context, props PropertySource) (Component, error) {
	driver := props.GetString("database.driver", "sqlite")
	dsn := props.GetString("database.dsn", ":memory:")

	// 构建组件
	component := &DBStoreComponent{
		BaseComponent: NewBaseComponent("dbstore", ComponentTypeDataSource),
		driver:        driver,
		dsn:           dsn,
	}

	return component, nil
}

// Dependencies 依赖
func (f *DBStoreComponentFactory) Dependencies() []string {
	return []string{"logger", "config"}
}

// ValidateConfig 验证配置
func (f *DBStoreComponentFactory) ValidateConfig(props PropertySource) error {
	if !props.GetBool("database.enabled", false) {
		return nil
	}

	driver := props.GetString("database.driver", "")
	if driver == "" {
		return NewConfigError("dbstore", "数据库驱动不能为空", nil)
	}

	validDrivers := []string{"mysql", "postgres", "sqlite"}
	isValid := false
	for _, validDriver := range validDrivers {
		if driver == validDriver {
			isValid = true
			break
		}
	}
	if !isValid {
		return NewConfigError("dbstore", fmt.Sprintf("不支持的数据库驱动: %s", driver), nil)
	}

	return nil
}

// GetConfigSchema 获取配置模式
func (f *DBStoreComponentFactory) GetConfigSchema() ConfigSchema {
	return ConfigSchema{
		RequiredProperties: []string{"database.driver"},
		Properties: map[string]PropertySchema{
			"database.enabled": {
				Type:         "bool",
				DefaultValue: false,
				Description:  "是否启用数据库",
				Required:     false,
			},
			"database.driver": {
				Type:         "string",
				DefaultValue: "sqlite",
				Description:  "数据库驱动",
				Required:     true,
			},
			"database.dsn": {
				Type:         "string",
				DefaultValue: ":memory:",
				Description:  "数据库连接字符串",
				Required:     false,
			},
		},
		Dependencies: []string{"logger", "config"},
	}
}

// CacheComponentFactory 缓存组件工厂
type CacheComponentFactory struct{}

// Create 创建缓存组件
func (f *CacheComponentFactory) Create(ctx context.Context, props PropertySource) (Component, error) {
	cacheType := props.GetString("cache.type", "memory")

	var cacheInstance cache.Cache
	var err error

	switch cacheType {
	case "memory":
		cacheInstance = cache.NewMemoryCache()
	case "redis":
		// 这里需要根据实际的redis配置创建redis缓存
		cacheInstance = cache.NewMemoryCache() // 临时使用内存缓存
	default:
		return nil, NewConfigError("cache", fmt.Sprintf("不支持的缓存类型: %s", cacheType), nil)
	}

	// 构建组件
	component := &CacheComponent{
		BaseComponent: NewBaseComponent("cache", ComponentTypeInfrastructure),
		cache:         cacheInstance,
		cacheType:     cacheType,
	}

	return component, err
}

// Dependencies 依赖
func (f *CacheComponentFactory) Dependencies() []string {
	return []string{"logger", "config"}
}

// ValidateConfig 验证配置
func (f *CacheComponentFactory) ValidateConfig(props PropertySource) error {
	if !props.GetBool("cache.enabled", true) {
		return nil
	}

	cacheType := props.GetString("cache.type", "memory")
	validTypes := []string{"memory", "redis"}
	isValid := false
	for _, validType := range validTypes {
		if cacheType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return NewConfigError("cache", fmt.Sprintf("不支持的缓存类型: %s", cacheType), nil)
	}

	return nil
}

// GetConfigSchema 获取配置模式
func (f *CacheComponentFactory) GetConfigSchema() ConfigSchema {
	return ConfigSchema{
		RequiredProperties: []string{},
		Properties: map[string]PropertySchema{
			"cache.enabled": {
				Type:         "bool",
				DefaultValue: true,
				Description:  "是否启用缓存",
				Required:     false,
			},
			"cache.type": {
				Type:         "string",
				DefaultValue: "memory",
				Description:  "缓存类型",
				Required:     false,
			},
		},
		Dependencies: []string{"logger", "config"},
	}
}

// WebComponentFactory Web组件工厂
type WebComponentFactory struct{}

// Create 创建Web组件
func (f *WebComponentFactory) Create(ctx context.Context, props PropertySource) (Component, error) {
	host := props.GetString("web.host", "0.0.0.0")
	port := props.GetInt("web.port", 8080)

	// 构建组件
	component := &WebComponent{
		BaseComponent: NewBaseComponent("web", ComponentTypeWeb),
		host:          host,
		port:          port,
	}

	return component, nil
}

// Dependencies 依赖
func (f *WebComponentFactory) Dependencies() []string {
	return []string{"logger", "config"}
}

// ValidateConfig 验证配置
func (f *WebComponentFactory) ValidateConfig(props PropertySource) error {
	if !props.GetBool("web.enabled", false) {
		return nil
	}

	port := props.GetInt("web.port", 8080)
	if port <= 0 || port > 65535 {
		return NewConfigError("web", fmt.Sprintf("无效的端口号: %d", port), nil)
	}

	return nil
}

// GetConfigSchema 获取配置模式
func (f *WebComponentFactory) GetConfigSchema() ConfigSchema {
	return ConfigSchema{
		RequiredProperties: []string{},
		Properties: map[string]PropertySchema{
			"web.enabled": {
				Type:         "bool",
				DefaultValue: false,
				Description:  "是否启用Web服务",
				Required:     false,
			},
			"web.host": {
				Type:         "string",
				DefaultValue: "0.0.0.0",
				Description:  "Web服务主机地址",
				Required:     false,
			},
			"web.port": {
				Type:         "int",
				DefaultValue: 8080,
				Description:  "Web服务端口",
				Required:     false,
			},
		},
		Dependencies: []string{"logger", "config"},
	}
}

// -------------------- 具体组件实现 --------------------

// LoggerComponent 日志组件
type LoggerComponent struct {
	*BaseComponent
	logger logger.Logger
}

// Initialize 初始化组件
func (c *LoggerComponent) Initialize(ctx context.Context) error {
	if err := c.BaseComponent.Initialize(ctx); err != nil {
		return err
	}
	c.SetMetric("logger_type", "zap")
	return nil
}

// Start 启动组件
func (c *LoggerComponent) Start(ctx context.Context) error {
	if err := c.BaseComponent.Start(ctx); err != nil {
		return err
	}
	c.logger.Info("日志组件已启动")
	return nil
}

// Stop 停止组件
func (c *LoggerComponent) Stop(ctx context.Context) error {
	c.logger.Info("日志组件正在停止")
	return c.BaseComponent.Stop(ctx)
}

// HealthCheck 健康检查
func (c *LoggerComponent) HealthCheck() error {
	if err := c.BaseComponent.HealthCheck(); err != nil {
		return err
	}
	// 测试日志器是否正常工作
	c.logger.Debug("健康检查")
	return nil
}

// GetLogger 获取日志器
func (c *LoggerComponent) GetLogger() logger.Logger {
	return c.logger
}

// ConfigComponent 配置组件
type ConfigComponent struct {
	*BaseComponent
	config config.Provider
}

// Initialize 初始化组件
func (c *ConfigComponent) Initialize(ctx context.Context) error {
	if err := c.BaseComponent.Initialize(ctx); err != nil {
		return err
	}
	c.SetMetric("config_type", "viper")
	return nil
}

// Start 启动组件
func (c *ConfigComponent) Start(ctx context.Context) error {
	return c.BaseComponent.Start(ctx)
}

// Stop 停止组件
func (c *ConfigComponent) Stop(ctx context.Context) error {
	return c.BaseComponent.Stop(ctx)
}

// GetConfig 获取配置
func (c *ConfigComponent) GetConfig() config.Provider {
	return c.config
}

// DBStoreComponent 数据库组件
type DBStoreComponent struct {
	*BaseComponent
	driver string
	dsn    string
	logger logger.Logger
	config config.Provider
}

// Initialize 初始化组件
func (c *DBStoreComponent) Initialize(ctx context.Context) error {
	if err := c.BaseComponent.Initialize(ctx); err != nil {
		return err
	}
	c.SetMetric("driver", c.driver)
	c.SetMetric("dsn_masked", "***") // 不暴露敏感信息
	return nil
}

// Start 启动组件
func (c *DBStoreComponent) Start(ctx context.Context) error {
	if err := c.BaseComponent.Start(ctx); err != nil {
		return err
	}
	if c.logger != nil {
		c.logger.Info("数据库组件已启动", logger.String("driver", c.driver))
	}
	return nil
}

// Stop 停止组件
func (c *DBStoreComponent) Stop(ctx context.Context) error {
	if c.logger != nil {
		c.logger.Info("数据库组件正在停止")
	}
	return c.BaseComponent.Stop(ctx)
}

// HealthCheck 健康检查
func (c *DBStoreComponent) HealthCheck() error {
	if err := c.BaseComponent.HealthCheck(); err != nil {
		return err
	}
	// 这里可以添加数据库连接检查
	return nil
}

// SetLogger 设置日志器
func (c *DBStoreComponent) SetLogger(logger logger.Logger) {
	c.logger = logger
}

// SetConfig 设置配置
func (c *DBStoreComponent) SetConfig(config config.Provider) {
	c.config = config
}

// CacheComponent 缓存组件
type CacheComponent struct {
	*BaseComponent
	cache     cache.Cache
	cacheType string
	logger    logger.Logger
	config    config.Provider
}

// Initialize 初始化组件
func (c *CacheComponent) Initialize(ctx context.Context) error {
	if err := c.BaseComponent.Initialize(ctx); err != nil {
		return err
	}
	c.SetMetric("cache_type", c.cacheType)
	return nil
}

// Start 启动组件
func (c *CacheComponent) Start(ctx context.Context) error {
	if err := c.BaseComponent.Start(ctx); err != nil {
		return err
	}
	if c.logger != nil {
		c.logger.Info("缓存组件已启动", logger.String("type", c.cacheType))
	}
	return nil
}

// Stop 停止组件
func (c *CacheComponent) Stop(ctx context.Context) error {
	if c.logger != nil {
		c.logger.Info("缓存组件正在停止")
	}
	return c.BaseComponent.Stop(ctx)
}

// HealthCheck 健康检查
func (c *CacheComponent) HealthCheck() error {
	if err := c.BaseComponent.HealthCheck(); err != nil {
		return err
	}
	// 测试缓存是否正常工作
	testKey := "health_check"
	ctx := context.Background()
	if err := c.cache.Set(ctx, testKey, "ok", time.Minute); err != nil {
		return fmt.Errorf("缓存写入测试失败: %v", err)
	}
	if _, found := c.cache.Get(ctx, testKey); !found {
		return fmt.Errorf("缓存读取测试失败: 键未找到")
	}
	return nil
}

// SetLogger 设置日志器
func (c *CacheComponent) SetLogger(logger logger.Logger) {
	c.logger = logger
}

// SetConfig 设置配置
func (c *CacheComponent) SetConfig(config config.Provider) {
	c.config = config
}

// GetCache 获取缓存
func (c *CacheComponent) GetCache() cache.Cache {
	return c.cache
}

// WebComponent Web组件
type WebComponent struct {
	*BaseComponent
	host   string
	port   int
	logger logger.Logger
	config config.Provider
}

// Initialize 初始化组件
func (c *WebComponent) Initialize(ctx context.Context) error {
	if err := c.BaseComponent.Initialize(ctx); err != nil {
		return err
	}
	c.SetMetric("host", c.host)
	c.SetMetric("port", c.port)
	c.SetMetric("address", fmt.Sprintf("%s:%d", c.host, c.port))
	return nil
}

// Start 启动组件
func (c *WebComponent) Start(ctx context.Context) error {
	if err := c.BaseComponent.Start(ctx); err != nil {
		return err
	}
	if c.logger != nil {
		c.logger.Info("Web组件已启动",
			logger.String("host", c.host),
			logger.Int("port", c.port))
	}
	return nil
}

// Stop 停止组件
func (c *WebComponent) Stop(ctx context.Context) error {
	if c.logger != nil {
		c.logger.Info("Web组件正在停止")
	}
	return c.BaseComponent.Stop(ctx)
}

// HealthCheck 健康检查
func (c *WebComponent) HealthCheck() error {
	if err := c.BaseComponent.HealthCheck(); err != nil {
		return err
	}
	// 这里可以添加Web服务器健康检查
	return nil
}

// SetLogger 设置日志器
func (c *WebComponent) SetLogger(logger logger.Logger) {
	c.logger = logger
}

// SetConfig 设置配置
func (c *WebComponent) SetConfig(config config.Provider) {
	c.config = config
}
