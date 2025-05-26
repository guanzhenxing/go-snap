package dbstore

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNew(t *testing.T) {
	// 使用 MySQL 进行测试
	config := Config{
		Driver:           "mysql",
		DSN:              "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
		MaxOpenConns:     10,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Hour,
		ConnMaxIdleTime:  time.Minute * 10,
		TablePrefix:      "test_",
		SingularTable:    true,
		Debug:            true,
		SlowThreshold:    time.Millisecond * 100,
		SkipDefaultTxn:   false,
		PrepareStmt:      true,
		DisableNestedTxn: false,
	}

	store, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, store)
	defer store.Close()

	// 验证连接是否正常
	err = store.Ping()
	assert.NoError(t, err)

	// 测试获取 DB 实例
	db := store.DB()
	assert.NotNil(t, db)

	// 测试统计信息
	stats := store.Stats()
	assert.NotNil(t, stats)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "mysql", config.Driver)
	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, 10, config.MaxIdleConns)
	assert.Equal(t, time.Hour, config.ConnMaxLifetime)
	assert.Equal(t, time.Minute*10, config.ConnMaxIdleTime)
	assert.Equal(t, "", config.TablePrefix)
	assert.False(t, config.SingularTable)
	assert.False(t, config.Debug)
	assert.Equal(t, time.Millisecond*200, config.SlowThreshold)
	assert.False(t, config.SkipDefaultTxn)
	assert.True(t, config.PrepareStmt)
	assert.False(t, config.DisableNestedTxn)
}

func TestStoreWithOptions(t *testing.T) {
	config := Config{
		Driver: "mysql",
		DSN:    "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
	}

	// 测试选项
	store, err := New(config,
		WithDebug(true),
		WithTablePrefix("prefix_"),
		WithSingularTable(true),
		WithPreparedStatement(false),
		WithSkipDefaultTransaction(true),
		WithDisableNestedTransaction(true),
	)
	require.NoError(t, err)
	require.NotNil(t, store)
	defer store.Close()

	// 验证选项是否被正确应用
	assert.True(t, store.config.Debug)
	assert.Equal(t, "prefix_", store.config.TablePrefix)
	assert.True(t, store.config.SingularTable)
	assert.False(t, store.config.PrepareStmt)
	assert.True(t, store.config.SkipDefaultTxn)
	assert.True(t, store.config.DisableNestedTxn)
}

func TestConcurrentOperations(t *testing.T) {
	// 使用 MySQL 进行测试
	config := Config{
		Driver:          "mysql",
		DSN:             "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
		MaxOpenConns:    20, // 增加连接数以支持并发
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 10,
		TablePrefix:     "test_",
		SingularTable:   true,
		Debug:           false, // 禁用调试以减少日志输出
	}

	store, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, store)
	defer store.Close()

	// 删除表（如果存在）
	store.DB().Exec("DROP TABLE IF EXISTS test_concurrent_test")

	// 创建测试表
	err = store.DB().Exec(`
		CREATE TABLE IF NOT EXISTS test_concurrent_test (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			counter INT NOT NULL DEFAULT 0,
			created_at DATETIME(3) NULL,
			updated_at DATETIME(3) NULL
		)
	`).Error
	require.NoError(t, err)

	// 创建初始记录
	err = store.DB().Exec("INSERT INTO test_concurrent_test (name, counter, created_at, updated_at) VALUES (?, ?, NOW(), NOW())", "concurrent-test", 0).Error
	require.NoError(t, err)

	// 并发操作次数
	concurrency := 10
	iterations := 5
	var wg sync.WaitGroup

	// 并发更新计数器
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				err := store.Transaction(func(tx *gorm.DB) error {
					// 读取当前计数
					var counter int
					err := tx.Raw("SELECT counter FROM test_concurrent_test WHERE name = ? FOR UPDATE", "concurrent-test").Scan(&counter).Error
					if err != nil {
						return err
					}

					// 增加计数并更新
					counter++
					return tx.Exec("UPDATE test_concurrent_test SET counter = ?, updated_at = NOW() WHERE name = ?", counter, "concurrent-test").Error
				})

				// 遇到错误就重试，但不要无限重试
				if err != nil {
					time.Sleep(10 * time.Millisecond)
					j--
					continue
				}
			}
		}(i)
	}

	// 等待所有并发操作完成
	wg.Wait()

	// 验证最终计数是否正确
	var result struct {
		Counter int
	}
	err = store.DB().Raw("SELECT counter FROM test_concurrent_test WHERE name = ?", "concurrent-test").Scan(&result).Error
	require.NoError(t, err)

	// 期望的计数值 = 并发数 * 迭代次数
	expectedCount := concurrency * iterations
	assert.Equal(t, expectedCount, result.Counter, "并发更新后的计数器值应该是%d", expectedCount)
}

// 测试Close、Ping和Stats方法的错误情况
func TestStoreErrorHandling(t *testing.T) {
	// 创建一个无效的数据库配置 - 使用一个非存在的MySQL服务器
	config := Config{
		Driver:          "mysql",
		DSN:             "root:wrongpassword@tcp(localhost:3307)/nonexistentdb", // 使用不存在的端口
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 10,
		TablePrefix:     "test_",
		SingularTable:   true,
	}

	// 尝试创建数据库实例，应该会失败
	store, err := New(config)
	// 因为无法连接到数据库，应该会返回错误
	assert.Error(t, err)
	assert.Nil(t, store)
}

// 测试Initialize方法的错误情况
func TestStoreInitializeError(t *testing.T) {
	// 测试不支持的驱动
	t.Run("unsupported_driver", func(t *testing.T) {
		config := Config{
			Driver:          "unsupported",
			DSN:             "test",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: time.Minute * 10,
		}

		store, err := New(config)
		assert.Error(t, err)
		assert.Nil(t, store)
		assert.Contains(t, err.Error(), "unsupported database driver")
	})

	// 测试无效的DSN
	t.Run("invalid_dsn", func(t *testing.T) {
		config := Config{
			Driver:          "mysql",
			DSN:             "invalid_dsn",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: time.Minute * 10,
		}

		store, err := New(config)
		assert.Error(t, err)
		assert.Nil(t, store)
		assert.Contains(t, err.Error(), "failed to connect to database")
	})
}
