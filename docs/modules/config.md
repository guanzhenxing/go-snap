# Config 模块

Config 模块是 Go-Snap 框架的配置管理组件，基于 [Viper](https://github.com/spf13/viper) 构建，提供强大而灵活的配置解决方案，支持多种配置格式、多环境配置、配置热重载等功能。

## 概述

Config 模块提供了一个统一的配置管理接口，简化了应用配置的读取、管理和验证。它支持从多种来源加载配置，并提供了配置优先级、环境变量覆盖、配置验证等企业级功能。

### 核心特性

- ✅ **多格式支持** - YAML、JSON、TOML、INI、HCL 等格式
- ✅ **多环境配置** - 轻松切换开发、测试、生产环境配置
- ✅ **配置优先级** - 命令行 > 环境变量 > 配置文件 > 默认值
- ✅ **热重载** - 支持配置文件变更监听和动态重载
- ✅ **环境变量** - 自动映射环境变量到配置项
- ✅ **配置验证** - 内置配置参数校验机制
- ✅ **远程配置** - 支持从远程配置中心加载配置
- ✅ **配置合并** - 支持多个配置文件的合并
- ✅ **默认值** - 提供合理的默认配置值

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/guanzhenxing/go-snap/boot"
)

func main() {
    // 启动应用（会自动加载配置）
    app := boot.NewBoot().SetConfigPath("configs")
    application, _ := app.Initialize()
    
    // 获取配置组件
    if configComp, found := application.GetComponent("config"); found {
        if cc, ok := configComp.(*boot.ConfigComponent); ok {
            config := cc.GetConfig()
            
            // 读取配置
            appName := config.GetString("app.name")
            port := config.GetInt("server.port")
            debug := config.GetBool("app.debug")
            
            fmt.Printf("应用: %s, 端口: %d, 调试: %t\n", appName, port, debug)
        }
    }
}
```

### 直接使用Config

```go
import "github.com/guanzhenxing/go-snap/config"

// 初始化配置
err := config.InitConfig(
    config.WithConfigPath("configs"),
    config.WithConfigName("application"),
    config.WithConfigType("yaml"),
)
if err != nil {
    panic(err)
}

// 读取配置
appName := config.GetString("app.name")
port := config.GetInt("server.port", 8080)
```

## 配置文件

### 目录结构

```
configs/
├── application.yaml          # 主配置文件
├── application-dev.yaml      # 开发环境配置
├── application-test.yaml     # 测试环境配置
├── application-prod.yaml     # 生产环境配置
└── database.yaml            # 数据库配置
```

### 主配置文件示例

```yaml
# configs/application.yaml

# 应用配置
app:
  name: "go-snap-app"
  version: "1.0.0"
  env: "development"
  debug: true
  
# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  
# 日志配置
logger:
  level: "info"
  format: "text"
  output: "stdout"
  
# 数据库配置
database:
  driver: "mysql"
  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
  max_idle_conns: 10
  max_open_conns: 100
  
# 缓存配置
cache:
  type: "redis"
  addr: "localhost:6379"
  password: ""
  db: 0
```

### 环境特定配置

```yaml
# configs/application-dev.yaml
app:
  debug: true
  
logger:
  level: "debug"
  
database:
  dsn: "user:password@tcp(localhost:3306)/dbname_dev?charset=utf8mb4&parseTime=True&loc=Local"

---
# configs/application-prod.yaml
app:
  debug: false
  
logger:
  level: "warn"
  format: "json"
  
database:
  dsn: "${DB_DSN}"
  max_idle_conns: 20
  max_open_conns: 200
```

## API 参考

### 初始化配置

```go
import "github.com/guanzhenxing/go-snap/config"

// 基础初始化
err := config.InitConfig()

// 自定义初始化
err := config.InitConfig(
    config.WithConfigPath("./configs"),
    config.WithConfigName("application"),
    config.WithConfigType("yaml"),
    config.WithEnvPrefix("MYAPP"),
)
```

### 配置选项

```go
// 可用的初始化选项
config.WithConfigPath(path)        // 设置配置文件路径
config.WithConfigName(name)        // 设置配置文件名（不含扩展名）
config.WithConfigType(typ)         // 设置配置文件类型
config.WithEnvPrefix(prefix)       // 设置环境变量前缀
config.WithAutomaticEnv()          // 启用自动环境变量映射
```

### 读取配置

```go
// 基础类型
value := config.Get("key")                    // 获取原始值
str := config.GetString("app.name")          // 字符串
num := config.GetInt("server.port")          // 整数
f := config.GetFloat64("ratio")              // 浮点数
b := config.GetBool("app.debug")             // 布尔值
dur := config.GetDuration("timeout")         // 时间间隔
t := config.GetTime("created_at")            // 时间

// 带默认值
port := config.GetInt("server.port", 8080)
timeout := config.GetDuration("timeout", time.Second*30)

// 数组和切片
hosts := config.GetStringSlice("database.hosts")
ports := config.GetIntSlice("server.ports")

// 映射
settings := config.GetStringMap("redis")
dbConfig := config.GetStringMapString("database")

// 子配置
dbViper := config.Sub("database")
if dbViper != nil {
    driver := dbViper.GetString("driver")
    dsn := dbViper.GetString("dsn")
}
```

### 设置配置

```go
// 设置配置值
config.Set("app.name", "new-app-name")
config.Set("server.port", 9090)

// 设置默认值
config.SetDefault("server.port", 8080)
config.SetDefault("app.debug", false)
```

### 检查配置

```go
// 检查配置是否存在
if config.IsSet("database.dsn") {
    dsn := config.GetString("database.dsn")
}

// 获取所有配置键
keys := config.AllKeys()
for _, key := range keys {
    fmt.Printf("%s = %v\n", key, config.Get(key))
}
```

### 配置绑定

```go
// 绑定到结构体
type AppConfig struct {
    Name    string `mapstructure:"name"`
    Version string `mapstructure:"version"`
    Debug   bool   `mapstructure:"debug"`
}

var appConfig AppConfig
err := config.UnmarshalKey("app", &appConfig)

// 绑定整个配置
type Config struct {
    App      AppConfig      `mapstructure:"app"`
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
}

var cfg Config
err := config.Unmarshal(&cfg)
```

## 环境变量

### 自动映射

```go
// 启用自动环境变量映射
config.InitConfig(config.WithAutomaticEnv())

// 环境变量会自动映射到配置
// 例如：APP_NAME -> app.name
// 例如：DATABASE_DSN -> database.dsn
```

### 自定义映射

```go
// 绑定特定环境变量
config.BindEnv("database.dsn", "DB_CONNECTION_STRING")

// 设置环境变量前缀
config.InitConfig(config.WithEnvPrefix("MYAPP"))
// MYAPP_DATABASE_DSN -> database.dsn
```

### 环境变量示例

```bash
# 设置环境变量
export APP_NAME="production-app"
export APP_DEBUG=false
export SERVER_PORT=8080
export DATABASE_DSN="user:pass@tcp(prod-db:3306)/mydb"
export LOGGER_LEVEL=warn
```

## 多环境配置

### 配置文件命名规则

```
application.yaml           # 基础配置
application-{env}.yaml     # 环境特定配置
```

### 环境切换

```go
// 通过环境变量指定环境
os.Setenv("GO_ENV", "production")

// 或在配置中指定
config.Set("app.env", "production")

// 框架会自动加载对应的配置文件
// application.yaml + application-production.yaml
```

### 配置合并规则

1. 加载基础配置文件 `application.yaml`
2. 根据环境加载特定配置 `application-{env}.yaml`
3. 环境特定配置会覆盖基础配置中的相同键
4. 环境变量会覆盖配置文件中的值

## 高级功能

### 配置热重载

```go
// 启用配置文件监听
config.WatchConfig()

// 设置配置变更回调
config.OnConfigChange(func(e fsnotify.Event) {
    fmt.Println("配置文件已更改:", e.Name)
    // 重新加载配置或执行其他操作
})
```

### 远程配置

```go
// 从远程配置中心加载（需要额外实现）
err := config.AddRemoteProvider("etcd", "http://127.0.0.1:4001", "/config/myapp.json")
err = config.ReadRemoteConfig()
```

### 配置验证

```go
// 验证必需配置
func ValidateConfig() error {
    required := []string{
        "app.name",
        "server.port",
        "database.dsn",
    }
    
    for _, key := range required {
        if !config.IsSet(key) {
            return fmt.Errorf("required config missing: %s", key)
        }
    }
    return nil
}

// 验证配置值范围
func ValidateValues() error {
    port := config.GetInt("server.port")
    if port <= 0 || port > 65535 {
        return fmt.Errorf("invalid port: %d", port)
    }
    
    logLevel := config.GetString("logger.level")
    validLevels := []string{"debug", "info", "warn", "error"}
    if !contains(validLevels, logLevel) {
        return fmt.Errorf("invalid log level: %s", logLevel)
    }
    
    return nil
}
```

### 配置加密

```go
// 敏感配置加密存储（需要自定义实现）
func DecryptConfig(encryptedValue string) string {
    // 实现解密逻辑
    return decryptedValue
}

// 在配置中使用
dsn := DecryptConfig(config.GetString("database.encrypted_dsn"))
```

## 配置模式

### 结构化配置

```go
// 定义配置结构
type Config struct {
    App struct {
        Name    string `mapstructure:"name" validate:"required"`
        Version string `mapstructure:"version" validate:"required"`
        Debug   bool   `mapstructure:"debug"`
    } `mapstructure:"app"`
    
    Server struct {
        Host         string        `mapstructure:"host" validate:"required"`
        Port         int           `mapstructure:"port" validate:"min=1,max=65535"`
        ReadTimeout  time.Duration `mapstructure:"read_timeout"`
        WriteTimeout time.Duration `mapstructure:"write_timeout"`
    } `mapstructure:"server"`
    
    Database struct {
        Driver       string `mapstructure:"driver" validate:"required"`
        DSN          string `mapstructure:"dsn" validate:"required"`
        MaxIdleConns int    `mapstructure:"max_idle_conns" validate:"min=1"`
        MaxOpenConns int    `mapstructure:"max_open_conns" validate:"min=1"`
    } `mapstructure:"database"`
}

// 加载和验证配置
func LoadConfig() (*Config, error) {
    var cfg Config
    
    // 绑定配置
    if err := config.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    // 验证配置
    if err := validate.Struct(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

### 配置工厂模式

```go
// 配置工厂
type ConfigFactory struct {
    config config.Provider
}

func NewConfigFactory(cfg config.Provider) *ConfigFactory {
    return &ConfigFactory{config: cfg}
}

func (f *ConfigFactory) GetDatabaseConfig() DatabaseConfig {
    return DatabaseConfig{
        Driver:       f.config.GetString("database.driver"),
        DSN:          f.config.GetString("database.dsn"),
        MaxIdleConns: f.config.GetInt("database.max_idle_conns", 10),
        MaxOpenConns: f.config.GetInt("database.max_open_conns", 100),
    }
}

func (f *ConfigFactory) GetServerConfig() ServerConfig {
    return ServerConfig{
        Host:         f.config.GetString("server.host", "0.0.0.0"),
        Port:         f.config.GetInt("server.port", 8080),
        ReadTimeout:  f.config.GetDuration("server.read_timeout", 30*time.Second),
        WriteTimeout: f.config.GetDuration("server.write_timeout", 30*time.Second),
    }
}
```

## 最佳实践

### 1. 配置组织

```yaml
# ✅ 好的做法：按功能分组
app:
  name: "myapp"
  version: "1.0.0"

server:
  port: 8080
  timeout: 30s

database:
  driver: "mysql"
  dsn: "..."

# ❌ 避免的做法：平铺所有配置
app_name: "myapp"
app_version: "1.0.0"
server_port: 8080
server_timeout: 30s
database_driver: "mysql"
database_dsn: "..."
```

### 2. 敏感信息处理

```yaml
# ✅ 使用环境变量
database:
  dsn: "${DB_DSN}"
  password: "${DB_PASSWORD}"

# ✅ 或使用配置引用
database:
  dsn: "{{.database_dsn}}"

# ❌ 避免硬编码敏感信息
database:
  dsn: "user:password@tcp(host:3306)/db"
```

### 3. 默认值设置

```go
// ✅ 在代码中设置默认值
config.SetDefault("server.port", 8080)
config.SetDefault("database.max_idle_conns", 10)
config.SetDefault("cache.ttl", time.Hour)

// ✅ 使用GetXxx方法的默认值参数
port := config.GetInt("server.port", 8080)
timeout := config.GetDuration("server.timeout", 30*time.Second)
```

### 4. 配置验证

```go
// ✅ 启动时验证关键配置
func ValidateCriticalConfig() error {
    // 验证必需配置
    if !config.IsSet("database.dsn") {
        return errors.New("database.dsn is required")
    }
    
    // 验证端口范围
    port := config.GetInt("server.port")
    if port <= 0 || port > 65535 {
        return fmt.Errorf("invalid port: %d", port)
    }
    
    return nil
}
```

### 5. 环境特定配置

```go
// ✅ 根据环境动态配置
env := config.GetString("app.env", "development")

switch env {
case "development":
    config.SetDefault("logger.level", "debug")
    config.SetDefault("app.debug", true)
case "production":
    config.SetDefault("logger.level", "warn")
    config.SetDefault("app.debug", false)
}
```

## 集成示例

### 与数据库集成

```go
func NewDatabaseConnection() (*gorm.DB, error) {
    // 从配置读取数据库参数
    driver := config.GetString("database.driver")
    dsn := config.GetString("database.dsn")
    maxIdle := config.GetInt("database.max_idle_conns", 10)
    maxOpen := config.GetInt("database.max_open_conns", 100)
    
    // 创建数据库连接
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    
    // 配置连接池
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    sqlDB.SetMaxIdleConns(maxIdle)
    sqlDB.SetMaxOpenConns(maxOpen)
    
    return db, nil
}
```

### 与Web服务集成

```go
func NewGinEngine() *gin.Engine {
    // 根据环境设置模式
    if config.GetBool("app.debug", false) {
        gin.SetMode(gin.DebugMode)
    } else {
        gin.SetMode(gin.ReleaseMode)
    }
    
    engine := gin.New()
    
    // 从配置读取中间件设置
    if config.GetBool("middleware.cors.enabled", false) {
        engine.Use(cors.Default())
    }
    
    if config.GetBool("middleware.logging.enabled", true) {
        engine.Use(gin.Logger())
    }
    
    return engine
}
```

### 与缓存集成

```go
func NewRedisClient() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     config.GetString("redis.addr", "localhost:6379"),
        Password: config.GetString("redis.password", ""),
        DB:       config.GetInt("redis.db", 0),
        PoolSize: config.GetInt("redis.pool_size", 10),
    })
}
```

## 性能优化

### 1. 配置缓存

```go
// 缓存频繁访问的配置
var configCache = make(map[string]interface{})
var cacheMutex sync.RWMutex

func GetCachedConfig(key string) interface{} {
    cacheMutex.RLock()
    if value, exists := configCache[key]; exists {
        cacheMutex.RUnlock()
        return value
    }
    cacheMutex.RUnlock()
    
    // 从配置中读取并缓存
    value := config.Get(key)
    cacheMutex.Lock()
    configCache[key] = value
    cacheMutex.Unlock()
    
    return value
}
```

### 2. 延迟加载

```go
// 延迟加载复杂配置
type LazyConfig struct {
    once   sync.Once
    config *ComplexConfig
}

func (lc *LazyConfig) Get() *ComplexConfig {
    lc.once.Do(func() {
        // 只在第一次访问时加载
        lc.config = LoadComplexConfig()
    })
    return lc.config
}
```

## 故障排除

### 常见问题

#### 1. 配置文件未找到

**错误**: `Config File "application" Not Found`

**解决方案**:
- 检查配置文件路径是否正确
- 确认配置文件名和扩展名
- 验证当前工作目录

#### 2. 环境变量未生效

**错误**: 环境变量设置后配置值没有变化

**解决方案**:
- 确认环境变量名称格式正确（大写，下划线分隔）
- 检查是否启用了自动环境变量映射
- 验证环境变量前缀设置

#### 3. 配置类型转换失败

**错误**: 配置值类型转换错误

**解决方案**:
- 检查配置文件中的值格式
- 使用正确的Get方法获取对应类型
- 提供合适的默认值

### 调试技巧

```go
// 打印所有配置
for _, key := range config.AllKeys() {
    fmt.Printf("%s = %v\n", key, config.Get(key))
}

// 检查配置来源
fmt.Println("配置文件:", config.ConfigFileUsed())

// 检查特定配置是否设置
if config.IsSet("problematic.key") {
    fmt.Println("配置已设置:", config.Get("problematic.key"))
} else {
    fmt.Println("配置未设置")
}
```

## 参考资料

- [Viper 官方文档](https://github.com/spf13/viper)
- [Go-Snap Boot 模块](boot.md)
- [Go-Snap 架构设计](../architecture.md)
- [配置文件示例](../examples/) 