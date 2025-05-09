package dbstore

import (
	"math"

	"gorm.io/gorm"
)

// Paginator 分页结果
type Paginator struct {
	TotalCount  int64       `json:"total_count"`  // 总记录数
	TotalPage   int         `json:"total_page"`   // 总页数
	CurrentPage int         `json:"current_page"` // 当前页
	PageSize    int         `json:"page_size"`    // 每页大小
	Data        interface{} `json:"data"`         // 数据
}

// Paginate 分页查询
// 示例：
//
//	var users []User
//	paginator, err := Paginate(db.Model(&User{}), 1, 10, &users)
func Paginate(db *gorm.DB, page, pageSize int, result interface{}) (*Paginator, error) {
	var totalCount int64
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// 防止页码和每页大小无效
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// 计算总页数
	totalPage := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	// 没有数据时直接返回空结果
	if totalCount == 0 {
		return &Paginator{
			TotalCount:  0,
			TotalPage:   0,
			CurrentPage: page,
			PageSize:    pageSize,
			Data:        result,
		}, nil
	}

	// 确保请求的页码有效
	if page > totalPage {
		page = totalPage
	}

	// 执行分页查询
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(result).Error; err != nil {
		return nil, err
	}

	return &Paginator{
		TotalCount:  totalCount,
		TotalPage:   totalPage,
		CurrentPage: page,
		PageSize:    pageSize,
		Data:        result,
	}, nil
}

// PaginateByQuery 根据请求参数进行分页查询
func PaginateByQuery(db *gorm.DB, query map[string]interface{}, result interface{}) (*Paginator, error) {
	page := 1
	pageSize := 10

	// 从查询参数中获取页码和每页大小
	if p, ok := query["page"]; ok {
		if pageNum, ok := p.(int); ok && pageNum > 0 {
			page = pageNum
		}
	}

	if ps, ok := query["page_size"]; ok {
		if size, ok := ps.(int); ok && size > 0 {
			pageSize = size
		}
	}

	// 应用查询条件
	for k, v := range query {
		if k != "page" && k != "page_size" && v != nil {
			db = db.Where(k+" = ?", v)
		}
	}

	return Paginate(db, page, pageSize, result)
}
