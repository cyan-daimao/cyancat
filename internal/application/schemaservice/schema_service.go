// Package schemaservice 提供数据库元数据浏览的应用层服务
package schemaservice

// SchemaService 元数据查询服务接口
type SchemaService interface {
	// ListDatabases 列出数据库
	ListDatabases(query *ListDatabasesQuery) ([]*DatabaseBO, error)
	// ListSchemas 列出指定数据库下的 schema
	ListSchemas(query *ListSchemasQuery) ([]*SchemaBO, error)
	// ListTables 列出指定 schema 下的表
	ListTables(query *ListTablesQuery) ([]*TableBO, error)
	// ListViews 列出视图
	ListViews(query *ListTablesQuery) ([]*ViewBO, error)
	// DescribeTable 描述表（字段+索引+外键）
	DescribeTable(query *DescribeTableQuery) (*TableDetailBO, error)
	// ListIndexes 列出索引
	ListIndexes(query *DescribeTableQuery) ([]*IndexBO, error)
	// ListForeignKeys 列出外键
	ListForeignKeys(query *DescribeTableQuery) ([]*ForeignKeyBO, error)
}
