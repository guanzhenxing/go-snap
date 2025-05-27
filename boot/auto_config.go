package boot

import (
	"sort"
)

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

// AutoConfig 自动配置引擎
type AutoConfig struct {
	configurers []AutoConfigurer
	activators  []ComponentActivator
}

// NewAutoConfig 创建自动配置引擎
func NewAutoConfig() *AutoConfig {
	return &AutoConfig{
		configurers: []AutoConfigurer{},
		activators:  []ComponentActivator{},
	}
}

// AddConfigurer 添加配置器
func (a *AutoConfig) AddConfigurer(configurer AutoConfigurer) {
	a.configurers = append(a.configurers, configurer)

	// 按优先级排序
	sort.Slice(a.configurers, func(i, j int) bool {
		return a.configurers[i].Order() < a.configurers[j].Order()
	})
}

// AddActivator 添加激活器
func (a *AutoConfig) AddActivator(activator ComponentActivator) {
	a.activators = append(a.activators, activator)
}

// Configure 执行自动配置
func (a *AutoConfig) Configure(registry *ComponentRegistry, props PropertySource) error {
	// 设置默认属性
	a.setDefaultProperties(props)

	// 执行所有配置器
	for _, configurer := range a.configurers {
		if err := configurer.Configure(registry, props); err != nil {
			return err
		}
	}

	return nil
}

// setDefaultProperties 设置默认属性
func (a *AutoConfig) setDefaultProperties(props PropertySource) {
	// 应用默认配置
	if !props.HasProperty("app.name") {
		props.SetProperty("app.name", "GoBootApp")
	}

	if !props.HasProperty("app.version") {
		props.SetProperty("app.version", "1.0.0")
	}

	if !props.HasProperty("app.env") {
		props.SetProperty("app.env", "development")
	}

	// 日志默认配置
	if !props.HasProperty("logger.level") {
		props.SetProperty("logger.level", "info")
	}

	// 数据库默认配置
	if props.HasProperty("database.enabled") && props.GetBool("database.enabled", false) {
		if !props.HasProperty("database.driver") {
			props.SetProperty("database.driver", "sqlite")
		}

		if !props.HasProperty("database.dsn") && props.GetString("database.driver", "") == "sqlite" {
			props.SetProperty("database.dsn", ":memory:")
		}
	}

	// 缓存默认配置
	if !props.HasProperty("cache.enabled") {
		props.SetProperty("cache.enabled", true)
	}

	if props.GetBool("cache.enabled", false) && !props.HasProperty("cache.type") {
		props.SetProperty("cache.type", "memory")
	}

	// Web服务器默认配置
	if props.HasProperty("web.enabled") && props.GetBool("web.enabled", false) {
		if !props.HasProperty("web.port") {
			props.SetProperty("web.port", 8080)
		}

		if !props.HasProperty("web.host") {
			props.SetProperty("web.host", "0.0.0.0")
		}
	}
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
