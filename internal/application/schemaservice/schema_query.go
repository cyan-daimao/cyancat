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
