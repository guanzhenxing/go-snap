package dbstore

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// 测试用的模型结构
type TestUser struct {
	ID        uint           `gorm:"primarykey"`
	Name      string         `gorm:"size:100;not null"`
	Email     string         `gorm:"size:100;uniqueIndex"`
	Age       int            `gorm:"default:0"`
	Active    bool           `gorm:"default:true"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 批量创建对象的测试模型
type BatchTestUser struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"size:100;not null"`
	Email string `gorm:"size:100;uniqueIndex"`
	Age   int
}

// 保存对象的测试模型
type SaveTestUser struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"size:100;not null"`
	Email string `gorm:"size:100;uniqueIndex"`
	Age   int
}

// ====================== Provider测试代码 ======================
// 模拟配置提供器
type mockProvider struct {
	data map[string]interface{}
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		data: make(map[string]interface{}),
	}
}

func (m *mockProvider) Get(key string) interface{} {
	return m.data[key]
}

func (m *mockProvider) GetString(key string) string {
	if v, ok := m.data[key].(string); ok {
		return v
	}
	return ""
}

func (m *mockProvider) GetBool(key string) bool {
	if v, ok := m.data[key].(bool); ok {
		return v
	}
	return false
}

func (m *mockProvider) GetInt(key string) int {
	if v, ok := m.data[key].(int); ok {
		return v
	}
	return 0
}

func (m *mockProvider) GetInt64(key string) int64 {
	if v, ok := m.data[key].(int64); ok {
		return v
	}
	return 0
}

func (m *mockProvider) GetFloat64(key string) float64 {
	if v, ok := m.data[key].(float64); ok {
		return v
	}
	return 0
}

func (m *mockProvider) GetTime(key string) time.Time {
	if v, ok := m.data[key].(time.Time); ok {
		return v
	}
	return time.Time{}
}

func (m *mockProvider) GetDuration(key string) time.Duration {
	if v, ok := m.data[key].(time.Duration); ok {
		return v
	}
	return 0
}

func (m *mockProvider) GetStringSlice(key string) []string {
	if v, ok := m.data[key].([]string); ok {
		return v
	}
	return nil
}

func (m *mockProvider) GetStringMap(key string) map[string]interface{} {
	if v, ok := m.data[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}

func (m *mockProvider) GetStringMapString(key string) map[string]string {
	if v, ok := m.data[key].(map[string]string); ok {
		return v
	}
	return nil
}

func (m *mockProvider) IsSet(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *mockProvider) Set(key string, value interface{}) {
	m.data[key] = value
}

func (m *mockProvider) UnmarshalKey(key string, rawVal interface{}) error {
	// 简化版：直接将整个map给目标结构
	if v, ok := m.data[key].(map[string]interface{}); ok {
		// 这里我们做一个简单的转换，将map中的数据复制到Config结构中
		if cfg, ok := rawVal.(*Config); ok {
			if driver, ok := v["driver"].(string); ok {
				cfg.Driver = driver
			}
			if dsn, ok := v["dsn"].(string); ok {
				cfg.DSN = dsn
			}
			if maxOpenConns, ok := v["max_open_conns"].(int); ok {
				cfg.MaxOpenConns = maxOpenConns
			}
			if maxIdleConns, ok := v["max_idle_conns"].(int); ok {
				cfg.MaxIdleConns = maxIdleConns
			}
			if tablePrefix, ok := v["table_prefix"].(string); ok {
				cfg.TablePrefix = tablePrefix
			}
			if debug, ok := v["debug"].(bool); ok {
				cfg.Debug = debug
			}
		}
	}
	return nil
}

func (m *mockProvider) Unmarshal(rawVal interface{}) error {
	// 简化版：直接将整个map给目标结构
	if cfg, ok := rawVal.(*Config); ok {
		if driver, ok := m.data["driver"].(string); ok {
			cfg.Driver = driver
		}
		if dsn, ok := m.data["dsn"].(string); ok {
			cfg.DSN = dsn
		}
		if maxOpenConns, ok := m.data["max_open_conns"].(int); ok {
			cfg.MaxOpenConns = maxOpenConns
		}
		if maxIdleConns, ok := m.data["max_idle_conns"].(int); ok {
			cfg.MaxIdleConns = maxIdleConns
		}
		if tablePrefix, ok := m.data["table_prefix"].(string); ok {
			cfg.TablePrefix = tablePrefix
		}
		if debug, ok := m.data["debug"].(bool); ok {
			cfg.Debug = debug
		}
	}
	return nil
}

func (m *mockProvider) LoadConfig() error {
	return nil
}

func (m *mockProvider) WatchConfig() {
}

func (m *mockProvider) OnConfigChange(run func()) {
}

func (m *mockProvider) ValidateConfig() error {
	return nil
}

// 模拟错误的配置提供器
type errorMockProvider struct {
	*mockProvider
}

func newErrorMockProvider() *errorMockProvider {
	return &errorMockProvider{
		mockProvider: newMockProvider(),
	}
}

func (m *errorMockProvider) Unmarshal(rawVal interface{}) error {
	return assert.AnError
}

func (m *errorMockProvider) UnmarshalKey(key string, rawVal interface{}) error {
	return assert.AnError
}

// 测试LoadFromProvider方法
func TestConfig_LoadFromProvider(t *testing.T) {
	// 创建模拟配置提供器
	p := newMockProvider()
	p.Set("driver", "mysql")
	p.Set("dsn", "user:pass@tcp(localhost:3306)/dbname")
	p.Set("max_open_conns", 50)
	p.Set("max_idle_conns", 10)
	p.Set("table_prefix", "test_")
	p.Set("debug", true)

	// 创建配置对象
	cfg := DefaultConfig()

	// 测试LoadFromProvider方法
	err := cfg.LoadFromProvider(p)
	require.NoError(t, err)

	// 验证配置是否正确加载
	assert.Equal(t, "mysql", cfg.Driver)
	assert.Equal(t, "user:pass@tcp(localhost:3306)/dbname", cfg.DSN)
	assert.Equal(t, 50, cfg.MaxOpenConns)
	assert.Equal(t, 10, cfg.MaxIdleConns)
	assert.Equal(t, "test_", cfg.TablePrefix)
	assert.Equal(t, true, cfg.Debug)
}

// 测试LoadFromProvider方法的错误处理
func TestConfig_LoadFromProvider_Error(t *testing.T) {
	// 创建模拟错误配置提供器
	p := newErrorMockProvider()

	// 创建配置对象
	cfg := DefaultConfig()

	// 测试LoadFromProvider方法返回错误
	err := cfg.LoadFromProvider(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal database config failed")
}

// 测试NewFromProvider方法
func TestNewFromProvider(t *testing.T) {
	// 测试直接加载
	t.Run("direct_load", func(t *testing.T) {
		// 创建模拟配置提供器
		p := newMockProvider()
		p.Set("driver", "mysql")
		p.Set("dsn", "user:pass@tcp(localhost:3306)/testdb")
		p.Set("max_open_conns", 20)
		p.Set("table_prefix", "prefix_")
		p.Set("debug", true)

		// 使用NewFromProvider创建数据库存储实例
		store, err := NewFromProvider(p, "")

		// 由于无法连接到真实数据库，这里会返回错误
		// 但我们仍然可以验证配置是否正确加载
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to database")

		// 在真实环境中，这将成功创建一个store实例
		if store != nil {
			assert.Equal(t, "mysql", store.config.Driver)
			assert.Equal(t, "user:pass@tcp(localhost:3306)/testdb", store.config.DSN)
			assert.Equal(t, 20, store.config.MaxOpenConns)
			assert.Equal(t, "prefix_", store.config.TablePrefix)
			assert.Equal(t, true, store.config.Debug)
		}
	})

	// 测试使用配置路径加载
	t.Run("with_config_path", func(t *testing.T) {
		// 创建模拟配置提供器
		p := newMockProvider()

		// 使用配置路径
		dbConfig := map[string]interface{}{
			"driver":         "mysql",
			"dsn":            "user:pass@tcp(localhost:3306)/pathdb",
			"max_open_conns": 30,
			"table_prefix":   "path_",
			"debug":          true,
		}
		p.Set("database", dbConfig)

		// 使用NewFromProvider创建数据库存储实例，指定配置路径
		store, err := NewFromProvider(p, "database")

		// 由于无法连接到真实数据库，这里会返回错误
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to database")

		// 验证配置是否正确加载
		if store != nil {
			assert.Equal(t, "mysql", store.config.Driver)
			assert.Equal(t, "user:pass@tcp(localhost:3306)/pathdb", store.config.DSN)
			assert.Equal(t, 30, store.config.MaxOpenConns)
			assert.Equal(t, "path_", store.config.TablePrefix)
			assert.Equal(t, true, store.config.Debug)
		}
	})
}

// 测试NewFromProvider方法的错误处理
func TestNewFromProvider_Error(t *testing.T) {
	// 测试直接解析错误
	t.Run("unmarshal_error", func(t *testing.T) {
		// 创建模拟错误配置提供器
		p := newErrorMockProvider()

		// 使用NewFromProvider创建数据库存储实例，预期失败
		_, err := NewFromProvider(p, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal database config")
	})

	// 测试使用配置路径解析错误
	t.Run("unmarshal_key_error", func(t *testing.T) {
		// 创建模拟错误配置提供器
		p := newErrorMockProvider()

		// 使用NewFromProvider创建数据库存储实例，指定配置路径，预期失败
		_, err := NewFromProvider(p, "database")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal database config")
	})
}

func setupTestDB(t *testing.T) (*Store, Repository) {
	// 使用 MySQL 测试数据库
	config := Config{
		Driver:        "mysql",
		DSN:           "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
		Debug:         true,
		SingularTable: true,
		TablePrefix:   "test_",
	}

	store, err := New(config)
	require.NoError(t, err)

	// 删除表（如果存在）
	store.DB().Exec("DROP TABLE IF EXISTS test_test_user")

	// 自动迁移测试模型
	err = store.Migrate(&TestUser{})
	require.NoError(t, err)

	// 创建仓储
	repo := NewRepository(store)
	require.NotNil(t, repo)

	return store, repo
}

func TestRepository_Create(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "测试用户",
		Email:  "test@example.com",
		Age:    25,
		Active: true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotZero(t, user.ID, "用户ID应该已自动生成")
}

func TestRepository_FindByID(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "测试用户",
		Email:  "test@example.com",
		Age:    25,
		Active: true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	userID := user.ID

	// 查找用户
	foundUser := &TestUser{}
	err = repo.FindByID(ctx, userID, foundUser)
	require.NoError(t, err)
	assert.Equal(t, userID, foundUser.ID)
	assert.Equal(t, "测试用户", foundUser.Name)
	assert.Equal(t, "test@example.com", foundUser.Email)
	assert.Equal(t, 25, foundUser.Age)
	assert.True(t, foundUser.Active)
}

func TestRepository_FindOneBy(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "测试用户",
		Email:  "test@example.com",
		Age:    25,
		Active: true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// 通过条件查找用户
	foundUser := &TestUser{}
	err = repo.FindOneBy(ctx, "email = ?", []interface{}{"test@example.com"}, foundUser)
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, "测试用户", foundUser.Name)
}

func TestRepository_FindAllBy(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建多个用户
	user1 := &TestUser{Name: "用户1", Email: "user1@example.com", Age: 20, Active: true}
	user2 := &TestUser{Name: "用户2", Email: "user2@example.com", Age: 25, Active: true}
	// 明确创建非活跃用户
	user3 := &TestUser{Name: "用户3", Email: "user3@example.com", Age: 30, Active: false}

	err := repo.Create(ctx, user1)
	require.NoError(t, err)
	err = repo.Create(ctx, user2)
	require.NoError(t, err)
	err = repo.Create(ctx, user3)
	require.NoError(t, err)

	// 确保user3是非活跃状态（直接更新数据库）
	err = store.DB().Exec("UPDATE test_test_user SET active = ? WHERE email = ?", false, "user3@example.com").Error
	require.NoError(t, err)

	// 验证user3是否确实设置为非活跃
	var checkUser TestUser
	err = repo.FindOneBy(ctx, "email = ?", []interface{}{"user3@example.com"}, &checkUser)
	require.NoError(t, err)
	require.False(t, checkUser.Active, "user3应该是非活跃状态")

	// 查询所有活跃用户
	var activeUsers []TestUser
	err = repo.FindAllBy(ctx, "active = ?", []interface{}{true}, &activeUsers)
	require.NoError(t, err)
	assert.Len(t, activeUsers, 2, "应该有2个活跃用户")

	// 查询所有用户
	var allUsers []TestUser
	err = repo.FindAllBy(ctx, nil, nil, &allUsers)
	require.NoError(t, err)
	assert.Len(t, allUsers, 3, "应该有3个用户")
}

func TestRepository_Update(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "测试用户",
		Email:  "test@example.com",
		Age:    25,
		Active: true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// 更新用户
	user.Name = "已更新的用户"
	user.Age = 30
	err = repo.Update(ctx, user)
	require.NoError(t, err)

	// 验证更新结果
	updatedUser := &TestUser{}
	err = repo.FindByID(ctx, user.ID, updatedUser)
	require.NoError(t, err)
	assert.Equal(t, "已更新的用户", updatedUser.Name)
	assert.Equal(t, 30, updatedUser.Age)
}

func TestRepository_UpdateBy(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "测试用户",
		Email:  "test@example.com",
		Age:    25,
		Active: true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// 通过条件更新
	updates := map[string]interface{}{
		"name":   "条件更新的用户",
		"active": false,
	}
	err = repo.UpdateBy(ctx, &TestUser{}, updates, "email = ?", []interface{}{"test@example.com"})
	require.NoError(t, err)

	// 验证更新结果
	updatedUser := &TestUser{}
	err = repo.FindByID(ctx, user.ID, updatedUser)
	require.NoError(t, err)
	assert.Equal(t, "条件更新的用户", updatedUser.Name)
	assert.False(t, updatedUser.Active)
}

func TestRepository_Delete(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "测试用户",
		Email:  "test@example.com",
		Age:    25,
		Active: true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// 删除用户
	err = repo.Delete(ctx, user)
	require.NoError(t, err)

	// 验证删除结果
	deletedUser := &TestUser{}
	err = repo.FindByID(ctx, user.ID, deletedUser)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRecordNotFound)
}

func TestRepository_DeleteByID(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "测试用户",
		Email:  "test@example.com",
		Age:    25,
		Active: true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	userID := user.ID

	// 通过ID删除
	err = repo.DeleteByID(ctx, &TestUser{}, userID)
	require.NoError(t, err)

	// 验证删除结果
	deletedUser := &TestUser{}
	err = repo.FindByID(ctx, userID, deletedUser)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRecordNotFound)
}

func TestRepository_DeleteBy(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建多个用户
	user1 := &TestUser{Name: "用户1", Email: "user1@example.com", Age: 20, Active: true}
	user2 := &TestUser{Name: "用户2", Email: "user2@example.com", Age: 25, Active: true}
	// 明确创建非活跃用户
	user3 := &TestUser{Name: "用户3", Email: "user3@example.com", Age: 30, Active: false}

	err := repo.Create(ctx, user1)
	require.NoError(t, err)
	err = repo.Create(ctx, user2)
	require.NoError(t, err)
	err = repo.Create(ctx, user3)
	require.NoError(t, err)

	// 确保user3是非活跃状态（直接更新数据库）
	err = store.DB().Exec("UPDATE test_test_user SET active = ? WHERE email = ?", false, "user3@example.com").Error
	require.NoError(t, err)

	// 验证user3是否确实设置为非活跃
	var checkUser TestUser
	err = repo.FindOneBy(ctx, "email = ?", []interface{}{"user3@example.com"}, &checkUser)
	require.NoError(t, err)
	require.False(t, checkUser.Active, "user3应该是非活跃状态")

	// 通过条件删除非活跃用户
	err = repo.DeleteBy(ctx, &TestUser{}, "active = ?", []interface{}{false})
	require.NoError(t, err)

	// 验证删除结果
	var remainingUsers []TestUser
	err = repo.FindAllBy(ctx, nil, nil, &remainingUsers)
	require.NoError(t, err)
	assert.Len(t, remainingUsers, 2, "删除不活跃用户后应该还有2个用户")
}

func TestRepository_Count(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建多个用户
	user1 := &TestUser{Name: "用户1", Email: "user1@example.com", Age: 20, Active: true}
	user2 := &TestUser{Name: "用户2", Email: "user2@example.com", Age: 25, Active: true}
	// 明确创建非活跃用户
	user3 := &TestUser{Name: "用户3", Email: "user3@example.com", Age: 30, Active: false}

	err := repo.Create(ctx, user1)
	require.NoError(t, err)
	err = repo.Create(ctx, user2)
	require.NoError(t, err)
	err = repo.Create(ctx, user3)
	require.NoError(t, err)

	// 确保user3是非活跃状态（直接更新数据库）
	err = store.DB().Exec("UPDATE test_test_user SET active = ? WHERE email = ?", false, "user3@example.com").Error
	require.NoError(t, err)

	// 验证user3是否确实设置为非活跃
	var checkUser TestUser
	err = repo.FindOneBy(ctx, "email = ?", []interface{}{"user3@example.com"}, &checkUser)
	require.NoError(t, err)
	require.False(t, checkUser.Active, "user3应该是非活跃状态")

	// 统计活跃用户数量
	count, err := repo.Count(ctx, &TestUser{}, "active = ?", []interface{}{true})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count, "应该有2个活跃用户")

	// 统计所有用户数量
	totalCount, err := repo.Count(ctx, &TestUser{}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(3), totalCount, "应该有3个用户")
}

func TestRepository_Exists(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "测试用户",
		Email:  "test@example.com",
		Age:    25,
		Active: true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// 检查存在性
	exists, err := repo.Exists(ctx, &TestUser{}, "email = ?", []interface{}{"test@example.com"})
	require.NoError(t, err)
	assert.True(t, exists)

	// 检查不存在的记录
	exists, err = repo.Exists(ctx, &TestUser{}, "email = ?", []interface{}{"nonexistent@example.com"})
	require.NoError(t, err)
	assert.False(t, exists)
}

// ====================== repository_more_test.go 测试代码 ======================

// 测试Repository的CreateInBatches方法
func TestRepository_CreateInBatches(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 删除表（如果存在）并重新创建
	err := store.DB().Exec("DROP TABLE IF EXISTS test_batch_test_user").Error
	require.NoError(t, err)

	// 创建测试表
	err = store.DB().AutoMigrate(&BatchTestUser{})
	require.NoError(t, err)

	// 生成唯一邮箱
	timestamp := time.Now().UnixNano()
	timestampStr := strconv.FormatInt(timestamp, 10)

	// 准备批量创建的数据
	users := []BatchTestUser{
		{Name: "批量用户1", Email: "batch1_" + timestampStr + "@example.com", Age: 21},
		{Name: "批量用户2", Email: "batch2_" + timestampStr + "@example.com", Age: 22},
		{Name: "批量用户3", Email: "batch3_" + timestampStr + "@example.com", Age: 23},
		{Name: "批量用户4", Email: "batch4_" + timestampStr + "@example.com", Age: 24},
		{Name: "批量用户5", Email: "batch5_" + timestampStr + "@example.com", Age: 25},
	}

	// 测试批量创建
	err = repo.CreateInBatches(ctx, &users, 2)
	require.NoError(t, err)

	// 验证数据是否创建成功
	var count int64
	err = store.DB().Model(&BatchTestUser{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)

	// 验证第一条和最后一条记录
	var firstUser, lastUser BatchTestUser
	err = store.DB().First(&firstUser).Error
	require.NoError(t, err)
	assert.Equal(t, "批量用户1", firstUser.Name)

	err = store.DB().Last(&lastUser).Error
	require.NoError(t, err)
	assert.Equal(t, "批量用户5", lastUser.Name)
}

// 测试Repository的Save方法
func TestRepository_Save(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 删除表（如果存在）并重新创建
	err := store.DB().Exec("DROP TABLE IF EXISTS test_save_test_user").Error
	require.NoError(t, err)

	// 创建测试表
	err = store.DB().AutoMigrate(&SaveTestUser{})
	require.NoError(t, err)

	// 生成唯一邮箱
	timestamp := time.Now().UnixNano()
	timestampStr := strconv.FormatInt(timestamp, 10)

	// 测试创建新记录
	t.Run("create_new", func(t *testing.T) {
		user := &SaveTestUser{
			Name:  "新用户",
			Email: "new_" + timestampStr + "@example.com",
			Age:   30,
		}

		// 使用Save创建
		err := repo.Save(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID, "ID应该在保存后被设置")

		// 验证记录是否创建成功
		var savedUser SaveTestUser
		err = store.DB().First(&savedUser, user.ID).Error
		require.NoError(t, err)
		assert.Equal(t, user.Name, savedUser.Name)
		assert.Equal(t, user.Email, savedUser.Email)
		assert.Equal(t, user.Age, savedUser.Age)
	})

	// 测试更新现有记录
	t.Run("update_existing", func(t *testing.T) {
		// 先创建一条记录
		user := &SaveTestUser{
			Name:  "待更新用户",
			Email: "update_" + timestampStr + "@example.com",
			Age:   25,
		}
		err := store.DB().Create(user).Error
		require.NoError(t, err)

		// 修改记录
		user.Name = "已更新用户"
		user.Age = 26

		// 使用Save更新
		err = repo.Save(ctx, user)
		require.NoError(t, err)

		// 验证记录是否更新成功
		var updatedUser SaveTestUser
		err = store.DB().First(&updatedUser, user.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "已更新用户", updatedUser.Name)
		assert.Equal(t, 26, updatedUser.Age)
		assert.Equal(t, user.Email, updatedUser.Email)
	})
}

// 测试Repository的异常情况
func TestRepository_EdgeCases(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 测试使用不存在的表名创建记录
	t.Run("create_with_non_existent_table", func(t *testing.T) {
		// 定义一个临时结构体，对应数据库中不存在的表
		type NonExistentTable struct {
			ID   uint   `gorm:"primarykey"`
			Name string `gorm:"size:100"`
		}

		// 尝试创建记录
		obj := &NonExistentTable{Name: "测试"}
		err := repo.Create(ctx, obj)
		assert.Error(t, err, "应该返回错误，因为表不存在")
	})

	// 测试使用空切片进行批量创建
	t.Run("create_in_batches_with_empty_slice", func(t *testing.T) {
		var emptyUsers []BatchTestUser
		err := repo.CreateInBatches(ctx, &emptyUsers, 10)
		// GORM处理空切片时通常不会返回错误，但不会执行实际操作
		assert.NoError(t, err)
	})

	// 测试使用无效参数保存
	t.Run("save_with_invalid_params", func(t *testing.T) {
		// 因为直接传递nil会导致panic，我们应该在Repository的Save方法中添加检查
		// 这里测试传递空结构体
		err := repo.Save(ctx, &struct{}{})
		assert.Error(t, err, "应该返回错误，因为参数无效")
	})
}

// 测试FindAllBy方法的错误情况
func TestRepository_FindAllBy_Errors(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 测试传递非指针类型
	t.Run("non_pointer_result", func(t *testing.T) {
		var users []TestUser
		err := repo.FindAllBy(ctx, "active = ?", []interface{}{true}, users)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "result must be a pointer to slice")
	})

	// 测试传递非切片指针
	t.Run("non_slice_pointer", func(t *testing.T) {
		var user TestUser
		err := repo.FindAllBy(ctx, "active = ?", []interface{}{true}, &user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "result must be a pointer to slice")
	})

	// 测试传递nil条件
	t.Run("nil_condition", func(t *testing.T) {
		var users []TestUser
		err := repo.FindAllBy(ctx, nil, nil, &users)
		assert.NoError(t, err)
	})
}

// ====================== transaction_test.go 测试代码 ======================

func TestStore_Transaction(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	// 测试成功的事务
	t.Run("成功的事务", func(t *testing.T) {
		err := store.Transaction(func(tx *gorm.DB) error {
			user1 := &TestUser{Name: "事务用户1", Email: "tx1@example.com", Age: 30}
			user2 := &TestUser{Name: "事务用户2", Email: "tx2@example.com", Age: 25}

			if err := tx.Create(user1).Error; err != nil {
				return err
			}

			if err := tx.Create(user2).Error; err != nil {
				return err
			}

			return nil
		})

		require.NoError(t, err)

		// 验证两个用户都被成功创建
		var count int64
		err = store.DB().Model(&TestUser{}).Where("email IN ?", []string{"tx1@example.com", "tx2@example.com"}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	// 测试失败回滚的事务
	t.Run("失败回滚的事务", func(t *testing.T) {
		// 这是一个将会失败的事务
		err := store.Transaction(func(tx *gorm.DB) error {
			user1 := &TestUser{Name: "回滚用户1", Email: "rollback1@example.com", Age: 30}
			user2 := &TestUser{Name: "回滚用户2", Email: "rollback2@example.com", Age: 25}

			if err := tx.Create(user1).Error; err != nil {
				return err
			}

			// 在创建第二个用户后故意返回错误，触发回滚
			if err := tx.Create(user2).Error; err != nil {
				return err
			}

			return errors.New("手动触发回滚")
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "手动触发回滚")

		// 验证两个用户都没有被创建（事务回滚）
		var count int64
		err = store.DB().Model(&TestUser{}).Where("email IN ?", []string{"rollback1@example.com", "rollback2@example.com"}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestStore_TransactionWithContext(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 测试带上下文的事务
	t.Run("带上下文的事务", func(t *testing.T) {
		err := store.TransactionWithContext(ctx, func(tx *gorm.DB) error {
			user := &TestUser{Name: "上下文事务用户", Email: "ctx_tx@example.com", Age: 28}
			return tx.Create(user).Error
		})

		require.NoError(t, err)

		// 验证用户被创建
		var user TestUser
		err = store.DB().Where("email = ?", "ctx_tx@example.com").First(&user).Error
		require.NoError(t, err)
		assert.Equal(t, "上下文事务用户", user.Name)
	})

	// 测试带上下文的事务回滚
	t.Run("带上下文的事务回滚", func(t *testing.T) {
		err := store.TransactionWithContext(ctx, func(tx *gorm.DB) error {
			user := &TestUser{Name: "上下文回滚用户", Email: "ctx_rollback@example.com", Age: 35}
			if err := tx.Create(user).Error; err != nil {
				return err
			}
			return errors.New("上下文事务手动回滚")
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "上下文事务手动回滚")

		// 验证用户没有被创建
		var count int64
		err = store.DB().Model(&TestUser{}).Where("email = ?", "ctx_rollback@example.com").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
