package main

import (
	"context"
	"log"

	"github.com/guanzhenxing/go-snap/appcore"
)

func main() {
	// 创建应用实例
	app := appcore.New("ExampleApp", "1.0.0")

	// 初始化配置组件
	configComponent := appcore.NewConfigComponent("./configs")
	if err := configComponent.Initialize(context.Background()); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	// 初始化日志组件
	loggerComponent := appcore.NewLoggerComponent()
	loggerComponent.SetConfig(configComponent.GetConfig())
	if err := loggerComponent.Initialize(context.Background()); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}

	// 设置应用配置和日志
	app.SetConfig(configComponent.GetConfig())
	app.SetLogger(loggerComponent.GetLogger())

	// 创建组件管理器
	manager := appcore.NewComponentManager()
	manager.RegisterComponent(configComponent)
	manager.RegisterComponent(loggerComponent)
	app.SetComponentManager(manager)

	// 添加启动前钩子
	app.WithHook(appcore.HookBeforeStart, func(ctx context.Context) error {
		log.Println("应用即将启动...")
		return nil
	})

	// 添加启动后钩子
	app.WithHook(appcore.HookAfterStart, func(ctx context.Context) error {
		log.Println("应用已成功启动!")
		return nil
	})

	// 运行应用
	if err := app.Run(); err != nil {
		log.Fatalf("应用运行失败: %v", err)
	}
}
