package dto

import (
	"cyancat/internal/application/schemaservice"
)

// --- 响应对象 ---

// DatabaseDTO 数据库
type DatabaseDTO struct {
	// Name 名称
	Name string `json:"name"`
	// Charset 字符集
	Charset string `json:"charset"`
	// Collation 排序规则
	Collation string `json:"collation"`
}

// SchemaDTO schema
type SchemaDTO struct {
	// Name 名称
	Name string `json:"name"`
	// Owner 所有者
	Owner string `json:"owner"`
}

// TableDTO 表
type TableDTO struct {
	// Name 名称
	Name string `json:"name"`
	// Type 类型
	Type string `json:"type"`
	// Comment 注释
	Comment string `json:"comment"`
	// RowCount 估算行数
	RowCount int64 `json:"rowCount"`
}

// ViewDTO 视图
type ViewDTO struct {
	// Name 名称
	Name string `json:"name"`
	// Definition 视图定义 SQL
	Definition string `json:"definition"`
}

// SchemaColumnDTO 表字段（避免与 query_dto 的 ColumnDTO 冲突）
type SchemaColumnDTO struct {
	// Name 列名
	Name string `json:"name"`
	// DatabaseType 数据库类型
	DatabaseType string `json:"databaseType"`
	// Nullable 是否可空
	Nullable bool `json:"nullable"`
	// IsPrimary 是否主键
	IsPrimary bool `json:"isPrimary"`
	// AutoIncrement 是否自增
	AutoIncrement bool `json:"autoIncrement"`
	// Unsigned 是否无符号
	Unsigned bool `json:"unsigned"`
	// DefaultValue 默认值
	DefaultValue *string `json:"defaultValue"`
	// Comment 注释
	Comment string `json:"comment"`
	// Extra 额外修饰
	Extra string `json:"extra"`
	// OrdinalPosition 列序号
	OrdinalPosition int `json:"ordinalPosition"`
	// TypeLength 长度
	TypeLength *int `json:"typeLength"`
	// Precision 精度
	Precision *int `json:"precision"`
	// Scale 小数位
	Scale *int `json:"scale"`
	// Collation 排序规则
	Collation string `json:"collation"`
}

// IndexDTO 索引
type IndexDTO struct {
	// Name 索引名
	Name string `json:"name"`
	// Columns 索引列
	Columns []string `json:"columns"`
	// Unique 是否唯一
	Unique bool `json:"unique"`
	// Primary 是否主键索引
	Primary bool `json:"primary"`
	// Comment 索引注释
	Comment string `json:"comment"`
}

// ForeignKeyDTO 外键
type ForeignKeyDTO struct {
	// Name 约束名
	Name string `json:"name"`
	// Columns 本表列
	Columns []string `json:"columns"`
	// ReferencedSchema 引用 schema
	ReferencedSchema string `json:"referencedSchema"`
	// ReferencedTable 引用表
	ReferencedTable string `json:"referencedTable"`
	// ReferencedColumns 引用列
	ReferencedColumns []string `json:"referencedColumns"`
	// OnUpdate 更新规则
	OnUpdate string `json:"onUpdate"`
	// OnDelete 删除规则
	OnDelete string `json:"onDelete"`
}

// TableDetailDTO 表详情
type TableDetailDTO struct {
	// Name 表名
	Name string `json:"name"`
	// Schema schema 名
	Schema string `json:"schema"`
	// Database 数据库名
	Database string `json:"database"`
	// Comment 注释
	Comment string `json:"comment"`
	// Columns 字段列表
	Columns []SchemaColumnDTO `json:"columns"`
	// Indexes 索引列表
	Indexes []IndexDTO `json:"indexes"`
	// ForeignKeys 外键列表
	ForeignKeys []ForeignKeyDTO `json:"foreignKeys"`
}

// CharsetDTO 字符集
type CharsetDTO struct {
	// Name 字符集名
	Name string `json:"name"`
	// Description 描述
	Description string `json:"description"`
	// DefaultCollation 默认排序规则
	DefaultCollation string `json:"defaultCollation"`
}

// CollationDTO 排序规则
type CollationDTO struct {
	// Name 排序规则名
	Name string `json:"name"`
	// Charset 所属字符集
	Charset string `json:"charset"`
	// IsDefault 是否字符集默认
	IsDefault bool `json:"isDefault"`
}

// DDLPreviewResultDTO DDL 预览结果
type DDLPreviewResultDTO struct {
	// SQL DDL 语句
	SQL string `json:"sql"`
}

// --- 请求对象 ---

// CreateDatabaseRequest 新建数据库请求
type CreateDatabaseRequest struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// Name 数据库名
	Name string `json:"name"`
	// Charset 字符集
	Charset string `json:"charset"`
	// Collation 排序规则
	Collation string `json:"collation"`
}

// ColumnSpecDTO 字段规格
type ColumnSpecDTO struct {
	// Name 字段名
	Name string `json:"name"`
	// DataType 数据类型
	DataType string `json:"dataType"`
	// TypeLength 长度
	TypeLength *int `json:"typeLength"`
	// Precision 精度
	Precision *int `json:"precision"`
	// Scale 小数位
	Scale *int `json:"scale"`
	// Nullable 是否可空
	Nullable bool `json:"nullable"`
	// AutoIncrement 是否自增
	AutoIncrement bool `json:"autoIncrement"`
	// Unsigned 是否无符号
	Unsigned bool `json:"unsigned"`
	// DefaultValue 默认值
	DefaultValue *string `json:"defaultValue"`
	// Comment 注释
	Comment string `json:"comment"`
	// Collation 排序规则
	Collation string `json:"collation"`
	// First 是否放在第一列（ALTER TABLE 用）
	First bool `json:"first"`
	// After 放在指定列之后（ALTER TABLE 用）
	After string `json:"after"`
}

// IndexSpecDTO 索引规格
type IndexSpecDTO struct {
	// Name 索引名
	Name string `json:"name"`
	// Type 索引类型（PRIMARY/UNIQUE/NORMAL/FULLTEXT）
	Type string `json:"type"`
	// Columns 索引列
	Columns []string `json:"columns"`
	// Comment 注释
	Comment string `json:"comment"`
}

// ForeignKeySpecDTO 外键规格
type ForeignKeySpecDTO struct {
	// Name 外键名
	Name string `json:"name"`
	// Columns 本表列
	Columns []string `json:"columns"`
	// ReferencedSchema 引用 schema
	ReferencedSchema string `json:"referencedSchema"`
	// ReferencedTable 引用表
	ReferencedTable string `json:"referencedTable"`
	// ReferencedColumns 引用列
	ReferencedColumns []string `json:"referencedColumns"`
	// OnDelete 删除规则
	OnDelete string `json:"onDelete"`
	// OnUpdate 更新规则
	OnUpdate string `json:"onUpdate"`
}

// ColumnRenameDTO 字段重命名
type ColumnRenameDTO struct {
	// Old 旧名
	Old string `json:"old"`
	// New 新名
	New string `json:"new"`
}

// CreateTableRequest 新建表请求
type CreateTableRequest struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// Database 数据库名
	Database string `json:"database"`
	// Schema schema 名
	Schema string `json:"schema"`
	// Name 表名
	Name string `json:"name"`
	// Columns 字段列表
	Columns []ColumnSpecDTO `json:"columns"`
	// PrimaryKey 主键列名
	PrimaryKey []string `json:"primaryKey"`
	// Indexes 索引列表
	Indexes []IndexSpecDTO `json:"indexes"`
	// ForeignKeys 外键列表
	ForeignKeys []ForeignKeySpecDTO `json:"foreignKeys"`
	// Engine 存储引擎
	Engine string `json:"engine"`
	// Charset 字符集
	Charset string `json:"charset"`
	// Collation 排序规则
	Collation string `json:"collation"`
	// Comment 表注释
	Comment string `json:"comment"`
}

// AlterTableRequest 修改表请求
type AlterTableRequest struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// Database 数据库名
	Database string `json:"database"`
	// Schema schema 名
	Schema string `json:"schema"`
	// Name 表名
	Name string `json:"name"`
	// AddColumns 新增字段
	AddColumns []ColumnSpecDTO `json:"addColumns"`
	// DropColumns 删除字段
	DropColumns []string `json:"dropColumns"`
	// RenameColumns 重命名字段
	RenameColumns []ColumnRenameDTO `json:"renameColumns"`
	// ModifyColumns 修改字段
	ModifyColumns []ColumnSpecDTO `json:"modifyColumns"`
	// AddIndexes 新增索引
	AddIndexes []IndexSpecDTO `json:"addIndexes"`
	// ModifyIndexes 修改索引
	ModifyIndexes []IndexSpecDTO `json:"modifyIndexes"`
	// DropIndexes 删除索引
	DropIndexes []string `json:"dropIndexes"`
	// AddForeignKeys 新增外键
	AddForeignKeys []ForeignKeySpecDTO `json:"addForeignKeys"`
	// DropForeignKeys 删除外键
	DropForeignKeys []string `json:"dropForeignKeys"`
	// Engine 存储引擎
	Engine string `json:"engine"`
	// Charset 字符集
	Charset string `json:"charset"`
	// Collation 排序规则
	Collation string `json:"collation"`
	// Comment 表注释
	Comment string `json:"comment"`
}

// GetCreateTableDDLRequest 获取建表 DDL 请求
type GetCreateTableDDLRequest struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// Database 数据库名
	Database string `json:"database"`
	// Schema schema 名
	Schema string `json:"schema"`
	// Table 表名
	Table string `json:"table"`
}

// DropTableRequest 删除表请求
type DropTableRequest struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// Database 数据库名
	Database string `json:"database"`
	// Schema schema 名
	Schema string `json:"schema"`
	// Name 表名
	Name string `json:"name"`
}

// ListCharsetsRequest 列出字符集请求
type ListCharsetsRequest struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
}

// ListCollationsRequest 列出排序规则请求
type ListCollationsRequest struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// Charset 字符集名（为空时列出全部）
	Charset string `json:"charset"`
}

// --- 转换 ---

// ToDatabaseDTO 单个
func ToDatabaseDTO(bo *schemaservice.DatabaseBO) *DatabaseDTO {
	if bo == nil {
		return nil
	}
	return &DatabaseDTO{
		Name:      bo.Name,
		Charset:   bo.Charset,
		Collation: bo.Collation,
	}
}

// ToDatabaseDTOs 批量
func ToDatabaseDTOs(bos []*schemaservice.DatabaseBO) []*DatabaseDTO {
	if len(bos) == 0 {
		return make([]*DatabaseDTO, 0)
	}
	result := make([]*DatabaseDTO, 0, len(bos))
	for _, bo := range bos {
		result = append(result, ToDatabaseDTO(bo))
	}
	return result
}

// ToSchemaDTO 单个
func ToSchemaDTO(bo *schemaservice.SchemaBO) *SchemaDTO {
	if bo == nil {
		return nil
	}
	return &SchemaDTO{Name: bo.Name, Owner: bo.Owner}
}

// ToSchemaDTOs 批量
func ToSchemaDTOs(bos []*schemaservice.SchemaBO) []*SchemaDTO {
	if len(bos) == 0 {
		return make([]*SchemaDTO, 0)
	}
	result := make([]*SchemaDTO, 0, len(bos))
	for _, bo := range bos {
		result = append(result, ToSchemaDTO(bo))
	}
	return result
}

// ToTableDTO 单个
func ToTableDTO(bo *schemaservice.TableBO) *TableDTO {
	if bo == nil {
		return nil
	}
	return &TableDTO{
		Name:     bo.Name,
		Type:     bo.Type,
		Comment:  bo.Comment,
		RowCount: bo.RowCount,
	}
}

// ToTableDTOs 批量
func ToTableDTOs(bos []*schemaservice.TableBO) []*TableDTO {
	if len(bos) == 0 {
		return make([]*TableDTO, 0)
	}
	result := make([]*TableDTO, 0, len(bos))
	for _, bo := range bos {
		result = append(result, ToTableDTO(bo))
	}
	return result
}

// ToViewDTO 单个
func ToViewDTO(bo *schemaservice.ViewBO) *ViewDTO {
	if bo == nil {
		return nil
	}
	return &ViewDTO{Name: bo.Name, Definition: bo.Definition}
}

// ToViewDTOs 批量
func ToViewDTOs(bos []*schemaservice.ViewBO) []*ViewDTO {
	if len(bos) == 0 {
		return make([]*ViewDTO, 0)
	}
	result := make([]*ViewDTO, 0, len(bos))
	for _, bo := range bos {
		result = append(result, ToViewDTO(bo))
	}
	return result
}

// ToTableDetailDTO 表详情转换
func ToTableDetailDTO(bo *schemaservice.TableDetailBO) *TableDetailDTO {
	if bo == nil {
		return nil
	}
	cols := make([]SchemaColumnDTO, 0, len(bo.Columns))
	for _, c := range bo.Columns {
		cols = append(cols, SchemaColumnDTO{
			Name:            c.Name,
			DatabaseType:    c.DatabaseType,
			Nullable:        c.Nullable,
			IsPrimary:       c.IsPrimary,
			AutoIncrement:   c.AutoIncrement,
			Unsigned:        c.Unsigned,
			DefaultValue:    c.DefaultValue,
			Comment:         c.Comment,
			Extra:           c.Extra,
			OrdinalPosition: c.OrdinalPosition,
			TypeLength:      c.TypeLength,
			Precision:       c.Precision,
			Scale:           c.Scale,
			Collation:       c.Collation,
		})
	}
	indexes := make([]IndexDTO, 0, len(bo.Indexes))
	for _, idx := range bo.Indexes {
		indexes = append(indexes, IndexDTO{
			Name:    idx.Name,
			Columns: idx.Columns,
			Unique:  idx.Unique,
			Primary: idx.Primary,
			Comment: idx.Comment,
		})
	}
	fks := make([]ForeignKeyDTO, 0, len(bo.ForeignKeys))
	for _, fk := range bo.ForeignKeys {
		fks = append(fks, ForeignKeyDTO{
			Name:              fk.Name,
			Columns:           fk.Columns,
			ReferencedSchema:  fk.ReferencedSchema,
			ReferencedTable:   fk.ReferencedTable,
			ReferencedColumns: fk.ReferencedColumns,
			OnUpdate:          fk.OnUpdate,
			OnDelete:          fk.OnDelete,
		})
	}
	return &TableDetailDTO{
		Name:        bo.Name,
		Schema:      bo.Schema,
		Database:    bo.Database,
		Comment:     bo.Comment,
		Columns:     cols,
		Indexes:     indexes,
		ForeignKeys: fks,
	}
}

// ToCharsetDTOs 批量转换字符集
func ToCharsetDTOs(bos []*schemaservice.CharsetBO) []*CharsetDTO {
	if len(bos) == 0 {
		return make([]*CharsetDTO, 0)
	}
	result := make([]*CharsetDTO, 0, len(bos))
	for _, bo := range bos {
		result = append(result, &CharsetDTO{
			Name:             bo.Name,
			Description:      bo.Description,
			DefaultCollation: bo.DefaultCollation,
		})
	}
	return result
}

// ToCollationDTOs 批量转换排序规则
func ToCollationDTOs(bos []*schemaservice.CollationBO) []*CollationDTO {
	if len(bos) == 0 {
		return make([]*CollationDTO, 0)
	}
	result := make([]*CollationDTO, 0, len(bos))
	for _, bo := range bos {
		result = append(result, &CollationDTO{
			Name:      bo.Name,
			Charset:   bo.Charset,
			IsDefault: bo.IsDefault,
		})
	}
	return result
}

// ToCreateDatabaseCmd 转换新建数据库命令
func ToCreateDatabaseCmd(req *CreateDatabaseRequest) *schemaservice.CreateDatabaseCmd {
	if req == nil {
		return nil
	}
	return &schemaservice.CreateDatabaseCmd{
		ConnID:    req.ConnID,
		Name:      req.Name,
		Charset:   req.Charset,
		Collation: req.Collation,
	}
}

// ToCreateTableCmd 转换新建表命令
func ToCreateTableCmd(req *CreateTableRequest) *schemaservice.CreateTableCmd {
	if req == nil {
		return nil
	}
	cmd := &schemaservice.CreateTableCmd{
		ConnID:    req.ConnID,
		Database:  req.Database,
		Schema:    req.Schema,
		Name:      req.Name,
		PK:        req.PrimaryKey,
		Engine:    req.Engine,
		Charset:   req.Charset,
		Collation: req.Collation,
		Comment:   req.Comment,
	}
	for _, c := range req.Columns {
		cmd.Columns = append(cmd.Columns, toColumnSpec(c))
	}
	for _, idx := range req.Indexes {
		cmd.Indexes = append(cmd.Indexes, toIndexSpec(idx))
	}
	for _, fk := range req.ForeignKeys {
		cmd.ForeignKeys = append(cmd.ForeignKeys, toForeignKeySpec(fk))
	}
	return cmd
}

// ToAlterTableCmd 转换修改表命令
func ToAlterTableCmd(req *AlterTableRequest) *schemaservice.AlterTableCmd {
	if req == nil {
		return nil
	}
	cmd := &schemaservice.AlterTableCmd{
		ConnID:    req.ConnID,
		Database:  req.Database,
		Schema:    req.Schema,
		Name:      req.Name,
		Engine:    req.Engine,
		Charset:   req.Charset,
		Collation: req.Collation,
		Comment:   req.Comment,
	}
	for _, c := range req.AddColumns {
		cmd.AddColumns = append(cmd.AddColumns, toColumnSpec(c))
	}
	cmd.DropColumns = req.DropColumns
	for _, r := range req.RenameColumns {
		cmd.RenameColumns = append(cmd.RenameColumns, schemaservice.ColumnRename{Old: r.Old, New: r.New})
	}
	for _, c := range req.ModifyColumns {
		cmd.ModifyColumns = append(cmd.ModifyColumns, toColumnSpec(c))
	}
	for _, idx := range req.AddIndexes {
		cmd.AddIndexes = append(cmd.AddIndexes, toIndexSpec(idx))
	}
	for _, idx := range req.ModifyIndexes {
		cmd.ModifyIndexes = append(cmd.ModifyIndexes, toIndexSpec(idx))
	}
	cmd.DropIndexes = req.DropIndexes
	for _, fk := range req.AddForeignKeys {
		cmd.AddForeignKeys = append(cmd.AddForeignKeys, toForeignKeySpec(fk))
	}
	cmd.DropForeignKeys = req.DropForeignKeys
	return cmd
}

func toColumnSpec(c ColumnSpecDTO) schemaservice.ColumnSpec {
	return schemaservice.ColumnSpec{
		Name:          c.Name,
		DataType:      c.DataType,
		TypeLength:    c.TypeLength,
		Precision:     c.Precision,
		Scale:         c.Scale,
		Nullable:      c.Nullable,
		AutoIncrement: c.AutoIncrement,
		Unsigned:      c.Unsigned,
		DefaultValue:  c.DefaultValue,
		Comment:       c.Comment,
		Collation:     c.Collation,
		First:         c.First,
		After:         c.After,
	}
}

func toIndexSpec(idx IndexSpecDTO) schemaservice.IndexSpec {
	return schemaservice.IndexSpec{
		Name:    idx.Name,
		Type:    idx.Type,
		Columns: idx.Columns,
		Comment: idx.Comment,
	}
}

func toForeignKeySpec(fk ForeignKeySpecDTO) schemaservice.ForeignKeySpec {
	return schemaservice.ForeignKeySpec{
		Name:              fk.Name,
		Columns:           fk.Columns,
		ReferencedSchema:  fk.ReferencedSchema,
		ReferencedTable:   fk.ReferencedTable,
		ReferencedColumns: fk.ReferencedColumns,
		OnDelete:          fk.OnDelete,
		OnUpdate:          fk.OnUpdate,
	}
}

// ToDropTableCmd 转换删除表命令
func ToDropTableCmd(req *DropTableRequest) *schemaservice.DropTableCmd {
	if req == nil {
		return nil
	}
	return &schemaservice.DropTableCmd{
		ConnID:   req.ConnID,
		Database: req.Database,
		Schema:   req.Schema,
		Name:     req.Name,
	}
}
