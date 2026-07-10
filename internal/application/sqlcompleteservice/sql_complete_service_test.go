package sqlcompleteservice

import (
	"testing"
	"time"

	"cyancat/internal/application/schemaservice"
)

type mockSchemaService struct {
	tables       []*schemaservice.TableBO
	columns      []schemaservice.ColumnBO
	schemas      []string
	listCalls    int
	searchCalls  int
	schemaCalls  int
}

func (m *mockSchemaService) ListDatabases(query *schemaservice.ListDatabasesQuery) ([]*schemaservice.DatabaseBO, error) {
	return nil, nil
}
func (m *mockSchemaService) ListSchemas(query *schemaservice.ListSchemasQuery) ([]*schemaservice.SchemaBO, error) {
	m.schemaCalls++
	list := make([]*schemaservice.SchemaBO, 0, len(m.schemas))
	for _, name := range m.schemas {
		list = append(list, &schemaservice.SchemaBO{Name: name})
	}
	return list, nil
}
func (m *mockSchemaService) ListTables(query *schemaservice.ListTablesQuery) ([]*schemaservice.TableBO, error) {
	m.listCalls++
	return m.tables, nil
}
func (m *mockSchemaService) ListViews(query *schemaservice.ListTablesQuery) ([]*schemaservice.ViewBO, error) {
	return nil, nil
}
func (m *mockSchemaService) SearchTables(query *schemaservice.SearchTablesQuery) ([]*schemaservice.TableBO, error) {
	m.searchCalls++
	return m.tables, nil
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

func TestCompleteTableCache(t *testing.T) {
	mock := &mockSchemaService{
		tables: []*schemaservice.TableBO{{Name: "users"}},
	}
	svc := NewServiceImpl(mock)
	q := &CompleteQuery{
		ConnID:         1,
		ConnectionType: "mysql",
		Database:       "db",
		Schema:         "db",
		SQL:            "SELECT * FROM ",
		CursorLine:     1,
		CursorColumn:   15,
	}
	if _, err := svc.Complete(q); err != nil {
		t.Fatalf("first complete failed: %v", err)
	}
	if _, err := svc.Complete(q); err != nil {
		t.Fatalf("second complete failed: %v", err)
	}
	if mock.listCalls != 1 {
		t.Errorf("expected ListTables called once due to cache, got %d", mock.listCalls)
	}

	// 不同 schema 应重新查询
	q2 := &CompleteQuery{
		ConnID:         1,
		ConnectionType: "mysql",
		Database:       "db",
		Schema:         "other",
		SQL:            "SELECT * FROM ",
		CursorLine:     1,
		CursorColumn:   15,
	}
	if _, err := svc.Complete(q2); err != nil {
		t.Fatalf("other schema complete failed: %v", err)
	}
	if mock.listCalls != 2 {
		t.Errorf("expected ListTables called twice for different schema, got %d", mock.listCalls)
	}

	// 不同连接 ID 应重新查询
	q3 := &CompleteQuery{
		ConnID:         2,
		ConnectionType: "mysql",
		Database:       "db",
		Schema:         "db",
		SQL:            "SELECT * FROM ",
		CursorLine:     1,
		CursorColumn:   15,
	}
	if _, err := svc.Complete(q3); err != nil {
		t.Fatalf("other conn complete failed: %v", err)
	}
	if mock.listCalls != 3 {
		t.Errorf("expected ListTables called three times for different connID, got %d", mock.listCalls)
	}
}

func TestCompleteTableSearchCache(t *testing.T) {
	mock := &mockSchemaService{
		tables: []*schemaservice.TableBO{{Name: "users"}},
	}
	svc := NewServiceImpl(mock)
	// SQL 停在 FROM 后触发 CtxTable，同时手动带前缀 u，走 SearchTables 分支
	q := &CompleteQuery{
		ConnID:         1,
		ConnectionType: "mysql",
		Database:       "db",
		Schema:         "db",
		SQL:            "SELECT * FROM ",
		CursorLine:     1,
		CursorColumn:   15,
		Prefix:         "u",
	}
	if _, err := svc.Complete(q); err != nil {
		t.Fatalf("first complete failed: %v", err)
	}
	if _, err := svc.Complete(q); err != nil {
		t.Fatalf("second complete failed: %v", err)
	}
	if mock.searchCalls != 1 {
		t.Errorf("expected SearchTables called once due to cache, got %d", mock.searchCalls)
	}

	// 清除该 connID:db:db 的缓存后应重新查询
	svc.ClearTableCache(1, "db", "db")
	if _, err := svc.Complete(q); err != nil {
		t.Fatalf("complete after cache clear failed: %v", err)
	}
	if mock.searchCalls != 2 {
		t.Errorf("expected SearchTables called twice after clear, got %d", mock.searchCalls)
	}
}

func TestCompletePostgresSchemaTable(t *testing.T) {
	mock := &mockSchemaService{
		schemas: []string{"public", "other"},
		tables:  []*schemaservice.TableBO{{Name: "users"}, {Name: "orders"}},
	}
	svc := NewServiceImpl(mock)
	res, err := svc.Complete(&CompleteQuery{
		ConnID:         1,
		ConnectionType: "postgres",
		Database:       "db",
		Schema:         "public",
		SQL:            "SELECT * FROM public.",
		CursorLine:     1,
		CursorColumn:   27,
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
		t.Fatalf("expected users table from public schema, got %+v", res.Candidates)
	}
	if mock.schemaCalls != 1 {
		t.Errorf("expected ListSchemas called once, got %d", mock.schemaCalls)
	}

	// 再次补全应走 schema 缓存，不再查 ListSchemas
	res2, err := svc.Complete(&CompleteQuery{
		ConnID:         1,
		ConnectionType: "postgres",
		Database:       "db",
		Schema:         "public",
		SQL:            "SELECT * FROM public.",
		CursorLine:     1,
		CursorColumn:   27,
	})
	if err != nil {
		t.Fatalf("second complete failed: %v", err)
	}
	if len(res2.Candidates) == 0 {
		t.Fatalf("expected candidates from cache, got none")
	}
	if mock.schemaCalls != 1 {
		t.Errorf("expected ListSchemas still called once due to schema cache, got %d", mock.schemaCalls)
	}
}

func TestCompletePostgresSchemaSuggestions(t *testing.T) {
	mock := &mockSchemaService{
		schemas: []string{"public", "dw_ai", "dw_cdp", "ads_bi"},
		tables:  []*schemaservice.TableBO{{Name: "bak_dwd_action_event"}},
	}
	svc := NewServiceImpl(mock)
	res, err := svc.Complete(&CompleteQuery{
		ConnID:         1,
		ConnectionType: "postgres",
		Database:       "db",
		Schema:         "public",
		SQL:            "SELECT * FROM dw",
		CursorLine:     1,
		CursorColumn:   19,
		Prefix:         "dw",
	})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	foundDWAI := false
	foundDWCDP := false
	for _, c := range res.Candidates {
		if c.Label == "dw_ai" && c.Kind == KindSchema {
			foundDWAI = true
		}
		if c.Label == "dw_cdp" && c.Kind == KindSchema {
			foundDWCDP = true
		}
	}
	if !foundDWAI || !foundDWCDP {
		t.Fatalf("expected dw_ai/dw_cdp schema candidates, got %+v", res.Candidates)
	}
}

func TestCompletePostgresSchemaTableWithPrefix(t *testing.T) {
	mock := &mockSchemaService{
		schemas: []string{"dw_ai"},
		tables:  []*schemaservice.TableBO{{Name: "action_event"}},
	}
	svc := NewServiceImpl(mock)
	res, err := svc.Complete(&CompleteQuery{
		ConnID:         1,
		ConnectionType: "postgres",
		Database:       "db",
		Schema:         "public",
		SQL:            "SELECT * FROM dw_ai.act",
		CursorLine:     1,
		CursorColumn:   26,
		Prefix:         "dw_ai.act",
	})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	found := false
	for _, c := range res.Candidates {
		if c.Label == "action_event" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected action_event table from dw_ai schema, got %+v", res.Candidates)
	}
}

func TestCompletePostgresEmptyPrefixShowsSchemas(t *testing.T) {
	mock := &mockSchemaService{
		schemas: []string{"public", "dw_ai", "dw_cdp"},
		tables:  []*schemaservice.TableBO{{Name: "users"}},
	}
	svc := NewServiceImpl(mock)
	res, err := svc.Complete(&CompleteQuery{
		ConnID:         1,
		ConnectionType: "postgres",
		Database:       "db",
		Schema:         "",
		SQL:            "SELECT * FROM ",
		CursorLine:     1,
		CursorColumn:   15,
		Prefix:         "",
	})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	foundDWAI := false
	foundUsers := false
	for _, c := range res.Candidates {
		if c.Label == "dw_ai" && c.Kind == KindSchema {
			foundDWAI = true
		}
		if c.Label == "users" {
			foundUsers = true
		}
	}
	if !foundDWAI {
		t.Fatalf("expected dw_ai schema candidate, got %+v", res.Candidates)
	}
	if foundUsers {
		t.Fatalf("expected no table candidates at database level, got %+v", res.Candidates)
	}
}

func TestCompletePostgresEmptyPrefixInSchemaShowsTables(t *testing.T) {
	mock := &mockSchemaService{
		schemas: []string{"public"},
		tables:  []*schemaservice.TableBO{{Name: "users"}},
	}
	svc := NewServiceImpl(mock)
	res, err := svc.Complete(&CompleteQuery{
		ConnID:         1,
		ConnectionType: "postgres",
		Database:       "db",
		Schema:         "public",
		SQL:            "SELECT * FROM ",
		CursorLine:     1,
		CursorColumn:   15,
		Prefix:         "",
	})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	found := false
	for _, c := range res.Candidates {
		if c.Label == "users" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected users table in schema context, got %+v", res.Candidates)
	}
}

func TestCompletePostgresSchemaTableUsesCacheAndFuzzyMatch(t *testing.T) {
	mock := &mockSchemaService{
		schemas: []string{"dw_cdp"},
		tables: []*schemaservice.TableBO{
			{Name: "ads_xqd_cdp_ai_analysis_result_user_profile"},
			{Name: "dws_xqd_welabel_job_people"},
			{Name: "dws_xqd_welabel_job_people_add"},
		},
	}
	svc := NewServiceImpl(mock)

	// 预先写入 dw_cdp schema 的全量表缓存
	svc.cacheMu.Lock()
	svc.tableCache["1:db:dw_cdp:"] = tableCacheEntry{
		tables:   mock.tables,
		cachedAt: time.Now(),
	}
	svc.cacheMu.Unlock()

	res, err := svc.Complete(&CompleteQuery{
		ConnID:         1,
		ConnectionType: "postgres",
		Database:       "db",
		Schema:         "public",
		SQL:            "SELECT * FROM dw_cdp.peo",
		CursorLine:     1,
		CursorColumn:   28,
		Prefix:         "dw_cdp.peo",
	})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	foundPeople := false
	for _, c := range res.Candidates {
		if c.Label == "dws_xqd_welabel_job_people" {
			foundPeople = true
		}
	}
	if !foundPeople {
		t.Fatalf("expected dws_xqd_welabel_job_people from cache fuzzy match, got %+v", res.Candidates)
	}
	if mock.searchCalls != 0 {
		t.Errorf("expected no SearchTables call when cache hit, got %d", mock.searchCalls)
	}
}

func TestCompleteMySQLMemberAccessStillColumn(t *testing.T) {
	mock := &mockSchemaService{
		columns: []schemaservice.ColumnBO{{Name: "id", DatabaseType: "int"}},
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
		t.Fatalf("expected column id for mysql alias access, got %+v", res.Candidates)
	}
}
