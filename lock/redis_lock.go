package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLock 基于Redis的分布式锁实现
type RedisLock struct {
	client     redis.UniversalClient
	key        string
	options    Options
	isAcquired bool
}

// NewRedisLock 创建新的Redis分布式锁
func NewRedisLock(client redis.UniversalClient, key string, opts ...Options) *RedisLock {
	var options Options
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = DefaultOptions()
	}

	// 如果未提供随机值，使用当前时间戳作为随机值
	if options.RandomValue == "" {
		options.RandomValue = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return &RedisLock{
		client:     client,
		key:        key,
		options:    options,
		isAcquired: false,
	}
}

// Acquire 获取锁
func (l *RedisLock) Acquire(ctx context.Context) (bool, error) {
	var retries int

	for {
		// 尝试使用SET NX命令获取锁
		ok, err := l.client.SetNX(ctx, l.key, l.options.RandomValue, l.options.Expiration).Result()
		if err != nil {
			return false, fmt.Errorf("failed to acquire lock: %w", err)
		}

		if ok {
			l.isAcquired = true
			return true, nil
		}

		// 如果达到最大重试次数，退出
		if l.options.MaxRetries > 0 && retries >= l.options.MaxRetries {
			return false, nil
		}

		// 否则等待一段时间再重试
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(l.options.RetryInterval):
			retries++
		}
	}
}

// Release 释放锁，只有持有锁的客户端才能释放
func (l *RedisLock) Release(ctx context.Context) (bool, error) {
	if !l.isAcquired {
		return false, ErrLockNotHeld
	}

	// 使用Lua脚本确保只有持有锁的客户端才能释放锁
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.options.RandomValue).Result()
	if err != nil {
		return false, fmt.Errorf("failed to release lock: %w", err)
	}

	l.isAcquired = false
	return result.(int64) == 1, nil
}

// Refresh 刷新锁的过期时间
func (l *RedisLock) Refresh(ctx context.Context) (bool, error) {
	if !l.isAcquired {
		return false, ErrLockNotHeld
	}

	// 使用Lua脚本确保只刷新自己持有的锁
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := l.client.Eval(
		ctx,
		script,
		[]string{l.key},
		l.options.RandomValue,
		int64(l.options.Expiration/time.Millisecond),
	).Result()

	if err != nil {
		return false, fmt.Errorf("failed to refresh lock: %w", err)
	}

	return result.(int64) == 1, nil
}

// GetClient 返回Redis客户端实例
func (l *RedisLock) GetClient() redis.UniversalClient {
	return l.client
}
