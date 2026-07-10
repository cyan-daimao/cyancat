package schemaservice

// ListDatabasesQuery 列出数据库的查询
type ListDatabasesQuery struct {
	// ConnID 目标连接 ID
	ConnID int64
}

// ListSchemasQuery 列出 schema 的查询
type ListSchemasQuery struct {
	// ConnID 目标连接 ID
	ConnID int64
	// Database 数据库名
	Database string
}

// ListTablesQuery 列出表/视图的查询
type ListTablesQuery struct {
	// ConnID 目标连接 ID
	ConnID int64
	// Database 数据库名
	Database string
	// Schema schema 名（MySQL 为空时与 Database 同名）
	Schema string
	// Limit 分页限制（<= 0 表示不限制）
	Limit int
	// Offset 分页偏移
	Offset int
}

// SearchTablesQuery 搜索表/视图的查询
type SearchTablesQuery struct {
	// ConnID 目标连接 ID
	ConnID int64
	// Database 数据库名
	Database string
	// Schema schema 名
	Schema string
	// Keyword 搜索关键字
	Keyword string
	// Limit 返回数量限制（<= 0 时使用默认值 50）
	Limit int
}

// DescribeTableQuery 描述表的查询
type DescribeTableQuery struct {
	// ConnID 目标连接 ID
	ConnID int64
	// Database 数据库名
	Database string
	// Schema schema 名
	Schema string
	// Table 表名
	Table string
}
