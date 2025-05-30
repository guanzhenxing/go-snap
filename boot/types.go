// Package boot 提供类似Spring Boot的自动配置框架，实现依赖注入、自动配置和组件生命周期管理
package boot

// PropertySource 属性源接口，负责管理应用配置属性，类似于Spring的PropertySource
type PropertySource interface {
	// GetProperty 获取属性值
	// 参数：
	//   key: 属性键名
	// 返回：
	//   属性值和是否存在的布尔值
	GetProperty(key string) (interface{}, bool)

	// GetString 获取字符串类型属性
	// 参数：
	//   key: 属性键名
	//   defaultValue: 属性不存在时返回的默认值
	// 返回：
	//   字符串类型的属性值
	GetString(key string, defaultValue string) string

	// GetBool 获取布尔类型属性
	// 参数：
	//   key: 属性键名
	//   defaultValue: 属性不存在时返回的默认值
	// 返回：
	//   布尔类型的属性值
	GetBool(key string, defaultValue bool) bool

	// GetInt 获取整型属性
	// 参数：
	//   key: 属性键名
	//   defaultValue: 属性不存在时返回的默认值
	// 返回：
	//   整型的属性值
	GetInt(key string, defaultValue int) int

	// GetFloat 获取浮点类型属性
	// 参数：
	//   key: 属性键名
	//   defaultValue: 属性不存在时返回的默认值
	// 返回：
	//   浮点类型的属性值
	GetFloat(key string, defaultValue float64) float64

	// HasProperty 判断属性是否存在
	// 参数：
	//   key: 属性键名
	// 返回：
	//   属性是否存在的布尔值
	HasProperty(key string) bool

	// SetProperty 设置属性值
	// 参数：
	//   key: 属性键名
	//   value: 要设置的属性值
	SetProperty(key string, value interface{})
}
