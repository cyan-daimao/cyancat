package sqlcompleteservice

import (
	"testing"

	"cyancat/internal/application/schemaservice"
)

type mockSchemaService struct {
	tables  []*schemaservice.TableBO
	columns []schemaservice.ColumnBO
}

func (m *mockSchemaService) ListDatabases(query *schemaservice.ListDatabasesQuery) ([]*schemaservice.DatabaseBO, error) {
	return nil, nil
}
func (m *mockSchemaService) ListSchemas(query *schemaservice.ListSchemasQuery) ([]*schemaservice.SchemaBO, error) {
	return nil, nil
}
func (m *mockSchemaService) ListTables(query *schemaservice.ListTablesQuery) ([]*schemaservice.TableBO, error) {
	return m.tables, nil
}
func (m *mockSchemaService) ListViews(query *schemaservice.ListTablesQuery) ([]*schemaservice.ViewBO, error) {
	return nil, nil
}
func (m *mockSchemaService) DescribeTable(query *schemaservice.DescribeTableQuery) (*schemaservice.TableDetailBO, error) {
	return &schemaservice.TableDetailBO{Columns: m.columns}, nil
}
func (m *mockSchemaService) ListIndexes(query *schemaservice.DescribeTableQuery) ([]*schemaservice.IndexBO, error) {
	return nil, nil
}
func (m *mockSchemaService) ListForeignKeys(query *schemaservice.DescribeTableQuery) ([]*schemaservice.ForeignKeyBO, error) {
	return nil, nil
}
func (m *mockSchemaService) ListCharsets(query *schemaservice.ListCharsetsQuery) ([]*schemaservice.CharsetBO, error) {
	return nil, nil
}
func (m *mockSchemaService) ListCollations(query *schemaservice.ListCollationsQuery) ([]*schemaservice.CollationBO, error) {
	return nil, nil
}
func (m *mockSchemaService) GetCreateTableDDL(query *schemaservice.GetCreateTableDDLQuery) (string, error) {
	return "", nil
}
func (m *mockSchemaService) PreviewCreateDatabase(cmd *schemaservice.CreateDatabaseCmd) (string, error) {
	return "", nil
}
func (m *mockSchemaService) CreateDatabase(cmd *schemaservice.CreateDatabaseCmd) error {
	return nil
}
func (m *mockSchemaService) PreviewDropDatabase(cmd *schemaservice.DropDatabaseCmd) (string, error) {
	return "", nil
}
func (m *mockSchemaService) DropDatabase(cmd *schemaservice.DropDatabaseCmd) error {
	return nil
}
func (m *mockSchemaService) PreviewCreateTable(cmd *schemaservice.CreateTableCmd) (string, error) {
	return "", nil
}
func (m *mockSchemaService) CreateTable(cmd *schemaservice.CreateTableCmd) error {
	return nil
}
func (m *mockSchemaService) PreviewAlterTable(cmd *schemaservice.AlterTableCmd) (string, error) {
	return "", nil
}
func (m *mockSchemaService) AlterTable(cmd *schemaservice.AlterTableCmd) error {
	return nil
}
func (m *mockSchemaService) PreviewDropTable(cmd *schemaservice.DropTableCmd) (string, error) {
	return "", nil
}
func (m *mockSchemaService) DropTable(cmd *schemaservice.DropTableCmd) error {
	return nil
}

func TestCompleteColumnWithAlias(t *testing.T) {
	mock := &mockSchemaService{
		tables: []*schemaservice.TableBO{{Name: "users"}},
		columns: []schemaservice.ColumnBO{
			{Name: "id", DatabaseType: "int"},
			{Name: "name", DatabaseType: "varchar"},
		},
	}
	svc := NewServiceImpl(mock)
	res, err := svc.Complete(&CompleteQuery{
		ConnID:         1,
		ConnectionType: "mysql",
		Database:       "db",
		Schema:         "db",
		SQL:            "SELECT a. FROM users a",
		CursorLine:     1,
		CursorColumn:   10,
	})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	foundID := false
	for _, c := range res.Candidates {
		if c.Label == "id" {
			foundID = true
		}
	}
	if !foundID {
		t.Fatalf("expected candidate id, got %+v", res.Candidates)
	}
}

func TestCompleteTable(t *testing.T) {
	mock := &mockSchemaService{
		tables: []*schemaservice.TableBO{{Name: "users"}, {Name: "orders"}},
	}
	svc := NewServiceImpl(mock)
	res, err := svc.Complete(&CompleteQuery{
		ConnID:         1,
		ConnectionType: "mysql",
		Database:       "db",
		Schema:         "db",
		SQL:            "SELECT * FROM ",
		CursorLine:     1,
		CursorColumn:   15,
	})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	foundUsers := false
	for _, c := range res.Candidates {
		if c.Label == "users" {
			foundUsers = true
		}
	}
	if !foundUsers {
		t.Fatalf("expected candidate users, got %+v", res.Candidates)
	}
}
