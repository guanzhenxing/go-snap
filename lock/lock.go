// Package lock 提供分布式锁实现
package lock

import (
	"context"
	"time"
)

// Lock 分布式锁接口
type Lock interface {
	// Acquire 获取锁
	Acquire(ctx context.Context) (bool, error)

	// Release 释放锁
	Release(ctx context.Context) (bool, error)

	// Refresh 刷新锁的过期时间
	Refresh(ctx context.Context) (bool, error)
}

// Options 分布式锁选项
type Options struct {
	// 锁的过期时间，防止死锁
	Expiration time.Duration

	// 重试获取锁的间隔
	RetryInterval time.Duration

	// 重试获取锁的最大次数，如果为0则表示无限重试
	MaxRetries int

	// 锁的随机值
	RandomValue string
}

// DefaultOptions 返回默认的锁选项
func DefaultOptions() Options {
	return Options{
		Expiration:    time.Second * 10, // 默认10秒过期
		RetryInterval: time.Millisecond * 100,
		MaxRetries:    50, // 5秒
	}
}

// WithLock 执行带锁的操作
func WithLock(ctx context.Context, l Lock, fn func() error) error {
	acquired, err := l.Acquire(ctx)
	if err != nil {
		return err
	}

	if !acquired {
		return ErrAcquireLockFailed
	}

	defer func() {
		_, _ = l.Release(ctx)
	}()

	return fn()
}
