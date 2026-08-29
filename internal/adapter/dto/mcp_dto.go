package dto

import "cyancat/internal/application/mcpservice"

// McpServerStatusDTO MCP Server 状态 DTO（全局单例）
type McpServerStatusDTO struct {
	// Enabled 是否已启用
	Enabled bool `json:"enabled"`
	// Address 访问地址
	Address string `json:"address"`
	// Port 监听端口
	Port int `json:"port"`
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

// StartMcpServerRequest 启动 MCP Server 请求（全局单例）
type StartMcpServerRequest struct {
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
	// ForceNewPort 是否强制使用新端口（忽略历史端口）
	ForceNewPort bool `json:"forceNewPort"`
}

// ToStartMcpServerCmd Request -> Cmd
func ToStartMcpServerCmd(req *StartMcpServerRequest) *mcpservice.StartMcpServerCmd {
	if req == nil {
		return nil
	}
	return &mcpservice.StartMcpServerCmd{
		AllowSelect:  req.AllowSelect,
		AllowInsert:  req.AllowInsert,
		AllowUpdate:  req.AllowUpdate,
		AllowDelete:  req.AllowDelete,
		AllowDDL:     req.AllowDDL,
		ForceNewPort: req.ForceNewPort,
	}
}

// ToMcpServerStatusDTO BO -> DTO
func ToMcpServerStatusDTO(bo *mcpservice.McpServerStatusBO) *McpServerStatusDTO {
	if bo == nil {
		return nil
	}
	return &McpServerStatusDTO{
		Enabled:     bo.Enabled,
		Address:     bo.Address,
		Port:        bo.Port,
		Token:       bo.Token,
		AllowSelect: bo.AllowSelect,
		AllowInsert: bo.AllowInsert,
		AllowUpdate: bo.AllowUpdate,
		AllowDelete: bo.AllowDelete,
		AllowDDL:    bo.AllowDDL,
	}
}
