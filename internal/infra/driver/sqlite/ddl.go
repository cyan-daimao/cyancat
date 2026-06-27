package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cyancat/internal/infra/driver"
)

type ddlGenerator struct {
	conn    *Conn
	dialect driver.Dialect
}

func newDDLGenerator(c *Conn) driver.DDLGenerator {
	return &ddlGenerator{conn: c, dialect: &sqliteDialect{}}
}

func (g *ddlGenerator) quote(ident string) string {
	return g.dialect.QuoteIdent(ident)
}

func (g *ddlGenerator) qualified(schema, name string) string {
	if schema != "" && schema != "main" {
		return g.quote(schema) + "." + g.quote(name)
	}
	return g.quote(name)
}

func (g *ddlGenerator) CreateDatabase(spec driver.DatabaseSpec) (string, error) {
	_ = spec
	return "", fmt.Errorf("sqlite/ddl: create database is not supported")
}

func (g *ddlGenerator) DropDatabase(name string) (string, error) {
	_ = name
	return "", fmt.Errorf("sqlite/ddl: drop database is not supported")
}

func (g *ddlGenerator) CreateTable(spec driver.TableSpec) (string, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return "", fmt.Errorf("sqlite/ddl: table name is required")
	}
	if len(spec.Columns) == 0 {
		return "", fmt.Errorf("sqlite/ddl: at least one column is required")
	}

	var b strings.Builder
	b.WriteString("CREATE TABLE ")
	b.WriteString(g.qualified(spec.Schema, spec.Name))
	b.WriteString(" (\n")

	parts := make([]string, 0, len(spec.Columns)+len(spec.Indexes)+len(spec.ForeignKeys)+1)
	autoIncrementPK := map[string]bool{}
	for _, c := range spec.Columns {
		def, err := g.renderColumnDef(c)
		if err != nil {
			return "", err
		}
		parts = append(parts, "  "+def)
		if c.AutoIncrement {
			autoIncrementPK[c.Name] = true
		}
	}

	pk := make([]string, 0, len(spec.PrimaryKey))
	for _, col := range spec.PrimaryKey {
		if !autoIncrementPK[col] {
			pk = append(pk, col)
		}
	}
	if len(pk) > 0 {
		cols := make([]string, len(pk))
		for i, c := range pk {
			cols[i] = g.quote(c)
		}
		parts = append(parts, "  PRIMARY KEY ("+strings.Join(cols, ", ")+")")
	}

	for _, fk := range spec.ForeignKeys {
		def, err := g.renderForeignKeyDef(fk)
		if err != nil {
			return "", err
		}
		parts = append(parts, "  "+def)
	}

	b.WriteString(strings.Join(parts, ",\n"))
	b.WriteString("\n)")

	stmts := []string{b.String()}
	for _, idx := range spec.Indexes {
		if strings.EqualFold(idx.Type, "PRIMARY") {
			continue
		}
		stmt, err := g.renderCreateIndex(idx, spec.Schema, spec.Name)
		if err != nil {
			return "", err
		}
		stmts = append(stmts, stmt)
	}

	return strings.Join(stmts, ";\n") + ";", nil
}

func (g *ddlGenerator) renderColumnDef(c driver.ColumnSpec) (string, error) {
	if strings.TrimSpace(c.Name) == "" {
		return "", fmt.Errorf("sqlite/ddl: column name is required")
	}
	if strings.TrimSpace(c.DataType) == "" {
		return "", fmt.Errorf("sqlite/ddl: column %q data type is required", c.Name)
	}

	dataType := strings.ToUpper(strings.TrimSpace(c.DataType))
	var b strings.Builder
	b.WriteString(g.quote(c.Name))
	b.WriteString(" ")
	if c.AutoIncrement {
		b.WriteString("INTEGER")
	} else {
		b.WriteString(dataType)
		switch {
		case c.Precision != nil && c.Scale != nil:
			b.WriteString(fmt.Sprintf("(%d,%d)", *c.Precision, *c.Scale))
		case c.Precision != nil:
			b.WriteString(fmt.Sprintf("(%d)", *c.Precision))
		case c.TypeLength != nil:
			b.WriteString(fmt.Sprintf("(%d)", *c.TypeLength))
		}
	}

	if c.AutoIncrement {
		b.WriteString(" PRIMARY KEY AUTOINCREMENT")
	} else if !c.Nullable {
		b.WriteString(" NOT NULL")
	}
	if c.DefaultValue != nil && !c.AutoIncrement {
		b.WriteString(" DEFAULT ")
		b.WriteString(formatSQLiteDefaultValue(*c.DefaultValue, dataType))
	}
	return b.String(), nil
}

func (g *ddlGenerator) renderForeignKeyDef(fk driver.ForeignKeySpec) (string, error) {
	if len(fk.Columns) == 0 || len(fk.ReferencedColumns) == 0 || fk.ReferencedTable == "" {
		return "", fmt.Errorf("sqlite/ddl: foreign key %q is incomplete", fk.Name)
	}
	cols := quoteList(g, fk.Columns)
	refCols := quoteList(g, fk.ReferencedColumns)
	out := ""
	if fk.Name != "" {
		out += "CONSTRAINT " + g.quote(fk.Name) + " "
	}
	out += "FOREIGN KEY (" + strings.Join(cols, ", ") + ")"
	out += " REFERENCES " + g.quote(fk.ReferencedTable)
	out += " (" + strings.Join(refCols, ", ") + ")"
	if fk.OnDelete != "" {
		out += " ON DELETE " + strings.ToUpper(fk.OnDelete)
	}
	if fk.OnUpdate != "" {
		out += " ON UPDATE " + strings.ToUpper(fk.OnUpdate)
	}
	return out, nil
}

func (g *ddlGenerator) renderCreateIndex(idx driver.IndexSpec, schema, table string) (string, error) {
	if len(idx.Columns) == 0 {
		return "", fmt.Errorf("sqlite/ddl: index %q has no columns", idx.Name)
	}
	cols := quoteList(g, idx.Columns)
	name := idx.Name
	if name == "" {
		name = table + "_" + strings.Join(idx.Columns, "_") + "_idx"
	}
	var b strings.Builder
	b.WriteString("CREATE ")
	if strings.EqualFold(idx.Type, "UNIQUE") {
		b.WriteString("UNIQUE ")
	}
	b.WriteString("INDEX ")
	b.WriteString(g.quote(name))
	b.WriteString(" ON ")
	b.WriteString(g.qualified(schema, table))
	b.WriteString(" (" + strings.Join(cols, ", ") + ")")
	return b.String(), nil
}

func (g *ddlGenerator) AlterTable(spec driver.AlterTableSpec) (string, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return "", fmt.Errorf("sqlite/ddl: table name is required")
	}
	if len(spec.ModifyColumns) > 0 || len(spec.DropColumns) > 0 || len(spec.ModifyIndexes) > 0 ||
		spec.Engine != "" || spec.Charset != "" || spec.Collation != "" || spec.Comment != "" {
		return "", fmt.Errorf("sqlite/ddl: complex alter table is not supported")
	}

	tbl := g.qualified(spec.Schema, spec.Name)
	var stmts []string
	for _, r := range spec.RenameColumns {
		if r.Old == "" || r.New == "" {
			return "", fmt.Errorf("sqlite/ddl: rename column requires both names")
		}
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tbl, g.quote(r.Old), g.quote(r.New)))
	}
	for _, c := range spec.AddColumns {
		def, err := g.renderColumnDef(c)
		if err != nil {
			return "", err
		}
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tbl, def))
	}
	for _, idx := range spec.DropIndexes {
		stmts = append(stmts, "DROP INDEX "+g.quote(idx))
	}
	for _, idx := range spec.AddIndexes {
		stmt, err := g.renderCreateIndex(idx, spec.Schema, spec.Name)
		if err != nil {
			return "", err
		}
		stmts = append(stmts, stmt)
	}
	if len(spec.DropForeignKeys) > 0 || len(spec.AddForeignKeys) > 0 {
		return "", fmt.Errorf("sqlite/ddl: alter foreign keys requires table rebuild and is not supported")
	}
	if len(stmts) == 0 {
		return "", fmt.Errorf("sqlite/ddl: alter table has no actions")
	}
	return strings.Join(stmts, ";\n") + ";", nil
}

func (g *ddlGenerator) DropTable(database, schema, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("sqlite/ddl: table name is required")
	}
	_ = database
	return "DROP TABLE " + g.qualified(schema, name), nil
}

func (g *ddlGenerator) RenameTable(database, schema, oldName, newName string) (string, error) {
	if oldName == "" || newName == "" {
		return "", fmt.Errorf("sqlite/ddl: rename requires both names")
	}
	_ = database
	return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", g.qualified(schema, oldName), g.quote(newName)), nil
}

func (g *ddlGenerator) GetCreateTableDDL(ctx context.Context, database, schema, table string) (string, error) {
	if strings.TrimSpace(table) == "" {
		return "", fmt.Errorf("sqlite/ddl: table name is required")
	}
	var ddl string
	err := g.conn.db.QueryRowContext(ctx,
		`SELECT sql FROM sqlite_master WHERE type IN ('table', 'view') AND name = ?`,
		table,
	).Scan(&ddl)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("sqlite/ddl: table not found")
		}
		return "", fmt.Errorf("sqlite/ddl: get create table: %w", err)
	}
	return ddl, nil
}

func quoteList(g *ddlGenerator, names []string) []string {
	out := make([]string, len(names))
	for i, name := range names {
		out[i] = g.quote(name)
	}
	return out
}

func formatSQLiteDefaultValue(v, dataType string) string {
	trim := strings.TrimSpace(v)
	upper := strings.ToUpper(trim)
	if upper == "NULL" || upper == "CURRENT_TIMESTAMP" || upper == "CURRENT_DATE" || upper == "CURRENT_TIME" {
		return upper
	}
	if upper == "TRUE" || upper == "FALSE" {
		return upper
	}
	if isSQLiteNumericType(dataType) && isNumericLiteral(trim) {
		return trim
	}
	if strings.HasPrefix(trim, "(") && strings.HasSuffix(trim, ")") {
		return trim
	}
	return "'" + escapeSQLiteString(v) + "'"
}

func isSQLiteNumericType(t string) bool {
	up := strings.ToUpper(strings.TrimSpace(t))
	switch up {
	case "INTEGER", "INT", "BIGINT", "SMALLINT", "TINYINT", "REAL", "DOUBLE", "FLOAT", "NUMERIC", "DECIMAL":
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

func escapeSQLiteString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
