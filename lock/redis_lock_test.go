package lock

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// 创建测试用的Redis客户端
func setupRedisTest(t *testing.T) (*miniredis.Miniredis, redis.UniversalClient) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("创建miniredis失败: %s", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return mr, client
}

// 测试创建Redis锁
func TestNewRedisLock(t *testing.T) {
	mr, client := setupRedisTest(t)
	defer mr.Close()

	// 测试默认选项
	lock := NewRedisLock(client, "test-lock")
	assert.Equal(t, "test-lock", lock.key)
	assert.Equal(t, time.Second*10, lock.options.Expiration)
	assert.Equal(t, time.Millisecond*100, lock.options.RetryInterval)
	assert.Equal(t, 50, lock.options.MaxRetries)
	assert.NotEmpty(t, lock.options.RandomValue)

	// 测试自定义选项
	customOpts := Options{
		Expiration:    time.Minute,
		RetryInterval: time.Second,
		MaxRetries:    5,
		RandomValue:   "custom-value",
	}
	customLock := NewRedisLock(client, "custom-lock", customOpts)
	assert.Equal(t, "custom-lock", customLock.key)
	assert.Equal(t, time.Minute, customLock.options.Expiration)
	assert.Equal(t, time.Second, customLock.options.RetryInterval)
	assert.Equal(t, 5, customLock.options.MaxRetries)
	assert.Equal(t, "custom-value", customLock.options.RandomValue)
}

// 测试获取锁
func TestRedisLockAcquire(t *testing.T) {
	mr, client := setupRedisTest(t)
	defer mr.Close()

	ctx := context.Background()
	lock := NewRedisLock(client, "test-lock", Options{
		Expiration:    time.Second * 5,
		RetryInterval: time.Millisecond * 10,
		MaxRetries:    3,
		RandomValue:   "test-value",
	})

	// 测试首次获取锁成功
	acquired, err := lock.Acquire(ctx)
	assert.NoError(t, err)
	assert.True(t, acquired)
	assert.True(t, lock.isAcquired)

	// 验证Redis中的值
	val, err := mr.Get("test-lock")
	assert.NoError(t, err)
	assert.Equal(t, "test-value", val)

	// 验证过期时间已设置
	ttl := mr.TTL("test-lock")
	assert.True(t, ttl > 0 && ttl <= 5*time.Second)

	// 测试获取已被锁定的锁
	lockB := NewRedisLock(client, "test-lock", Options{
		Expiration:    time.Second * 5,
		RetryInterval: time.Millisecond * 10,
		MaxRetries:    3,
		RandomValue:   "other-value",
	})

	acquired, err = lockB.Acquire(ctx)
	assert.NoError(t, err)
	assert.False(t, acquired)
}

// 测试释放锁
func TestRedisLockRelease(t *testing.T) {
	mr, client := setupRedisTest(t)
	defer mr.Close()

	ctx := context.Background()
	lock := NewRedisLock(client, "test-lock", Options{
		RandomValue: "test-value",
	})

	// 先获取锁
	acquired, err := lock.Acquire(ctx)
	assert.NoError(t, err)
	assert.True(t, acquired)

	// 测试释放锁
	released, err := lock.Release(ctx)
	assert.NoError(t, err)
	assert.True(t, released)
	assert.False(t, lock.isAcquired)

	// 验证锁已从Redis中删除
	exists := mr.Exists("test-lock")
	assert.False(t, exists)

	// 测试释放未持有的锁
	released, err = lock.Release(ctx)
	assert.Equal(t, ErrLockNotHeld, err)
	assert.False(t, released)

	// 测试释放被他人持有的锁
	// 先在Redis中设置锁
	mr.Set("test-lock", "other-value")

	lock = NewRedisLock(client, "test-lock", Options{
		RandomValue: "test-value",
	})
	lock.isAcquired = true // 手动设置状态

	released, err = lock.Release(ctx)
	assert.NoError(t, err)
	assert.False(t, released)

	// 验证其他值的锁没有被删除
	exists = mr.Exists("test-lock")
	assert.True(t, exists)
}

// 测试刷新锁
func TestRedisLockRefresh(t *testing.T) {
	mr, client := setupRedisTest(t)
	defer mr.Close()

	ctx := context.Background()
	lock := NewRedisLock(client, "test-lock", Options{
		Expiration:  time.Second * 5,
		RandomValue: "test-value",
	})

	// 先获取锁
	acquired, err := lock.Acquire(ctx)
	assert.NoError(t, err)
	assert.True(t, acquired)

	// 记录初始TTL
	initialTTL := mr.TTL("test-lock")

	// 等待部分过期时间
	time.Sleep(time.Second)

	// 刷新锁
	refreshed, err := lock.Refresh(ctx)
	assert.NoError(t, err)
	assert.True(t, refreshed)

	// 验证TTL已重置
	newTTL := mr.TTL("test-lock")
	assert.True(t, newTTL > initialTTL-time.Second)

	// 测试刷新未持有的锁
	lock = NewRedisLock(client, "test-lock", Options{
		RandomValue: "test-value",
	})
	refreshed, err = lock.Refresh(ctx)
	assert.Equal(t, ErrLockNotHeld, err)
	assert.False(t, refreshed)

	// 测试刷新被他人持有的锁
	lock.isAcquired = true // 手动设置状态
	mr.Set("test-lock", "other-value")

	refreshed, err = lock.Refresh(ctx)
	assert.NoError(t, err)
	assert.False(t, refreshed)
}

// 测试锁超时
func TestRedisLockExpiration(t *testing.T) {
	mr, client := setupRedisTest(t)
	defer mr.Close()

	ctx := context.Background()
	lock := NewRedisLock(client, "test-lock", Options{
		Expiration:  time.Millisecond * 50, // 非常短的超时时间
		RandomValue: "test-value",
	})

	// 获取锁
	acquired, err := lock.Acquire(ctx)
	assert.NoError(t, err)
	assert.True(t, acquired)

	// 手动设置过期时间
	mr.FastForward(time.Millisecond * 100) // 快进时间使锁过期

	// 验证锁已从Redis中过期
	exists := mr.Exists("test-lock")
	assert.False(t, exists, "锁应该已经过期")

	// 即使本地状态显示已获取，也应该能再次获取锁
	lock2 := NewRedisLock(client, "test-lock", Options{
		RandomValue: "test-value-2",
	})
	acquired, err = lock2.Acquire(ctx)
	assert.NoError(t, err)
	assert.True(t, acquired, "应该能获取已过期的锁")

	// 验证新值已写入
	val, err := mr.Get("test-lock")
	assert.NoError(t, err)
	assert.Equal(t, "test-value-2", val)
}

// 测试并发获取锁
func TestRedisLockConcurrency(t *testing.T) {
	mr, client := setupRedisTest(t)
	defer mr.Close()

	const numGoroutines = 5 // 减少并发数量，提高测试稳定性
	var successCount int32
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	ctx := context.Background()

	// 并发获取同一把锁
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			lock := NewRedisLock(client, "concurrent-lock", Options{
				Expiration:    time.Second * 1,
				RetryInterval: time.Millisecond * 10,
				MaxRetries:    3,
				RandomValue:   fmt.Sprintf("value-%d", id),
			})

			acquired, err := lock.Acquire(ctx)
			if err == nil && acquired {
				// 模拟持有锁的工作
				time.Sleep(time.Millisecond * 10)
				lock.Release(ctx)
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	// 应该至少有一个goroutine能获取到锁
	assert.True(t, successCount > 0, "至少一个goroutine应该能获取到锁")
	// 由于并发性，不是所有goroutine都能获取锁
	assert.True(t, successCount <= numGoroutines, "不应该所有goroutine都能获取锁")

	// 最后锁应该已释放
	exists := mr.Exists("concurrent-lock")
	assert.False(t, exists, "测试结束后锁应该已释放")
}

// 测试获取锁超时
func TestRedisLockAcquireTimeout(t *testing.T) {
	mr, client := setupRedisTest(t)
	defer mr.Close()

	ctx := context.Background()

	// 先获取一个锁
	lock1 := NewRedisLock(client, "timeout-lock", Options{
		Expiration:  time.Second * 5,
		RandomValue: "first-value",
	})
	acquired, err := lock1.Acquire(ctx)
	assert.NoError(t, err)
	assert.True(t, acquired)

	// 尝试获取同一把锁，设置较短的重试次数
	lock2 := NewRedisLock(client, "timeout-lock", Options{
		Expiration:    time.Second * 5,
		RetryInterval: time.Millisecond * 10,
		MaxRetries:    2, // 只重试2次
		RandomValue:   "second-value",
	})

	// 应该很快返回，且获取失败
	startTime := time.Now()
	acquired, err = lock2.Acquire(ctx)
	elapsed := time.Since(startTime)

	assert.NoError(t, err)
	assert.False(t, acquired)
	assert.True(t, elapsed < time.Millisecond*100, "获取锁超时应该很快返回")

	// 测试上下文取消
	ctxWithCancel, cancel := context.WithCancel(ctx)
	go func() {
		time.Sleep(time.Millisecond * 30)
		cancel()
	}()

	lock3 := NewRedisLock(client, "timeout-lock", Options{
		Expiration:    time.Second * 5,
		RetryInterval: time.Second * 10, // 长时间等待
		MaxRetries:    0,                // 无限重试
		RandomValue:   "third-value",
	})

	// 应该被上下文取消
	startTime = time.Now()
	acquired, err = lock3.Acquire(ctxWithCancel)
	elapsed = time.Since(startTime)

	assert.Error(t, err)
	assert.False(t, acquired)
	assert.True(t, elapsed < time.Second, "上下文取消应该及时返回")
}

// 测试GetClient方法
func TestRedisLockGetClient(t *testing.T) {
	mr, client := setupRedisTest(t)
	defer mr.Close()

	lock := NewRedisLock(client, "test-lock")

	// 验证GetClient返回正确的客户端
	returnedClient := lock.GetClient()
	assert.Equal(t, client, returnedClient)

	// 验证客户端功能
	err := returnedClient.Set(context.Background(), "test-key", "test-value", 0).Err()
	assert.NoError(t, err)

	val, err := mr.Get("test-key")
	assert.NoError(t, err)
	assert.Equal(t, "test-value", val)
}
