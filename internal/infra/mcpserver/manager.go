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
	// Address 访问地址
	Address string
	// Token 访问令牌
	Token string
	// Enabled 是否已启用
	Enabled bool
}

// Manager 全局 MCP Server 管理器接口（单例模式）
type Manager interface {
	// Start 启动全局 MCP Server
	Start(cfg ServerConfig) (*ServerInfo, error)
	// Stop 停止全局 MCP Server
	Stop() error
	// GetStatus 获取全局 MCP Server 运行状态
	GetStatus() *ServerInfo
	// StopAll 停止全局 MCP Server（等价于 Stop，保持接口兼容）
	StopAll() error
}

// NewManager 创建全局单例 Manager 实现
func NewManager() Manager {
	return &managerImpl{}
}

type managerImpl struct {
	mu      sync.RWMutex
	server  *Server
	address string
	token   string
	cancel  context.CancelFunc
}

// Start 启动全局 MCP Server，若已存在则先停止
func (m *managerImpl) Start(cfg ServerConfig) (*ServerInfo, error) {
	if cfg.Executor == nil {
		return nil, errors.New("mcpserver: executor cannot be nil")
	}

	// 先停止已有服务
	_ = m.Stop()

	allow := map[string]bool{
		"select": cfg.AllowSelect,
		"insert": cfg.AllowInsert,
		"update": cfg.AllowUpdate,
		"delete": cfg.AllowDelete,
		"ddl":    cfg.AllowDDL,
	}

	srv, err := NewServer(cfg.Token, allow, cfg.Executor)
	if err != nil {
		return nil, err
	}
	srv.desiredPort = cfg.Port

	address, err := srv.Start()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	m.mu.Lock()
	m.server = srv
	m.address = address
	m.token = cfg.Token
	m.cancel = cancel
	m.mu.Unlock()

	// 监听服务退出，自动清理
	go m.watch(ctx, srv)

	return &ServerInfo{
		Address: address,
		Token:   cfg.Token,
		Enabled: true,
	}, nil
}

// Stop 停止全局 MCP Server
func (m *managerImpl) Stop() error {
	m.mu.Lock()
	srv := m.server
	cancel := m.cancel
	m.server = nil
	m.address = ""
	m.token = ""
	m.cancel = nil
	m.mu.Unlock()

	if srv == nil {
		return nil
	}

	if cancel != nil {
		cancel()
	}
	ctx, c := context.WithTimeout(context.Background(), 5*time.Second)
	defer c()
	return srv.Stop(ctx)
}

// GetStatus 获取运行状态
func (m *managerImpl) GetStatus() *ServerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.server == nil {
		return nil
	}
	return &ServerInfo{
		Address: m.address,
		Token:   m.token,
		Enabled: true,
	}
}

// StopAll 停止全局 MCP Server
func (m *managerImpl) StopAll() error {
	return m.Stop()
}

func (m *managerImpl) watch(ctx context.Context, srv *Server) {
	// 等待外部取消信号
	<-ctx.Done()

	m.mu.Lock()
	if m.server == srv {
		m.server = nil
		m.address = ""
		m.token = ""
	}
	m.mu.Unlock()
}
