package lock

import (
	"github.com/guanzhenxing/go-snap/errors"
)

// 锁相关错误定义
var (

	// ErrAcquireLockFailed 表示获取锁失败
	ErrAcquireLockFailed = errors.New("failed to acquire lock after retries")

	// ErrInvalidLockClient 表示无效的锁客户端
	ErrInvalidLockClient = errors.New("invalid lock client")

	// ErrLockNotHeld 表示未持有锁
	ErrLockNotHeld = errors.New("lock is not held")
)
