package starrocks

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cyancat/internal/infra/driver"
)

// ddlGenerator StarRocks DDL 生成器。
// 与 MySQL 大部分一致，区别是 SHOW CREATE TABLE / DROP TABLE 等需要拼
// catalog.database.table 三层限定名。
type ddlGenerator struct {
	dialect driver.Dialect
	db      *sql.DB
}

func newDDLGenerator(c driver.Conn) driver.DDLGenerator {
	return &ddlGenerator{
		dialect: &mysqlDialect{},
		db:      db(c),
	}
}

func (g *ddlGenerator) quote(ident string) string {
	return g.dialect.QuoteIdent(ident)
}

// threePartName 返回 `catalog`.`database`.`table`
func threePartName(catalog, database, table string) string {
	parts := []string{}
	if catalog != "" {
		parts = append(parts, "`"+strings.ReplaceAll(catalog, "`", "``")+"`")
	}
	if database != "" {
		parts = append(parts, "`"+strings.ReplaceAll(database, "`", "``")+"`")
	}
	if table != "" {
		parts = append(parts, "`"+strings.ReplaceAll(table, "`", "``")+"`")
	}
	return strings.Join(parts, ".")
}

func (g *ddlGenerator) CreateDatabase(spec driver.DatabaseSpec) (string, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return "", fmt.Errorf("starrocks/ddl: database name is required")
	}
	return "CREATE DATABASE " + g.quote(spec.Name), nil
}

func (g *ddlGenerator) DropDatabase(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("starrocks/ddl: database name is required")
	}
	return "DROP DATABASE " + g.quote(name), nil
}

func (g *ddlGenerator) CreateTable(spec driver.TableSpec) (string, error) {
	return "", fmt.Errorf("starrocks/ddl: CreateTable not supported yet")
}

func (g *ddlGenerator) AlterTable(spec driver.AlterTableSpec) (string, error) {
	return "", fmt.Errorf("starrocks/ddl: AlterTable not supported yet")
}

func (g *ddlGenerator) DropTable(database, schema, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("starrocks/ddl: table name is required")
	}
	catalog, dbName := pickCatalogDatabase(database, schema)
	return "DROP TABLE " + threePartName(catalog, dbName, name), nil
}

func (g *ddlGenerator) RenameTable(database, schema, oldName, newName string) (string, error) {
	if oldName == "" || newName == "" {
		return "", fmt.Errorf("starrocks/ddl: rename requires both names")
	}
	catalog, dbName := pickCatalogDatabase(database, schema)
	return "ALTER TABLE " + threePartName(catalog, dbName, oldName) + " RENAME " + threePartName("", "", newName), nil
}

func (g *ddlGenerator) GetCreateTableDDL(ctx context.Context, database, schema, table string) (string, error) {
	if table == "" {
		return "", fmt.Errorf("starrocks/ddl: table name is required")
	}
	catalog, dbName := pickCatalogDatabase(database, schema)

	q := "SHOW CREATE TABLE " + threePartName(catalog, dbName, table)
	row := g.db.QueryRowContext(ctx, q)
	var tname, ddl string
	if err := row.Scan(&tname, &ddl); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("starrocks/ddl: table not found")
		}
		return "", fmt.Errorf("starrocks/ddl: show create table: %w", err)
	}
	return ddl, nil
}

// mysqlDialect 复用 mysql 包的反引号方言。
type mysqlDialect struct{}

func (d *mysqlDialect) QuoteIdent(ident string) string {
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

func (d *mysqlDialect) Placeholder(n int) string {
	_ = n
	return "?"
}

func (d *mysqlDialect) DefaultLimit() int {
	return 1000
}
