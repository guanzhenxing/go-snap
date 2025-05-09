package main

import (
	"context"
	"fmt"
	"time"

	"github.com/guanzhenxing/go-snap/cache"
)

func main() {
	// 创建内存缓存
	memoryCache := cache.NewMemoryCache()

	// 创建一个基本上下文
	ctx := context.Background()

	// 基本缓存操作
	fmt.Println("===== 基本缓存操作 =====")

	// 设置缓存
	err := memoryCache.Set(ctx, "key1", "value1", time.Minute)
	if err != nil {
		fmt.Printf("设置缓存失败: %v\n", err)
		return
	}
	fmt.Println("设置缓存成功: key1 = value1, TTL = 1分钟")

	// 获取缓存
	value, found := memoryCache.Get(ctx, "key1")
	if found {
		fmt.Printf("获取缓存成功: key1 = %v\n", value)
	} else {
		fmt.Println("获取缓存失败: key1 不存在")
	}

	// 检查键是否存在
	exists, _ := memoryCache.Exists(ctx, "key1")
	fmt.Printf("key1 是否存在: %v\n", exists)

	// 获取带TTL的缓存
	value, ttl, found := memoryCache.GetWithTTL(ctx, "key1")
	if found {
		fmt.Printf("获取缓存成功: key1 = %v, 剩余TTL = %v\n", value, ttl)
	}

	// 带标签的缓存
	item := &cache.Item{
		Value:      "value with tags",
		Expiration: time.Minute,
		Tags:       []string{"tag1", "tag2"},
	}

	err = memoryCache.SetItem(ctx, "key2", item)
	if err != nil {
		fmt.Printf("设置带标签缓存失败: %v\n", err)
		return
	}
	fmt.Println("设置带标签缓存成功: key2 = value with tags, Tags = [tag1, tag2]")

	// 根据标签删除缓存
	err = memoryCache.DeleteByTag(ctx, "tag1")
	if err != nil {
		fmt.Printf("根据标签删除缓存失败: %v\n", err)
		return
	}
	fmt.Println("根据标签删除缓存成功: tag = tag1")

	// 检查是否删除成功
	_, found = memoryCache.Get(ctx, "key2")
	fmt.Printf("key2 是否仍然存在: %v\n", found)

	// 数值操作
	memoryCache.Set(ctx, "counter", 0, 0)
	fmt.Println("设置计数器: counter = 0")

	// 增加计数器
	newValue, err := memoryCache.Increment(ctx, "counter", 5)
	if err != nil {
		fmt.Printf("增加计数器失败: %v\n", err)
	} else {
		fmt.Printf("增加计数器成功: counter = %v\n", newValue)
	}

	// 减少计数器
	newValue, err = memoryCache.Decrement(ctx, "counter", 2)
	if err != nil {
		fmt.Printf("减少计数器失败: %v\n", err)
	} else {
		fmt.Printf("减少计数器成功: counter = %v\n", newValue)
	}

	// 使用构建器模式创建缓存
	fmt.Println("\n===== 使用构建器模式 =====")

	memBuilder := cache.NewBuilder().
		WithType(cache.CacheTypeMemory).
		WithOptions(cache.Options{
			DefaultTTL:      time.Minute * 30,
			CleanupInterval: time.Minute,
		})

	memCache, err := memBuilder.Build()
	if err != nil {
		fmt.Printf("创建内存缓存失败: %v\n", err)
		return
	}

	memCache.Set(ctx, "builder-key", "builder-value", 0)
	value, found = memCache.Get(ctx, "builder-key")
	if found {
		fmt.Printf("通过构建器创建的缓存工作正常: builder-key = %v\n", value)
	}

	// 清空缓存
	err = memCache.Flush(ctx)
	if err != nil {
		fmt.Printf("清空缓存失败: %v\n", err)
	} else {
		fmt.Println("清空缓存成功")
	}

	// 关闭缓存
	memCache.Close()
	fmt.Println("缓存已关闭")
}
