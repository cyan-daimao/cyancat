// Package adapter 聚合所有暴露给前端的 Wails API
package adapter

import (
	"cyancat/internal/adapter/http"
	"cyancat/internal/application/connectionservice"
	"cyancat/internal/application/queryservice"
	"cyancat/internal/application/schemaservice"
)

// App 应用聚合根，所有 API 都挂载在此结构体上，由 Wails 一次性绑定给前端
type App struct {
	// ConnectionAPI 连接管理 API
	ConnectionAPI *http.ConnectionAPI
	// QueryAPI SQL 查询 API
	QueryAPI *http.QueryAPI
	// SchemaAPI 元数据查询 API
	SchemaAPI *http.SchemaAPI
}

// NewApp 构造 App
func NewApp(
	connectionService connectionservice.ConnectionService,
	queryService queryservice.QueryService,
	schemaService schemaservice.SchemaService,
) *App {
	return &App{
		ConnectionAPI: http.NewConnectionAPI(connectionService),
		QueryAPI:      http.NewQueryAPI(queryService),
		SchemaAPI:     http.NewSchemaAPI(schemaService),
	}
}
