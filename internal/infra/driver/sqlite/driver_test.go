package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"cyancat/internal/infra/driver"
)

func TestConnectRequiresExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.sqlite")
	_, err := New().Connect(context.Background(), driver.ConnConfig{
		Type: driver.SQLite,
		Host: path,
	})
	if err == nil {
		t.Fatalf("expected missing sqlite file to fail")
	}
}

func TestSQLiteDriverMetadataAndDDL(t *testing.T) {
	path := createSQLiteFixture(t)

	conn, err := New().Connect(context.Background(), driver.ConnConfig{
		Type:           driver.SQLite,
		Host:           path,
		ConnectTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("connect sqlite: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	version, err := conn.ServerVersion(ctx)
	if err != nil || version == "" {
		t.Fatalf("expected sqlite version, version=%q err=%v", version, err)
	}

	result, err := conn.Execute(ctx, `SELECT amount FROM child WHERE name = 'row'`)
	if err != nil {
		t.Fatalf("query sqlite: %v", err)
	}
	if got := result.Rows[0][0]; got != int64(9223372036854775807) {
		t.Fatalf("expected bigint value to round-trip as int64, got %#v", got)
	}

	dbs, err := conn.Inspector().ListDatabases(ctx)
	if err != nil {
		t.Fatalf("list databases: %v", err)
	}
	if len(dbs) != 1 || dbs[0].Name != "main" {
		t.Fatalf("expected single main database, got %#v", dbs)
	}

	schemas, err := conn.Inspector().ListSchemas(ctx, "main")
	if err != nil {
		t.Fatalf("list schemas: %v", err)
	}
	if len(schemas) != 1 || schemas[0].Name != "main" {
		t.Fatalf("expected single main schema, got %#v", schemas)
	}

	tables, err := conn.Inspector().ListTables(ctx, "main", "main")
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	if !hasTable(tables, "child") || hasTable(tables, "sqlite_sequence") {
		t.Fatalf("expected user tables without sqlite internals, got %#v", tables)
	}

	views, err := conn.Inspector().ListViews(ctx, "main", "main")
	if err != nil {
		t.Fatalf("list views: %v", err)
	}
	if len(views) != 1 || views[0].Name != "child_view" || views[0].Definition == "" {
		t.Fatalf("expected child_view definition, got %#v", views)
	}

	detail, err := conn.Inspector().DescribeTable(ctx, "main", "main", "child")
	if err != nil {
		t.Fatalf("describe table: %v", err)
	}
	if !hasColumn(detail.Columns, "id", true) || !hasColumn(detail.Columns, "amount", false) {
		t.Fatalf("expected child columns, got %#v", detail.Columns)
	}
	if !hasIndex(detail.Indexes, "idx_child_name") {
		t.Fatalf("expected idx_child_name, got %#v", detail.Indexes)
	}
	if len(detail.ForeignKeys) != 1 || detail.ForeignKeys[0].ReferencedTable != "parent" {
		t.Fatalf("expected parent foreign key, got %#v", detail.ForeignKeys)
	}

	ddl, err := conn.DDL().GetCreateTableDDL(ctx, "main", "main", "child")
	if err != nil {
		t.Fatalf("get ddl: %v", err)
	}
	if ddl == "" || !containsFold(ddl, "CREATE TABLE child") {
		t.Fatalf("expected create table ddl, got %q", ddl)
	}
}

func createSQLiteFixture(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.sqlite")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer db.Close()

	stmts := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE parent (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL)`,
		`CREATE TABLE child (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			parent_id INTEGER NOT NULL,
			name TEXT,
			amount BIGINT,
			FOREIGN KEY(parent_id) REFERENCES parent(id) ON DELETE CASCADE ON UPDATE NO ACTION
		)`,
		`CREATE INDEX idx_child_name ON child(name)`,
		`CREATE VIEW child_view AS SELECT id, name FROM child`,
		`INSERT INTO parent (name) VALUES ('parent')`,
		`INSERT INTO child (parent_id, name, amount) VALUES (1, 'row', 9223372036854775807)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec fixture stmt %q: %v", stmt, err)
		}
	}
	return path
}

func hasTable(tables []driver.Table, name string) bool {
	for _, table := range tables {
		if table.Name == name {
			return true
		}
	}
	return false
}

func hasColumn(columns []driver.Column, name string, primary bool) bool {
	for _, column := range columns {
		if column.Name == name && column.IsPrimary == primary {
			return true
		}
	}
	return false
}

func hasIndex(indexes []driver.Index, name string) bool {
	for _, index := range indexes {
		if index.Name == name {
			return true
		}
	}
	return false
}

func containsFold(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		if equalFoldASCII(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		x, y := a[i], b[i]
		if x >= 'a' && x <= 'z' {
			x -= 32
		}
		if y >= 'a' && y <= 'z' {
			y -= 32
		}
		if x != y {
			return false
		}
	}
	return true
}
