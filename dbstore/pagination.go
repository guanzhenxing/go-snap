// Package dbstore 提供了基于GORM的数据库操作封装，简化项目中的数据库操作
//
// # 分页机制设计
//
// 本文件实现了数据库查询的分页功能，支持标准的基于偏移量的分页方案。
// 分页系统的主要设计目标是：
//
// 1. 简化分页查询实现，减少重复代码
// 2. 提供标准化的分页结果格式，便于前端处理
// 3. 自动处理边界情况和参数验证
// 4. 与GORM查询构建器无缝集成
//
// # 分页性能考虑
//
// 基于偏移量的分页（OFFSET/LIMIT）在数据量大的情况下可能存在性能问题，
// 特别是当页码较大时，数据库需要先扫描跳过的记录。针对这种情况，有以下建议：
//
// 1. 对于小型数据集，当前实现的偏移分页完全够用
// 2. 对于大型数据集，考虑使用游标分页（基于上次查询的最后一个ID继续查询）
// 3. 在查询中添加合适的索引，尤其是排序字段
// 4. 考虑在Repository层添加游标分页的特定实现
//
// # 使用场景
//
// 分页功能主要用于以下场景：
//
// 1. 列表页面展示大量数据
// 2. API接口返回分批数据
// 3. 数据导出时分批处理
// 4. 无限滚动或"加载更多"功能
package dbstore

import (
	"math"

	"gorm.io/gorm"
)

// Paginator 分页结果结构体
// 包含分页查询的结果和元数据，用于向客户端返回分页信息
// 此结构体可直接序列化为JSON响应返回给前端
type Paginator struct {
	// TotalCount 查询条件匹配的总记录数
	// 用于计算总页数和显示记录总数
	TotalCount int64 `json:"total_count"`

	// TotalPage 总页数
	// 根据总记录数和每页大小计算得出
	// 前端可用于生成分页控件
	TotalPage int `json:"total_page"`

	// CurrentPage 当前页码
	// 从1开始计数，表示当前返回的是第几页数据
	CurrentPage int `json:"current_page"`

	// PageSize 每页记录数
	// 控制每页返回的最大记录数量
	PageSize int `json:"page_size"`

	// Data 分页数据
	// 当前页的实际数据，类型由调用者决定
	// 在JSON序列化时会包含具体的数据内容
	Data interface{} `json:"data"`
}

// Paginate 执行分页查询并返回分页结果
// 这是分页功能的核心方法，处理计数查询和数据查询
//
// 参数：
//   - db: 已配置查询条件的GORM DB实例
//   - page: 请求的页码，从1开始
//   - pageSize: 每页记录数
//   - result: 结果接收器，必须是指向切片的指针
//
// 返回：
//   - *Paginator: 包含分页元数据和查询结果的分页器
//   - error: 查询过程中遇到的错误，成功则为nil
//
// 工作流程：
//  1. 首先执行COUNT查询获取总记录数
//  2. 验证并调整页码和每页大小参数
//  3. 计算总页数和查询偏移量
//  4. 执行带LIMIT/OFFSET的查询获取当前页数据
//  5. 返回包含结果和元数据的Paginator对象
//
// 示例：
//
//	var users []User
//	// 准备查询，添加条件但不执行
//	query := db.Model(&User{}).Where("status = ?", "active")
//	// 执行分页查询
//	paginator, err := Paginate(query, 1, 10, &users)
//	if err != nil {
//	    return err
//	}
//	// 使用结果
//	fmt.Printf("总记录数: %d, 总页数: %d\n", paginator.TotalCount, paginator.TotalPage)
//	fmt.Printf("当前页数据: %v\n", users)
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

// PaginateByQuery 根据请求参数映射进行分页查询
// 便于直接使用API请求参数进行分页查询
//
// 参数：
//   - db: 已配置部分查询条件的GORM DB实例
//   - query: 包含分页参数和过滤条件的映射
//   - result: 结果接收器，必须是指向切片的指针
//
// 返回：
//   - *Paginator: 包含分页元数据和查询结果的分页器
//   - error: 查询过程中遇到的错误，成功则为nil
//
// 特性：
//   - 自动从query中提取"page"和"page_size"参数
//   - 将其他参数作为查询条件应用（简单的等值条件）
//   - 对于复杂查询条件，应在调用前通过db参数配置
//
// 示例：
//
//	var products []Product
//	// 从HTTP请求或其他源获取查询参数
//	queryParams := map[string]interface{}{
//	    "page": 2,
//	    "page_size": 20,
//	    "category": "electronics",
//	    "in_stock": true,
//	}
//	// 执行分页查询
//	paginator, err := PaginateByQuery(db.Model(&Product{}), queryParams, &products)
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
