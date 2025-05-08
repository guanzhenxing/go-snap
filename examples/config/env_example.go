package main

import (
	"fmt"
	"os"

	"github.com/guanzhenxing/go-snap/config"
)

// RunEnvExample 运行环境变量配置示例
func RunEnvExample() {
	// 设置环境变量（在实际应用中，这些应该是系统环境变量）
	os.Setenv("APP_DATABASE_DSN", "env_user:env_password@tcp(env-db.example.com:3306)/env_db")
	os.Setenv("APP_SERVER_PORT", "9090")
	os.Setenv("APP_REDIS_PASSWORD", "env_redis_password")
	os.Setenv("APP_APP_ENV", "production") // 设置为生产环境

	// 初始化配置，启用环境变量支持
	err := config.InitConfig(
		config.WithConfigName("app"),
		config.WithConfigType("yaml"),
		config.WithConfigPath("./configs"),
		config.WithEnvPrefix("APP"),    // 设置环境变量前缀
		config.WithAutomaticEnv(true),  // 启用环境变量自动加载
		config.WithEnvKeyReplacer(nil), // 使用默认替换器（将.替换为_）
	)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 打印环境信息
	fmt.Println("=== 环境变量示例 ===")
	fmt.Println("当前环境:", config.GetCurrentEnvironment())

	// 使用环境函数
	fmt.Println("是开发环境?", config.IsDevelopment())
	fmt.Println("是测试环境?", config.IsTesting())
	fmt.Println("是预发环境?", config.IsStaging())
	fmt.Println("是生产环境?", config.IsProduction())

	// 打印从环境变量覆盖的配置值
	fmt.Println("\n=== 环境变量覆盖的配置 ===")
	fmt.Println("数据源:", config.Config.GetString("database.dsn"))
	fmt.Println("服务器端口:", config.Config.GetInt("server.port"))
	fmt.Println("Redis密码:", config.Config.GetString("redis.password"))

	// 根据环境加载不同的配置
	if config.IsProduction() {
		fmt.Println("\n=== 生产环境特定配置 ===")
		fmt.Println("日志级别:", config.Config.GetString("logging.level"))
		fmt.Println("日志输出:", config.Config.GetString("logging.output"))
		fmt.Println("日志文件:", config.Config.GetString("logging.file_path"))
	} else {
		fmt.Println("\n=== 非生产环境配置 ===")
		fmt.Println("调试模式:", config.Config.GetBool("app.debug"))
	}
}
