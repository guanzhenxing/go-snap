package main

import (
	"fmt"
	"time"

	"github.com/guanzhenxing/go-snap/config"
)

// RunBasicExample 运行基本配置示例
func RunBasicExample() {
	// 初始化配置系统
	err := config.InitConfig(
		config.WithConfigName("app"),                       // 配置文件名称为app
		config.WithConfigType("yaml"),                      // 配置文件类型为yaml
		config.WithConfigPath("./configs"),                 // 配置文件路径
		config.WithDefaultValue("app.name", "default-app"), // 设置默认值
	)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 打印配置信息
	fmt.Println("=== 基本配置示例 ===")
	fmt.Println("应用名称:", config.Config.GetString("app.name"))
	fmt.Println("环境:", config.Config.GetString("app.env"))
	fmt.Println("调试模式:", config.Config.GetBool("app.debug"))
	fmt.Println("版本:", config.Config.GetString("app.version"))

	// 读取服务器配置
	fmt.Println("\n=== 服务器配置 ===")
	fmt.Println("主机:", config.Config.GetString("server.host"))
	fmt.Println("端口:", config.Config.GetInt("server.port"))
	fmt.Println("超时:", config.Config.GetDuration("server.timeout"))

	// 读取数据库配置
	fmt.Println("\n=== 数据库配置 ===")
	fmt.Println("驱动:", config.Config.GetString("database.driver"))
	fmt.Println("数据源:", config.Config.GetString("database.dsn"))
	fmt.Println("最大连接数:", config.Config.GetInt("database.max_open_conns"))
	fmt.Println("最大空闲连接:", config.Config.GetInt("database.max_idle_conns"))
	fmt.Println("连接生命周期:", config.Config.GetDuration("database.conn_max_lifetime"))

	// 使用结构体解析配置
	type ServerConfig struct {
		Host    string        `json:"host"`
		Port    int           `json:"port"`
		Timeout time.Duration `json:"timeout"`
	}

	var serverConfig ServerConfig
	if err := config.Config.UnmarshalKey("server", &serverConfig); err != nil {
		fmt.Printf("解析server配置错误: %v\n", err)
	} else {
		fmt.Println("\n=== 使用结构体解析的服务器配置 ===")
		fmt.Printf("主机: %s, 端口: %d, 超时: %v\n",
			serverConfig.Host,
			serverConfig.Port,
			serverConfig.Timeout)
	}

	// 检查配置项是否存在
	if config.Config.IsSet("redis.addr") {
		fmt.Println("\n=== Redis配置 ===")
		fmt.Println("地址:", config.Config.GetString("redis.addr"))
		fmt.Println("密码:", config.Config.GetString("redis.password"))
		fmt.Println("数据库:", config.Config.GetInt("redis.db"))
	}

	// 设置新的配置值
	config.Config.Set("custom.key", "自定义值")
	fmt.Println("\n=== 自定义配置 ===")
	fmt.Println("自定义键值:", config.Config.GetString("custom.key"))
}
