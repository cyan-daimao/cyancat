package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cyancat/internal/infra/driver"
)

// inspector 基于 INFORMATION_SCHEMA 的元数据查询器
type inspector struct {
	conn *Conn
}

func newInspector(c *Conn) driver.Inspector {
	return &inspector{conn: c}
}

// ListDatabases 列出所有数据库
func (i *inspector) ListDatabases(ctx context.Context) ([]driver.Database, error) {
	const q = `SELECT SCHEMA_NAME, DEFAULT_CHARACTER_SET_NAME, DEFAULT_COLLATION_NAME
		FROM INFORMATION_SCHEMA.SCHEMATA
		WHERE SCHEMA_NAME NOT IN ('information_schema', 'performance_schema', 'mysql', 'sys')
		ORDER BY SCHEMA_NAME`

	rows, err := i.conn.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("mysql/inspector: list databases: %w", err)
	}
	defer rows.Close()

	var result []driver.Database
	for rows.Next() {
		var d driver.Database
		var charset, collation sql.NullString
		if err := rows.Scan(&d.Name, &charset, &collation); err != nil {
			return nil, err
		}
		d.Charset = charset.String
		d.Collation = collation.String
		result = append(result, d)
	}
	return result, rows.Err()
}

// ListSchemas MySQL 中 schema == database，返回同名条目
func (i *inspector) ListSchemas(ctx context.Context, database string) ([]driver.Schema, error) {
	if database == "" {
		return nil, fmt.Errorf("mysql/inspector: database is required")
	}
	return []driver.Schema{{Name: database}}, nil
}

// ListTables 列出指定数据库下的表
func (i *inspector) ListTables(ctx context.Context, database, schema string) ([]driver.Table, error) {
	target := pickSchema(database, schema)
	const q = `SELECT TABLE_NAME, TABLE_TYPE, IFNULL(TABLE_COMMENT, ''), IFNULL(TABLE_ROWS, 0)
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME`

	rows, err := i.conn.db.QueryContext(ctx, q, target)
	if err != nil {
		return nil, fmt.Errorf("mysql/inspector: list tables: %w", err)
	}
	defer rows.Close()

	var result []driver.Table
	for rows.Next() {
		var t driver.Table
		if err := rows.Scan(&t.Name, &t.Type, &t.Comment, &t.RowCount); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

// ListViews 列出视图
func (i *inspector) ListViews(ctx context.Context, database, schema string) ([]driver.View, error) {
	target := pickSchema(database, schema)
	const q = `SELECT TABLE_NAME, IFNULL(VIEW_DEFINITION, '')
		FROM INFORMATION_SCHEMA.VIEWS
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME`

	rows, err := i.conn.db.QueryContext(ctx, q, target)
	if err != nil {
		return nil, fmt.Errorf("mysql/inspector: list views: %w", err)
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

// DescribeTable 描述表
func (i *inspector) DescribeTable(ctx context.Context, database, schema, table string) (*driver.TableDetail, error) {
	target := pickSchema(database, schema)
	if table == "" {
		return nil, fmt.Errorf("mysql/inspector: table name is required")
	}

	detail := &driver.TableDetail{
		Name:     table,
		Schema:   target,
		Database: target,
	}

	// 表信息（注释+引擎+字符集+排序规则）
	row := i.conn.db.QueryRowContext(ctx,
		`SELECT IFNULL(TABLE_COMMENT, ''), IFNULL(ENGINE, ''), IFNULL(TABLE_COLLATION, '')
			FROM INFORMATION_SCHEMA.TABLES
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`,
		target, table,
	)
	var engine, collation string
	if err := row.Scan(&detail.Comment, &engine, &collation); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("mysql/inspector: describe table info: %w", err)
	}
	_ = engine
	_ = collation

	// 列定义（带所有扩展字段）
	colRows, err := i.conn.db.QueryContext(ctx,
		`SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY,
				IFNULL(CHARACTER_MAXIMUM_LENGTH, 0),
				IFNULL(NUMERIC_PRECISION, 0),
				IFNULL(NUMERIC_SCALE, 0),
				COLUMN_DEFAULT,
				IFNULL(COLUMN_COMMENT, ''),
				EXTRA,
				ORDINAL_POSITION,
				IFNULL(COLLATION_NAME, '')
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION`,
		target, table,
	)
	if err != nil {
		return nil, fmt.Errorf("mysql/inspector: describe columns: %w", err)
	}
	defer colRows.Close()

	for colRows.Next() {
		var col driver.Column
		var nullable, key string
		var charLen, numericPrec, numericScale sql.NullInt64
		var colDefault sql.NullString
		if err := colRows.Scan(
			&col.Name, &col.DatabaseType, &nullable, &key,
			&charLen, &numericPrec, &numericScale,
			&colDefault, &col.Comment, &col.Extra,
			&col.OrdinalPosition, &col.Collation,
		); err != nil {
			return nil, err
		}
		col.Nullable = nullable == "YES"
		col.IsPrimary = key == "PRI"
		col.AutoIncrement = strings.Contains(col.Extra, "auto_increment")
		col.Unsigned = strings.Contains(col.DatabaseType, "unsigned")
		if colDefault.Valid {
			v := colDefault.String
			col.DefaultValue = &v
		}
		if charLen.Valid {
			v := int(charLen.Int64)
			col.TypeLength = &v
		}
		if numericPrec.Valid && numericPrec.Int64 > 0 {
			v := int(numericPrec.Int64)
			col.Precision = &v
		}
		if numericScale.Valid && numericScale.Int64 > 0 {
			v := int(numericScale.Int64)
			col.Scale = &v
		}
		detail.Columns = append(detail.Columns, col)
	}

	// 索引
	indexes, err := i.ListIndexes(ctx, target, target, table)
	if err != nil {
		return nil, err
	}
	detail.Indexes = indexes

	// 外键
	fks, err := i.ListForeignKeys(ctx, target, target, table)
	if err != nil {
		return nil, err
	}
	detail.ForeignKeys = fks

	return detail, nil
}

// ListIndexes 列出表的索引
func (i *inspector) ListIndexes(ctx context.Context, database, schema, table string) ([]driver.Index, error) {
	target := pickSchema(database, schema)
	const q = `SELECT INDEX_NAME, NON_UNIQUE, COLUMN_NAME, SEQ_IN_INDEX
		FROM INFORMATION_SCHEMA.STATISTICS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX`

	rows, err := i.conn.db.QueryContext(ctx, q, target, table)
	if err != nil {
		return nil, fmt.Errorf("mysql/inspector: list indexes: %w", err)
	}
	defer rows.Close()

	idxMap := make(map[string]*driver.Index)
	var order []string
	for rows.Next() {
		var name, col string
		var nonUnique int
		var seq int
		if err := rows.Scan(&name, &nonUnique, &col, &seq); err != nil {
			return nil, err
		}
		if _, ok := idxMap[name]; !ok {
			idxMap[name] = &driver.Index{
				Name:    name,
				Unique:  nonUnique == 0,
				Primary: name == "PRIMARY",
			}
			order = append(order, name)
		}
		idxMap[name].Columns = append(idxMap[name].Columns, col)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]driver.Index, 0, len(order))
	for _, name := range order {
		result = append(result, *idxMap[name])
	}
	return result, nil
}

// ListForeignKeys 列出外键
func (i *inspector) ListForeignKeys(ctx context.Context, database, schema, table string) ([]driver.ForeignKey, error) {
	target := pickSchema(database, schema)
	const q = `SELECT kcu.CONSTRAINT_NAME, kcu.COLUMN_NAME, kcu.REFERENCED_TABLE_SCHEMA,
			kcu.REFERENCED_TABLE_NAME, kcu.REFERENCED_COLUMN_NAME,
			IFNULL(rc.UPDATE_RULE, ''), IFNULL(rc.DELETE_RULE, '')
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
		LEFT JOIN INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS rc
			ON rc.CONSTRAINT_SCHEMA = kcu.CONSTRAINT_SCHEMA AND rc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
		WHERE kcu.TABLE_SCHEMA = ? AND kcu.TABLE_NAME = ? AND kcu.REFERENCED_TABLE_NAME IS NOT NULL
		ORDER BY kcu.CONSTRAINT_NAME, kcu.ORDINAL_POSITION`

	rows, err := i.conn.db.QueryContext(ctx, q, target, table)
	if err != nil {
		return nil, fmt.Errorf("mysql/inspector: list fks: %w", err)
	}
	defer rows.Close()

	fkMap := make(map[string]*driver.ForeignKey)
	var order []string
	for rows.Next() {
		var name, col, refSchema, refTable, refCol, onUpd, onDel string
		if err := rows.Scan(&name, &col, &refSchema, &refTable, &refCol, &onUpd, &onDel); err != nil {
			return nil, err
		}
		if _, ok := fkMap[name]; !ok {
			fkMap[name] = &driver.ForeignKey{
				Name:             name,
				ReferencedSchema: refSchema,
				ReferencedTable:  refTable,
				OnUpdate:         onUpd,
				OnDelete:         onDel,
			}
			order = append(order, name)
		}
		fkMap[name].Columns = append(fkMap[name].Columns, col)
		fkMap[name].ReferencedColumns = append(fkMap[name].ReferencedColumns, refCol)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]driver.ForeignKey, 0, len(order))
	for _, name := range order {
		result = append(result, *fkMap[name])
	}
	return result, nil
}

// ListCharsets 列出可用字符集
func (i *inspector) ListCharsets(ctx context.Context) ([]driver.Charset, error) {
	rows, err := i.conn.db.QueryContext(ctx,
		`SELECT CHARACTER_SET_NAME, IFNULL(DESCRIPTION, ''), IFNULL(DEFAULT_COLLATE_NAME, '')
			FROM INFORMATION_SCHEMA.CHARACTER_SETS
			ORDER BY CHARACTER_SET_NAME`)
	if err != nil {
		return nil, fmt.Errorf("mysql/inspector: list charsets: %w", err)
	}
	defer rows.Close()

	result := make([]driver.Charset, 0)
	for rows.Next() {
		var cs driver.Charset
		if err := rows.Scan(&cs.Name, &cs.Description, &cs.DefaultCollation); err != nil {
			return nil, err
		}
		result = append(result, cs)
	}
	return result, rows.Err()
}

// ListCollations 列出指定字符集下的排序规则
func (i *inspector) ListCollations(ctx context.Context, charset string) ([]driver.Collation, error) {
	q := `SELECT COLLATION_NAME, CHARACTER_SET_NAME, IS_DEFAULT
		FROM INFORMATION_SCHEMA.COLLATIONS`
	args := []any{}
	if charset != "" {
		q += " WHERE CHARACTER_SET_NAME = ?"
		args = append(args, charset)
	}
	q += " ORDER BY COLLATION_NAME"

	rows, err := i.conn.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("mysql/inspector: list collations: %w", err)
	}
	defer rows.Close()

	result := make([]driver.Collation, 0)
	for rows.Next() {
		var c driver.Collation
		var isDefault string
		if err := rows.Scan(&c.Name, &c.Charset, &isDefault); err != nil {
			return nil, err
		}
		c.IsDefault = isDefault == "Yes"
		result = append(result, c)
	}
	return result, rows.Err()
}

// pickSchema 优先使用 schema，否则用 database（MySQL 两者等价）
func pickSchema(database, schema string) string {
	if schema != "" {
		return schema
	}
	return database
}
