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
	// ListCharsets 列出可用字符集
	ListCharsets(query *ListCharsetsQuery) ([]*CharsetBO, error)
	// ListCollations 列出指定字符集下的排序规则
	ListCollations(query *ListCollationsQuery) ([]*CollationBO, error)
	// GetCreateTableDDL 获取建表 DDL
	GetCreateTableDDL(query *GetCreateTableDDLQuery) (string, error)
	// PreviewCreateDatabase 预览 CREATE DATABASE DDL（不执行）
	PreviewCreateDatabase(cmd *CreateDatabaseCmd) (string, error)
	// CreateDatabase 创建数据库
	CreateDatabase(cmd *CreateDatabaseCmd) error
	// PreviewDropDatabase 预览 DROP DATABASE DDL（不执行）
	PreviewDropDatabase(cmd *DropDatabaseCmd) (string, error)
	// DropDatabase 删除数据库
	DropDatabase(cmd *DropDatabaseCmd) error
	// PreviewCreateTable 预览 CREATE TABLE DDL（不执行）
	PreviewCreateTable(cmd *CreateTableCmd) (string, error)
	// CreateTable 创建表
	CreateTable(cmd *CreateTableCmd) error
	// PreviewAlterTable 预览 ALTER TABLE DDL（不执行）
	PreviewAlterTable(cmd *AlterTableCmd) (string, error)
	// AlterTable 修改表
	AlterTable(cmd *AlterTableCmd) error
	// PreviewDropTable 预览 DROP TABLE DDL（不执行）
	PreviewDropTable(cmd *DropTableCmd) (string, error)
	// DropTable 删除表
	DropTable(cmd *DropTableCmd) error
}
