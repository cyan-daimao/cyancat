package starrocks

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cyancat/internal/infra/driver"
)

// inspector StarRocks 元数据查询器。
// StarRocks 采用 catalog -> database -> table 三层结构，映射到 driver 接口：
//   ListDatabases   -> SHOW CATALOGS        (catalog)
//   ListSchemas     -> SHOW DATABASES FROM  (database under catalog)
//   ListTables      -> SHOW TABLES FROM     (table under catalog.database)
//   DescribeTable   -> DESCRIBE / COLUMNS   (catalog.database.table)
type inspector struct {
	db *sql.DB
}

func newInspector(c driver.Conn) driver.Inspector {
	return &inspector{db: db(c)}
}

// ListDatabases 列出所有 catalog
func (i *inspector) ListDatabases(ctx context.Context) ([]driver.Database, error) {
	if i.db == nil {
		return nil, fmt.Errorf("starrocks/inspector: underlying db not available")
	}

	const q = `SHOW CATALOGS`
	rows, err := i.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("starrocks/inspector: list catalogs: %w", err)
	}
	defer rows.Close()

	var result []driver.Database
	for rows.Next() {
		var d driver.Database
		var catalogName, catalogType, comment sql.NullString
		if err := rows.Scan(&catalogName, &catalogType, &comment); err != nil {
			return nil, err
		}
		d.Name = catalogName.String
		result = append(result, d)
	}
	return result, rows.Err()
}

// ListSchemas 列出指定 catalog 下的 database
func (i *inspector) ListSchemas(ctx context.Context, database string) ([]driver.Schema, error) {
	catalog := strings.TrimSpace(database)
	if catalog == "" {
		return nil, fmt.Errorf("starrocks/inspector: catalog is required")
	}
	if i.db == nil {
		return nil, fmt.Errorf("starrocks/inspector: underlying db not available")
	}

	q := fmt.Sprintf("SHOW DATABASES FROM %s", quoteIdent(catalog))
	rows, err := i.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("starrocks/inspector: list databases from catalog %q: %w", catalog, err)
	}
	defer rows.Close()

	var result []driver.Schema
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result = append(result, driver.Schema{Name: name})
	}
	return result, rows.Err()
}

// ListTables 列出指定 catalog.database 下的表
func (i *inspector) ListTables(ctx context.Context, database, schema string) ([]driver.Table, error) {
	catalog, dbName := pickCatalogDatabase(database, schema)
	if catalog == "" {
		return nil, fmt.Errorf("starrocks/inspector: catalog is required")
	}
	if dbName == "" {
		return nil, fmt.Errorf("starrocks/inspector: database is required")
	}

	const q = `SELECT TABLE_NAME, IFNULL(TABLE_COMMENT, ''), IFNULL(TABLE_ROWS, 0)
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_CATALOG = ? AND TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME`

	rows, err := i.db.QueryContext(ctx, q, catalog, dbName)
	if err != nil {
		return nil, fmt.Errorf("starrocks/inspector: list tables from %s.%s: %w", catalog, dbName, err)
	}
	defer rows.Close()

	var result []driver.Table
	for rows.Next() {
		var t driver.Table
		if err := rows.Scan(&t.Name, &t.Comment, &t.RowCount); err != nil {
			return nil, err
		}
		t.Type = "BASE TABLE"
		result = append(result, t)
	}
	return result, rows.Err()
}

// ListViews 列出指定 catalog.database 下的视图
func (i *inspector) ListViews(ctx context.Context, database, schema string) ([]driver.View, error) {
	catalog, dbName := pickCatalogDatabase(database, schema)
	if catalog == "" || dbName == "" {
		return nil, nil
	}

	const q = `SELECT TABLE_NAME, IFNULL(TABLE_VIEW_DEFINITION, '')
		FROM INFORMATION_SCHEMA.VIEWS
		WHERE TABLE_CATALOG = ? AND TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME`

	rows, err := i.db.QueryContext(ctx, q, catalog, dbName)
	if err != nil {
		return nil, fmt.Errorf("starrocks/inspector: list views from %s.%s: %w", catalog, dbName, err)
	}
	defer rows.Close()

	var result []driver.View
	for rows.Next() {
		var v driver.View
		var def sql.NullString
		if err := rows.Scan(&v.Name, &def); err != nil {
			return nil, err
		}
		if def.Valid {
			v.Definition = def.String
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

// DescribeTable 描述表结构
func (i *inspector) DescribeTable(ctx context.Context, database, schema, table string) (*driver.TableDetail, error) {
	catalog, dbName := pickCatalogDatabase(database, schema)
	if table == "" {
		return nil, fmt.Errorf("starrocks/inspector: table name is required")
	}

	detail := &driver.TableDetail{
		Name:     table,
		Database: catalog,
		Schema:   dbName,
	}

	// 表注释
	if catalog != "" && dbName != "" {
		row := i.db.QueryRowContext(ctx,
			`SELECT IFNULL(TABLE_COMMENT, '')
				FROM INFORMATION_SCHEMA.TABLES
				WHERE TABLE_CATALOG = ? AND TABLE_SCHEMA = ? AND TABLE_NAME = ?`,
			catalog, dbName, table,
		)
		_ = row.Scan(&detail.Comment)
	}

	// 列定义
	colRows, err := i.db.QueryContext(ctx,
		`SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_KEY,
				IFNULL(CHARACTER_MAXIMUM_LENGTH, 0),
				IFNULL(NUMERIC_PRECISION, 0),
				IFNULL(NUMERIC_SCALE, 0),
				COLUMN_DEFAULT,
				IFNULL(COLUMN_COMMENT, ''),
				EXTRA,
				ORDINAL_POSITION,
				IFNULL(COLLATION_NAME, '')
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_CATALOG = ? AND TABLE_SCHEMA = ? AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION`,
		catalog, dbName, table,
	)
	if err != nil {
		return nil, fmt.Errorf("starrocks/inspector: describe columns: %w", err)
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
	indexes, err := i.ListIndexes(ctx, database, schema, table)
	if err != nil {
		return nil, err
	}
	detail.Indexes = indexes

	// 外键（StarRocks 当前不支持外键，返回空）
	detail.ForeignKeys = []driver.ForeignKey{}

	return detail, nil
}

// ListIndexes 列出表索引
func (i *inspector) ListIndexes(ctx context.Context, database, schema, table string) ([]driver.Index, error) {
	catalog, dbName := pickCatalogDatabase(database, schema)
	if table == "" {
		return nil, fmt.Errorf("starrocks/inspector: table name is required")
	}

	const q = `SELECT INDEX_NAME, INDEX_TYPE, COLUMN_NAME
		FROM INFORMATION_SCHEMA.STATISTICS
		WHERE TABLE_CATALOG = ? AND TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX`

	rows, err := i.db.QueryContext(ctx, q, catalog, dbName, table)
	if err != nil {
		return nil, fmt.Errorf("starrocks/inspector: list indexes: %w", err)
	}
	defer rows.Close()

	idxMap := make(map[string]*driver.Index)
	var order []string
	for rows.Next() {
		var name, idxType, col string
		if err := rows.Scan(&name, &idxType, &col); err != nil {
			return nil, err
		}
		if _, ok := idxMap[name]; !ok {
			idxMap[name] = &driver.Index{
				Name:    name,
				Unique:  idxType == "UNIQUE",
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

// ListForeignKeys StarRocks 暂不支持外键
func (i *inspector) ListForeignKeys(ctx context.Context, database, schema, table string) ([]driver.ForeignKey, error) {
	return []driver.ForeignKey{}, nil
}

// ListCharsets StarRocks 字符集信息有限，返回 utf8mb4
func (i *inspector) ListCharsets(ctx context.Context) ([]driver.Charset, error) {
	return []driver.Charset{
		{Name: "utf8mb4", Description: "UTF-8 Unicode", DefaultCollation: "utf8mb4_general_ci"},
	}, nil
}

// ListCollations StarRocks 排序规则信息有限
func (i *inspector) ListCollations(ctx context.Context, charset string) ([]driver.Collation, error) {
	return []driver.Collation{
		{Name: "utf8mb4_general_ci", Charset: "utf8mb4", IsDefault: true},
	}, nil
}

// pickCatalogDatabase 把 driver 层的 database/schema 解析为 StarRocks 的 catalog/database。
// 约定：database 字段表示 catalog，schema 字段表示 database；如果 schema 为空，则尝试从 database 按 "." 解析。
func pickCatalogDatabase(database, schema string) (catalog, dbName string) {
	catalog = strings.TrimSpace(database)
	dbName = strings.TrimSpace(schema)

	if dbName != "" {
		return catalog, dbName
	}
	// schema 为空时，database 可能是 "catalog.database" 形式
	if idx := strings.Index(catalog, "."); idx > 0 {
		return catalog[:idx], catalog[idx+1:]
	}
	return catalog, ""
}

// quoteIdent 使用 MySQL 反引号引用标识符
func quoteIdent(ident string) string {
	out := make([]byte, 0, len(ident)+2)
	out = append(out, '`')
	for i := 0; i < len(ident); i++ {
		if ident[i] == '`' {
			out = append(out, '`', '`')
		} else {
			out = append(out, ident[i])
		}
	}
	out = append(out, '`')
	return string(out)
}
