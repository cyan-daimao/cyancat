// Package connectionservice 提供连接管理的应用层服务
package connectionservice

import (
	"cyancat/internal/infra/api"
)

// ConnectionService 连接管理应用层服务接口
type ConnectionService interface {
	// List 列出连接
	List(query *ListConnectionQuery) ([]*ConnectionBO, error)
	// Page 分页查询连接
	Page(query *PageConnectionQuery) (*api.Page[*ConnectionBO], error)
	// GetByID 按 ID 查询连接
	GetByID(id int64) (*ConnectionBO, error)
	// Create 创建连接
	Create(cmd *CreateConnectionCmd) (*ConnectionBO, error)
	// Update 更新连接
	Update(id int64, cmd *UpdateConnectionCmd) (*ConnectionBO, error)
	// Delete 删除连接
	Delete(cmd *DeleteConnectionCmd) error
	// Test 测试连接（仅 Ping，不缓存到 SessionManager）
	Test(cmd *TestConnectionCmd) (*TestConnectionResult, error)
	// Open 打开已保存的连接（建立长连接，存入 SessionManager）
	Open(id int64) (*ConnectionBO, error)
	// Close 关闭已打开的连接
	Close(id int64) error
}

