package connectionservice

// ListConnectionQuery 列表查询条件
type ListConnectionQuery struct {
	// Group 按分组过滤
	Group string
	// Type 按驱动类型过滤
	Type string
	// Keyword 按名称/主机关键字模糊匹配
	Keyword string
}

// PageConnectionQuery 分页查询条件
type PageConnectionQuery struct {
	// Group 按分组过滤
	Group string
	// Type 按驱动类型过滤
	Type string
	// Keyword 按名称/主机关键字模糊匹配
	Keyword string
	// Page 页码（从 1 开始）
	Page int
	// PageSize 每页条数
	PageSize int
}
