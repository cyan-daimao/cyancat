// Package http 提供 Wails 暴露给前端的 McpAPI
package http

import (
	"errors"

	"cyancat/internal/adapter/dto"
	"cyancat/internal/application/mcpservice"
	"cyancat/internal/infra/api"
)

// McpAPI MCP Server 管理 API（通过 Wails Bindings 暴露给前端，全局单例模式）
type McpAPI struct {
	svc mcpservice.McpService
}

// NewMcpAPI 创建 McpAPI
func NewMcpAPI(svc mcpservice.McpService) *McpAPI {
	return &McpAPI{svc: svc}
}

// GetStatus 获取全局 MCP Server 状态
func (a *McpAPI) GetStatus() *api.Response[*dto.McpServerStatusDTO] {
	bo, err := a.svc.GetStatus()
	if err != nil {
		return api.Fail[*dto.McpServerStatusDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToMcpServerStatusDTO(bo))
}

// Start 启动全局 MCP Server
func (a *McpAPI) Start(req *dto.StartMcpServerRequest) *api.Response[*dto.McpServerStatusDTO] {
	if req == nil {
		return api.Fail[*dto.McpServerStatusDTO](api.BadRequestCode, nil, "request cannot be nil")
	}

	bo, err := a.svc.Start(dto.ToStartMcpServerCmd(req))
	if err != nil {
		var conflict *mcpservice.PortConflictError
		if errors.As(err, &conflict) {
			return api.Fail[*dto.McpServerStatusDTO](
				api.ConflictCode,
				&dto.McpServerStatusDTO{Port: conflict.Port},
				conflict.Error(),
			)
		}
		return api.Fail[*dto.McpServerStatusDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToMcpServerStatusDTO(bo))
}

// Stop 停止全局 MCP Server
func (a *McpAPI) Stop() *api.Response[bool] {
	if err := a.svc.Stop(); err != nil {
		return api.Fail[bool](api.ErrorCode, false, err.Error())
	}
	return api.Success(true)
}
