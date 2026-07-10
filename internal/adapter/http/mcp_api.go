// Package http 提供 Wails 暴露给前端的 McpAPI
package http

import (
	"errors"

	"cyancat/internal/adapter/dto"
	"cyancat/internal/application/mcpservice"
	"cyancat/internal/infra/api"
)

// McpAPI MCP Server 管理 API（通过 Wails Bindings 暴露给前端）
type McpAPI struct {
	svc mcpservice.McpService
}

// NewMcpAPI 创建 McpAPI
func NewMcpAPI(svc mcpservice.McpService) *McpAPI {
	return &McpAPI{svc: svc}
}

// GetStatus 获取 MCP Server 状态
func (a *McpAPI) GetStatus(connID int64) *api.Response[*dto.McpServerStatusDTO] {
	if connID <= 0 {
		return api.Fail[*dto.McpServerStatusDTO](api.BadRequestCode, nil, "connID must be positive")
	}

	bo, err := a.svc.GetStatus(connID)
	if err != nil {
		return api.Fail[*dto.McpServerStatusDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToMcpServerStatusDTO(bo))
}

// Start 启动 MCP Server
func (a *McpAPI) Start(req *dto.StartMcpServerRequest) *api.Response[*dto.McpServerStatusDTO] {
	if req == nil {
		return api.Fail[*dto.McpServerStatusDTO](api.BadRequestCode, nil, "request cannot be nil")
	}
	if req.ConnID <= 0 {
		return api.Fail[*dto.McpServerStatusDTO](api.BadRequestCode, nil, "connID must be positive")
	}

	bo, err := a.svc.Start(dto.ToStartMcpServerCmd(req))
	if err != nil {
		var conflict *mcpservice.PortConflictError
		if errors.As(err, &conflict) {
			return api.Fail[*dto.McpServerStatusDTO](
				api.ConflictCode,
				&dto.McpServerStatusDTO{ConnID: req.ConnID, Port: conflict.Port},
				conflict.Error(),
			)
		}
		return api.Fail[*dto.McpServerStatusDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToMcpServerStatusDTO(bo))
}

// Stop 停止 MCP Server
func (a *McpAPI) Stop(connID int64) *api.Response[bool] {
	if connID <= 0 {
		return api.Fail[bool](api.BadRequestCode, false, "connID must be positive")
	}

	if err := a.svc.Stop(connID); err != nil {
		return api.Fail[bool](api.ErrorCode, false, err.Error())
	}
	return api.Success(true)
}
