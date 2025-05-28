package lock

import (
	"context"
	"testing"
	"time"
)

// MockLock 模拟锁实现，用于测试
type MockLock struct {
	key          string
	ttl          time.Duration
	isLocked     bool
	lockCount    int
	releaseError error
	acquireError error
}

// NewMockLock 创建模拟锁
func NewMockLock(key string, ttl time.Duration) *MockLock {
	return &MockLock{
		key:       key,
		ttl:       ttl,
		isLocked:  false,
		lockCount: 0,
	}
}

func (m *MockLock) Acquire(ctx context.Context) (bool, error) {
	if m.acquireError != nil {
		return false, m.acquireError
	}
	if !m.isLocked {
		m.isLocked = true
		m.lockCount++
		return true, nil
	}
	return false, nil
}

func (m *MockLock) Release(ctx context.Context) (bool, error) {
	if m.releaseError != nil {
		return false, m.releaseError
	}
	if m.isLocked {
		m.isLocked = false
		return true, nil
	}
	return false, nil
}

func (m *MockLock) Refresh(ctx context.Context) (bool, error) {
	if m.isLocked {
		return true, nil
	}
	return false, nil
}

func (m *MockLock) IsHeldByCurrentThread() bool {
	return m.isLocked
}

// 测试模拟锁的基本功能
func TestMockLock(t *testing.T) {
	lock := NewMockLock("test-key", time.Minute)
	ctx := context.Background()

	// 测试初始状态
	if lock.IsHeldByCurrentThread() {
		t.Error("新创建的锁不应该被持有")
	}

	// 测试获取锁
	acquired, err := lock.Acquire(ctx)
	if err != nil {
		t.Errorf("Acquire() 失败: %v", err)
	}
	if !acquired {
		t.Error("第一次 Acquire() 应该成功")
	}

	if !lock.IsHeldByCurrentThread() {
		t.Error("锁应该被持有")
	}

	// 测试再次获取锁应该失败
	acquired, err = lock.Acquire(ctx)
	if err != nil {
		t.Errorf("Acquire() 失败: %v", err)
	}
	if acquired {
		t.Error("重复 Acquire() 应该失败")
	}

	// 测试释放锁
	released, err := lock.Release(ctx)
	if err != nil {
		t.Errorf("Release() 失败: %v", err)
	}
	if !released {
		t.Error("Release() 应该成功")
	}

	if lock.IsHeldByCurrentThread() {
		t.Error("锁应该已被释放")
	}

	// 测试释放未锁定的锁
	released, err = lock.Release(ctx)
	if err != nil {
		t.Errorf("Release() 失败: %v", err)
	}
	if released {
		t.Error("释放未锁定的锁应该返回false")
	}
}

// 测试刷新功能
func TestMockLockRefresh(t *testing.T) {
	lock := NewMockLock("test-key", time.Minute)
	ctx := context.Background()

	// 未获取锁时刷新应该失败
	refreshed, err := lock.Refresh(ctx)
	if err != nil {
		t.Errorf("Refresh() 失败: %v", err)
	}
	if refreshed {
		t.Error("未获取锁时 Refresh() 应该返回false")
	}

	// 获取锁
	acquired, err := lock.Acquire(ctx)
	if err != nil {
		t.Errorf("Acquire() 失败: %v", err)
	}
	if !acquired {
		t.Error("Acquire() 应该成功")
	}

	// 刷新锁
	refreshed, err = lock.Refresh(ctx)
	if err != nil {
		t.Errorf("Refresh() 失败: %v", err)
	}
	if !refreshed {
		t.Error("持有锁时 Refresh() 应该成功")
	}
}

// 测试错误处理
func TestMockLockErrors(t *testing.T) {
	lock := NewMockLock("test-key", time.Minute)
	ctx := context.Background()

	// 测试获取错误
	expectedError := ErrAcquireLockFailed
	lock.acquireError = expectedError

	acquired, err := lock.Acquire(ctx)
	if err != expectedError {
		t.Errorf("Acquire() 应该返回预期错误: got %v, want %v", err, expectedError)
	}
	if acquired {
		t.Error("出错时 Acquire() 应该返回false")
	}

	// 重置获取错误，设置释放错误
	lock.acquireError = nil
	lock.releaseError = ErrLockNotHeld

	// 先成功获取锁
	acquired, err = lock.Acquire(ctx)
	if err != nil {
		t.Errorf("Acquire() 失败: %v", err)
	}
	if !acquired {
		t.Error("Acquire() 应该成功")
	}

	// 释放时应该失败
	released, err := lock.Release(ctx)
	if err != ErrLockNotHeld {
		t.Errorf("Release() 应该返回预期错误: got %v, want %v", err, ErrLockNotHeld)
	}
	if released {
		t.Error("出错时 Release() 应该返回false")
	}
}

// 测试WithLock功能
func TestWithLock(t *testing.T) {
	lock := NewMockLock("test-key", time.Minute)
	ctx := context.Background()

	executed := false
	err := WithLock(ctx, lock, func() error {
		executed = true
		if !lock.IsHeldByCurrentThread() {
			t.Error("在WithLock回调中锁应该被持有")
		}
		return nil
	})

	if err != nil {
		t.Errorf("WithLock() 失败: %v", err)
	}

	if !executed {
		t.Error("WithLock回调应该被执行")
	}

	if lock.IsHeldByCurrentThread() {
		t.Error("WithLock完成后锁应该被释放")
	}
}

// 测试WithLock获取锁失败
func TestWithLockAcquireFailed(t *testing.T) {
	lock := NewMockLock("test-key", time.Minute)
	ctx := context.Background()

	// 先获取锁使其无法再次获取
	lock.Acquire(ctx)

	executed := false
	err := WithLock(ctx, lock, func() error {
		executed = true
		return nil
	})

	if err != ErrAcquireLockFailed {
		t.Errorf("WithLock() 应该返回 ErrAcquireLockFailed, got %v", err)
	}

	if executed {
		t.Error("获取锁失败时回调不应该被执行")
	}
}

// 测试WithLock回调错误
func TestWithLockCallbackError(t *testing.T) {
	lock := NewMockLock("test-key", time.Minute)
	ctx := context.Background()

	expectedError := &LockError{
		Op:      "callback",
		Key:     "test-key",
		Message: "回调错误",
	}

	err := WithLock(ctx, lock, func() error {
		return expectedError
	})

	if err != expectedError {
		t.Errorf("WithLock() 应该返回回调错误: got %v, want %v", err, expectedError)
	}

	if lock.IsHeldByCurrentThread() {
		t.Error("即使回调出错，锁也应该被释放")
	}
}

// 测试默认选项
func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Expiration != time.Second*10 {
		t.Errorf("默认过期时间错误: got %v, want %v", opts.Expiration, time.Second*10)
	}

	if opts.RetryInterval != time.Millisecond*100 {
		t.Errorf("默认重试间隔错误: got %v, want %v", opts.RetryInterval, time.Millisecond*100)
	}

	if opts.MaxRetries != 50 {
		t.Errorf("默认最大重试次数错误: got %d, want 50", opts.MaxRetries)
	}
}

// LockError 锁错误类型（用于测试）
type LockError struct {
	Op      string
	Key     string
	Message string
	Err     error
}

func (e *LockError) Error() string {
	if e.Err != nil {
		return e.Op + " " + e.Key + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Op + " " + e.Key + ": " + e.Message
}

func (e *LockError) Unwrap() error {
	return e.Err
}

// 基准测试
func BenchmarkMockLockAcquireRelease(b *testing.B) {
	lock := NewMockLock("bench-key", time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		acquired, _ := lock.Acquire(ctx)
		if acquired {
			lock.Release(ctx)
		}
	}
}

func BenchmarkWithLock(b *testing.B) {
	lock := NewMockLock("bench-key", time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithLock(ctx, lock, func() error {
			return nil
		})
	}
}
