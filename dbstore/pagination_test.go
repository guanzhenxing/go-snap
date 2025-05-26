package dbstore

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginate(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建测试数据 - 前10个用户为活跃状态，后10个为非活跃状态
	for i := 0; i < 20; i++ {
		user := &TestUser{
			Name:   "用户" + strconv.Itoa(i),
			Email:  "user" + strconv.Itoa(i) + "@example.com",
			Age:    20 + i,
			Active: i < 10, // 前10个用户设置为活跃
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// 确保后10个用户为非活跃状态
		if i >= 10 {
			// 通过更新来确保非活跃状态正确设置
			updates := map[string]interface{}{
				"active": false,
			}
			err = repo.UpdateBy(ctx, &TestUser{}, updates, "email = ?", []interface{}{"user" + strconv.Itoa(i) + "@example.com"})
			require.NoError(t, err)
		}
	}

	// 验证活跃用户数量
	activeCount, err := repo.Count(ctx, &TestUser{}, "active = ?", []interface{}{true})
	require.NoError(t, err)
	assert.Equal(t, int64(10), activeCount, "应该有10个活跃用户")

	t.Run("基本分页", func(t *testing.T) {
		// 第一页，每页5条
		var users []TestUser
		paginator, err := repo.Paginate(ctx, 1, 5, &TestUser{}, &users, nil, nil)
		require.NoError(t, err)

		assert.Equal(t, int64(20), paginator.TotalCount)
		assert.Equal(t, 4, paginator.TotalPage)
		assert.Equal(t, 1, paginator.CurrentPage)
		assert.Equal(t, 5, paginator.PageSize)
		assert.Len(t, users, 5)
	})

	t.Run("带条件的分页", func(t *testing.T) {
		// 查询活跃用户
		var users []TestUser
		paginator, err := repo.Paginate(ctx, 1, 5, &TestUser{}, &users, "active = ?", []interface{}{true})
		require.NoError(t, err)

		assert.Equal(t, int64(10), paginator.TotalCount) // 应该有10个活跃用户
		assert.Equal(t, 2, paginator.TotalPage)
		assert.Equal(t, 1, paginator.CurrentPage)
		assert.Equal(t, 5, paginator.PageSize)
		assert.Len(t, users, 5)

		// 检查所有返回的用户是否都是活跃的
		for _, user := range users {
			assert.True(t, user.Active)
		}
	})

	t.Run("无效页码", func(t *testing.T) {
		// 页码小于1
		var users []TestUser
		paginator, err := repo.Paginate(ctx, 0, 5, &TestUser{}, &users, nil, nil)
		require.NoError(t, err)

		assert.Equal(t, int64(20), paginator.TotalCount)
		assert.Equal(t, 4, paginator.TotalPage)
		assert.Equal(t, 1, paginator.CurrentPage) // 应该自动调整为第1页
		assert.Equal(t, 5, paginator.PageSize)
		assert.Len(t, users, 5)

		// 页码大于总页数
		paginator, err = repo.Paginate(ctx, 10, 5, &TestUser{}, &users, nil, nil)
		require.NoError(t, err)

		assert.Equal(t, int64(20), paginator.TotalCount)
		assert.Equal(t, 4, paginator.TotalPage)
		assert.Equal(t, 4, paginator.CurrentPage) // 应该自动调整为最后一页
		assert.Equal(t, 5, paginator.PageSize)
		assert.Len(t, users, 5)
	})

	t.Run("无效页大小", func(t *testing.T) {
		// 每页大小小于1
		var users []TestUser
		paginator, err := repo.Paginate(ctx, 1, 0, &TestUser{}, &users, nil, nil)
		require.NoError(t, err)

		assert.Equal(t, int64(20), paginator.TotalCount)
		assert.Equal(t, 2, paginator.TotalPage)
		assert.Equal(t, 1, paginator.CurrentPage)
		assert.Equal(t, 10, paginator.PageSize) // 应该自动调整为默认的10
		assert.Len(t, users, 10)
	})

	t.Run("空结果", func(t *testing.T) {
		// 查询不存在的记录
		var users []TestUser
		paginator, err := repo.Paginate(ctx, 1, 5, &TestUser{}, &users, "age > ?", []interface{}{100})
		require.NoError(t, err)

		assert.Equal(t, int64(0), paginator.TotalCount)
		assert.Equal(t, 0, paginator.TotalPage)
		assert.Equal(t, 1, paginator.CurrentPage)
		assert.Equal(t, 5, paginator.PageSize)
		assert.Len(t, users, 0)
	})
}

func TestPaginateByQuery(t *testing.T) {
	store, repo := setupTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// 创建测试数据 - 前10个用户为活跃状态，后10个为非活跃状态
	for i := 0; i < 20; i++ {
		user := &TestUser{
			Name:   "用户" + strconv.Itoa(i),
			Email:  "user" + strconv.Itoa(i) + "@example.com",
			Age:    20 + i,
			Active: i < 10, // 前10个用户设置为活跃
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// 确保后10个用户为非活跃状态
		if i >= 10 {
			// 通过更新来确保非活跃状态正确设置
			updates := map[string]interface{}{
				"active": false,
			}
			err = repo.UpdateBy(ctx, &TestUser{}, updates, "email = ?", []interface{}{"user" + strconv.Itoa(i) + "@example.com"})
			require.NoError(t, err)
		}
	}

	// 验证活跃用户数量
	activeCount, err := repo.Count(ctx, &TestUser{}, "active = ?", []interface{}{true})
	require.NoError(t, err)
	assert.Equal(t, int64(10), activeCount, "应该有10个活跃用户")

	// 从仓储获取数据库
	db := store.WithContext(ctx).Model(&TestUser{})

	// 构建查询参数
	query := map[string]interface{}{
		"page":      2,
		"page_size": 5,
		"active":    true,
	}

	// 执行分页查询
	var users []TestUser
	paginator, err := PaginateByQuery(db, query, &users)
	require.NoError(t, err)

	// 验证结果
	assert.Equal(t, int64(10), paginator.TotalCount) // 应该有10个活跃用户
	assert.Equal(t, 2, paginator.TotalPage)
	assert.Equal(t, 2, paginator.CurrentPage)
	assert.Equal(t, 5, paginator.PageSize)
	assert.Len(t, users, 5)

	// 检查所有返回的用户是否都是活跃的
	for _, user := range users {
		assert.True(t, user.Active)
	}
}
