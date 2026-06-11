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
			Name:         c.Name,
			DatabaseType: c.DatabaseType,
			Nullable:     c.Nullable,
			IsPrimary:    c.IsPrimary,
		})
	}
	indexes := make([]IndexDTO, 0, len(bo.Indexes))
	for _, idx := range bo.Indexes {
		indexes = append(indexes, IndexDTO{
			Name:    idx.Name,
			Columns: idx.Columns,
			Unique:  idx.Unique,
			Primary: idx.Primary,
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
