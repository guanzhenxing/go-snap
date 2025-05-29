package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/guanzhenxing/go-snap/boot"
	"github.com/guanzhenxing/go-snap/examples/admin/internal/components"
)

func main() {
	// 创建启动器
	bootApp := boot.NewBoot()

	// 设置配置路径
	bootApp.SetConfigPath("examples/admin/configs")

	// 添加管理面板组件
	bootApp.AddComponent(components.NewAdminComponent())

	// 初始化应用（不启动）
	app, err := bootApp.Initialize()
	if err != nil {
		log.Fatalf("应用初始化失败: %v", err)
	}

	// 订阅事件
	eventBus := app.GetEventBus()
	eventBus.Subscribe("application.started", func(eventName string, eventData interface{}) {
		log.Println("应用已启动，管理面板可以访问")
	})

	eventBus.Subscribe("application.health_check.failed", func(eventName string, eventData interface{}) {
		if healthResults, ok := eventData.(map[string]error); ok {
			log.Printf("健康检查失败: %v", healthResults)
		}
	})

	// 使用Boot启动应用
	go func() {
		if err := bootApp.Run(); err != nil {
			log.Fatalf("应用运行失败: %v", err)
		}
	}()

	// 等待中断信号优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭应用...")

	// 设置关闭超时并通知应用停止
	// 注意：由于app没有直接提供stop方法，这里通过退出来停止
	time.Sleep(1 * time.Second)
	log.Println("应用已成功关闭")
}
