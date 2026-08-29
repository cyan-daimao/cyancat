package postgres

import (
	"context"
	"fmt"
	"strings"

	"cyancat/internal/infra/driver"
)

// ddlGenerator PostgreSQL DDL 生成器实现
type ddlGenerator struct {
	conn    *Conn
	dialect driver.Dialect
}

func newDDLGenerator(c *Conn) driver.DDLGenerator {
	return &ddlGenerator{
		conn:    c,
		dialect: &postgresDialect{},
	}
}

// quote 简写
func (g *ddlGenerator) quote(ident string) string {
	return g.dialect.QuoteIdent(ident)
}

// qualified 拼接 "schema"."table" 形式
func (g *ddlGenerator) qualified(schema, name string) string {
	if schema != "" {
		return g.quote(schema) + "." + g.quote(name)
	}
	return g.quote(name)
}

// CreateDatabase 生成 CREATE DATABASE DDL（PG 中较少用，通常用 CREATE SCHEMA）
func (g *ddlGenerator) CreateDatabase(spec driver.DatabaseSpec) (string, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return "", fmt.Errorf("postgres/ddl: database name is required")
	}
	var b strings.Builder
	b.WriteString("CREATE DATABASE ")
	b.WriteString(g.quote(spec.Name))
	// PG 的 CREATE DATABASE 支持 ENCODING 和 LC_COLLATE
	if spec.Charset != "" {
		b.WriteString(" ENCODING ")
		b.WriteString("'" + escapePGString(spec.Charset) + "'")
	}
	if spec.Collation != "" {
		b.WriteString(" LC_COLLATE ")
		b.WriteString("'" + escapePGString(spec.Collation) + "'")
	}
	return b.String(), nil
}

// DropDatabase 生成 DROP DATABASE DDL
func (g *ddlGenerator) DropDatabase(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("postgres/ddl: database name is required")
	}
	return "DROP DATABASE " + g.quote(name), nil
}

// renderColumnDef 生成单个字段定义片段
// 形如： "name" VARCHAR(255) NOT NULL DEFAULT ''
func (g *ddlGenerator) renderColumnDef(c driver.ColumnSpec) (string, error) {
	if strings.TrimSpace(c.Name) == "" {
		return "", fmt.Errorf("postgres/ddl: column name is required")
	}
	if strings.TrimSpace(c.DataType) == "" {
		return "", fmt.Errorf("postgres/ddl: column %q data type is required", c.Name)
	}
	var b strings.Builder
	b.WriteString(g.quote(c.Name))
	b.WriteString(" ")

	// PG 不支持 UNSIGNED，忽略 c.Unsigned
	dataType := strings.ToUpper(c.DataType)

	// 自增类型映射：SERIAL / BIGSERIAL / SMALLSERIAL
	if c.AutoIncrement {
		switch dataType {
		case "BIGINT", "INT8":
			dataType = "BIGSERIAL"
		case "SMALLINT", "INT2":
			dataType = "SMALLSERIAL"
		default:
			dataType = "SERIAL"
		}
	}

	b.WriteString(dataType)

	// 拼长度/精度（SERIAL 系列类型不支持长度修饰）
	if !isPGSerialType(dataType) {
		switch {
		case c.Precision != nil && c.Scale != nil:
			b.WriteString(fmt.Sprintf("(%d,%d)", *c.Precision, *c.Scale))
		case c.Precision != nil:
			b.WriteString(fmt.Sprintf("(%d)", *c.Precision))
		case c.TypeLength != nil:
			b.WriteString(fmt.Sprintf("(%d)", *c.TypeLength))
		}
	}

	if !c.Nullable {
		b.WriteString(" NOT NULL")
	}
	if c.DefaultValue != nil && !c.AutoIncrement {
		b.WriteString(" DEFAULT ")
		b.WriteString(formatPGDefaultValue(*c.DefaultValue, dataType))
	}
	if c.Comment != "" {
		// PG 的列注释需要单独的 COMMENT ON 语句，在 CreateTable 中处理
	}
	return b.String(), nil
}

// formatPGDefaultValue 格式化 PG 默认值
func formatPGDefaultValue(v, dataType string) string {
	trim := strings.TrimSpace(v)
	upper := strings.ToUpper(trim)
	if upper == "NULL" {
		return "NULL"
	}
	if upper == "CURRENT_TIMESTAMP" || upper == "NOW()" {
		return upper
	}
	if upper == "TRUE" || upper == "FALSE" {
		return upper
	}
	// 数字类型不加引号
	if isPGNumericType(dataType) && isNumericLiteral(trim) {
		return trim
	}
	return "'" + escapePGString(v) + "'"
}

func isPGSerialType(t string) bool {
	up := strings.ToUpper(strings.TrimSpace(t))
	return up == "SERIAL" || up == "BIGSERIAL" || up == "SMALLSERIAL"
}

func isPGNumericType(t string) bool {
	up := strings.ToUpper(strings.TrimSpace(t))
	switch up {
	case "SMALLINT", "INT2", "INTEGER", "INT", "INT4", "BIGINT", "INT8",
		"DECIMAL", "NUMERIC", "REAL", "FLOAT4", "DOUBLE PRECISION", "FLOAT8",
		"SERIAL", "BIGSERIAL", "SMALLSERIAL":
		return true
	}
	return false
}

func isNumericLiteral(s string) bool {
	if s == "" {
		return false
	}
	hasDot := false
	for i, c := range s {
		if i == 0 && (c == '-' || c == '+') {
			continue
		}
		if c == '.' {
			if hasDot {
				return false
			}
			hasDot = true
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func escapePGString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// CreateTable 生成 CREATE TABLE DDL
func (g *ddlGenerator) CreateTable(spec driver.TableSpec) (string, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return "", fmt.Errorf("postgres/ddl: table name is required")
	}
	if len(spec.Columns) == 0 {
		return "", fmt.Errorf("postgres/ddl: at least one column is required")
	}

	var b strings.Builder
	b.WriteString("CREATE TABLE ")
	b.WriteString(g.qualified(spec.Schema, spec.Name))
	b.WriteString(" (\n")

	parts := make([]string, 0, len(spec.Columns)+len(spec.Indexes)+len(spec.ForeignKeys)+1)

	// 字段定义
	pkFromColumns := []string{}
	commentStmts := []string{}
	for _, c := range spec.Columns {
		def, err := g.renderColumnDef(c)
		if err != nil {
			return "", err
		}
		parts = append(parts, "  "+def)
		if c.AutoIncrement {
			pkFromColumns = append(pkFromColumns, c.Name)
		}
		// PG 列注释单独生成
		if c.Comment != "" {
			commentStmts = append(commentStmts,
				fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s'",
					g.qualified(spec.Schema, spec.Name),
					g.quote(c.Name),
					escapePGString(c.Comment)))
		}
	}

	// 主键
	pk := spec.PrimaryKey
	if len(pk) == 0 {
		pk = pkFromColumns
	}
	if len(pk) > 0 {
		cols := make([]string, len(pk))
		for i, c := range pk {
			cols[i] = g.quote(c)
		}
		parts = append(parts, "  PRIMARY KEY ("+strings.Join(cols, ", ")+")")
	}

	// 外键（PG 中索引需要在表外单独创建）
	for _, fk := range spec.ForeignKeys {
		def, err := g.renderForeignKeyDef(fk, spec.Schema)
		if err != nil {
			return "", err
		}
		parts = append(parts, "  "+def)
	}

	b.WriteString(strings.Join(parts, ",\n"))
	b.WriteString("\n)")

	// PG 表选项较少
	if spec.Comment != "" {
		commentStmts = append([]string{
			fmt.Sprintf("COMMENT ON TABLE %s IS '%s'",
				g.qualified(spec.Schema, spec.Name),
				escapePGString(spec.Comment)),
		}, commentStmts...)
	}

	// 索引需要单独 CREATE INDEX 语句
	indexStmts := make([]string, 0, len(spec.Indexes))
	for _, idx := range spec.Indexes {
		if strings.EqualFold(idx.Type, "PRIMARY") {
			continue
		}
		stmt, err := g.renderCreateIndex(idx, spec.Schema, spec.Name)
		if err != nil {
			return "", err
		}
		indexStmts = append(indexStmts, stmt)
	}

	// 拼接：CREATE TABLE + CREATE INDEX + COMMENT ON
	allStmts := []string{b.String()}
	allStmts = append(allStmts, indexStmts...)
	allStmts = append(allStmts, commentStmts...)

	return strings.Join(allStmts, ";\n") + ";", nil
}

// renderForeignKeyDef 生成外键定义片段（在 CREATE TABLE 内使用）
func (g *ddlGenerator) renderForeignKeyDef(fk driver.ForeignKeySpec, schema string) (string, error) {
	if len(fk.Columns) == 0 || len(fk.ReferencedColumns) == 0 || fk.ReferencedTable == "" {
		return "", fmt.Errorf("postgres/ddl: foreign key %q is incomplete", fk.Name)
	}
	cols := make([]string, len(fk.Columns))
	for i, c := range fk.Columns {
		cols[i] = g.quote(c)
	}
	refCols := make([]string, len(fk.ReferencedColumns))
	for i, c := range fk.ReferencedColumns {
		refCols[i] = g.quote(c)
	}
	refSchema := fk.ReferencedSchema
	if refSchema == "" {
		refSchema = schema
	}
	out := "CONSTRAINT"
	if fk.Name != "" {
		out += " " + g.quote(fk.Name)
	}
	out += " FOREIGN KEY (" + strings.Join(cols, ", ") + ")"
	out += " REFERENCES " + g.qualified(refSchema, fk.ReferencedTable)
	out += " (" + strings.Join(refCols, ", ") + ")"
	if fk.OnDelete != "" {
		out += " ON DELETE " + strings.ToUpper(fk.OnDelete)
	}
	if fk.OnUpdate != "" {
		out += " ON UPDATE " + strings.ToUpper(fk.OnUpdate)
	}
	return out, nil
}

// renderCreateIndex 生成 CREATE INDEX 语句
func (g *ddlGenerator) renderCreateIndex(idx driver.IndexSpec, schema, table string) (string, error) {
	if len(idx.Columns) == 0 {
		return "", fmt.Errorf("postgres/ddl: index %q has no columns", idx.Name)
	}
	cols := make([]string, len(idx.Columns))
	for i, c := range idx.Columns {
		cols[i] = g.quote(c)
	}
	var b strings.Builder
	b.WriteString("CREATE ")
	if strings.EqualFold(idx.Type, "UNIQUE") {
		b.WriteString("UNIQUE ")
	}
	b.WriteString("INDEX ")
	if idx.Name != "" {
		b.WriteString(g.quote(idx.Name))
	}
	b.WriteString(" ON ")
	b.WriteString(g.qualified(schema, table))
	b.WriteString(" (" + strings.Join(cols, ", ") + ")")
	if idx.Comment != "" {
		// PG 不支持索引注释语法，但可以在创建后 COMMENT ON INDEX
		_ = idx.Comment
	}
	return b.String(), nil
}

// AlterTable 生成 ALTER TABLE DDL（多语句，PG 不支持单语句多 action）
func (g *ddlGenerator) AlterTable(spec driver.AlterTableSpec) (string, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return "", fmt.Errorf("postgres/ddl: table name is required")
	}

	var stmts []string
	tbl := g.qualified(spec.Schema, spec.Name)

	// 1. DROP FOREIGN KEY
	for _, fk := range spec.DropForeignKeys {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tbl, g.quote(fk)))
	}

	// 2. DROP INDEX
	for _, idx := range spec.DropIndexes {
		stmts = append(stmts, "DROP INDEX IF EXISTS "+g.quote(idx))
	}

	// 3. DROP COLUMN
	for _, col := range spec.DropColumns {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tbl, g.quote(col)))
	}

	// 4. RENAME COLUMN
	for _, r := range spec.RenameColumns {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tbl, g.quote(r.Old), g.quote(r.New)))
	}

	// 5. MODIFY COLUMN（ALTER COLUMN 系列）
	for _, c := range spec.ModifyColumns {
		// PG ALTER TABLE 不支持 MODIFY COLUMN，需要拆成多条 ALTER COLUMN
		pgAlterCols := g.renderAlterColumnActions(c, spec.Schema, spec.Name)
		stmts = append(stmts, pgAlterCols...)
	}

	// 6. ADD COLUMN
	for _, c := range spec.AddColumns {
		def, err := g.renderColumnDef(c)
		if err != nil {
			return "", err
		}
		add := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tbl, def)
		stmts = append(stmts, add)
	}

	// 7. ADD INDEX
	for _, idx := range spec.AddIndexes {
		stmt, err := g.renderCreateIndex(idx, spec.Schema, spec.Name)
		if err != nil {
			return "", err
		}
		stmts = append(stmts, stmt)
	}

	// 8. ADD FOREIGN KEY
	for _, fk := range spec.AddForeignKeys {
		def, err := g.renderForeignKeyDef(fk, spec.Schema)
		if err != nil {
			return "", err
		}
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD %s", tbl, def))
	}

	// 9. COMMENT
	if spec.Comment != "" {
		stmts = append(stmts, fmt.Sprintf("COMMENT ON TABLE %s IS '%s'", tbl, escapePGString(spec.Comment)))
	}

	if len(stmts) == 0 {
		return "", fmt.Errorf("postgres/ddl: alter table has no actions")
	}

	return strings.Join(stmts, ";\n") + ";", nil
}

// renderAlterColumnActions 将一个字段修改拆成 PG 的 ALTER COLUMN 语句列表
func (g *ddlGenerator) renderAlterColumnActions(c driver.ColumnSpec, schema, table string) []string {
	tbl := g.qualified(schema, table)
	var stmts []string
	colName := g.quote(c.Name)

	// 类型修改
	dataType := strings.ToUpper(c.DataType)
	if c.AutoIncrement {
		switch dataType {
		case "BIGINT", "INT8":
			dataType = "BIGSERIAL"
		case "SMALLINT", "INT2":
			dataType = "SMALLSERIAL"
		default:
			dataType = "SERIAL"
		}
	}
	typeStr := dataType
	if !isPGSerialType(dataType) {
		switch {
		case c.Precision != nil && c.Scale != nil:
			typeStr += fmt.Sprintf("(%d,%d)", *c.Precision, *c.Scale)
		case c.Precision != nil:
			typeStr += fmt.Sprintf("(%d)", *c.Precision)
		case c.TypeLength != nil:
			typeStr += fmt.Sprintf("(%d)", *c.TypeLength)
		}
	}
	stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", tbl, colName, typeStr))

	// NOT NULL
	if !c.Nullable {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", tbl, colName))
	} else {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", tbl, colName))
	}

	// DEFAULT（SERIAL 类型自带 nextval 默认值，无需显式 SET DEFAULT）
	if c.DefaultValue != nil && !c.AutoIncrement && !isPGSerialType(dataType) {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s",
			tbl, colName, formatPGDefaultValue(*c.DefaultValue, dataType)))
	}

	// COMMENT
	if c.Comment != "" {
		stmts = append(stmts, fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s'",
			tbl, colName, escapePGString(c.Comment)))
	}

	return stmts
}

// DropTable 生成 DROP TABLE DDL
func (g *ddlGenerator) DropTable(database, schema, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("postgres/ddl: table name is required")
	}
	_ = database
	return "DROP TABLE " + g.qualified(schema, name), nil
}

// RenameTable 生成 ALTER TABLE RENAME DDL
func (g *ddlGenerator) RenameTable(database, schema, oldName, newName string) (string, error) {
	if oldName == "" || newName == "" {
		return "", fmt.Errorf("postgres/ddl: rename requires both names")
	}
	_ = database
	return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", g.qualified(schema, oldName), g.quote(newName)), nil
}

// GetCreateTableDDL 使用 pg_dump 风格重建 DDL（PG 没有 SHOW CREATE TABLE）
// 方案：从 INFORMATION_SCHEMA + pg_catalog 查询，拼接 DDL
func (g *ddlGenerator) GetCreateTableDDL(ctx context.Context, database, schema, table string) (string, error) {
	if table == "" {
		return "", fmt.Errorf("postgres/ddl: table name is required")
	}
	sch := schema
	if sch == "" {
		sch = "public"
	}

	// 获取表详情
	detail, err := g.conn.Inspector().DescribeTable(ctx, database, sch, table)
	if err != nil {
		return "", fmt.Errorf("postgres/ddl: describe table: %w", err)
	}

	// 用 CreateTable 生成
	spec := driver.TableSpec{
		Name:     detail.Name,
		Schema:   sch,
		Database: database,
		Comment:  detail.Comment,
	}
	for _, col := range detail.Columns {
		spec.Columns = append(spec.Columns, driver.ColumnSpec{
			Name:          col.Name,
			DataType:      col.DatabaseType,
			Nullable:      col.Nullable,
			AutoIncrement: col.AutoIncrement,
			DefaultValue:  col.DefaultValue,
			Comment:       col.Comment,
			TypeLength:    col.TypeLength,
			Precision:     col.Precision,
			Scale:         col.Scale,
		})
		if col.IsPrimary {
			spec.PrimaryKey = append(spec.PrimaryKey, col.Name)
		}
	}
	for _, idx := range detail.Indexes {
		if idx.Primary {
			continue
		}
		spec.Indexes = append(spec.Indexes, driver.IndexSpec{
			Name:    idx.Name,
			Columns: idx.Columns,
			Type:    "NORMAL",
		})
		if idx.Unique {
			spec.Indexes[len(spec.Indexes)-1].Type = "UNIQUE"
		}
	}
	for _, fk := range detail.ForeignKeys {
		spec.ForeignKeys = append(spec.ForeignKeys, driver.ForeignKeySpec{
			Name:              fk.Name,
			Columns:           fk.Columns,
			ReferencedSchema:  fk.ReferencedSchema,
			ReferencedTable:   fk.ReferencedTable,
			ReferencedColumns: fk.ReferencedColumns,
			OnDelete:          fk.OnDelete,
			OnUpdate:          fk.OnUpdate,
		})
	}

	return g.CreateTable(spec)
}
