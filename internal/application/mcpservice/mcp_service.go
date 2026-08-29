// Package mcpservice 提供 MCP 服务器管理的应用层服务
package mcpservice

// McpService MCP 服务器应用层服务接口（全局单例模式）
type McpService interface {
	// GetStatus 获取全局 MCP Server 状态
	GetStatus() (*McpServerStatusBO, error)
	// Start 启动全局 MCP Server
	Start(cmd *StartMcpServerCmd) (*McpServerStatusBO, error)
	// Stop 停止全局 MCP Server
	Stop() error
}
