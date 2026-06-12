package postgres

import (
	"context"
	"fmt"

	"cyancat/internal/infra/driver"
)

// inspector 基于 pg_catalog 的元数据查询器
type inspector struct {
	conn *Conn
}

func newInspector(c *Conn) driver.Inspector {
	return &inspector{conn: c}
}

// ListDatabases 列出所有数据库
func (i *inspector) ListDatabases(ctx context.Context) ([]driver.Database, error) {
	const q = `SELECT datname, pg_encoding_to_char(encoding), datcollate
		FROM pg_database WHERE datistemplate = false ORDER BY datname`

	rows, err := i.conn.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("pg/inspector: list databases: %w", err)
	}
	defer rows.Close()

	var result []driver.Database
	for rows.Next() {
		var d driver.Database
		var coll sqlStr
		if err := rows.Scan(&d.Name, &d.Charset, &coll); err != nil {
			return nil, err
		}
		d.Collation = string(coll)
		result = append(result, d)
	}
	return result, rows.Err()
}

// ListSchemas 列出指定数据库下的 schema（当前 PG 默认取 public 即可，这里列出全部）
func (i *inspector) ListSchemas(ctx context.Context, database string) ([]driver.Schema, error) {
	const q = `SELECT nspname, nspowner::regrole::text
		FROM pg_catalog.pg_namespace
		WHERE nspname NOT LIKE 'pg_%' AND nspname != 'information_schema'
		ORDER BY nspname`

	rows, err := i.conn.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("pg/inspector: list schemas: %w", err)
	}
	defer rows.Close()

	var result []driver.Schema
	for rows.Next() {
		var s driver.Schema
		if err := rows.Scan(&s.Name, &s.Owner); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// ListTables 列出指定 schema 下的基表
func (i *inspector) ListTables(ctx context.Context, database, schema string) ([]driver.Table, error) {
	const q = `SELECT c.relname, c.relkind, COALESCE(d.description, ''), 0
		FROM pg_catalog.pg_class c
		LEFT JOIN pg_catalog.pg_description d ON d.objoid = c.oid AND d.objsubid = 0
		WHERE c.relnamespace = (SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = $1)
			AND c.relkind IN ('r', 'p')
		ORDER BY c.relname`

	rows, err := i.conn.pool.Query(ctx, q, schema)
	if err != nil {
		return nil, fmt.Errorf("pg/inspector: list tables: %w", err)
	}
	defer rows.Close()

	var result []driver.Table
	for rows.Next() {
		var t driver.Table
		var kind string
		if err := rows.Scan(&t.Name, &kind, &t.Comment, &t.RowCount); err != nil {
			return nil, err
		}
		switch kind {
		case "r":
			t.Type = "TABLE"
		case "p":
			t.Type = "PARTITIONED TABLE"
		default:
			t.Type = kind
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

// ListViews 列出指定 schema 下的视图
func (i *inspector) ListViews(ctx context.Context, database, schema string) ([]driver.View, error) {
	const q = `SELECT c.relname, pg_get_viewdef(c.oid)
		FROM pg_catalog.pg_class c
		WHERE c.relnamespace = (SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = $1)
			AND c.relkind = 'v'
		ORDER BY c.relname`

	rows, err := i.conn.pool.Query(ctx, q, schema)
	if err != nil {
		return nil, fmt.Errorf("pg/inspector: list views: %w", err)
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

// DescribeTable 描述表（含字段、索引、外键）
func (i *inspector) DescribeTable(ctx context.Context, database, schema, table string) (*driver.TableDetail, error) {
	if table == "" {
		return nil, fmt.Errorf("pg/inspector: table is required")
	}
	sch := schema
	if sch == "" {
		sch = "public"
	}

	detail := &driver.TableDetail{
		Name:   table,
		Schema: sch,
	}

	// 表注释
	if err := i.conn.pool.QueryRow(ctx,
		`SELECT COALESCE(d.description, '')
		FROM pg_catalog.pg_class c
		LEFT JOIN pg_catalog.pg_description d ON d.objoid = c.oid AND d.objsubid = 0
		WHERE c.relnamespace = (SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = $1)
			AND c.relname = $2`,
		sch, table,
	).Scan(&detail.Comment); err != nil {
		// not found 就算了
	}

	// 列定义
	colRows, err := i.conn.pool.Query(ctx,
		`SELECT a.attname,
				pg_catalog.format_type(a.atttypid, a.atttypmod),
				a.attnotnull,
				COALESCE((SELECT TRUE FROM pg_catalog.pg_index i WHERE i.indrelid = a.attrelid AND i.indisprimary AND a.attnum = ANY(i.indkey)), FALSE)
		FROM pg_catalog.pg_attribute a
		WHERE a.attrelid = (SELECT c.oid FROM pg_catalog.pg_class c
			WHERE c.relnamespace = (SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = $1) AND c.relname = $2)
			AND a.attnum > 0 AND NOT a.attisdropped
		ORDER BY a.attnum`,
		sch, table,
	)
	if err != nil {
		return nil, fmt.Errorf("pg/inspector: describe columns: %w", err)
	}
	defer colRows.Close()

	for colRows.Next() {
		var col driver.Column
		var notNull bool
		if err := colRows.Scan(&col.Name, &col.DatabaseType, &notNull, &col.IsPrimary); err != nil {
			return nil, err
		}
		col.Nullable = !notNull
		detail.Columns = append(detail.Columns, col)
	}

	// 索引
	idxRows, err := i.conn.pool.Query(ctx,
		`SELECT i.relname,
				ARRAY(SELECT a.attname FROM pg_catalog.pg_attribute a WHERE a.attrelid = idx.indrelid AND a.attnum = ANY(idx.indkey) AND a.attnum > 0 ORDER BY a.attnum),
				idx.indisunique,
				idx.indisprimary
		FROM pg_catalog.pg_index idx
		JOIN pg_catalog.pg_class i ON i.oid = idx.indexrelid
		WHERE idx.indrelid = (SELECT c.oid FROM pg_catalog.pg_class c
			WHERE c.relnamespace = (SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = $1) AND c.relname = $2)
		ORDER BY i.relname`,
		sch, table,
	)
	if err == nil {
		defer idxRows.Close()
		for idxRows.Next() {
			var idx driver.Index
			var cols []string
			if err := idxRows.Scan(&idx.Name, &cols, &idx.Unique, &idx.Primary); err == nil {
				idx.Columns = cols
				detail.Indexes = append(detail.Indexes, idx)
			}
		}
	}

	// 外键
	fkRows, err := i.conn.pool.Query(ctx,
		`SELECT con.conname,
				ARRAY(SELECT a.attname FROM pg_catalog.pg_attribute a WHERE a.attrelid = con.conrelid AND a.attnum = ANY(con.conkey) ORDER BY a.attnum),
				ns.nspname,
				cls.relname,
				ARRAY(SELECT a.attname FROM pg_catalog.pg_attribute a WHERE a.attrelid = con.confrelid AND a.attnum = ANY(con.confkey) ORDER BY a.attnum),
				COALESCE(con.confupdtype::text, ''),
				COALESCE(con.confdeltype::text, '')
		FROM pg_catalog.pg_constraint con
		JOIN pg_catalog.pg_class cls ON cls.oid = con.confrelid
		JOIN pg_catalog.pg_namespace ns ON ns.oid = cls.relnamespace
		WHERE con.conrelid = (SELECT c.oid FROM pg_catalog.pg_class c
			WHERE c.relnamespace = (SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = $1) AND c.relname = $2)
			AND con.contype = 'f'`,
		sch, table,
	)
	if err == nil {
		defer fkRows.Close()
		for fkRows.Next() {
			var fk driver.ForeignKey
			if err := fkRows.Scan(&fk.Name, &fk.Columns, &fk.ReferencedSchema, &fk.ReferencedTable, &fk.ReferencedColumns, &fk.OnUpdate, &fk.OnDelete); err == nil {
				fk.OnUpdate = pgFKRule(fk.OnUpdate)
				fk.OnDelete = pgFKRule(fk.OnDelete)
				detail.ForeignKeys = append(detail.ForeignKeys, fk)
			}
		}
	}

	return detail, nil
}

// ListIndexes 列出表的索引
func (i *inspector) ListIndexes(ctx context.Context, database, schema, table string) ([]driver.Index, error) {
	detail, err := i.DescribeTable(ctx, database, schema, table)
	if err != nil {
		return nil, err
	}
	return detail.Indexes, nil
}

// ListForeignKeys 列出表的外键
func (i *inspector) ListForeignKeys(ctx context.Context, database, schema, table string) ([]driver.ForeignKey, error) {
	detail, err := i.DescribeTable(ctx, database, schema, table)
	if err != nil {
		return nil, err
	}
	return detail.ForeignKeys, nil
}

// ListCharsets 列出 PG 可用 encoding
func (i *inspector) ListCharsets(ctx context.Context) ([]driver.Charset, error) {
	// PG 中字符集对应 encoding，pg_encoding_to_char 配合 pg_database 列出
	rows, err := i.conn.pool.Query(ctx,
		`SELECT DISTINCT pg_encoding_to_char(encoding) AS name FROM pg_database WHERE encoding IS NOT NULL ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("pg/inspector: list charsets: %w", err)
	}
	defer rows.Close()
	result := make([]driver.Charset, 0)
	for rows.Next() {
		var cs driver.Charset
		if err := rows.Scan(&cs.Name); err != nil {
			return nil, err
		}
		result = append(result, cs)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// 常用兜底
	if len(result) == 0 {
		result = []driver.Charset{
			{Name: "UTF8"},
			{Name: "LATIN1"},
			{Name: "SQL_ASCII"},
		}
	}
	return result, nil
}

// ListCollations 列出 PG 排序规则
func (i *inspector) ListCollations(ctx context.Context, charset string) ([]driver.Collation, error) {
	_ = charset
	rows, err := i.conn.pool.Query(ctx,
		`SELECT collname, COALESCE(pg_encoding_to_char(collencoding), '') AS charset
			FROM pg_collation
			ORDER BY collname`)
	if err != nil {
		return nil, fmt.Errorf("pg/inspector: list collations: %w", err)
	}
	defer rows.Close()
	result := make([]driver.Collation, 0)
	for rows.Next() {
		var c driver.Collation
		if err := rows.Scan(&c.Name, &c.Charset); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

// pgFKRule 把 PG 单字母规则码转可读字符串
func pgFKRule(code string) string {
	switch code {
	case "a":
		return "NO ACTION"
	case "r":
		return "RESTRICT"
	case "c":
		return "CASCADE"
	case "n":
		return "SET NULL"
	case "d":
		return "SET DEFAULT"
	default:
		return code
	}
}

// Helper type to avoid direct dependency on pgx types
type sqlStr string