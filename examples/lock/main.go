package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/guanzhenxing/go-snap/cache"
	"github.com/guanzhenxing/go-snap/lock"
)

func main() {
	ctx := context.Background()
	fmt.Println("===== 分布式锁示例 =====")

	// 示例1: 直接使用Redis客户端
	fmt.Println("\n1. 直接使用Redis客户端创建锁")

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 测试连接
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Redis连接失败: %v\n", err)
		fmt.Println("继续展示代码结构...")
	} else {
		fmt.Println("Redis连接成功!")

		// 创建Redis锁
		redisLock := lock.NewRedisLock(redisClient, "direct-lock", lock.Options{
			Expiration:    time.Second * 10,
			RetryInterval: time.Millisecond * 100,
			MaxRetries:    5,
		})

		// 使用锁执行操作
		baseLockDemo(ctx, redisLock)
	}

	// 示例2: 通过缓存创建锁
	fmt.Println("\n2. 通过缓存创建锁")

	// 创建Redis缓存
	redisOpts := cache.RedisOptions{
		Mode:       "standalone",
		Addr:       "localhost:6379",
		DB:         0,
		KeyPrefix:  "lock-demo",
		DefaultTTL: time.Minute,
	}

	redisCache, err := cache.NewRedisCache(redisOpts, nil)
	if err != nil {
		fmt.Printf("Redis缓存创建失败: %v\n", err)
		fmt.Println("继续展示代码结构...")
	} else {
		fmt.Println("Redis缓存创建成功!")

		// 从缓存创建锁
		cacheLock, err := lock.FromCache(redisCache, "cache-lock", lock.Options{
			Expiration: time.Second * 10,
		})

		if err != nil {
			fmt.Printf("从缓存创建锁失败: %v\n", err)
		} else {
			baseLockDemo(ctx, cacheLock)
		}

		// 使用便捷的WithCacheLock函数
		fmt.Println("\n使用WithCacheLock便捷函数:")
		err = lock.WithCacheLock(ctx, redisCache, "quick-lock", func() error {
			fmt.Println("在锁保护下执行操作...")
			time.Sleep(time.Second)
			return nil
		})

		if err != nil {
			fmt.Printf("带锁操作失败: %v\n", err)
		} else {
			fmt.Println("带锁操作成功完成")
		}
	}

	fmt.Println("\n===== 示例结束 =====")
}

// baseLockDemo 展示基本的锁操作
func baseLockDemo(ctx context.Context, l lock.Lock) {
	// 尝试获取锁
	acquired, err := l.Acquire(ctx)
	if err != nil {
		fmt.Printf("获取锁失败: %v\n", err)
		return
	}

	if acquired {
		fmt.Println("成功获取锁")

		// 模拟执行受保护的操作
		fmt.Println("执行受锁保护的操作...")
		time.Sleep(time.Second)

		// 刷新锁的过期时间
		refreshed, err := l.Refresh(ctx)
		if err != nil {
			fmt.Printf("刷新锁失败: %v\n", err)
		} else if refreshed {
			fmt.Println("成功刷新锁")
		} else {
			fmt.Println("锁已不存在，无法刷新")
		}

		// 释放锁
		released, err := l.Release(ctx)
		if err != nil {
			fmt.Printf("释放锁失败: %v\n", err)
		} else if released {
			fmt.Println("成功释放锁")
		} else {
			fmt.Println("锁已不存在，无法释放")
		}
	} else {
		fmt.Println("无法获取锁，可能已被其他客户端持有")
	}
}
