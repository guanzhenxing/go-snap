package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/guanzhenxing/go-snap/boot"
	"github.com/guanzhenxing/go-snap/logger"
)

// 自定义组件
type CustomComponent struct {
	name   string
	logger logger.Logger
}

// Name 返回组件名称
func (c *CustomComponent) Name() string {
	return c.name
}

// Type 返回组件类型
func (c *CustomComponent) Type() boot.ComponentType {
	return boot.ComponentTypeCore
}

// Initialize 初始化组件
func (c *CustomComponent) Initialize(ctx context.Context) error {
	c.logger.Debug("初始化自定义组件")
	return nil
}

// Start 启动组件
func (c *CustomComponent) Start(ctx context.Context) error {
	c.logger.Info("启动自定义组件")

	// 启动一个后台协程
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				c.logger.Info("自定义组件后台任务停止")
				return
			case <-ticker.C:
				c.logger.Info("自定义组件执行定时任务")
			}
		}
	}()

	return nil
}

// Stop 停止组件
func (c *CustomComponent) Stop(ctx context.Context) error {
	c.logger.Info("停止自定义组件")
	return nil
}

// SetLogger 设置日志器
func (c *CustomComponent) SetLogger(logger logger.Logger) {
	c.logger = logger
}

// 自定义组件工厂
type CustomComponentFactory struct{}

// Create 创建自定义组件
func (f *CustomComponentFactory) Create(ctx context.Context, props boot.PropertySource) (boot.Component, error) {
	// 创建组件
	component := &CustomComponent{
		name:   "custom",
		logger: logger.New(), // 使用空参数创建日志器
	}

	return component, nil
}

// Dependencies 依赖
func (f *CustomComponentFactory) Dependencies() []string {
	return []string{"logger"}
}

// 自定义配置器
type CustomConfigurer struct{}

// Configure 配置自定义组件
func (c *CustomConfigurer) Configure(registry *boot.ComponentRegistry, props boot.PropertySource) error {
	// 检查是否启用自定义组件
	enabled := props.GetBool("custom.enabled", true)
	if !enabled {
		return nil
	}

	// 注册自定义组件工厂
	registry.RegisterFactory("custom", &CustomComponentFactory{})

	return nil
}

// Order 配置顺序
func (c *CustomConfigurer) Order() int {
	return 1000 // 在标准组件之后配置
}

func main() {
	// 创建启动器
	bootApp := boot.NewBoot()

	// 添加自定义配置器
	bootApp.AddConfigurer(&CustomConfigurer{})

	// 运行应用
	err := bootApp.Run()
	if err != nil {
		log.Fatalf("应用运行失败: %v", err)
	}

	fmt.Println("应用已退出")
}
