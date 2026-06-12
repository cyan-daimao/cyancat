package schemaservice

import (
	"context"
	"errors"
	"strings"
	"time"

	"cyancat/internal/infra/driver"
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

// ListCharsets 列出字符集
func (s *SchemaServiceImpl) ListCharsets(query *ListCharsetsQuery) ([]*CharsetBO, error) {
	if query == nil || query.ConnID <= 0 {
		return nil, errors.New("schemaservice: connID is required")
	}
	conn, err := s.sessionMgr.Get(query.ConnID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := conn.Inspector().ListCharsets(ctx)
	if err != nil {
		return nil, err
	}
	return ToCharsetBOs(list), nil
}

// ListCollations 列出排序规则
func (s *SchemaServiceImpl) ListCollations(query *ListCollationsQuery) ([]*CollationBO, error) {
	if query == nil || query.ConnID <= 0 {
		return nil, errors.New("schemaservice: connID is required")
	}
	conn, err := s.sessionMgr.Get(query.ConnID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := conn.Inspector().ListCollations(ctx, query.Charset)
	if err != nil {
		return nil, err
	}
	return ToCollationBOs(list), nil
}

// GetCreateTableDDL 获取建表 DDL
func (s *SchemaServiceImpl) GetCreateTableDDL(query *GetCreateTableDDLQuery) (string, error) {
	if query == nil || query.ConnID <= 0 {
		return "", errors.New("schemaservice: connID is required")
	}
	if query.Table == "" {
		return "", errors.New("schemaservice: table is required")
	}
	conn, err := s.sessionMgr.Get(query.ConnID)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return conn.DDL().GetCreateTableDDL(ctx, query.Database, query.Schema, query.Table)
}

// PreviewCreateDatabase 预览 CREATE DATABASE DDL（不执行）
func (s *SchemaServiceImpl) PreviewCreateDatabase(cmd *CreateDatabaseCmd) (string, error) {
	if cmd == nil || cmd.ConnID <= 0 {
		return "", errors.New("schemaservice: connID is required")
	}
	if strings.TrimSpace(cmd.Name) == "" {
		return "", errors.New("schemaservice: database name is required")
	}
	conn, err := s.sessionMgr.Get(cmd.ConnID)
	if err != nil {
		return "", err
	}
	return conn.DDL().CreateDatabase(driver.DatabaseSpec{
		Name:      cmd.Name,
		Charset:   cmd.Charset,
		Collation: cmd.Collation,
	})
}

// CreateDatabase 创建数据库
func (s *SchemaServiceImpl) CreateDatabase(cmd *CreateDatabaseCmd) error {
	sql, err := s.PreviewCreateDatabase(cmd)
	if err != nil {
		return err
	}
	conn, err := s.sessionMgr.Get(cmd.ConnID)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = conn.Execute(ctx, sql)
	return err
}

// PreviewDropDatabase 预览 DROP DATABASE DDL（不执行）
func (s *SchemaServiceImpl) PreviewDropDatabase(cmd *DropDatabaseCmd) (string, error) {
	if cmd == nil || cmd.ConnID <= 0 {
		return "", errors.New("schemaservice: connID is required")
	}
	if strings.TrimSpace(cmd.Name) == "" {
		return "", errors.New("schemaservice: database name is required")
	}
	conn, err := s.sessionMgr.Get(cmd.ConnID)
	if err != nil {
		return "", err
	}
	return conn.DDL().DropDatabase(cmd.Name)
}

// DropDatabase 删除数据库
func (s *SchemaServiceImpl) DropDatabase(cmd *DropDatabaseCmd) error {
	sql, err := s.PreviewDropDatabase(cmd)
	if err != nil {
		return err
	}
	conn, err := s.sessionMgr.Get(cmd.ConnID)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = conn.Execute(ctx, sql)
	return err
}

// PreviewCreateTable 预览 CREATE TABLE DDL（不执行）
func (s *SchemaServiceImpl) PreviewCreateTable(cmd *CreateTableCmd) (string, error) {
	if cmd == nil || cmd.ConnID <= 0 {
		return "", errors.New("schemaservice: connID is required")
	}
	if strings.TrimSpace(cmd.Name) == "" {
		return "", errors.New("schemaservice: table name is required")
	}
	if len(cmd.Columns) == 0 {
		return "", errors.New("schemaservice: at least one column is required")
	}
	conn, err := s.sessionMgr.Get(cmd.ConnID)
	if err != nil {
		return "", err
	}
	spec := driver.TableSpec{
		Database:   cmd.Database,
		Schema:     cmd.Schema,
		Name:       cmd.Name,
		Engine:     cmd.Engine,
		Charset:    cmd.Charset,
		Collation:  cmd.Collation,
		Comment:    cmd.Comment,
		PrimaryKey: cmd.PK,
	}
	spec.Columns = toDriverColumnSpecs(cmd.Columns)
	spec.Indexes = toDriverIndexSpecs(cmd.Indexes)
	spec.ForeignKeys = toDriverForeignKeySpecs(cmd.ForeignKeys)
	return conn.DDL().CreateTable(spec)
}

// CreateTable 创建表
func (s *SchemaServiceImpl) CreateTable(cmd *CreateTableCmd) error {
	sql, err := s.PreviewCreateTable(cmd)
	if err != nil {
		return err
	}
	conn, err := s.sessionMgr.Get(cmd.ConnID)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// 多语句以 ";" 分隔时拆分执行
	stmts := splitDDLStatements(sql)
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := conn.Execute(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

// PreviewAlterTable 预览 ALTER TABLE DDL（不执行）
func (s *SchemaServiceImpl) PreviewAlterTable(cmd *AlterTableCmd) (string, error) {
	if cmd == nil || cmd.ConnID <= 0 {
		return "", errors.New("schemaservice: connID is required")
	}
	if strings.TrimSpace(cmd.Name) == "" {
		return "", errors.New("schemaservice: table name is required")
	}
	conn, err := s.sessionMgr.Get(cmd.ConnID)
	if err != nil {
		return "", err
	}
	spec := driver.AlterTableSpec{
		Database:        cmd.Database,
		Schema:          cmd.Schema,
		Name:            cmd.Name,
		Engine:          cmd.Engine,
		Charset:         cmd.Charset,
		Collation:       cmd.Collation,
		Comment:         cmd.Comment,
		DropColumns:     cmd.DropColumns,
		DropIndexes:     cmd.DropIndexes,
		DropForeignKeys: cmd.DropForeignKeys,
	}
	spec.AddColumns = toDriverColumnSpecs(cmd.AddColumns)
	spec.ModifyColumns = toDriverColumnSpecs(cmd.ModifyColumns)
	spec.AddIndexes = toDriverIndexSpecs(cmd.AddIndexes)
	spec.ModifyIndexes = toDriverIndexSpecs(cmd.ModifyIndexes)
	spec.AddForeignKeys = toDriverForeignKeySpecs(cmd.AddForeignKeys)
	for _, r := range cmd.RenameColumns {
		spec.RenameColumns = append(spec.RenameColumns, driver.ColumnRename{Old: r.Old, New: r.New})
	}
	return conn.DDL().AlterTable(spec)
}

// AlterTable 修改表
func (s *SchemaServiceImpl) AlterTable(cmd *AlterTableCmd) error {
	sql, err := s.PreviewAlterTable(cmd)
	if err != nil {
		return err
	}
	conn, err := s.sessionMgr.Get(cmd.ConnID)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	stmts := splitDDLStatements(sql)
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := conn.Execute(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

// PreviewDropTable 预览 DROP TABLE DDL（不执行）
func (s *SchemaServiceImpl) PreviewDropTable(cmd *DropTableCmd) (string, error) {
	if cmd == nil || cmd.ConnID <= 0 {
		return "", errors.New("schemaservice: connID is required")
	}
	if strings.TrimSpace(cmd.Name) == "" {
		return "", errors.New("schemaservice: table name is required")
	}
	conn, err := s.sessionMgr.Get(cmd.ConnID)
	if err != nil {
		return "", err
	}
	return conn.DDL().DropTable(cmd.Database, cmd.Schema, cmd.Name)
}

// DropTable 删除表
func (s *SchemaServiceImpl) DropTable(cmd *DropTableCmd) error {
	sql, err := s.PreviewDropTable(cmd)
	if err != nil {
		return err
	}
	conn, err := s.sessionMgr.Get(cmd.ConnID)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = conn.Execute(ctx, sql)
	return err
}

// splitDDLStatements 拆分多语句 DDL（按 ";\n" 简单拆分；不处理字符串中的分号）
// 实际生产环境应使用更稳健的解析；此处假设 DDL 由本地 generator 生成，可控
func splitDDLStatements(s string) []string {
	return strings.Split(s, ";\n")
}
