package dbstore

import (
	"context"
	"errors"
	"reflect"

	"gorm.io/gorm"
)

var (
	// ErrRecordNotFound 记录未找到错误
	ErrRecordNotFound = gorm.ErrRecordNotFound
	// ErrEmptySlice 空切片错误
	ErrEmptySlice = errors.New("empty slice")
)

// Repository 通用仓储接口，定义基本的CRUD操作
type Repository interface {
	// Create 创建记录
	Create(ctx context.Context, model interface{}) error
	// CreateInBatches 批量创建记录
	CreateInBatches(ctx context.Context, models interface{}, batchSize int) error
	// Save 保存记录（如果主键存在则更新，否则创建）
	Save(ctx context.Context, model interface{}) error
	// FindByID 通过ID查找记录
	FindByID(ctx context.Context, id interface{}, result interface{}) error
	// FindOneBy 通过条件查找单个记录
	FindOneBy(ctx context.Context, query interface{}, args []interface{}, result interface{}) error
	// FindAllBy 通过条件查找所有记录
	FindAllBy(ctx context.Context, query interface{}, args []interface{}, result interface{}) error
	// DeleteByID 通过ID删除记录
	DeleteByID(ctx context.Context, model interface{}, id interface{}) error
	// Delete 删除记录
	Delete(ctx context.Context, model interface{}) error
	// DeleteBy 通过条件删除记录
	DeleteBy(ctx context.Context, model interface{}, query interface{}, args []interface{}) error
	// Update 更新记录
	Update(ctx context.Context, model interface{}) error
	// UpdateBy 通过条件更新记录的指定字段
	UpdateBy(ctx context.Context, model interface{}, values map[string]interface{}, query interface{}, args []interface{}) error
	// Count 统计记录数
	Count(ctx context.Context, model interface{}, query interface{}, args []interface{}) (int64, error)
	// Exists 检查记录是否存在
	Exists(ctx context.Context, model interface{}, query interface{}, args []interface{}) (bool, error)
	// Paginate 分页查询
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
