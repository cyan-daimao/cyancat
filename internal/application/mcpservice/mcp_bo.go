package mcpservice

// McpServerStatusBO MCP Server 状态业务对象
type McpServerStatusBO struct {
	// ConnID 连接 ID
	ConnID int64
	// Enabled 是否已启用
	Enabled bool
	// Address 访问地址，如 http://127.0.0.1:12345/sse
	Address string
	// Port 监听端口
	Port int
	// Token 访问令牌
	Token string
	// AllowSelect 允许 SELECT
	AllowSelect bool
	// AllowInsert 允许 INSERT
	AllowInsert bool
	// AllowUpdate 允许 UPDATE
	AllowUpdate bool
	// AllowDelete 允许 DELETE
	AllowDelete bool
	// AllowDDL 允许 DDL
	AllowDDL bool
}
