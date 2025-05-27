package boot

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/guanzhenxing/go-snap/config"
)

// DefaultPropertySource 默认属性源实现
type DefaultPropertySource struct {
	properties map[string]interface{}
}

// NewDefaultPropertySource 创建默认属性源
func NewDefaultPropertySource() *DefaultPropertySource {
	return &DefaultPropertySource{
		properties: make(map[string]interface{}),
	}
}

// GetProperty 获取属性值
func (p *DefaultPropertySource) GetProperty(key string) (interface{}, bool) {
	value, exists := p.properties[key]
	return value, exists
}

// GetString 获取字符串属性
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
func (p *DefaultPropertySource) HasProperty(key string) bool {
	_, exists := p.properties[key]
	return exists
}

// SetProperty 设置属性值
func (p *DefaultPropertySource) SetProperty(key string, value interface{}) {
	p.properties[key] = value
}

// FilePropertySource 文件属性源
type FilePropertySource struct {
	*DefaultPropertySource
	configProvider config.Provider
}

// NewFilePropertySource 从文件创建属性源
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

// loadProperties 加载属性
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

// GetConfigProvider 获取配置提供者，供其他方法使用
func (p *FilePropertySource) GetConfigProvider() config.Provider {
	return p.configProvider
}

// GetProperty 获取属性值，优先从配置文件获取
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

// 加载环境变量
func LoadEnvironmentVariables(source PropertySource) {
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			// 转换环境变量名称为配置键
			// 例如: APP_NAME -> app.name
			configKey := strings.ToLower(strings.Replace(key, "_", ".", -1))

			source.SetProperty(configKey, value)
		}
	}
}
