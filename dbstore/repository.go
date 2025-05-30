// Package dbstore 提供了基于GORM的数据库操作封装，简化项目中的数据库操作
//
// # Repository模式设计
//
// Repository模式是数据访问层的设计模式，它的核心思想是将数据访问逻辑与业务逻辑分离，
// 提供一个面向对象的接口来操作领域对象，同时隐藏数据源的具体实现细节。
//
// 主要优势：
//
// 1. 关注点分离：业务逻辑专注于业务规则，不需要关心数据如何存储和获取
// 2. 可测试性：易于模拟Repository进行单元测试
// 3. 代码复用：通用数据访问操作被封装在基础Repository中
// 4. 维护性：数据访问逻辑集中在一处，易于维护和修改
// 5. 抽象数据源：可以轻松切换底层数据存储而不影响业务逻辑
//
// # 实现架构
//
// 本包提供了以下核心组件：
//
// 1. Repository接口：定义通用的CRUD操作方法
// 2. BaseRepository：实现Repository接口的基础实现
// 3. 错误常量：标准化错误处理
// 4. 分页工具：简化分页查询
//
// 开发者可以直接使用BaseRepository，也可以通过嵌入或组合创建特定领域的Repository，
// 添加特定业务逻辑的方法。
//
// # 代码组织最佳实践
//
// 推荐按以下方式组织Repository相关代码：
//
// 1. 每个领域实体定义自己的Repository接口，继承基础Repository接口
// 2. 实现特定于该领域的Repository实现类
// 3. 使用依赖注入方式提供Repository实例
//
// 示例：
//
//	// 用户仓储接口
//	type UserRepository interface {
//	    Repository
//	    FindByEmail(ctx context.Context, email string) (*User, error)
//	    FindActiveUsers(ctx context.Context) ([]*User, error)
//	}
//
//	// 用户仓储实现
//	type userRepository struct {
//	    BaseRepository
//	}
//
//	func NewUserRepository(store *Store) UserRepository {
//	    return &userRepository{
//	        BaseRepository: BaseRepository{
//	            store: store,
//	        },
//	    }
//	}
//
//	func (r *userRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
//	    var user User
//	    err := r.FindOneBy(ctx, "email = ?", []interface{}{email}, &user)
//	    return &user, err
//	}
//
// # 事务处理
//
// Repository可以在事务中使用：
//
//	err := store.Transaction(func(tx *gorm.DB) error {
//	    // 创建事务级Repository
//	    txRepo := NewTransactionalRepository(tx)
//
//	    // 使用事务级Repository执行操作
//	    if err := txRepo.Create(ctx, user); err != nil {
//	        return err
//	    }
//
//	    if err := txRepo.Create(ctx, profile); err != nil {
//	        return err
//	    }
//
//	    return nil
//	})
package dbstore

import (
	"context"
	"errors"
	"reflect"

	"gorm.io/gorm"
)

var (
	// ErrRecordNotFound 记录未找到错误
	// 当查询没有找到任何匹配记录时返回此错误
	// 应用层可以通过errors.Is(err, ErrRecordNotFound)检查此错误
	ErrRecordNotFound = gorm.ErrRecordNotFound

	// ErrEmptySlice 空切片错误
	// 当尝试批量操作一个空切片时返回此错误
	// 用于避免不必要的数据库调用
	ErrEmptySlice = errors.New("empty slice")
)

// Repository 通用仓储接口，定义基本的CRUD操作
// 所有实体特定的仓储都应该实现此接口或嵌入此接口
// 提供对数据库表的基本操作抽象，隐藏底层实现细节
type Repository interface {
	// Create 创建记录
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - model: 要创建的模型实例，必须是指针类型
	// 返回：
	//   - error: 创建过程中遇到的错误，成功则为nil
	// 示例：
	//   err := repo.Create(ctx, &User{Name: "张三", Age: 30})
	Create(ctx context.Context, model interface{}) error

	// CreateInBatches 批量创建记录
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - models: 要创建的模型切片，必须是指针类型的切片
	//   - batchSize: 批处理大小，控制每批插入的记录数
	// 返回：
	//   - error: 创建过程中遇到的错误，成功则为nil
	// 注意：
	//   - batchSize建议设置为100-1000之间，太小影响性能，太大可能导致数据库压力过大
	// 示例：
	//   users := []*User{{Name: "张三"}, {Name: "李四"}}
	//   err := repo.CreateInBatches(ctx, users, 100)
	CreateInBatches(ctx context.Context, models interface{}, batchSize int) error

	// Save 保存记录（如果主键存在则更新，否则创建）
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - model: 要保存的模型实例，必须是指针类型
	// 返回：
	//   - error: 保存过程中遇到的错误，成功则为nil
	// 注意：
	//   - Save会更新所有字段，包括零值字段，如果只想更新特定字段，请使用Update方法
	// 示例：
	//   user.Name = "新名称"
	//   err := repo.Save(ctx, user)
	Save(ctx context.Context, model interface{}) error

	// FindByID 通过ID查找记录
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - id: 记录ID，可以是整数或字符串
	//   - result: 查询结果接收器，必须是指针类型
	// 返回：
	//   - error: 查询过程中遇到的错误，未找到记录会返回ErrRecordNotFound
	// 示例：
	//   var user User
	//   err := repo.FindByID(ctx, 123, &user)
	FindByID(ctx context.Context, id interface{}, result interface{}) error

	// FindOneBy 通过条件查找单个记录
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - query: 查询条件，可以是字符串或gorm.Expr
	//   - args: 查询参数，对应query中的占位符
	//   - result: 查询结果接收器，必须是指针类型
	// 返回：
	//   - error: 查询过程中遇到的错误，未找到记录会返回ErrRecordNotFound
	// 示例：
	//   var user User
	//   err := repo.FindOneBy(ctx, "email = ?", []interface{}{"test@example.com"}, &user)
	FindOneBy(ctx context.Context, query interface{}, args []interface{}, result interface{}) error

	// FindAllBy 通过条件查找所有记录
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - query: 查询条件，可以是字符串或gorm.Expr，传nil表示查询所有记录
	//   - args: 查询参数，对应query中的占位符
	//   - result: 查询结果接收器，必须是指向切片的指针
	// 返回：
	//   - error: 查询过程中遇到的错误，未找到记录不会返回错误而是空切片
	// 示例：
	//   var users []User
	//   err := repo.FindAllBy(ctx, "age > ?", []interface{}{18}, &users)
	FindAllBy(ctx context.Context, query interface{}, args []interface{}, result interface{}) error

	// DeleteByID 通过ID删除记录
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - model: 模型类型，用于确定表名
	//   - id: 要删除的记录ID
	// 返回：
	//   - error: 删除过程中遇到的错误，成功则为nil
	// 注意：
	//   - 此方法执行物理删除，如需逻辑删除请使用UpdateBy更新删除标记
	// 示例：
	//   err := repo.DeleteByID(ctx, &User{}, 123)
	DeleteByID(ctx context.Context, model interface{}, id interface{}) error

	// Delete 删除记录
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - model: 要删除的模型实例，必须包含主键值
	// 返回：
	//   - error: 删除过程中遇到的错误，成功则为nil
	// 示例：
	//   err := repo.Delete(ctx, user) // user必须有ID值
	Delete(ctx context.Context, model interface{}) error

	// DeleteBy 通过条件删除记录
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - model: 模型类型，用于确定表名
	//   - query: 删除条件，不能为nil
	//   - args: 删除参数，对应query中的占位符
	// 返回：
	//   - error: 删除过程中遇到的错误，成功则为nil
	// 警告：
	//   - 此方法可能会删除多条记录，使用时需谨慎
	// 示例：
	//   err := repo.DeleteBy(ctx, &User{}, "created_at < ?", []interface{}{lastWeek})
	DeleteBy(ctx context.Context, model interface{}, query interface{}, args []interface{}) error

	// Update 更新记录
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - model: 要更新的模型实例，必须包含主键值
	// 返回：
	//   - error: 更新过程中遇到的错误，成功则为nil
	// 注意：
	//   - 此方法只更新非零值字段
	// 示例：
	//   user.Name = "新名称"
	//   err := repo.Update(ctx, user) // 只会更新Name字段
	Update(ctx context.Context, model interface{}) error

	// UpdateBy 通过条件更新记录的指定字段
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - model: 模型类型，用于确定表名
	//   - values: 要更新的字段和值的映射
	//   - query: 更新条件，不能为nil
	//   - args: 更新参数，对应query中的占位符
	// 返回：
	//   - error: 更新过程中遇到的错误，成功则为nil
	// 警告：
	//   - 此方法可能会更新多条记录，使用时需谨慎
	// 示例：
	//   values := map[string]interface{}{"status": "inactive", "updated_at": time.Now()}
	//   err := repo.UpdateBy(ctx, &User{}, values, "last_login < ?", []interface{}{sixMonthsAgo})
	UpdateBy(ctx context.Context, model interface{}, values map[string]interface{}, query interface{}, args []interface{}) error

	// Count 统计记录数
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - model: 模型类型，用于确定表名
	//   - query: 统计条件，可为nil表示统计所有记录
	//   - args: 统计参数，对应query中的占位符
	// 返回：
	//   - int64: 符合条件的记录数
	//   - error: 统计过程中遇到的错误，成功则为nil
	// 示例：
	//   count, err := repo.Count(ctx, &User{}, "status = ?", []interface{}{"active"})
	Count(ctx context.Context, model interface{}, query interface{}, args []interface{}) (int64, error)

	// Exists 检查记录是否存在
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - model: 模型类型，用于确定表名
	//   - query: 检查条件，不能为nil
	//   - args: 检查参数，对应query中的占位符
	// 返回：
	//   - bool: 记录是否存在
	//   - error: 检查过程中遇到的错误，成功则为nil
	// 示例：
	//   exists, err := repo.Exists(ctx, &User{}, "email = ?", []interface{}{"test@example.com"})
	Exists(ctx context.Context, model interface{}, query interface{}, args []interface{}) (bool, error)

	// Paginate 分页查询
	// 参数：
	//   - ctx: 上下文，用于传递请求级信息和控制超时
	//   - page: 页码，从1开始
	//   - pageSize: 每页记录数
	//   - model: 模型类型，用于确定表名
	//   - result: 查询结果接收器，必须是指向切片的指针
	//   - query: 查询条件，可为nil表示查询所有记录
	//   - args: 查询参数，对应query中的占位符
	// 返回：
	//   - *Paginator: 分页信息，包含总记录数、总页数等
	//   - error: 查询过程中遇到的错误，成功则为nil
	// 示例：
	//   var users []User
	//   paginator, err := repo.Paginate(ctx, 1, 10, &User{}, &users, "status = ?", []interface{}{"active"})
	//   fmt.Printf("总记录数：%d, 总页数：%d\n", paginator.Total, paginator.TotalPages)
	Paginate(ctx context.Context, page, pageSize int, model interface{}, result interface{}, query interface{}, args []interface{}) (*Paginator, error)
}

// BaseRepository 基础仓储实现
type BaseRepository struct {
	store *Store
}

// NewRepository 创建仓储实例
func NewRepository(store *Store) Repository {
	return &BaseRepository{
		store: store,
	}
}

// Create 创建记录
func (r *BaseRepository) Create(ctx context.Context, model interface{}) error {
	return r.store.WithContext(ctx).Create(model).Error
}

// CreateInBatches 批量创建记录
func (r *BaseRepository) CreateInBatches(ctx context.Context, models interface{}, batchSize int) error {
	return r.store.WithContext(ctx).CreateInBatches(models, batchSize).Error
}

// Save 保存记录
func (r *BaseRepository) Save(ctx context.Context, model interface{}) error {
	return r.store.WithContext(ctx).Save(model).Error
}

// FindByID 通过ID查找记录
func (r *BaseRepository) FindByID(ctx context.Context, id interface{}, result interface{}) error {
	return r.store.WithContext(ctx).First(result, id).Error
}

// FindOneBy 通过条件查找单个记录
func (r *BaseRepository) FindOneBy(ctx context.Context, query interface{}, args []interface{}, result interface{}) error {
	return r.store.WithContext(ctx).Where(query, args...).First(result).Error
}

// FindAllBy 通过条件查找所有记录
func (r *BaseRepository) FindAllBy(ctx context.Context, query interface{}, args []interface{}, result interface{}) error {
	if reflect.TypeOf(result).Kind() != reflect.Ptr || reflect.TypeOf(result).Elem().Kind() != reflect.Slice {
		return errors.New("result must be a pointer to slice")
	}
	if query == nil {
		return r.store.WithContext(ctx).Find(result).Error
	}
	return r.store.WithContext(ctx).Where(query, args...).Find(result).Error
}

// DeleteByID 通过ID删除记录
func (r *BaseRepository) DeleteByID(ctx context.Context, model interface{}, id interface{}) error {
	return r.store.WithContext(ctx).Delete(model, id).Error
}

// Delete 删除记录
func (r *BaseRepository) Delete(ctx context.Context, model interface{}) error {
	return r.store.WithContext(ctx).Delete(model).Error
}

// DeleteBy 通过条件删除记录
func (r *BaseRepository) DeleteBy(ctx context.Context, model interface{}, query interface{}, args []interface{}) error {
	return r.store.WithContext(ctx).Where(query, args...).Delete(model).Error
}

// Update 更新记录
func (r *BaseRepository) Update(ctx context.Context, model interface{}) error {
	return r.store.WithContext(ctx).Updates(model).Error
}

// UpdateBy 通过条件更新记录的指定字段
func (r *BaseRepository) UpdateBy(ctx context.Context, model interface{}, values map[string]interface{}, query interface{}, args []interface{}) error {
	return r.store.WithContext(ctx).Model(model).Where(query, args...).Updates(values).Error
}

// Count 统计记录数
func (r *BaseRepository) Count(ctx context.Context, model interface{}, query interface{}, args []interface{}) (int64, error) {
	var count int64
	db := r.store.WithContext(ctx).Model(model)
	if query != nil {
		db = db.Where(query, args...)
	}
	err := db.Count(&count).Error
	return count, err
}

// Exists 检查记录是否存在
func (r *BaseRepository) Exists(ctx context.Context, model interface{}, query interface{}, args []interface{}) (bool, error) {
	count, err := r.Count(ctx, model, query, args)
	return count > 0, err
}

// Paginate 分页查询
func (r *BaseRepository) Paginate(ctx context.Context, page, pageSize int, model interface{}, result interface{}, query interface{}, args []interface{}) (*Paginator, error) {
	db := r.store.WithContext(ctx).Model(model)
	if query != nil {
		db = db.Where(query, args...)
	}
	return Paginate(db, page, pageSize, result)
}
