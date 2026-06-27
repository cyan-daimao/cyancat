package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cyancat/internal/infra/driver"
)

type inspector struct {
	conn *Conn
}

func newInspector(c *Conn) driver.Inspector {
	return &inspector{conn: c}
}

func (i *inspector) ListDatabases(ctx context.Context) ([]driver.Database, error) {
	_ = ctx
	return []driver.Database{{Name: "main"}}, nil
}

func (i *inspector) ListSchemas(ctx context.Context, database string) ([]driver.Schema, error) {
	_ = ctx
	_ = database
	return []driver.Schema{{Name: "main"}}, nil
}

func (i *inspector) ListTables(ctx context.Context, database, schema string) ([]driver.Table, error) {
	const q = `SELECT name FROM sqlite_master
		WHERE type = 'table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name`
	rows, err := i.conn.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("sqlite/inspector: list tables: %w", err)
	}
	defer rows.Close()

	var result []driver.Table
	for rows.Next() {
		var t driver.Table
		if err := rows.Scan(&t.Name); err != nil {
			return nil, err
		}
		t.Type = "TABLE"
		result = append(result, t)
	}
	return result, rows.Err()
}

func (i *inspector) ListViews(ctx context.Context, database, schema string) ([]driver.View, error) {
	const q = `SELECT name, COALESCE(sql, '') FROM sqlite_master
		WHERE type = 'view' AND name NOT LIKE 'sqlite_%'
		ORDER BY name`
	rows, err := i.conn.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("sqlite/inspector: list views: %w", err)
	}
	defer rows.Close()

	var result []driver.View
	for rows.Next() {
		var v driver.View
		if err := rows.Scan(&v.Name, &v.Definition); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (i *inspector) DescribeTable(ctx context.Context, database, schema, table string) (*driver.TableDetail, error) {
	if strings.TrimSpace(table) == "" {
		return nil, fmt.Errorf("sqlite/inspector: table name is required")
	}
	detail := &driver.TableDetail{
		Name:     table,
		Schema:   "main",
		Database: "main",
	}

	cols, err := i.listColumns(ctx, table)
	if err != nil {
		return nil, err
	}
	detail.Columns = cols

	indexes, err := i.ListIndexes(ctx, "main", "main", table)
	if err != nil {
		return nil, err
	}
	detail.Indexes = indexes

	fks, err := i.ListForeignKeys(ctx, "main", "main", table)
	if err != nil {
		return nil, err
	}
	detail.ForeignKeys = fks

	return detail, nil
}

func (i *inspector) listColumns(ctx context.Context, table string) ([]driver.Column, error) {
	rows, err := i.conn.db.QueryContext(ctx, "PRAGMA table_info("+quoteIdent(table)+")")
	if err != nil {
		return nil, fmt.Errorf("sqlite/inspector: describe columns: %w", err)
	}
	defer rows.Close()

	var result []driver.Column
	for rows.Next() {
		var (
			cid       int
			name      string
			dataType  string
			notNull   int
			dfltValue sql.NullString
			pk        int
		)
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		col := driver.Column{
			Name:            name,
			DatabaseType:    dataType,
			Nullable:        notNull == 0 && pk == 0,
			IsPrimary:       pk > 0,
			AutoIncrement:   false,
			OrdinalPosition: cid + 1,
		}
		if dfltValue.Valid {
			v := dfltValue.String
			col.DefaultValue = &v
		}
		fillTypeHints(&col)
		result = append(result, col)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	i.markAutoIncrement(ctx, table, result)
	return result, nil
}

func (i *inspector) ListIndexes(ctx context.Context, database, schema, table string) ([]driver.Index, error) {
	if strings.TrimSpace(table) == "" {
		return nil, fmt.Errorf("sqlite/inspector: table name is required")
	}
	rows, err := i.conn.db.QueryContext(ctx, "PRAGMA index_list("+quoteIdent(table)+")")
	if err != nil {
		return nil, fmt.Errorf("sqlite/inspector: list indexes: %w", err)
	}
	defer rows.Close()

	var result []driver.Index
	for rows.Next() {
		var (
			seq     int
			name    string
			unique  int
			origin  string
			partial int
		)
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return nil, err
		}
		idx := driver.Index{
			Name:    name,
			Unique:  unique == 1,
			Primary: origin == "pk",
		}
		result = append(result, idx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	for idx := range result {
		cols, err := i.indexColumns(ctx, result[idx].Name)
		if err != nil {
			return nil, err
		}
		result[idx].Columns = cols
	}
	return result, nil
}

func (i *inspector) indexColumns(ctx context.Context, indexName string) ([]string, error) {
	rows, err := i.conn.db.QueryContext(ctx, "PRAGMA index_info("+quoteIdent(indexName)+")")
	if err != nil {
		return nil, fmt.Errorf("sqlite/inspector: index columns: %w", err)
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var seqno, cid int
		var name string
		if err := rows.Scan(&seqno, &cid, &name); err != nil {
			return nil, err
		}
		cols = append(cols, name)
	}
	return cols, rows.Err()
}

func (i *inspector) ListForeignKeys(ctx context.Context, database, schema, table string) ([]driver.ForeignKey, error) {
	if strings.TrimSpace(table) == "" {
		return nil, fmt.Errorf("sqlite/inspector: table name is required")
	}
	rows, err := i.conn.db.QueryContext(ctx, "PRAGMA foreign_key_list("+quoteIdent(table)+")")
	if err != nil {
		return nil, fmt.Errorf("sqlite/inspector: list fks: %w", err)
	}
	defer rows.Close()

	type fkPart struct {
		seq      int
		from     string
		to       string
		refTable string
		onUpdate string
		onDelete string
	}
	byID := map[int][]fkPart{}
	var order []int
	for rows.Next() {
		var (
			id       int
			seq      int
			refTable string
			from     string
			to       string
			onUpdate string
			onDelete string
			match    string
		)
		if err := rows.Scan(&id, &seq, &refTable, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			return nil, err
		}
		if _, ok := byID[id]; !ok {
			order = append(order, id)
		}
		byID[id] = append(byID[id], fkPart{
			seq: seq, from: from, to: to, refTable: refTable,
			onUpdate: onUpdate, onDelete: onDelete,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]driver.ForeignKey, 0, len(order))
	for _, id := range order {
		parts := byID[id]
		fk := driver.ForeignKey{
			Name:             fmt.Sprintf("fk_%s_%d", table, id),
			ReferencedSchema: "main",
			ReferencedTable:  parts[0].refTable,
			OnUpdate:         parts[0].onUpdate,
			OnDelete:         parts[0].onDelete,
		}
		for _, p := range parts {
			fk.Columns = append(fk.Columns, p.from)
			fk.ReferencedColumns = append(fk.ReferencedColumns, p.to)
		}
		result = append(result, fk)
	}
	return result, nil
}

func (i *inspector) ListCharsets(ctx context.Context) ([]driver.Charset, error) {
	_ = ctx
	return []driver.Charset{{Name: "UTF-8", Description: "SQLite text encoding", DefaultCollation: "BINARY"}}, nil
}

func (i *inspector) ListCollations(ctx context.Context, charset string) ([]driver.Collation, error) {
	_ = ctx
	_ = charset
	return []driver.Collation{
		{Name: "BINARY", Charset: "UTF-8", IsDefault: true},
		{Name: "NOCASE", Charset: "UTF-8"},
		{Name: "RTRIM", Charset: "UTF-8"},
	}, nil
}

func quoteIdent(ident string) string {
	var b strings.Builder
	b.Grow(len(ident) + 2)
	b.WriteByte('"')
	for i := 0; i < len(ident); i++ {
		if ident[i] == '"' {
			b.WriteString(`""`)
		} else {
			b.WriteByte(ident[i])
		}
	}
	b.WriteByte('"')
	return b.String()
}

func fillTypeHints(col *driver.Column) {
	raw := strings.ToUpper(strings.TrimSpace(col.DatabaseType))
	if raw == "" {
		return
	}
	if open := strings.IndexByte(raw, '('); open > 0 && strings.HasSuffix(raw, ")") {
		args := strings.TrimSuffix(raw[open+1:], ")")
		col.DatabaseType = raw[:open] + "(" + args + ")"
		parts := strings.Split(args, ",")
		if len(parts) == 1 {
			if v, ok := parsePositiveInt(parts[0]); ok {
				col.TypeLength = &v
			}
		} else if len(parts) == 2 {
			if v, ok := parsePositiveInt(parts[0]); ok {
				col.Precision = &v
			}
			if v, ok := parsePositiveInt(parts[1]); ok {
				col.Scale = &v
			}
		}
	}
}

func parsePositiveInt(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int(s[i]-'0')
	}
	return n, n > 0
}

func (i *inspector) markAutoIncrement(ctx context.Context, table string, columns []driver.Column) {
	var ddl sql.NullString
	err := i.conn.db.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type IN ('table', 'view') AND name = ?`, table).Scan(&ddl)
	if err != nil || !ddl.Valid {
		return
	}
	upperDDL := strings.ToUpper(ddl.String)
	for idx := range columns {
		if columns[idx].IsPrimary &&
			strings.Contains(strings.ToUpper(columns[idx].DatabaseType), "INTEGER") &&
			strings.Contains(upperDDL, "AUTOINCREMENT") {
			columns[idx].AutoIncrement = true
		}
	}
}
