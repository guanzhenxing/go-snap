package boot

import "context"

// ComponentStatus 组件状态枚举，表示组件在其生命周期中的不同状态
type ComponentStatus int

const (
	// ComponentStatusUnknown 未知状态，表示组件状态未确定
	ComponentStatusUnknown ComponentStatus = iota
	// ComponentStatusCreated 已创建状态，表示组件已实例化但尚未初始化
	ComponentStatusCreated
	// ComponentStatusInitialized 已初始化状态，表示组件已完成初始化但尚未启动
	ComponentStatusInitialized
	// ComponentStatusStarted 已启动状态，表示组件已启动并正在运行
	ComponentStatusStarted
	// ComponentStatusStopped 已停止状态，表示组件已停止运行
	ComponentStatusStopped
	// ComponentStatusFailed 失败状态，表示组件在初始化、启动或运行过程中遇到错误
	ComponentStatusFailed
)

// String 返回组件状态的字符串表示
// 返回：
//
//	表示组件状态的可读字符串
func (s ComponentStatus) String() string {
	switch s {
	case ComponentStatusCreated:
		return "Created"
	case ComponentStatusInitialized:
		return "Initialized"
	case ComponentStatusStarted:
		return "Started"
	case ComponentStatusStopped:
		return "Stopped"
	case ComponentStatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// Component 组件接口，定义了应用组件的基本操作和生命周期方法
type Component interface {
	// Name 返回组件的唯一名称
	// 返回：
	//   组件的唯一标识名称
	Name() string

	// Type 返回组件类型
	// 返回：
	//   组件的类型，用于分类和管理
	Type() ComponentType

	// Initialize 初始化组件，在组件启动前进行必要的准备工作
	// 参数：
	//   ctx: 上下文，可用于传递取消信号或超时
	// 返回：
	//   初始化过程中遇到的错误，如果初始化成功则返回nil
	Initialize(ctx context.Context) error

	// Start 启动组件，使组件开始提供服务
	// 参数：
	//   ctx: 上下文，可用于传递取消信号或超时
	// 返回：
	//   启动过程中遇到的错误，如果启动成功则返回nil
	Start(ctx context.Context) error

	// Stop 停止组件，清理资源并停止提供服务
	// 参数：
	//   ctx: 上下文，可用于传递取消信号或超时
	// 返回：
	//   停止过程中遇到的错误，如果停止成功则返回nil
	Stop(ctx context.Context) error

	// HealthCheck 执行组件健康检查，确认组件是否正常运行
	// 返回：
	//   健康检查结果，如果组件健康则返回nil，否则返回错误信息
	HealthCheck() error

	// GetStatus 获取组件当前状态
	// 返回：
	//   组件的当前状态
	GetStatus() ComponentStatus

	// GetMetrics 获取组件运行指标
	// 返回：
	//   包含组件运行指标的键值对映射
	GetMetrics() map[string]interface{}
}

// ComponentType 组件类型枚举，用于对组件进行分类
type ComponentType int

const (
	// ComponentTypeInfrastructure 基础设施组件，如配置、日志等基础服务
	ComponentTypeInfrastructure ComponentType = iota

	// ComponentTypeDataSource 数据源组件，如数据库、缓存等
	ComponentTypeDataSource

	// ComponentTypeCore 核心业务组件，实现应用的主要业务逻辑
	ComponentTypeCore

	// ComponentTypeWeb Web服务组件，提供HTTP/WebSocket等Web服务
	ComponentTypeWeb
)

// ConfigSchema 配置模式，描述组件所需的配置结构
type ConfigSchema struct {
	// RequiredProperties 必需的属性列表
	RequiredProperties []string `json:"required_properties"`
	// Properties 属性定义映射，键为属性名，值为属性模式
	Properties map[string]PropertySchema `json:"properties"`
	// Dependencies 依赖的其他组件列表
	Dependencies []string `json:"dependencies"`
}

// PropertySchema 属性模式，描述单个配置属性的特征
type PropertySchema struct {
	// Type 属性类型，如string、int、bool等
	Type string `json:"type"`
	// DefaultValue 属性默认值
	DefaultValue interface{} `json:"default_value"`
	// Description 属性描述
	Description string `json:"description"`
	// Required 属性是否必需
	Required bool `json:"required"`
}

// ComponentFactory 组件工厂接口，负责创建和配置组件实例
type ComponentFactory interface {
	// Create 创建组件实例
	// 参数：
	//   ctx: 上下文，可用于传递取消信号或超时
	//   props: 属性源，提供组件配置属性
	// 返回：
	//   创建的组件实例和可能的错误
	Create(ctx context.Context, props PropertySource) (Component, error)

	// Dependencies 返回依赖的组件名称列表
	// 返回：
	//   当前组件依赖的其他组件名称列表
	Dependencies() []string

	// ValidateConfig 验证配置是否满足组件需求
	// 参数：
	//   props: 属性源，提供组件配置属性
	// 返回：
	//   验证过程中遇到的错误，如果验证通过则返回nil
	ValidateConfig(props PropertySource) error

	// GetConfigSchema 获取组件配置模式
	// 返回：
	//   描述组件配置需求的模式
	GetConfigSchema() ConfigSchema
}

// BeanProvider Bean提供者接口，类似于依赖注入容器
type BeanProvider interface {
	// GetBean 获取Bean实例
	// 参数：
	//   name: Bean的名称
	//   bean: 用于存储获取到的Bean的指针
	// 返回：
	//   获取过程中遇到的错误，如果获取成功则返回nil
	GetBean(name string, bean interface{}) error
}

// AutoConfigurer 自动配置器接口，负责自动配置组件
type AutoConfigurer interface {
	// Configure 配置组件
	// 参数：
	//   registry: 组件注册表
	//   props: 属性源，提供配置属性
	// 返回：
	//   配置过程中遇到的错误，如果配置成功则返回nil
	Configure(registry *ComponentRegistry, props PropertySource) error

	// Order 配置顺序，数字越小优先级越高
	// 返回：
	//   配置器的优先级顺序值
	Order() int

	// GetName 获取配置器名称
	// 返回：
	//   配置器的唯一名称
	GetName() string
}

// ComponentActivator 组件激活器接口，决定组件是否应该被激活
type ComponentActivator interface {
	// ShouldActivate 判断组件是否应该激活
	// 参数：
	//   props: 属性源，提供配置属性
	// 返回：
	//   组件是否应该被激活的布尔值
	ShouldActivate(props PropertySource) bool

	// ComponentType 获取组件类型
	// 返回：
	//   组件类型的字符串表示
	ComponentType() string
}

// Condition 条件接口，用于条件化配置
type Condition interface {
	// Matches 判断条件是否匹配
	// 参数：
	//   props: 属性源，提供配置属性
	// 返回：
	//   条件是否匹配的布尔值
	Matches(props PropertySource) bool
}

// PropertyCondition 基于属性的条件实现
type PropertyCondition struct {
	// Key 属性键
	Key string
	// Value 期望的属性值
	Value interface{}
	// Operator 比较操作符，支持equals、not-equals、exists、not-exists
	Operator string
}

// Matches 判断属性条件是否匹配
// 参数：
//
//	props: 属性源，提供配置属性
//
// 返回：
//
//	条件是否匹配的布尔值
func (c *PropertyCondition) Matches(props PropertySource) bool {
	switch c.Operator {
	case "equals":
		val, exists := props.GetProperty(c.Key)
		return exists && val == c.Value
	case "not-equals":
		val, exists := props.GetProperty(c.Key)
		return !exists || val != c.Value
	case "exists":
		return props.HasProperty(c.Key)
	case "not-exists":
		return !props.HasProperty(c.Key)
	default:
		return false
	}
}

// ConditionalOnProperty 创建属性值匹配条件
// 参数：
//
//	key: 属性键
//	value: 期望的属性值
//
// 返回：
//
//	配置为"equals"操作的PropertyCondition实例
func ConditionalOnProperty(key string, value interface{}) *PropertyCondition {
	return &PropertyCondition{
		Key:      key,
		Value:    value,
		Operator: "equals",
	}
}

// ConditionalOnPropertyExists 创建属性存在条件
// 参数：
//
//	key: 属性键
//
// 返回：
//
//	配置为"exists"操作的PropertyCondition实例
func ConditionalOnPropertyExists(key string) *PropertyCondition {
	return &PropertyCondition{
		Key:      key,
		Operator: "exists",
	}
}

// ConditionalOnMissingProperty 创建属性不存在条件
// 参数：
//
//	key: 属性键
//
// 返回：
//
//	配置为"not-exists"操作的PropertyCondition实例
func ConditionalOnMissingProperty(key string) *PropertyCondition {
	return &PropertyCondition{
		Key:      key,
		Operator: "not-exists",
	}
}

// ComponentHealthChecker 组件健康检查器接口
type ComponentHealthChecker interface {
	// CheckHealth 检查组件健康状态
	// 参数：
	//   component: 要检查的组件
	// 返回：
	//   健康检查结果，如果组件健康则返回nil，否则返回错误信息
	CheckHealth(component Component) error
}

// DefaultHealthChecker 默认健康检查器实现
type DefaultHealthChecker struct{}

// CheckHealth 检查组件健康状态，通过调用组件自身的HealthCheck方法
// 参数：
//
//	component: 要检查的组件
//
// 返回：
//
//	健康检查结果，如果组件健康则返回nil，否则返回错误信息
func (h *DefaultHealthChecker) CheckHealth(component Component) error {
	return component.HealthCheck()
}
