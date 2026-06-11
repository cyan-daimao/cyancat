package api

// Page 统一分页结构
type Page[T any] struct {
	// Page 当前页码（从 1 开始）
	Page int `json:"page"`
	// PageSize 每页条数
	PageSize int `json:"pageSize"`
	// Total 总数
	Total int64 `json:"total"`
	// List 数据列表
	List []T `json:"list"`
}

// NewPage 构造分页结果
func NewPage[T any](list []T, total int64, page, pageSize int) *Page[T] {
	return &Page[T]{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		List:     list,
	}
}