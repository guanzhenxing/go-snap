package boot

import (
	"log"
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
	app, err := NewApplication(b.configPath)
	if err != nil {
		return err
	}

	// 添加标准配置器
	app.AddConfigurer(&ConfigConfigurer{})
	app.AddConfigurer(&LoggerConfigurer{})
	// 注意：这里我们忽略了其他配置器的实现细节

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

	// 运行应用
	return app.Run()
}

// Initialize 初始化应用并返回应用实例（不启动）
func (b *Boot) Initialize() (*Application, error) {
	// 创建应用
	app, err := NewApplication(b.configPath)
	if err != nil {
		return nil, err
	}

	// 添加标准配置器
	app.AddConfigurer(&ConfigConfigurer{})
	app.AddConfigurer(&LoggerConfigurer{})
	// 注意：这里我们忽略了其他配置器的实现细节

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

	// 初始化应用
	if err := app.Initialize(); err != nil {
		return nil, err
	}

	return app, nil
}

// DefaultConfigurations 获取默认配置
func DefaultConfigurations() []AutoConfigurer {
	return []AutoConfigurer{
		&ConfigConfigurer{},
		&LoggerConfigurer{},
		// 注意：这里我们忽略了其他配置器的实现细节
	}
}

// WebConfigurations 获取Web应用配置
func WebConfigurations() []AutoConfigurer {
	return append(
		DefaultConfigurations(),
		// 添加Web配置器
		&WebConfigurer{},
	)
}

// StorageConfigurations 获取存储应用配置
func StorageConfigurations() []AutoConfigurer {
	return append(
		DefaultConfigurations(),
		// 添加存储配置器
		&DBStoreConfigurer{},
		&CacheConfigurer{},
	)
}

// FullConfigurations 获取完整应用配置
func FullConfigurations() []AutoConfigurer {
	return append(
		StorageConfigurations(),
		// 添加Web配置器
		&WebConfigurer{},
	)
}
