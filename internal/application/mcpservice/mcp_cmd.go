package mcpservice

// StartMcpServerCmd 启动 MCP Server 命令
type StartMcpServerCmd struct {
	// ConnID 连接 ID
	ConnID int64
	// AllowSelect 允许 SELECT / WITH / SHOW / DESC / EXPLAIN
	AllowSelect bool
	// AllowInsert 允许 INSERT
	AllowInsert bool
	// AllowUpdate 允许 UPDATE
	AllowUpdate bool
	// AllowDelete 允许 DELETE
	AllowDelete bool
	// AllowDDL 允许 CREATE / ALTER / DROP / TRUNCATE / RENAME 等 DDL
	AllowDDL bool
}
