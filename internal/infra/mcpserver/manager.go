package mcpserver

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ServerConfig MCP Server 启动配置
type ServerConfig struct {
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
	// Token 访问令牌
	Token string
	// Port 指定监听端口，0 表示随机端口
	Port int
	// Executor SQL 执行器
	Executor SQLExecutor
}

// ServerInfo 运行中的 MCP Server 信息
type ServerInfo struct {
	// ConnID 连接 ID
	ConnID int64
	// Address 访问地址
	Address string
	// Token 访问令牌
	Token string
	// Enabled 是否已启用
	Enabled bool
}

// Manager MCP Server 管理器接口
type Manager interface {
	// Start 启动指定连接的 MCP Server
	Start(connID int64, cfg ServerConfig) (*ServerInfo, error)
	// Stop 停止指定连接的 MCP Server
	Stop(connID int64) error
	// GetStatus 获取指定连接的 MCP Server 运行状态
	GetStatus(connID int64) *ServerInfo
	// StopAll 停止所有运行的 MCP Server
	StopAll() error
}

// NewManager 创建默认 Manager 实现
func NewManager() Manager {
	return &managerImpl{
		servers: make(map[int64]*serverEntry),
	}
}

type serverEntry struct {
	server  *Server
	address string
	token   string
	cancel  context.CancelFunc
}

type managerImpl struct {
	mu      sync.RWMutex
	servers map[int64]*serverEntry
}

// Start 启动 MCP Server，若已存在则先停止
func (m *managerImpl) Start(connID int64, cfg ServerConfig) (*ServerInfo, error) {
	if connID <= 0 {
		return nil, errors.New("mcpserver: connID must be positive")
	}
	if cfg.Executor == nil {
		return nil, errors.New("mcpserver: executor cannot be nil")
	}

	// 先停止已有服务
	_ = m.Stop(connID)

	allow := map[string]bool{
		"select": cfg.AllowSelect,
		"insert": cfg.AllowInsert,
		"update": cfg.AllowUpdate,
		"delete": cfg.AllowDelete,
		"ddl":    cfg.AllowDDL,
	}

	srv, err := NewServer(connID, cfg.Token, allow, cfg.Executor)
	if err != nil {
		return nil, err
	}
	srv.desiredPort = cfg.Port

	address, err := srv.Start()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	entry := &serverEntry{
		server:  srv,
		address: address,
		token:   cfg.Token,
		cancel:  cancel,
	}

	m.mu.Lock()
	m.servers[connID] = entry
	m.mu.Unlock()

	// 监听服务退出，自动清理
	go m.watch(ctx, connID, srv)

	return &ServerInfo{
		ConnID:  connID,
		Address: address,
		Token:   cfg.Token,
		Enabled: true,
	}, nil
}

// Stop 停止指定连接的 MCP Server
func (m *managerImpl) Stop(connID int64) error {
	m.mu.Lock()
	entry, ok := m.servers[connID]
	if ok {
		delete(m.servers, connID)
	}
	m.mu.Unlock()

	if !ok {
		return nil
	}

	if entry.cancel != nil {
		entry.cancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return entry.server.Stop(ctx)
}

// GetStatus 获取运行状态
func (m *managerImpl) GetStatus(connID int64) *ServerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, ok := m.servers[connID]
	if !ok {
		return nil
	}
	return &ServerInfo{
		ConnID:  connID,
		Address: entry.address,
		Token:   entry.token,
		Enabled: true,
	}
}

// StopAll 停止所有 MCP Server
func (m *managerImpl) StopAll() error {
	m.mu.Lock()
	entries := make([]*serverEntry, 0, len(m.servers))
	for _, entry := range m.servers {
		entries = append(entries, entry)
	}
	m.servers = make(map[int64]*serverEntry)
	m.mu.Unlock()

	var firstErr error
	for _, entry := range entries {
		if entry.cancel != nil {
			entry.cancel()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := entry.server.Stop(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
		cancel()
	}
	return firstErr
}

func (m *managerImpl) watch(ctx context.Context, connID int64, srv *Server) {
	// 等待外部取消信号或服务自然退出
	select {
	case <-ctx.Done():
	}

	m.mu.Lock()
	entry, ok := m.servers[connID]
	if ok && entry.server == srv {
		delete(m.servers, connID)
	}
	m.mu.Unlock()
}
