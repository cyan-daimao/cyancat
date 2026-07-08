// Package mcpservice 提供 MCP 服务器管理的应用层服务
package mcpservice

// McpService MCP 服务器应用层服务接口
type McpService interface {
	// GetStatus 获取指定连接的 MCP Server 状态
	GetStatus(connID int64) (*McpServerStatusBO, error)
	// Start 启动指定连接的 MCP Server
	Start(cmd *StartMcpServerCmd) (*McpServerStatusBO, error)
	// Stop 停止指定连接的 MCP Server
	Stop(connID int64) error
}
