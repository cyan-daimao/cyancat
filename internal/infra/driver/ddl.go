package driver

import "context"

// DDLGenerator 方言专属的 DDL 生成器，从 Conn.DDL() 获取。
// DDL 生成器是纯函数层——只返回 SQL 字符串，不执行。
// 执行仍然通过 Conn.Execute() 完成。
type DDLGenerator interface {
	// CreateDatabase 生成 CREATE DATABASE / CREATE SCHEMA DDL
	CreateDatabase(spec DatabaseSpec) (string, error)
	// DropDatabase 生成 DROP DATABASE / DROP SCHEMA DDL
	DropDatabase(name string) (string, error)
	// CreateTable 生成 CREATE TABLE DDL
	CreateTable(spec TableSpec) (string, error)
	// AlterTable 生成 ALTER TABLE DDL（可含多个 action）
	AlterTable(spec AlterTableSpec) (string, error)
	// DropTable 生成 DROP TABLE DDL
	DropTable(database, schema, name string) (string, error)
	// GetCreateTableDDL 获取已有表的 CREATE TABLE DDL（MySQL: SHOW CREATE TABLE；PG: pg_dump 风格）
	GetCreateTableDDL(ctx context.Context, database, schema, table string) (string, error)
	// RenameTable 生成 RENAME TABLE / ALTER TABLE RENAME DDL
	RenameTable(database, schema, oldName, newName string) (string, error)
}

// DatabaseSpec 创建数据库的规格
type DatabaseSpec struct {
	// Name 数据库名
	Name string
	// Charset 字符集（MySQL 适用）
	Charset string
	// Collation 排序规则（MySQL 适用）
	Collation string
}

// TableSpec 创建表的规格
type TableSpec struct {
	// Database 数据库名
	Database string
	// Schema schema 名（PG 适用）
	Schema string
	// Name 表名
	Name string
	// Columns 字段列表
	Columns []ColumnSpec
	// PrimaryKey 主键字段列表（为空则无主键）
	PrimaryKey []string
	// Indexes 索引列表
	Indexes []IndexSpec
	// ForeignKeys 外键列表
	ForeignKeys []ForeignKeySpec
	// Engine 存储引擎（MySQL 适用，如 InnoDB）
	Engine string
	// Charset 字符集
	Charset string
	// Collation 排序规则
	Collation string
	// Comment 表注释
	Comment string
}

// ColumnSpec 字段规格
type ColumnSpec struct {
	// Name 字段名
	Name string
	// DataType 数据类型（如 VARCHAR, INT, DECIMAL, TEXT 等）
	DataType string
	// TypeLength 类型长度（如 varchar(255) 的 255）
	TypeLength *int
	// Precision 精度（如 decimal(10,2) 的 10）
	Precision *int
	// Scale 小数位（如 decimal(10,2) 的 2）
	Scale *int
	// Nullable 是否可空
	Nullable bool
	// Unsigned 是否无符号（MySQL 适用）
	Unsigned bool
	// AutoIncrement 是否自增
	AutoIncrement bool
	// DefaultValue 默认值（nil 表示无默认值）
	DefaultValue *string
	// Comment 字段注释
	Comment string
	// Collation 字段级排序规则
	Collation string
	// After 在指定字段之后（ALTER TABLE ADD COLUMN 时使用）
	After string
	// First 是否作为第一列（ALTER TABLE ADD COLUMN 时使用）
	First bool
}

// IndexSpec 索引规格
type IndexSpec struct {
	// Name 索引名
	Name string
	// Type 索引类型（PRIMARY / UNIQUE / NORMAL / FULLTEXT）
	Type string
	// Columns 索引列
	Columns []string
	// Comment 索引注释
	Comment string
}

// ForeignKeySpec 外键规格
type ForeignKeySpec struct {
	// Name 外键名
	Name string
	// Columns 本表外键列
	Columns []string
	// ReferencedSchema 引用 schema
	ReferencedSchema string
	// ReferencedTable 引用表
	ReferencedTable string
	// ReferencedColumns 引用列
	ReferencedColumns []string
	// OnUpdate 更新规则
	OnUpdate string
	// OnDelete 删除规则
	OnDelete string
}

// AlterTableSpec ALTER TABLE 规格
type AlterTableSpec struct {
	// Database 数据库名
	Database string
	// Schema schema 名
	Schema string
	// Name 表名
	Name string
	// AddColumns 新增字段
	AddColumns []ColumnSpec
	// DropColumns 删除字段
	DropColumns []string
	// ModifyColumns 修改字段（完整定义覆盖）
	ModifyColumns []ColumnSpec
	// RenameColumns 重命名字段
	RenameColumns []ColumnRename
	// AddIndexes 新增索引
	AddIndexes []IndexSpec
	// DropIndexes 删除索引
	DropIndexes []string
	// AddForeignKeys 新增外键
	AddForeignKeys []ForeignKeySpec
	// DropForeignKeys 删除外键
	DropForeignKeys []string
	// Charset 修改字符集
	Charset string
	// Collation 修改排序规则
	Collation string
	// Comment 修改表注释
	Comment string
	// Engine 修改存储引擎
	Engine string
}

// ColumnRename 字段重命名
type ColumnRename struct {
	// Old 旧名
	Old string
	// New 新名
	New string
}
