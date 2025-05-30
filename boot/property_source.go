package boot

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/guanzhenxing/go-snap/config"
)

// DefaultPropertySource 默认属性源实现
// 提供基本的内存属性存储和访问功能
type DefaultPropertySource struct {
	// properties 存储所有配置属性的内存映射
	properties map[string]interface{}
}

// NewDefaultPropertySource 创建默认属性源实例
// 返回：
//
//	初始化的DefaultPropertySource实例
func NewDefaultPropertySource() *DefaultPropertySource {
	return &DefaultPropertySource{
		properties: make(map[string]interface{}),
	}
}

// GetProperty 获取属性值
// 参数：
//
//	key: 属性键名
//
// 返回：
//
//	属性值和是否存在的布尔值
func (p *DefaultPropertySource) GetProperty(key string) (interface{}, bool) {
	value, exists := p.properties[key]
	return value, exists
}

// GetString 获取字符串属性
// 参数：
//
//	key: 属性键名
//	defaultValue: 属性不存在时返回的默认值
//
// 返回：
//
//	字符串类型的属性值，如果属性不存在或类型不匹配则返回默认值
func (p *DefaultPropertySource) GetString(key string, defaultValue string) string {
	value, exists := p.GetProperty(key)
	if !exists {
		return defaultValue
	}

	switch v := value.(type) {
	case string:
		return v
	default:
		return defaultValue
	}
}

// GetBool 获取布尔属性
// 参数：
//
//	key: 属性键名
//	defaultValue: 属性不存在时返回的默认值
//
// 返回：
//
//	布尔类型的属性值，如果属性不存在或类型不匹配则返回默认值
//
// 注意：
//
//	字符串类型的值会尝试解析为布尔值，例如"true"、"false"
func (p *DefaultPropertySource) GetBool(key string, defaultValue bool) bool {
	value, exists := p.GetProperty(key)
	if !exists {
		return defaultValue
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return defaultValue
		}
		return b
	default:
		return defaultValue
	}
}

// GetInt 获取整型属性
// 参数：
//
//	key: 属性键名
//	defaultValue: 属性不存在时返回的默认值
//
// 返回：
//
//	整型的属性值，如果属性不存在或类型不匹配则返回默认值
//
// 注意：
//
//	字符串和浮点类型的值会尝试转换为整型
func (p *DefaultPropertySource) GetInt(key string, defaultValue int) int {
	value, exists := p.GetProperty(key)
	if !exists {
		return defaultValue
	}

	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return defaultValue
		}
		return i
	default:
		return defaultValue
	}
}

// GetFloat 获取浮点属性
// 参数：
//
//	key: 属性键名
//	defaultValue: 属性不存在时返回的默认值
//
// 返回：
//
//	浮点类型的属性值，如果属性不存在或类型不匹配则返回默认值
//
// 注意：
//
//	字符串类型的值会尝试解析为浮点值
func (p *DefaultPropertySource) GetFloat(key string, defaultValue float64) float64 {
	value, exists := p.GetProperty(key)
	if !exists {
		return defaultValue
	}

	switch v := value.(type) {
	case float64:
		return v
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return defaultValue
		}
		return f
	default:
		return defaultValue
	}
}

// HasProperty 判断属性是否存在
// 参数：
//
//	key: 属性键名
//
// 返回：
//
//	属性是否存在的布尔值
func (p *DefaultPropertySource) HasProperty(key string) bool {
	_, exists := p.properties[key]
	return exists
}

// SetProperty 设置属性值
// 参数：
//
//	key: 属性键名
//	value: 要设置的属性值
func (p *DefaultPropertySource) SetProperty(key string, value interface{}) {
	p.properties[key] = value
}

// FilePropertySource 文件属性源，基于配置文件实现
// 扩展DefaultPropertySource，支持从文件加载配置
type FilePropertySource struct {
	*DefaultPropertySource
	// configProvider 底层配置提供者，负责实际的文件读取和解析
	configProvider config.Provider
}

// NewFilePropertySource 从指定配置路径创建文件属性源
// 参数：
//
//	configPath: 配置文件路径，如果为空则使用当前目录下的configs目录
//
// 返回：
//
//	初始化的FilePropertySource实例和可能的错误
func NewFilePropertySource(configPath string) (*FilePropertySource, error) {
	// 设置默认值
	if configPath == "" {
		cwd, _ := os.Getwd()
		configPath = filepath.Join(cwd, "configs")
	}

	// 初始化配置
	err := config.InitConfig(config.WithConfigPath(configPath))
	if err != nil {
		return nil, &ConfigError{Message: "加载配置文件失败", Cause: err}
	}

	source := &FilePropertySource{
		DefaultPropertySource: NewDefaultPropertySource(),
		configProvider:        config.Config,
	}

	// 从配置文件加载属性
	if err := source.loadProperties(); err != nil {
		return nil, err
	}

	return source, nil
}

// loadProperties 加载属性到内存缓存
// 返回：
//
//	加载过程中遇到的错误，如果加载成功则返回nil
func (p *FilePropertySource) loadProperties() error {
	// 加载一些常见键的配置
	commonKeys := []string{
		"app.name", "app.version", "app.env", "app.debug",
		"logger.enabled", "logger.level", "logger.json", "logger.file.path",
		"database.enabled", "database.driver", "database.dsn",
		"cache.enabled", "cache.type",
		"web.enabled", "web.port", "web.host",
	}

	// 遍历这些键，获取值并设置到属性源中
	for _, key := range commonKeys {
		if p.configProvider.IsSet(key) {
			value := p.configProvider.Get(key)
			p.SetProperty(key, value)
		}
	}

	return nil
}

// GetConfigProvider 获取底层配置提供者
// 返回：
//
//	底层配置提供者实例
func (p *FilePropertySource) GetConfigProvider() config.Provider {
	return p.configProvider
}

// GetProperty 获取属性值，优先从内存缓存获取，如果没有则从配置文件获取
// 参数：
//
//	key: 属性键名
//
// 返回：
//
//	属性值和是否存在的布尔值
func (p *FilePropertySource) GetProperty(key string) (interface{}, bool) {
	// 先从本地缓存获取
	value, exists := p.DefaultPropertySource.GetProperty(key)
	if exists {
		return value, true
	}

	// 再从配置提供者获取
	if p.configProvider.IsSet(key) {
		value = p.configProvider.Get(key)
		p.SetProperty(key, value) // 缓存到本地
		return value, true
	}

	return nil, false
}

// LoadEnvironmentVariables 加载环境变量到属性源
// 参数：
//
//	source: 要加载环境变量的属性源
//
// 注意：
//
//	环境变量会被转换为小写并用点号替换下划线，例如APP_NAME会变为app.name
func LoadEnvironmentVariables(source PropertySource) {
	// 遍历所有环境变量
	for _, env := range os.Environ() {
		// 分割环境变量名和值
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// 转换环境变量名为配置键格式（小写，用点号替换下划线）
		key = strings.ToLower(key)
		key = strings.ReplaceAll(key, "_", ".")

		// 设置到属性源中
		source.SetProperty(key, value)
	}
}
