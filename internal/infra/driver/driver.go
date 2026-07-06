// Package driver 定义数据库驱动抽象，负责屏蔽不同数据库（MySQL / PostgreSQL 等）的底层差异。
//
// 驱动抽象属于基础设施层，提供给 application 层用于实际连接、执行 SQL 与查询元数据。
// domain 层只感知 Driver 接口，不感知具体实现。
package driver

import (
	"context"
	"time"
)

// DriverType 驱动类型枚举
type DriverType string

const (
	// MySQL MySQL 驱动
	MySQL DriverType = "mysql"
	// PostgreSQL PostgreSQL 驱动
	PostgreSQL DriverType = "postgres"
	// SQLite SQLite 文件数据库驱动
	SQLite DriverType = "sqlite"
	// StarRocks StarRocks 驱动（MySQL 协议兼容）
	StarRocks DriverType = "starrocks"
	// Kafka Kafka 驱动
	Kafka DriverType = "kafka"
)

// IsValid 校验驱动类型是否合法
func (t DriverType) IsValid() bool {
	switch t {
	case MySQL, PostgreSQL, SQLite, StarRocks, Kafka:
		return true
	default:
		return false
	}
}

// String 返回驱动类型字符串
func (t DriverType) String() string {
	return string(t)
}

// ConnConfig 通用连接配置
type ConnConfig struct {
	// Type 驱动类型
	Type DriverType
	// Host 主机
	Host string
	// Port 端口
	Port int
	// User 用户名
	User string
	// Password 密码
	Password string
	// Database 默认数据库
	Database string
	// SSL 是否启用 SSL
	SSL bool
	// ConnectTimeout 连接超时
	ConnectTimeout time.Duration
	// Params 额外驱动参数（如 charset、sslmode）
	Params map[string]string
}

// Driver 数据库驱动接口
type Driver interface {
	// Type 驱动类型
	Type() DriverType
	// Connect 建立连接
	Connect(ctx context.Context, cfg ConnConfig) (Conn, error)
	// Dialect 方言（引号、占位符、默认 schema 等）
	Dialect() Dialect
}

// Conn 单个数据库连接（屏蔽底层 sql.DB / pgx.Pool 差异）
type Conn interface {
	// Ping 测试连接
	Ping(ctx context.Context) error
	// WithDatabase 返回用于指定数据库的执行连接；cleanup 必须在使用后调用
	WithDatabase(ctx context.Context, database string) (conn Conn, cleanup func(), err error)
	// Execute 执行 SQL（适用于小结果集或 DML/DDL）
	Execute(ctx context.Context, sql string, args ...any) (*Result, error)
	// Stream 流式执行（适用于大结果集，返回游标）
	Stream(ctx context.Context, sql string, args ...any) (RowStream, error)
	// Inspector 元数据查询器
	Inspector() Inspector
	// DDL 返回方言专属的 DDL 生成器
	DDL() DDLGenerator
	// ServerVersion 返回数据库服务端版本字符串
	ServerVersion(ctx context.Context) (string, error)
	// Close 关闭连接
	Close() error
}

// Column 结果集/表字段元数据
type Column struct {
	// Name 列名
	Name string
	// DatabaseType 数据库原生类型（如 VARCHAR(64)、int4、jsonb）
	DatabaseType string
	// Nullable 是否可空（元数据未知时为 true）
	Nullable bool
	// IsPrimary 是否主键（仅在 DescribeTable 等元数据查询时填充）
	IsPrimary bool
	// AutoIncrement 是否自增（MySQL: auto_increment；PG: identity/serial）
	AutoIncrement bool
	// Unsigned 是否无符号（MySQL 专有）
	Unsigned bool
	// DefaultValue 默认值字符串（NULL 用 nil 区分「无默认」与「DEFAULT NULL」）
	DefaultValue *string
	// Comment 列注释
	Comment string
	// Extra 额外修饰符（MySQL EXTRA 列；PG 暂未使用）
	Extra string
	// OrdinalPosition 列序号
	OrdinalPosition int
	// TypeLength 字符/字节最大长度（如 varchar 长度）
	TypeLength *int
	// Precision 数字精度（decimal 总位数）
	Precision *int
	// Scale 数字小数位
	Scale *int
	// Collation 列级排序规则
	Collation string
}

// Result 同步执行结果（适合小结果集和 DML/DDL）
type Result struct {
	// Columns 结果集列定义（DML/DDL 时为空）
	Columns []Column
	// Rows 结果数据（每行 []any，与 Columns 一一对应；DML/DDL 时为空）
	Rows [][]any
	// RowsAffected 受影响行数（仅 DML 有效）
	RowsAffected int64
	// LastInsertID 最后插入 ID（仅 MySQL DML 有效，PG 始终 0）
	LastInsertID int64
}

// RowStream 流式行游标
type RowStream interface {
	// Columns 返回结果集列定义
	Columns() []Column
	// Next 移动到下一行，返回 false 表示读取完毕
	Next() bool
	// Scan 把当前行扫描到 []any
	Scan() ([]any, error)
	// Err 读取过程中的错误
	Err() error
	// Close 关闭游标
	Close() error
}

// Dialect SQL 方言适配（引号、占位符等差异）
type Dialect interface {
	// QuoteIdent 用方言对应的引号包裹标识符（如 MySQL 反引号、PG 双引号）
	QuoteIdent(ident string) string
	// Placeholder 返回第 N 个参数占位符（MySQL: "?"，PG: "$1"）
	Placeholder(n int) string
	// DefaultLimit 默认 LIMIT 行数（用于未带 LIMIT 的 SELECT 兜底）
	DefaultLimit() int
}

// Inspector 数据库元数据查询接口
type Inspector interface {
	// ListDatabases 列出所有数据库
	ListDatabases(ctx context.Context) ([]Database, error)
	// ListSchemas 列出指定数据库下的 schema（MySQL 概念为 database 同名）
	ListSchemas(ctx context.Context, database string) ([]Schema, error)
	// ListTables 列出指定 schema 下的表
	ListTables(ctx context.Context, database, schema string) ([]Table, error)
	// ListViews 列出指定 schema 下的视图
	ListViews(ctx context.Context, database, schema string) ([]View, error)
	// DescribeTable 描述表结构（含字段、索引、外键）
	DescribeTable(ctx context.Context, database, schema, table string) (*TableDetail, error)
	// ListIndexes 列出表的索引
	ListIndexes(ctx context.Context, database, schema, table string) ([]Index, error)
	// ListForeignKeys 列出表的外键
	ListForeignKeys(ctx context.Context, database, schema, table string) ([]ForeignKey, error)
	// ListCharsets 列出可用字符集（用于新建数据库/表时选择）
	ListCharsets(ctx context.Context) ([]Charset, error)
	// ListCollations 列出指定字符集下的排序规则（charset 为空时列出全部）
	ListCollations(ctx context.Context, charset string) ([]Collation, error)
}

// Charset 字符集元数据
type Charset struct {
	// Name 字符集名（utf8mb4 / UTF8 ...）
	Name string
	// Description 描述
	Description string
	// DefaultCollation 默认排序规则
	DefaultCollation string
}

// Collation 排序规则元数据
type Collation struct {
	// Name 排序规则名（utf8mb4_general_ci）
	Name string
	// Charset 所属字符集
	Charset string
	// IsDefault 是否字符集默认
	IsDefault bool
}

// Database 数据库元数据
type Database struct {
	// Name 数据库名称
	Name string
	// Charset 字符集
	Charset string
	// Collation 排序规则
	Collation string
}

// Schema schema 元数据
type Schema struct {
	// Name schema 名称
	Name string
	// Owner 所有者（PG 适用）
	Owner string
}

// Table 表元数据
type Table struct {
	// Name 表名
	Name string
	// Type 类型（BASE TABLE / TABLE / VIEW 等）
	Type string
	// Comment 表注释
	Comment string
	// RowCount 估算行数（可能为 0）
	RowCount int64
}

// View 视图元数据
type View struct {
	// Name 视图名
	Name string
	// Definition 视图定义 SQL
	Definition string
}

// TableDetail 表详情（字段+索引+外键）
type TableDetail struct {
	// Name 表名
	Name string
	// Schema schema 名
	Schema string
	// Database 数据库名
	Database string
	// Comment 表注释
	Comment string
	// Columns 字段定义
	Columns []Column
	// Indexes 索引列表
	Indexes []Index
	// ForeignKeys 外键列表
	ForeignKeys []ForeignKey
}

// Index 索引元数据
type Index struct {
	// Name 索引名
	Name string
	// Columns 索引列
	Columns []string
	// Unique 是否唯一索引
	Unique bool
	// Primary 是否主键索引
	Primary bool
	// Comment 索引注释
	Comment string
}

// ForeignKey 外键元数据
type ForeignKey struct {
	// Name 外键约束名
	Name string
	// Columns 本表外键列
	Columns []string
	// ReferencedSchema 引用 schema
	ReferencedSchema string
	// ReferencedTable 引用表
	ReferencedTable string
	// ReferencedColumns 引用列
	ReferencedColumns []string
	// OnUpdate 更新规则（CASCADE/SET NULL/...）
	OnUpdate string
	// OnDelete 删除规则
	OnDelete string
}
