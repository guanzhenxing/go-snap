package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/guanzhenxing/go-snap/config"
)

// RunWatcherExample 运行配置监听示例
func RunWatcherExample() {
	// 创建临时配置文件
	tempDir, err := os.MkdirTemp("", "config-watch-example")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 创建配置文件
	configFile := filepath.Join(tempDir, "watch-config.yaml")
	initialConfig := []byte(`
app:
  name: "watch-app"
  version: "1.0.0"
server:
  port: 8080
`)
	if err := os.WriteFile(configFile, initialConfig, 0644); err != nil {
		fmt.Printf("写入配置文件失败: %v\n", err)
		return
	}

	// 初始化配置
	err = config.InitConfig(
		config.WithConfigName("watch-config"),
		config.WithConfigType("yaml"),
		config.WithConfigPath(tempDir),
		config.WithWatchConfigFile(true), // 启用配置文件监听
	)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 创建配置监听器
	watcher := config.NewWatcher(config.Config)

	// 监听特定配置项
	changeCount := 0
	watcher.WatchFunc("app.name", func(event config.ConfigChangeEvent) {
		changeCount++
		fmt.Printf("\n检测到配置变更 #%d: %s 从 '%v' 变更为 '%v'\n",
			changeCount, event.Key, event.OldValue, event.NewValue)
	})

	watcher.WatchFunc("server.port", func(event config.ConfigChangeEvent) {
		fmt.Printf("\n服务器端口变更: 从 %v 变更为 %v\n", event.OldValue, event.NewValue)
	})

	// 打印初始配置
	fmt.Println("=== 配置监听示例 ===")
	fmt.Println("应用名称:", config.Config.GetString("app.name"))
	fmt.Println("版本:", config.Config.GetString("app.version"))
	fmt.Println("服务器端口:", config.Config.GetInt("server.port"))

	// 注册全局配置变更回调
	config.Config.OnConfigChange(func() {
		fmt.Println("\n全局配置变更通知: 配置文件已重新加载")
	})

	fmt.Println("\n正在修改配置文件...")

	// 修改配置文件
	updatedConfig := []byte(`
app:
  name: "watch-app-updated"
  version: "1.0.1"
server:
  port: 9090
`)
	if err := os.WriteFile(configFile, updatedConfig, 0644); err != nil {
		fmt.Printf("更新配置文件失败: %v\n", err)
		return
	}

	// 等待一段时间以确保配置变更被检测到
	fmt.Println("等待配置变更被检测...")
	time.Sleep(1 * time.Second)

	// 模拟再次修改配置
	// 注意：因为handleConfigChange是未导出方法，所以我们通过直接修改配置来触发变更
	config.Config.Set("app.name", "watch-app-manual-update")
	config.Config.Set("server.port", 10080)

	// 打印更新后的配置
	fmt.Println("\n=== 更新后的配置 ===")
	fmt.Println("应用名称:", config.Config.GetString("app.name"))
	fmt.Println("版本:", config.Config.GetString("app.version"))
	fmt.Println("服务器端口:", config.Config.GetInt("server.port"))

	// 取消监听
	fmt.Println("\n取消配置监听...")
	watcher.UnwatchAll()

	// 一些清理工作
	fmt.Println("配置监听示例完成")
}
