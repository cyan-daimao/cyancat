// Package http 提供 Wails 暴露给前端的 SchemaAPI
package http

import (
	"cyancat/internal/adapter/dto"
	"cyancat/internal/application/schemaservice"
	"cyancat/internal/infra/api"
)

// SchemaAPI 元数据查询 API（通过 Wails Bindings 暴露给前端）
type SchemaAPI struct {
	svc schemaservice.SchemaService
}

// NewSchemaAPI 创建 SchemaAPI
func NewSchemaAPI(svc schemaservice.SchemaService) *SchemaAPI {
	return &SchemaAPI{svc: svc}
}

// ListDatabases 列出数据库
func (a *SchemaAPI) ListDatabases(connID int64) *api.Response[[]*dto.DatabaseDTO] {
	if connID <= 0 {
		return api.Fail[[]*dto.DatabaseDTO](api.BadRequestCode, nil, "connID must be positive")
	}
	list, err := a.svc.ListDatabases(&schemaservice.ListDatabasesQuery{ConnID: connID})
	if err != nil {
		return api.Fail[[]*dto.DatabaseDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToDatabaseDTOs(list))
}

// ListSchemas 列出 schema
func (a *SchemaAPI) ListSchemas(connID int64, database string) *api.Response[[]*dto.SchemaDTO] {
	if connID <= 0 {
		return api.Fail[[]*dto.SchemaDTO](api.BadRequestCode, nil, "connID must be positive")
	}
	list, err := a.svc.ListSchemas(&schemaservice.ListSchemasQuery{ConnID: connID, Database: database})
	if err != nil {
		return api.Fail[[]*dto.SchemaDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToSchemaDTOs(list))
}

// ListTables 列出表
func (a *SchemaAPI) ListTables(connID int64, database, schema string) *api.Response[[]*dto.TableDTO] {
	if connID <= 0 {
		return api.Fail[[]*dto.TableDTO](api.BadRequestCode, nil, "connID must be positive")
	}
	list, err := a.svc.ListTables(&schemaservice.ListTablesQuery{ConnID: connID, Database: database, Schema: schema})
	if err != nil {
		return api.Fail[[]*dto.TableDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToTableDTOs(list))
}

// ListViews 列出视图
func (a *SchemaAPI) ListViews(connID int64, database, schema string) *api.Response[[]*dto.ViewDTO] {
	if connID <= 0 {
		return api.Fail[[]*dto.ViewDTO](api.BadRequestCode, nil, "connID must be positive")
	}
	list, err := a.svc.ListViews(&schemaservice.ListTablesQuery{ConnID: connID, Database: database, Schema: schema})
	if err != nil {
		return api.Fail[[]*dto.ViewDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToViewDTOs(list))
}

// DescribeTable 描述表结构
func (a *SchemaAPI) DescribeTable(connID int64, database, schema, table string) *api.Response[*dto.TableDetailDTO] {
	if connID <= 0 {
		return api.Fail[*dto.TableDetailDTO](api.BadRequestCode, nil, "connID must be positive")
	}
	if table == "" {
		return api.Fail[*dto.TableDetailDTO](api.BadRequestCode, nil, "table is required")
	}
	detail, err := a.svc.DescribeTable(&schemaservice.DescribeTableQuery{ConnID: connID, Database: database, Schema: schema, Table: table})
	if err != nil {
		return api.Fail[*dto.TableDetailDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToTableDetailDTO(detail))
}

// --- DDL 相关 API ---

// ListCharsets 列出可用字符集
func (a *SchemaAPI) ListCharsets(req *dto.ListCharsetsRequest) *api.Response[[]*dto.CharsetDTO] {
	if req == nil || req.ConnID <= 0 {
		return api.Fail[[]*dto.CharsetDTO](api.BadRequestCode, nil, "connID must be positive")
	}
	list, err := a.svc.ListCharsets(&schemaservice.ListCharsetsQuery{ConnID: req.ConnID})
	if err != nil {
		return api.Fail[[]*dto.CharsetDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToCharsetDTOs(list))
}

// ListCollations 列出排序规则
func (a *SchemaAPI) ListCollations(req *dto.ListCollationsRequest) *api.Response[[]*dto.CollationDTO] {
	if req == nil || req.ConnID <= 0 {
		return api.Fail[[]*dto.CollationDTO](api.BadRequestCode, nil, "connID must be positive")
	}
	list, err := a.svc.ListCollations(&schemaservice.ListCollationsQuery{ConnID: req.ConnID, Charset: req.Charset})
	if err != nil {
		return api.Fail[[]*dto.CollationDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToCollationDTOs(list))
}

// GetCreateTableDDL 获取建表 DDL
func (a *SchemaAPI) GetCreateTableDDL(req *dto.GetCreateTableDDLRequest) *api.Response[string] {
	if req == nil || req.ConnID <= 0 {
		return api.Fail[string](api.BadRequestCode, "", "connID must be positive")
	}
	if req.Table == "" {
		return api.Fail[string](api.BadRequestCode, "", "table is required")
	}
	ddl, err := a.svc.GetCreateTableDDL(&schemaservice.GetCreateTableDDLQuery{
		ConnID:   req.ConnID,
		Database: req.Database,
		Schema:   req.Schema,
		Table:    req.Table,
	})
	if err != nil {
		return api.Fail[string](api.ErrorCode, "", err.Error())
	}
	return api.Success(ddl)
}

// PreviewCreateTable 预览新建表 DDL
func (a *SchemaAPI) PreviewCreateTable(req *dto.CreateTableRequest) *api.Response[string] {
	if req == nil || req.ConnID <= 0 {
		return api.Fail[string](api.BadRequestCode, "", "connID must be positive")
	}
	cmd := dto.ToCreateTableCmd(req)
	ddl, err := a.svc.PreviewCreateTable(cmd)
	if err != nil {
		return api.Fail[string](api.ErrorCode, "", err.Error())
	}
	return api.Success(ddl)
}

// PreviewAlterTable 预览修改表 DDL
func (a *SchemaAPI) PreviewAlterTable(req *dto.AlterTableRequest) *api.Response[string] {
	if req == nil || req.ConnID <= 0 {
		return api.Fail[string](api.BadRequestCode, "", "connID must be positive")
	}
	cmd := dto.ToAlterTableCmd(req)
	ddl, err := a.svc.PreviewAlterTable(cmd)
	if err != nil {
		return api.Fail[string](api.ErrorCode, "", err.Error())
	}
	return api.Success(ddl)
}

// CreateDatabase 新建数据库
func (a *SchemaAPI) CreateDatabase(req *dto.CreateDatabaseRequest) *api.Response[bool] {
	if req == nil || req.ConnID <= 0 {
		return api.Fail[bool](api.BadRequestCode, false, "connID must be positive")
	}
	if req.Name == "" {
		return api.Fail[bool](api.BadRequestCode, false, "database name is required")
	}
	cmd := dto.ToCreateDatabaseCmd(req)
	if err := a.svc.CreateDatabase(cmd); err != nil {
		return api.Fail[bool](api.ErrorCode, false, err.Error())
	}
	return api.Success(true)
}

// CreateTable 新建表
func (a *SchemaAPI) CreateTable(req *dto.CreateTableRequest) *api.Response[bool] {
	if req == nil || req.ConnID <= 0 {
		return api.Fail[bool](api.BadRequestCode, false, "connID must be positive")
	}
	if req.Name == "" {
		return api.Fail[bool](api.BadRequestCode, false, "table name is required")
	}
	cmd := dto.ToCreateTableCmd(req)
	if err := a.svc.CreateTable(cmd); err != nil {
		return api.Fail[bool](api.ErrorCode, false, err.Error())
	}
	return api.Success(true)
}

// AlterTable 修改表
func (a *SchemaAPI) AlterTable(req *dto.AlterTableRequest) *api.Response[bool] {
	if req == nil || req.ConnID <= 0 {
		return api.Fail[bool](api.BadRequestCode, false, "connID must be positive")
	}
	if req.Name == "" {
		return api.Fail[bool](api.BadRequestCode, false, "table name is required")
	}
	cmd := dto.ToAlterTableCmd(req)
	if err := a.svc.AlterTable(cmd); err != nil {
		return api.Fail[bool](api.ErrorCode, false, err.Error())
	}
	return api.Success(true)
}
