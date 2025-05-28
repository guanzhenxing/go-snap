package boot

import (
	"sort"
)

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
