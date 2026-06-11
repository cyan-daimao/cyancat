package schemaservice

// DatabaseBO 数据库业务对象
type DatabaseBO struct {
	// Name 数据库名称
	Name string
	// Charset 字符集
	Charset string
	// Collation 排序规则
	Collation string
}

// SchemaBO Schema 业务对象
type SchemaBO struct {
	// Name schema 名称
	Name string
	// Owner 所有者
	Owner string
}

// TableBO 表业务对象
type TableBO struct {
	// Name 表名
	Name string
	// Type 类型
	Type string
	// Comment 注释
	Comment string
	// RowCount 估算行数
	RowCount int64
}

// ViewBO 视图业务对象
type ViewBO struct {
	// Name 视图名
	Name string
	// Definition 视图定义
	Definition string
}

// ColumnBO 字段业务对象
type ColumnBO struct {
	// Name 字段名
	Name string
	// DatabaseType 数据库原生类型
	DatabaseType string
	// Nullable 是否可空
	Nullable bool
	// IsPrimary 是否主键
	IsPrimary bool
}

// IndexBO 索引业务对象
type IndexBO struct {
	// Name 索引名
	Name string
	// Columns 索引列
	Columns []string
	// Unique 是否唯一
	Unique bool
	// Primary 是否主键索引
	Primary bool
}

// ForeignKeyBO 外键业务对象
type ForeignKeyBO struct {
	// Name 约束名
	Name string
	// Columns 本表列
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

// TableDetailBO 表详情业务对象
type TableDetailBO struct {
	// Name 表名
	Name string
	// Schema schema 名
	Schema string
	// Database 数据库名
	Database string
	// Comment 注释
	Comment string
	// Columns 字段列表
	Columns []ColumnBO
	// Indexes 索引列表
	Indexes []IndexBO
	// ForeignKeys 外键列表
	ForeignKeys []ForeignKeyBO
}
