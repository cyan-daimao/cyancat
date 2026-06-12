package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cyancat/internal/infra/driver"
)

// ddlGenerator MySQL DDL 生成器实现
type ddlGenerator struct {
	conn    *Conn
	dialect driver.Dialect
}

func newDDLGenerator(c *Conn) driver.DDLGenerator {
	return &ddlGenerator{
		conn:    c,
		dialect: &mysqlDialect{},
	}
}

// quote 简写
func (g *ddlGenerator) quote(ident string) string {
	return g.dialect.QuoteIdent(ident)
}

// qualified 拼接 `db`.`table` 形式
func (g *ddlGenerator) qualified(database, name string) string {
	if database != "" {
		return g.quote(database) + "." + g.quote(name)
	}
	return g.quote(name)
}

// CreateDatabase 生成 CREATE DATABASE DDL
func (g *ddlGenerator) CreateDatabase(spec driver.DatabaseSpec) (string, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return "", fmt.Errorf("mysql/ddl: database name is required")
	}
	var b strings.Builder
	b.WriteString("CREATE DATABASE ")
	b.WriteString(g.quote(spec.Name))
	if spec.Charset != "" {
		b.WriteString(" DEFAULT CHARACTER SET ")
		b.WriteString(spec.Charset)
	}
	if spec.Collation != "" {
		b.WriteString(" COLLATE ")
		b.WriteString(spec.Collation)
	}
	return b.String(), nil
}

// DropDatabase 生成 DROP DATABASE DDL
func (g *ddlGenerator) DropDatabase(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("mysql/ddl: database name is required")
	}
	return "DROP DATABASE " + g.quote(name), nil
}

// renderColumnDef 生成单个字段定义片段
// 形如：`name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '...'
func (g *ddlGenerator) renderColumnDef(c driver.ColumnSpec) (string, error) {
	if strings.TrimSpace(c.Name) == "" {
		return "", fmt.Errorf("mysql/ddl: column name is required")
	}
	if strings.TrimSpace(c.DataType) == "" {
		return "", fmt.Errorf("mysql/ddl: column %q data type is required", c.Name)
	}
	var b strings.Builder
	b.WriteString(g.quote(c.Name))
	b.WriteString(" ")
	b.WriteString(strings.ToUpper(c.DataType))

	// 拼长度/精度
	switch {
	case c.Precision != nil && c.Scale != nil:
		b.WriteString(fmt.Sprintf("(%d,%d)", *c.Precision, *c.Scale))
	case c.Precision != nil:
		b.WriteString(fmt.Sprintf("(%d)", *c.Precision))
	case c.TypeLength != nil:
		b.WriteString(fmt.Sprintf("(%d)", *c.TypeLength))
	}

	if c.Unsigned {
		b.WriteString(" UNSIGNED")
	}
	if c.Collation != "" {
		b.WriteString(" COLLATE ")
		b.WriteString(c.Collation)
	}
	if !c.Nullable {
		b.WriteString(" NOT NULL")
	} else {
		b.WriteString(" NULL")
	}
	if c.DefaultValue != nil {
		b.WriteString(" DEFAULT ")
		b.WriteString(formatDefaultValue(*c.DefaultValue, c.DataType))
	}
	if c.AutoIncrement {
		b.WriteString(" AUTO_INCREMENT")
	}
	if c.Comment != "" {
		b.WriteString(" COMMENT '")
		b.WriteString(escapeMySQLString(c.Comment))
		b.WriteString("'")
	}
	return b.String(), nil
}

// formatDefaultValue 格式化默认值。NULL/CURRENT_TIMESTAMP/数字直接写入，否则加引号
func formatDefaultValue(v, dataType string) string {
	trim := strings.TrimSpace(v)
	upper := strings.ToUpper(trim)
	if upper == "NULL" || upper == "CURRENT_TIMESTAMP" || upper == "CURRENT_TIMESTAMP()" {
		return upper
	}
	// 数字类型不加引号
	if isNumericType(dataType) && isNumericLiteral(trim) {
		return trim
	}
	return "'" + escapeMySQLString(v) + "'"
}

func isNumericType(t string) bool {
	up := strings.ToUpper(strings.TrimSpace(t))
	switch up {
	case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT",
		"DECIMAL", "NUMERIC", "FLOAT", "DOUBLE", "REAL", "BIT":
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

func escapeMySQLString(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\\"), "'", "''")
}

// CreateTable 生成 CREATE TABLE DDL
func (g *ddlGenerator) CreateTable(spec driver.TableSpec) (string, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return "", fmt.Errorf("mysql/ddl: table name is required")
	}
	if len(spec.Columns) == 0 {
		return "", fmt.Errorf("mysql/ddl: at least one column is required")
	}

	var b strings.Builder
	b.WriteString("CREATE TABLE ")
	b.WriteString(g.qualified(spec.Database, spec.Name))
	b.WriteString(" (\n")

	parts := make([]string, 0, len(spec.Columns)+len(spec.Indexes)+len(spec.ForeignKeys)+1)

	// 字段定义
	pkFromColumns := []string{}
	for _, c := range spec.Columns {
		def, err := g.renderColumnDef(c)
		if err != nil {
			return "", err
		}
		parts = append(parts, "  "+def)
		// 自增字段隐式带主键（如果未在 PrimaryKey 中声明）
		if c.AutoIncrement {
			pkFromColumns = append(pkFromColumns, c.Name)
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

	// 索引
	for _, idx := range spec.Indexes {
		if strings.EqualFold(idx.Type, "PRIMARY") {
			continue // 主键已经处理
		}
		segs, err := g.renderIndexDef(idx)
		if err != nil {
			return "", err
		}
		parts = append(parts, "  "+segs)
	}

	// 外键
	for _, fk := range spec.ForeignKeys {
		segs, err := g.renderForeignKeyDef(fk)
		if err != nil {
			return "", err
		}
		parts = append(parts, "  "+segs)
	}

	b.WriteString(strings.Join(parts, ",\n"))
	b.WriteString("\n)")

	// 表选项
	if spec.Engine != "" {
		b.WriteString(" ENGINE=")
		b.WriteString(spec.Engine)
	}
	if spec.Charset != "" {
		b.WriteString(" DEFAULT CHARSET=")
		b.WriteString(spec.Charset)
	}
	if spec.Collation != "" {
		b.WriteString(" COLLATE=")
		b.WriteString(spec.Collation)
	}
	if spec.Comment != "" {
		b.WriteString(" COMMENT='")
		b.WriteString(escapeMySQLString(spec.Comment))
		b.WriteString("'")
	}
	return b.String(), nil
}

// renderIndexDef 生成索引定义片段
func (g *ddlGenerator) renderIndexDef(idx driver.IndexSpec) (string, error) {
	if len(idx.Columns) == 0 {
		return "", fmt.Errorf("mysql/ddl: index %q has no columns", idx.Name)
	}
	cols := make([]string, len(idx.Columns))
	for i, c := range idx.Columns {
		cols[i] = g.quote(c)
	}
	var prefix string
	switch strings.ToUpper(idx.Type) {
	case "UNIQUE":
		prefix = "UNIQUE KEY"
	case "FULLTEXT":
		prefix = "FULLTEXT KEY"
	default:
		prefix = "KEY"
	}
	out := prefix
	if idx.Name != "" {
		out += " " + g.quote(idx.Name)
	}
	out += " (" + strings.Join(cols, ", ") + ")"
	if idx.Comment != "" {
		out += " COMMENT '" + escapeMySQLString(idx.Comment) + "'"
	}
	return out, nil
}

// renderForeignKeyDef 生成外键定义片段
func (g *ddlGenerator) renderForeignKeyDef(fk driver.ForeignKeySpec) (string, error) {
	if len(fk.Columns) == 0 || len(fk.ReferencedColumns) == 0 || fk.ReferencedTable == "" {
		return "", fmt.Errorf("mysql/ddl: foreign key %q is incomplete", fk.Name)
	}
	cols := make([]string, len(fk.Columns))
	for i, c := range fk.Columns {
		cols[i] = g.quote(c)
	}
	refCols := make([]string, len(fk.ReferencedColumns))
	for i, c := range fk.ReferencedColumns {
		refCols[i] = g.quote(c)
	}
	out := "CONSTRAINT"
	if fk.Name != "" {
		out += " " + g.quote(fk.Name)
	}
	out += " FOREIGN KEY (" + strings.Join(cols, ", ") + ")"
	out += " REFERENCES " + g.qualified(fk.ReferencedSchema, fk.ReferencedTable)
	out += " (" + strings.Join(refCols, ", ") + ")"
	if fk.OnDelete != "" {
		out += " ON DELETE " + strings.ToUpper(fk.OnDelete)
	}
	if fk.OnUpdate != "" {
		out += " ON UPDATE " + strings.ToUpper(fk.OnUpdate)
	}
	return out, nil
}

// AlterTable 生成 ALTER TABLE DDL（单语句多 action）
// 动作顺序：DROP FK -> DROP INDEX -> DROP COLUMN -> RENAME COLUMN -> MODIFY COLUMN -> ADD COLUMN -> ADD INDEX -> ADD FK -> 表选项
func (g *ddlGenerator) AlterTable(spec driver.AlterTableSpec) (string, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return "", fmt.Errorf("mysql/ddl: table name is required")
	}
	var actions []string

	for _, fk := range spec.DropForeignKeys {
		actions = append(actions, "DROP FOREIGN KEY "+g.quote(fk))
	}
	for _, idx := range spec.DropIndexes {
		actions = append(actions, "DROP INDEX "+g.quote(idx))
	}
	// 修改索引 = DROP + ADD（MySQL 不支持 ALTER INDEX COMMENT）
	for _, idx := range spec.ModifyIndexes {
		if idx.Name == "" {
			return "", fmt.Errorf("mysql/ddl: modify index requires name")
		}
		actions = append(actions, "DROP INDEX "+g.quote(idx.Name))
	}
	for _, col := range spec.DropColumns {
		actions = append(actions, "DROP COLUMN "+g.quote(col))
	}
	for _, r := range spec.RenameColumns {
		actions = append(actions, "RENAME COLUMN "+g.quote(r.Old)+" TO "+g.quote(r.New))
	}
	for _, c := range spec.ModifyColumns {
		def, err := g.renderColumnDef(c)
		if err != nil {
			return "", err
		}
		actions = append(actions, "MODIFY COLUMN "+def)
	}
	for _, c := range spec.AddColumns {
		def, err := g.renderColumnDef(c)
		if err != nil {
			return "", err
		}
		add := "ADD COLUMN " + def
		switch {
		case c.First:
			add += " FIRST"
		case c.After != "":
			add += " AFTER " + g.quote(c.After)
		}
		actions = append(actions, add)
	}
	for _, idx := range spec.AddIndexes {
		def, err := g.renderIndexDef(idx)
		if err != nil {
			return "", err
		}
		actions = append(actions, "ADD "+def)
	}
	for _, fk := range spec.AddForeignKeys {
		def, err := g.renderForeignKeyDef(fk)
		if err != nil {
			return "", err
		}
		actions = append(actions, "ADD "+def)
	}
	if spec.Engine != "" {
		actions = append(actions, "ENGINE="+spec.Engine)
	}
	if spec.Charset != "" {
		actions = append(actions, "DEFAULT CHARSET="+spec.Charset)
	}
	if spec.Collation != "" {
		actions = append(actions, "COLLATE="+spec.Collation)
	}
	if spec.Comment != "" {
		actions = append(actions, "COMMENT='"+escapeMySQLString(spec.Comment)+"'")
	}

	if len(actions) == 0 {
		return "", fmt.Errorf("mysql/ddl: alter table has no actions")
	}

	var b strings.Builder
	b.WriteString("ALTER TABLE ")
	b.WriteString(g.qualified(spec.Database, spec.Name))
	b.WriteString("\n  ")
	b.WriteString(strings.Join(actions, ",\n  "))
	return b.String(), nil
}

// DropTable 生成 DROP TABLE DDL
func (g *ddlGenerator) DropTable(database, schema, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("mysql/ddl: table name is required")
	}
	_ = schema
	return "DROP TABLE " + g.qualified(database, name), nil
}

// RenameTable 生成 RENAME TABLE DDL
func (g *ddlGenerator) RenameTable(database, schema, oldName, newName string) (string, error) {
	if oldName == "" || newName == "" {
		return "", fmt.Errorf("mysql/ddl: rename requires both names")
	}
	_ = schema
	return "RENAME TABLE " + g.qualified(database, oldName) + " TO " + g.qualified(database, newName), nil
}

// GetCreateTableDDL 调 SHOW CREATE TABLE
func (g *ddlGenerator) GetCreateTableDDL(ctx context.Context, database, schema, table string) (string, error) {
	if table == "" {
		return "", fmt.Errorf("mysql/ddl: table name is required")
	}
	_ = schema
	q := "SHOW CREATE TABLE " + g.qualified(database, table)
	row := g.conn.db.QueryRowContext(ctx, q)
	var name, ddl string
	if err := row.Scan(&name, &ddl); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("mysql/ddl: table not found")
		}
		return "", fmt.Errorf("mysql/ddl: show create table: %w", err)
	}
	return ddl, nil
}
