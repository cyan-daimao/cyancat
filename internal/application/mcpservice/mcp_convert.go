package mcpservice

import "cyancat/internal/infra/db/mcprepo"

// ToMcpServerStatusBO 把 DO 转换为 MCP Server 状态 BO
func ToMcpServerStatusBO(do *mcprepo.McpServerDO) *McpServerStatusBO {
	if do == nil {
		return nil
	}
	return &McpServerStatusBO{
		Enabled:     do.Enabled,
		Port:        do.Port,
		AllowSelect: do.AllowSelect,
		AllowInsert: do.AllowInsert,
		AllowUpdate: do.AllowUpdate,
		AllowDelete: do.AllowDelete,
		AllowDDL:    do.AllowDDL,
		Token:       do.Token,
	}
}

// ToMcpServerDO 把 BO 转换为 DO
func ToMcpServerDO(bo *McpServerStatusBO) *mcprepo.McpServerDO {
	if bo == nil {
		return nil
	}
	return &mcprepo.McpServerDO{
		Enabled:     bo.Enabled,
		Port:        bo.Port,
		AllowSelect: bo.AllowSelect,
		AllowInsert: bo.AllowInsert,
		AllowUpdate: bo.AllowUpdate,
		AllowDelete: bo.AllowDelete,
		AllowDDL:    bo.AllowDDL,
		Token:       bo.Token,
	}
}
