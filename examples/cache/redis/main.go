package main

import (
	"context"
	"fmt"
	"time"

	"github.com/guanzhenxing/go-snap/cache"
)

func main() {
	// 创建一个基本上下文
	ctx := context.Background()

	// 创建Redis缓存选项
	redisOpts := cache.RedisOptions{
		Mode:       cache.RedisModeStandalone,
		Addr:       "localhost:6379",
		DB:         0,
		KeyPrefix:  "example",
		DefaultTTL: time.Minute,
	}

	// 创建Redis缓存
	fmt.Println("===== Redis缓存示例 =====")
	fmt.Println("正在连接到Redis...")

	redisCache, err := cache.NewRedisCache(redisOpts, nil)
	if err != nil {
		fmt.Printf("创建Redis缓存失败: %v\n", err)
		fmt.Println("请确保Redis服务已启动并可访问。如果Redis未启动，本示例将仅展示代码结构。")
		return
	}

	fmt.Println("Redis连接成功!")

	// 基本缓存操作
	err = redisCache.Set(ctx, "redis-key", "redis-value", time.Minute)
	if err != nil {
		fmt.Printf("设置Redis缓存失败: %v\n", err)
		return
	}
	fmt.Println("设置Redis缓存成功: redis-key = redis-value, TTL = 1分钟")

	// 获取缓存
	value, found := redisCache.Get(ctx, "redis-key")
	if found {
		fmt.Printf("获取Redis缓存成功: redis-key = %v\n", value)
	} else {
		fmt.Println("获取Redis缓存失败: redis-key 不存在")
	}

	// 关闭缓存连接
	redisCache.Close()
	fmt.Println("\nRedis缓存已关闭")
}
