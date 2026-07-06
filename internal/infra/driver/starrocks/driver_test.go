package starrocks

import (
	"context"
	"testing"

	"cyancat/internal/infra/driver"
)

// mockConn 用于验证 starrocks.conn 的方法覆盖，不依赖真实数据库。
type mockConn struct {
	ddl driver.DDLGenerator
}

func (m *mockConn) Ping(ctx context.Context) error                         { return nil }
func (m *mockConn) WithDatabase(ctx context.Context, database string) (driver.Conn, func(), error) {
	return m, func() {}, nil
}
func (m *mockConn) Execute(ctx context.Context, sql string, args ...any) (*driver.Result, error) {
	return nil, nil
}
func (m *mockConn) Stream(ctx context.Context, sql string, args ...any) (driver.RowStream, error) {
	return nil, nil
}
func (m *mockConn) Inspector() driver.Inspector                            { return nil }
func (m *mockConn) DDL() driver.DDLGenerator                               { return m.ddl }
func (m *mockConn) ServerVersion(ctx context.Context) (string, error)      { return "", nil }
func (m *mockConn) Close() error                                           { return nil }

func TestStarRocksDriverType(t *testing.T) {
	d := New()
	if got := d.Type(); got != driver.StarRocks {
		t.Fatalf("expected type %q, got %q", driver.StarRocks, got)
	}
}

func TestStarRocksDriverTypeIsValid(t *testing.T) {
	if !driver.StarRocks.IsValid() {
		t.Fatalf("expected starrocks driver type to be valid")
	}
}

func TestStarRocksConnDDLUsesStarRocksGenerator(t *testing.T) {
	want := new(ddlGenerator)
	c := &conn{
		Conn:      &mockConn{ddl: new(ddlGenerator)},
		inspector: nil,
		ddl:       want,
	}
	if got := c.DDL(); got != want {
		t.Fatalf("conn.DDL() should return starrocks ddlGenerator, got %T", got)
	}
}

func TestPickCatalogDatabase(t *testing.T) {
	cases := []struct {
		db, schema string
		wantCat    string
		wantDB     string
	}{
		{"iceberg", "dws", "iceberg", "dws"},
		{"iceberg.dws", "", "iceberg", "dws"},
		{"default_catalog", "test", "default_catalog", "test"},
		{"default_catalog.test", "", "default_catalog", "test"},
	}
	for _, tc := range cases {
		cat, dbName := pickCatalogDatabase(tc.db, tc.schema)
		if cat != tc.wantCat || dbName != tc.wantDB {
			t.Errorf("pickCatalogDatabase(%q, %q) = (%q, %q), want (%q, %q)",
				tc.db, tc.schema, cat, dbName, tc.wantCat, tc.wantDB)
		}
	}
}

func TestThreePartName(t *testing.T) {
	if got := threePartName("iceberg", "dws", "t"); got != "`iceberg`.`dws`.`t`" {
		t.Fatalf("unexpected three part name: %s", got)
	}
	if got := threePartName("iceberg", "", "t"); got != "`iceberg`.`t`" {
		t.Fatalf("unexpected three part name: %s", got)
	}
}
