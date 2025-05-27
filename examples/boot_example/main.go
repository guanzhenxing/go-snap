package main

import (
	"context"
	"log"

	"github.com/guanzhenxing/go-snap/boot"
)

// 自定义组件示例
type ExampleComponent struct {
	name string
}

func NewExampleComponent() *ExampleComponent {
	return &ExampleComponent{
		name: "example",
	}
}

func (c *ExampleComponent) Name() string {
	return c.name
}

func (c *ExampleComponent) Type() boot.ComponentType {
	return boot.ComponentTypeCore
}

func (c *ExampleComponent) Initialize(ctx context.Context) error {
	log.Println("Initializing example component")
	return nil
}

func (c *ExampleComponent) Start(ctx context.Context) error {
	log.Println("Starting example component")
	return nil
}

func (c *ExampleComponent) Stop(ctx context.Context) error {
	log.Println("Stopping example component")
	return nil
}

func main() {
	// 创建启动器
	bootApp := boot.NewBoot()

	// 添加自定义组件
	bootApp.AddComponent(NewExampleComponent())

	// 设置配置路径
	bootApp.SetConfigPath("configs")

	// 运行应用
	if err := bootApp.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
