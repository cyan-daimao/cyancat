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
