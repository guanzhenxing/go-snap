package main

import (
	"fmt"

	"github.com/guanzhenxing/go-snap/config"
)

// AppConfig 应用配置结构体
type AppConfig struct {
	App struct {
		Name    string `json:"name" validate:"required"`
		Version string `json:"version" validate:"required"`
		Env     string `json:"env" validate:"oneof=development testing staging production"`
		Debug   bool   `json:"debug"`
	} `json:"app"`

	Server struct {
		Host    string `json:"host" validate:"required"`
		Port    int    `json:"port" validate:"required,min=1,max=65535"`
		Timeout string `json:"timeout" validate:"required"`
	} `json:"server"`

	Database struct {
		Driver       string `json:"driver" validate:"required,oneof=mysql postgres sqlite"`
		DSN          string `json:"dsn" validate:"required"`
		MaxOpenConns int    `json:"max_open_conns" validate:"min=1,max=1000"`
		MaxIdleConns int    `json:"max_idle_conns" validate:"min=1,max=1000"`
	} `json:"database"`
}

// RunValidatorExample 运行配置验证示例
func RunValidatorExample() {
	// 初始化配置
	err := config.InitConfig(
		config.WithConfigName("app"),
		config.WithConfigType("yaml"),
		config.WithConfigPath("./configs"),
		// 添加配置验证器
		config.WithConfigValidator(config.RequiredConfigValidator(
			"app.name",
			"app.env",
			"server.host",
			"server.port",
		)),
		// 添加端口范围验证
		config.WithConfigValidator(config.RangeValidator("server.port", 1, 65535)),
		// 添加环境验证
		config.WithConfigValidator(config.EnvironmentValidator()),
		// 结构体验证
		config.WithConfigValidator(config.CreateStructValidator(&AppConfig{})),
	)

	if err != nil {
		fmt.Printf("配置初始化失败: %v\n", err)
		return
	}

	fmt.Println("=== 配置验证示例 ===")
	fmt.Println("配置验证通过!")

	// 尝试设置无效值并验证
	fmt.Println("\n尝试设置无效的端口号...")
	config.Config.Set("server.port", 70000) // 超出端口范围

	// 手动验证配置
	err = config.Config.ValidateConfig()
	if err != nil {
		fmt.Printf("配置验证失败: %v\n", err)
	} else {
		fmt.Println("配置验证通过")
	}

	// 恢复有效值
	config.Config.Set("server.port", 8080)

	// 使用结构体接收并验证配置
	var appConfig AppConfig
	err = config.Config.Unmarshal(&appConfig)
	if err != nil {
		fmt.Printf("配置解析失败: %v\n", err)
		return
	}

	fmt.Println("\n=== 解析后的配置结构体 ===")
	fmt.Printf("应用名称: %s\n", appConfig.App.Name)
	fmt.Printf("环境: %s\n", appConfig.App.Env)
	fmt.Printf("服务器地址: %s:%d\n", appConfig.Server.Host, appConfig.Server.Port)
	fmt.Printf("数据库驱动: %s\n", appConfig.Database.Driver)

	// 验证解析后的结构体
	err = config.ValidateStruct(appConfig)
	if err != nil {
		fmt.Printf("结构体验证失败: %v\n", err)
	} else {
		fmt.Println("\n结构体验证通过!")
	}
}
