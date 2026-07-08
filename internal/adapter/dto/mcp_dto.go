package dto

import "cyancat/internal/application/mcpservice"

// McpServerStatusDTO MCP Server 状态 DTO
type McpServerStatusDTO struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// Enabled 是否已启用
	Enabled bool `json:"enabled"`
	// Address 访问地址
	Address string `json:"address"`
	// Token 访问令牌
	Token string `json:"token"`
	// AllowSelect 允许 SELECT
	AllowSelect bool `json:"allowSelect"`
	// AllowInsert 允许 INSERT
	AllowInsert bool `json:"allowInsert"`
	// AllowUpdate 允许 UPDATE
	AllowUpdate bool `json:"allowUpdate"`
	// AllowDelete 允许 DELETE
	AllowDelete bool `json:"allowDelete"`
	// AllowDDL 允许 DDL
	AllowDDL bool `json:"allowDDL"`
}

// StartMcpServerRequest 启动 MCP Server 请求
type StartMcpServerRequest struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// AllowSelect 允许 SELECT
	AllowSelect bool `json:"allowSelect"`
	// AllowInsert 允许 INSERT
	AllowInsert bool `json:"allowInsert"`
	// AllowUpdate 允许 UPDATE
	AllowUpdate bool `json:"allowUpdate"`
	// AllowDelete 允许 DELETE
	AllowDelete bool `json:"allowDelete"`
	// AllowDDL 允许 DDL
	AllowDDL bool `json:"allowDDL"`
}

// ToStartMcpServerCmd Request -> Cmd
func ToStartMcpServerCmd(req *StartMcpServerRequest) *mcpservice.StartMcpServerCmd {
	if req == nil {
		return nil
	}
	return &mcpservice.StartMcpServerCmd{
		ConnID:      req.ConnID,
		AllowSelect: req.AllowSelect,
		AllowInsert: req.AllowInsert,
		AllowUpdate: req.AllowUpdate,
		AllowDelete: req.AllowDelete,
		AllowDDL:    req.AllowDDL,
	}
}

// ToMcpServerStatusDTO BO -> DTO
func ToMcpServerStatusDTO(bo *mcpservice.McpServerStatusBO) *McpServerStatusDTO {
	if bo == nil {
		return nil
	}
	return &McpServerStatusDTO{
		ConnID:      bo.ConnID,
		Enabled:     bo.Enabled,
		Address:     bo.Address,
		Token:       bo.Token,
		AllowSelect: bo.AllowSelect,
		AllowInsert: bo.AllowInsert,
		AllowUpdate: bo.AllowUpdate,
		AllowDelete: bo.AllowDelete,
		AllowDDL:    bo.AllowDDL,
	}
}
