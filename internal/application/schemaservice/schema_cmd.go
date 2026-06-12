package schemaservice

// --- DDL 命令类型 ---

// CreateDatabaseCmd 创建数据库命令
type CreateDatabaseCmd struct {
	ConnID    int64
	Name      string
	Charset   string
	Collation string
}

// CreateTableCmd 创建表命令
type CreateTableCmd struct {
	ConnID      int64
	Database    string
	Schema      string
	Name        string
	Columns     []ColumnSpec
	PK          []string
	Indexes     []IndexSpec
	ForeignKeys []ForeignKeySpec
	Engine      string
	Charset     string
	Collation   string
	Comment     string
}

// AlterTableCmd 修改表命令
type AlterTableCmd struct {
	ConnID          int64
	Database        string
	Schema          string
	Name            string
	AddColumns      []ColumnSpec
	DropColumns     []string
	RenameColumns   []ColumnRename
	ModifyColumns   []ColumnSpec
	AddIndexes      []IndexSpec
	ModifyIndexes   []IndexSpec
	DropIndexes     []string
	AddForeignKeys  []ForeignKeySpec
	DropForeignKeys []string
	Engine          string
	Charset         string
	Collation       string
	Comment         string
}

// DropTableCmd 删除表命令
type DropTableCmd struct {
	ConnID   int64
	Database string
	Schema   string
	Name     string
}

// ColumnSpec 应用层字段规格（用于 DDL 命令）
type ColumnSpec struct {
	Name          string
	DataType      string
	TypeLength    *int
	Precision     *int
	Scale         *int
	Nullable      bool
	Unsigned      bool
	AutoIncrement bool
	DefaultValue  *string
	Comment       string
	Collation     string
	First         bool
	After         string
}

// IndexSpec 应用层索引规格
type IndexSpec struct {
	Name    string
	Type    string
	Columns []string
	Comment string
}

// ForeignKeySpec 应用层外键规格
type ForeignKeySpec struct {
	Name              string
	Columns           []string
	ReferencedSchema  string
	ReferencedTable   string
	ReferencedColumns []string
	OnUpdate          string
	OnDelete          string
}

// ColumnRename 字段重命名
type ColumnRename struct {
	Old string
	New string
}

// --- 查询类型 ---

// ListCharsetsQuery 列出字符集查询
type ListCharsetsQuery struct {
	ConnID int64
}

// ListCollationsQuery 列出排序规则查询
type ListCollationsQuery struct {
	ConnID  int64
	Charset string
}

// GetCreateTableDDLQuery 获取建表 DDL 查询
type GetCreateTableDDLQuery struct {
	ConnID   int64
	Database string
	Schema   string
	Table    string
}
