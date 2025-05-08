package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ViperProvider 基于Viper的配置提供者实现
type ViperProvider struct {
	v               *viper.Viper
	opts            Options
	configFile      string
	changeCallbacks []func()
	mu              sync.RWMutex
}

// NewViperProvider 创建一个基于Viper的配置提供者
func NewViperProvider(opts ...Option) Provider {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	v := viper.New()

	// 设置配置名称、类型和搜索路径
	v.SetConfigName(options.ConfigName)
	v.SetConfigType(options.ConfigType)
	for _, path := range options.ConfigPaths {
		v.AddConfigPath(path)
	}

	// 设置环境变量相关配置
	if options.AutomaticEnv {
		v.AutomaticEnv()
	}
	if options.EnvPrefix != "" {
		v.SetEnvPrefix(options.EnvPrefix)
	}
	if options.EnvKeyReplacer != nil {
		v.SetEnvKeyReplacer(options.EnvKeyReplacer)
	}

	// 设置默认值
	for key, value := range options.DefaultValues {
		v.SetDefault(key, value)
	}

	provider := &ViperProvider{
		v:               v,
		opts:            options,
		changeCallbacks: make([]func(), 0),
	}

	return provider
}

// LoadConfig 加载配置文件
func (p *ViperProvider) LoadConfig() error {
	// 尝试查找配置文件
	configFile := FindConfigFile(p.opts)
	if configFile != "" {
		p.configFile = configFile
		p.v.SetConfigFile(configFile)
	}

	// 尝试读取配置文件
	err := p.v.ReadInConfig()
	if err != nil {
		// 配置文件不存在时不报错，使用默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 验证配置
	if err := p.ValidateConfig(); err != nil {
		return err
	}

	// 设置配置监听
	if p.opts.WatchConfigFile {
		p.WatchConfig()
	}

	return nil
}

// BindFlags 将配置绑定到命令行标志
func (p *ViperProvider) BindFlags(flags *pflag.FlagSet) {
	if flags == nil {
		return
	}

	flags.VisitAll(func(flag *pflag.Flag) {
		// 将标志名称转换为配置键格式（'-'替换为'.'）
		configKey := strings.ReplaceAll(flag.Name, "-", ".")

		// 如果配置中已经存在该值，则不绑定
		if !p.IsSet(configKey) {
			// 将标志绑定到配置
			if err := p.v.BindPFlag(configKey, flag); err != nil {
				// 只记录绑定错误，不中断程序
				fmt.Printf("Error binding flag '%s' to config: %v\n", flag.Name, err)
			}
		}
	})
}

// Get 获取指定键的配置值
func (p *ViperProvider) Get(key string) interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.Get(key)
}

// GetString 获取字符串类型的配置
func (p *ViperProvider) GetString(key string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetString(key)
}

// GetBool 获取布尔类型的配置
func (p *ViperProvider) GetBool(key string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetBool(key)
}

// GetInt 获取整型的配置
func (p *ViperProvider) GetInt(key string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetInt(key)
}

// GetInt64 获取int64类型的配置
func (p *ViperProvider) GetInt64(key string) int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetInt64(key)
}

// GetFloat64 获取浮点类型的配置
func (p *ViperProvider) GetFloat64(key string) float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetFloat64(key)
}

// GetTime 获取时间类型的配置
func (p *ViperProvider) GetTime(key string) time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetTime(key)
}

// GetDuration 获取时间段类型的配置
func (p *ViperProvider) GetDuration(key string) time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetDuration(key)
}

// GetStringSlice 获取字符串切片类型的配置
func (p *ViperProvider) GetStringSlice(key string) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetStringSlice(key)
}

// GetStringMap 获取字符串映射类型的配置
func (p *ViperProvider) GetStringMap(key string) map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetStringMap(key)
}

// GetStringMapString 获取字符串-字符串映射类型的配置
func (p *ViperProvider) GetStringMapString(key string) map[string]string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetStringMapString(key)
}

// IsSet 判断配置项是否存在
func (p *ViperProvider) IsSet(key string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.IsSet(key)
}

// Set 设置配置项
func (p *ViperProvider) Set(key string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.v.Set(key, value)
}

// UnmarshalKey 将指定键下的配置值解析到结构体中
func (p *ViperProvider) UnmarshalKey(key string, rawVal interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.UnmarshalKey(key, rawVal)
}

// Unmarshal 将所有配置解析到结构体中
func (p *ViperProvider) Unmarshal(rawVal interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.Unmarshal(rawVal)
}

// WatchConfig 启用配置热重载
func (p *ViperProvider) WatchConfig() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.v.WatchConfig()
	p.v.OnConfigChange(func(e fsnotify.Event) {
		// 重新验证配置
		if err := p.ValidateConfig(); err != nil {
			fmt.Printf("Config validation failed after change: %v\n", err)
			return
		}

		// 调用所有注册的配置变更回调
		p.mu.RLock()
		callbacks := make([]func(), len(p.changeCallbacks))
		copy(callbacks, p.changeCallbacks)
		p.mu.RUnlock()

		for _, callback := range callbacks {
			callback()
		}
	})
}

// OnConfigChange 注册配置变更回调函数
func (p *ViperProvider) OnConfigChange(run func()) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.changeCallbacks = append(p.changeCallbacks, run)
}

// ValidateConfig 验证配置是否合法
func (p *ViperProvider) ValidateConfig() error {
	// 执行所有注册的验证器
	for _, validator := range p.opts.ConfigValidators {
		if err := validator(p); err != nil {
			return err
		}
	}
	return nil
}

// SaveConfig 保存当前配置到文件
func (p *ViperProvider) SaveConfig() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.configFile == "" {
		// 如果没有指定配置文件，使用默认路径
		dir := p.opts.ConfigPaths[0]
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		p.configFile = fmt.Sprintf("%s/%s.%s", dir, p.opts.ConfigName, p.opts.ConfigType)
	}

	return p.v.WriteConfigAs(p.configFile)
}

// MergeConfig 合并另一个配置文件
func (p *ViperProvider) MergeConfig(configFile string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	v := viper.New()
	v.SetConfigFile(configFile)
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config for merge: %w", err)
	}

	if err := p.v.MergeConfigMap(v.AllSettings()); err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}

	return nil
}

// MergeConfigFromReader 从reader合并配置
func (p *ViperProvider) MergeConfigFromReader(in io.Reader) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 读取内容到缓冲区
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(in); err != nil {
		return fmt.Errorf("failed to read config data: %w", err)
	}

	// 创建临时Viper实例
	v := viper.New()
	v.SetConfigType(p.opts.ConfigType)

	if err := v.ReadConfig(bytes.NewReader(buf.Bytes())); err != nil {
		return fmt.Errorf("failed to parse config data: %w", err)
	}

	if err := p.v.MergeConfigMap(v.AllSettings()); err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}

	return nil
}
