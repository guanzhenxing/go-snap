// Package boot 提供类似Spring Boot的自动配置框架
package boot

// PropertySource 属性源接口
type PropertySource interface {
	// GetProperty 获取属性值
	GetProperty(key string) (interface{}, bool)

	// GetString 获取字符串属性
	GetString(key string, defaultValue string) string

	// GetBool 获取布尔属性
	GetBool(key string, defaultValue bool) bool

	// GetInt 获取整型属性
	GetInt(key string, defaultValue int) int

	// GetFloat 获取浮点属性
	GetFloat(key string, defaultValue float64) float64

	// HasProperty 判断属性是否存在
	HasProperty(key string) bool

	// SetProperty 设置属性值
	SetProperty(key string, value interface{})
}
