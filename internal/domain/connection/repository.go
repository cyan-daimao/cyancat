package connection

// Repository 连接配置仓储接口（由 infra 层实现）
type Repository interface {
	// List 列出所有未删除连接
	List(query *ListQuery) ([]*Connection, error)
	// Page 分页查询
	Page(query *PageQuery) ([]*Connection, int64, error)
	// GetByID 按 ID 查询
	GetByID(id int64) (*Connection, error)
	// GetByName 按名称查询（用于唯一性校验）
	GetByName(name string) (*Connection, error)
	// Save 新建连接
	Save(conn *Connection) error
	// Update 更新连接
	Update(conn *Connection) error
	// Delete 软删除连接
	Delete(id int64) error
}

// ListQuery 列表查询条件
type ListQuery struct {
	// Group 按分组过滤（空字符串表示不过滤）
	Group string
	// Type 按驱动类型过滤（空字符串表示不过滤）
	Type string
	// Keyword 按名称/主机关键字模糊匹配
	Keyword string
}

// PageQuery 分页查询条件
type PageQuery struct {
	// ListQuery 复用列表查询条件
	ListQuery
	// Page 页码（从 1 开始）
	Page int
	// PageSize 每页条数
	PageSize int
}
