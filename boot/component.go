package boot

import "context"

// Component 组件接口
type Component interface {
	// Name 返回组件名称
	Name() string

	// Type 返回组件类型
	Type() ComponentType

	// Initialize 初始化组件
	Initialize(ctx context.Context) error

	// Start 启动组件
	Start(ctx context.Context) error

	// Stop 停止组件
	Stop(ctx context.Context) error
}

// ComponentType 组件类型
type ComponentType int

const (
	// ComponentTypeInfrastructure 基础设施组件
	ComponentTypeInfrastructure ComponentType = iota

	// ComponentTypeDataSource 数据源组件
	ComponentTypeDataSource

	// ComponentTypeCore 核心业务组件
	ComponentTypeCore

	// ComponentTypeWeb Web服务组件
	ComponentTypeWeb
)

// ComponentFactory 组件工厂接口
type ComponentFactory interface {
	// Create 创建组件实例
	Create(ctx context.Context, props PropertySource) (Component, error)

	// Dependencies 返回依赖的组件名称
	Dependencies() []string
}

// BeanProvider Bean提供者接口
type BeanProvider interface {
	// GetBean 获取Bean实例
	GetBean(name string, bean interface{}) error
}

// AutoConfigurer 自动配置器接口
type AutoConfigurer interface {
	// Configure 配置组件
	Configure(registry *ComponentRegistry, props PropertySource) error

	// Order 配置顺序，数字越小优先级越高
	Order() int
}

// ComponentActivator 组件激活器接口
type ComponentActivator interface {
	// ShouldActivate 判断组件是否应该激活
	ShouldActivate(props PropertySource) bool

	// ComponentType 组件类型
	ComponentType() string
}

// Condition 条件接口
type Condition interface {
	// Matches 判断条件是否匹配
	Matches(props PropertySource) bool
}

// 基于属性的条件
type PropertyCondition struct {
	Key      string
	Value    interface{}
	Operator string // equals, not-equals, exists, not-exists
}

// Matches 判断条件是否匹配
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

// ConditionalOnProperty 属性条件
func ConditionalOnProperty(key string, value interface{}) *PropertyCondition {
	return &PropertyCondition{
		Key:      key,
		Value:    value,
		Operator: "equals",
	}
}

// ConditionalOnPropertyExists 属性存在条件
func ConditionalOnPropertyExists(key string) *PropertyCondition {
	return &PropertyCondition{
		Key:      key,
		Operator: "exists",
	}
}

// ConditionalOnMissingProperty 属性不存在条件
func ConditionalOnMissingProperty(key string) *PropertyCondition {
	return &PropertyCondition{
		Key:      key,
		Operator: "not-exists",
	}
}
