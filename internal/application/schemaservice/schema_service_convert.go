package schemaservice

import "cyancat/internal/infra/driver"

// ToDatabaseBOs 转换数据库列表
func ToDatabaseBOs(list []driver.Database) []*DatabaseBO {
	if len(list) == 0 {
		return make([]*DatabaseBO, 0)
	}
	result := make([]*DatabaseBO, 0, len(list))
	for _, d := range list {
		result = append(result, &DatabaseBO{
			Name:      d.Name,
			Charset:   d.Charset,
			Collation: d.Collation,
		})
	}
	return result
}

// ToSchemaBOs 转换 schema 列表
func ToSchemaBOs(list []driver.Schema) []*SchemaBO {
	if len(list) == 0 {
		return make([]*SchemaBO, 0)
	}
	result := make([]*SchemaBO, 0, len(list))
	for _, s := range list {
		result = append(result, &SchemaBO{
			Name:  s.Name,
			Owner: s.Owner,
		})
	}
	return result
}

// ToTableBOs 转换表列表
func ToTableBOs(list []driver.Table) []*TableBO {
	if len(list) == 0 {
		return make([]*TableBO, 0)
	}
	result := make([]*TableBO, 0, len(list))
	for _, t := range list {
		result = append(result, &TableBO{
			Name:     t.Name,
			Type:     t.Type,
			Comment:  t.Comment,
			RowCount: t.RowCount,
		})
	}
	return result
}

// ToViewBOs 转换视图列表
func ToViewBOs(list []driver.View) []*ViewBO {
	if len(list) == 0 {
		return make([]*ViewBO, 0)
	}
	result := make([]*ViewBO, 0, len(list))
	for _, v := range list {
		result = append(result, &ViewBO{
			Name:       v.Name,
			Definition: v.Definition,
		})
	}
	return result
}

// ToColumnBOs 转换字段列表
func ToColumnBOs(list []driver.Column) []ColumnBO {
	if len(list) == 0 {
		return make([]ColumnBO, 0)
	}
	result := make([]ColumnBO, 0, len(list))
	for _, c := range list {
		result = append(result, ColumnBO{
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
	return result
}

// ToCharsetBOs 转换字符集列表
func ToCharsetBOs(list []driver.Charset) []*CharsetBO {
	result := make([]*CharsetBO, 0, len(list))
	for _, c := range list {
		result = append(result, &CharsetBO{
			Name:             c.Name,
			Description:      c.Description,
			DefaultCollation: c.DefaultCollation,
		})
	}
	return result
}

// ToCollationBOs 转换排序规则列表
func ToCollationBOs(list []driver.Collation) []*CollationBO {
	result := make([]*CollationBO, 0, len(list))
	for _, c := range list {
		result = append(result, &CollationBO{
			Name:      c.Name,
			Charset:   c.Charset,
			IsDefault: c.IsDefault,
		})
	}
	return result
}

// --- 应用层 spec 转 driver spec ---

// toDriverColumnSpecs 转换字段规格列表
func toDriverColumnSpecs(list []ColumnSpec) []driver.ColumnSpec {
	if len(list) == 0 {
		return nil
	}
	out := make([]driver.ColumnSpec, 0, len(list))
	for _, c := range list {
		out = append(out, driver.ColumnSpec{
			Name:          c.Name,
			DataType:      c.DataType,
			TypeLength:    c.TypeLength,
			Precision:     c.Precision,
			Scale:         c.Scale,
			Nullable:      c.Nullable,
			Unsigned:      c.Unsigned,
			AutoIncrement: c.AutoIncrement,
			DefaultValue:  c.DefaultValue,
			Comment:       c.Comment,
			Collation:     c.Collation,
			First:         c.First,
			After:         c.After,
		})
	}
	return out
}

// toDriverIndexSpecs 转换索引规格列表
func toDriverIndexSpecs(list []IndexSpec) []driver.IndexSpec {
	if len(list) == 0 {
		return nil
	}
	out := make([]driver.IndexSpec, 0, len(list))
	for _, idx := range list {
		out = append(out, driver.IndexSpec{
			Name:    idx.Name,
			Type:    idx.Type,
			Columns: idx.Columns,
			Comment: idx.Comment,
		})
	}
	return out
}

// toDriverForeignKeySpecs 转换外键规格列表
func toDriverForeignKeySpecs(list []ForeignKeySpec) []driver.ForeignKeySpec {
	if len(list) == 0 {
		return nil
	}
	out := make([]driver.ForeignKeySpec, 0, len(list))
	for _, fk := range list {
		out = append(out, driver.ForeignKeySpec{
			Name:              fk.Name,
			Columns:           fk.Columns,
			ReferencedSchema:  fk.ReferencedSchema,
			ReferencedTable:   fk.ReferencedTable,
			ReferencedColumns: fk.ReferencedColumns,
			OnUpdate:          fk.OnUpdate,
			OnDelete:          fk.OnDelete,
		})
	}
	return out
}

// ToIndexBOs 转换索引列表
func ToIndexBOs(list []driver.Index) []*IndexBO {
	if len(list) == 0 {
		return make([]*IndexBO, 0)
	}
	result := make([]*IndexBO, 0, len(list))
	for _, idx := range list {
		result = append(result, &IndexBO{
			Name:    idx.Name,
			Columns: idx.Columns,
			Unique:  idx.Unique,
			Primary: idx.Primary,
			Comment: idx.Comment,
		})
	}
	return result
}

// ToForeignKeyBOs 转换外键列表
func ToForeignKeyBOs(list []driver.ForeignKey) []*ForeignKeyBO {
	if len(list) == 0 {
		return make([]*ForeignKeyBO, 0)
	}
	result := make([]*ForeignKeyBO, 0, len(list))
	for _, fk := range list {
		result = append(result, &ForeignKeyBO{
			Name:              fk.Name,
			Columns:           fk.Columns,
			ReferencedSchema:  fk.ReferencedSchema,
			ReferencedTable:   fk.ReferencedTable,
			ReferencedColumns: fk.ReferencedColumns,
			OnUpdate:          fk.OnUpdate,
			OnDelete:          fk.OnDelete,
		})
	}
	return result
}

// ToTableDetailBO 转换表详情
func ToTableDetailBO(detail *driver.TableDetail) *TableDetailBO {
	if detail == nil {
		return nil
	}
	bo := &TableDetailBO{
		Name:     detail.Name,
		Schema:   detail.Schema,
		Database: detail.Database,
		Comment:  detail.Comment,
		Columns:  ToColumnBOs(detail.Columns),
	}
	// indexes
	if len(detail.Indexes) > 0 {
		bo.Indexes = make([]IndexBO, 0, len(detail.Indexes))
		for _, idx := range detail.Indexes {
			bo.Indexes = append(bo.Indexes, IndexBO{
				Name:    idx.Name,
				Columns: idx.Columns,
				Unique:  idx.Unique,
				Primary: idx.Primary,
				Comment: idx.Comment,
			})
		}
	}
	// fks
	if len(detail.ForeignKeys) > 0 {
		bo.ForeignKeys = make([]ForeignKeyBO, 0, len(detail.ForeignKeys))
		for _, fk := range detail.ForeignKeys {
			bo.ForeignKeys = append(bo.ForeignKeys, ForeignKeyBO{
				Name:              fk.Name,
				Columns:           fk.Columns,
				ReferencedSchema:  fk.ReferencedSchema,
				ReferencedTable:   fk.ReferencedTable,
				ReferencedColumns: fk.ReferencedColumns,
				OnUpdate:          fk.OnUpdate,
				OnDelete:          fk.OnDelete,
			})
		}
	}
	return bo
}
