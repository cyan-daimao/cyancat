package schemaservice

import (
	"context"
	"errors"
	"time"

	"cyancat/internal/infra/session"
)

// SchemaServiceImpl 元数据服务实现
type SchemaServiceImpl struct {
	sessionMgr session.Manager
}

// NewSchemaServiceImpl 创建元数据服务
func NewSchemaServiceImpl(sessionMgr session.Manager) *SchemaServiceImpl {
	return &SchemaServiceImpl{sessionMgr: sessionMgr}
}

// ListDatabases 列出数据库
func (s *SchemaServiceImpl) ListDatabases(query *ListDatabasesQuery) ([]*DatabaseBO, error) {
	if query == nil || query.ConnID <= 0 {
		return nil, errors.New("schemaservice: connID is required")
	}
	conn, err := s.sessionMgr.Get(query.ConnID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := conn.Inspector().ListDatabases(ctx)
	if err != nil {
		return nil, err
	}
	return ToDatabaseBOs(list), nil
}

// ListSchemas 列出 schema
func (s *SchemaServiceImpl) ListSchemas(query *ListSchemasQuery) ([]*SchemaBO, error) {
	if query == nil || query.ConnID <= 0 {
		return nil, errors.New("schemaservice: connID is required")
	}
	conn, err := s.sessionMgr.Get(query.ConnID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := conn.Inspector().ListSchemas(ctx, query.Database)
	if err != nil {
		return nil, err
	}
	return ToSchemaBOs(list), nil
}

// ListTables 列出表
func (s *SchemaServiceImpl) ListTables(query *ListTablesQuery) ([]*TableBO, error) {
	if query == nil || query.ConnID <= 0 {
		return nil, errors.New("schemaservice: connID is required")
	}
	conn, err := s.sessionMgr.Get(query.ConnID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := conn.Inspector().ListTables(ctx, query.Database, query.Schema)
	if err != nil {
		return nil, err
	}
	return ToTableBOs(list), nil
}

// ListViews 列出视图
func (s *SchemaServiceImpl) ListViews(query *ListTablesQuery) ([]*ViewBO, error) {
	if query == nil || query.ConnID <= 0 {
		return nil, errors.New("schemaservice: connID is required")
	}
	conn, err := s.sessionMgr.Get(query.ConnID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := conn.Inspector().ListViews(ctx, query.Database, query.Schema)
	if err != nil {
		return nil, err
	}
	return ToViewBOs(list), nil
}

// DescribeTable 描述表
func (s *SchemaServiceImpl) DescribeTable(query *DescribeTableQuery) (*TableDetailBO, error) {
	if query == nil || query.ConnID <= 0 {
		return nil, errors.New("schemaservice: connID is required")
	}
	if query.Table == "" {
		return nil, errors.New("schemaservice: table is required")
	}
	conn, err := s.sessionMgr.Get(query.ConnID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	detail, err := conn.Inspector().DescribeTable(ctx, query.Database, query.Schema, query.Table)
	if err != nil {
		return nil, err
	}
	return ToTableDetailBO(detail), nil
}

// ListIndexes 列出索引
func (s *SchemaServiceImpl) ListIndexes(query *DescribeTableQuery) ([]*IndexBO, error) {
	if query == nil || query.ConnID <= 0 {
		return nil, errors.New("schemaservice: connID is required")
	}
	if query.Table == "" {
		return nil, errors.New("schemaservice: table is required")
	}
	conn, err := s.sessionMgr.Get(query.ConnID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := conn.Inspector().ListIndexes(ctx, query.Database, query.Schema, query.Table)
	if err != nil {
		return nil, err
	}
	return ToIndexBOs(list), nil
}

// ListForeignKeys 列出外键
func (s *SchemaServiceImpl) ListForeignKeys(query *DescribeTableQuery) ([]*ForeignKeyBO, error) {
	if query == nil || query.ConnID <= 0 {
		return nil, errors.New("schemaservice: connID is required")
	}
	if query.Table == "" {
		return nil, errors.New("schemaservice: table is required")
	}
	conn, err := s.sessionMgr.Get(query.ConnID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := conn.Inspector().ListForeignKeys(ctx, query.Database, query.Schema, query.Table)
	if err != nil {
		return nil, err
	}
	return ToForeignKeyBOs(list), nil
}
