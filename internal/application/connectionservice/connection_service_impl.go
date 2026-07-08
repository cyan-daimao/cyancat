package connectionservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cyancat/internal/domain/connection"
	"cyancat/internal/infra/api"
	"cyancat/internal/infra/driver"
	"cyancat/internal/infra/session"
)

// McpServerStopper MCP Server 停止器（由 mcpserver.Manager 实现）
type McpServerStopper interface {
	Stop(connID int64) error
}

// ConnectionServiceImpl 连接管理服务实现
type ConnectionServiceImpl struct {
	connectionRepository connection.Repository
	sessionMgr           session.Manager
	mcpStopper           McpServerStopper
}

// NewConnectionServiceImpl 构造 ConnectionServiceImpl
// sessionMgr 可为 nil（V1.0 早期场景仅做配置 CRUD），但建议总是注入
func NewConnectionServiceImpl(repo connection.Repository, sessionMgr session.Manager) *ConnectionServiceImpl {
	return &ConnectionServiceImpl{
		connectionRepository: repo,
		sessionMgr:           sessionMgr,
	}
}

// SetMcpStopper 设置 MCP Server 停止器，用于关闭连接时联动停止 MCP Server
func (s *ConnectionServiceImpl) SetMcpStopper(stopper McpServerStopper) {
	s.mcpStopper = stopper
}

func (s *ConnectionServiceImpl) stopMcpServer(connID int64) {
	if s.mcpStopper != nil {
		_ = s.mcpStopper.Stop(connID)
	}
}

// List 列出连接
func (s *ConnectionServiceImpl) List(query *ListConnectionQuery) ([]*ConnectionBO, error) {
	list, err := s.connectionRepository.List(ToListQuery(query))
	if err != nil {
		return nil, err
	}
	return ToConnectionBOs(list), nil
}

// Page 分页查询
func (s *ConnectionServiceImpl) Page(query *PageConnectionQuery) (*api.Page[*ConnectionBO], error) {
	if query == nil {
		query = &PageConnectionQuery{}
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	list, total, err := s.connectionRepository.Page(ToPageQuery(query))
	if err != nil {
		return nil, err
	}
	return api.NewPage(ToConnectionBOs(list), total, query.Page, query.PageSize), nil
}

// GetByID 按 ID 查询
func (s *ConnectionServiceImpl) GetByID(id int64) (*ConnectionBO, error) {
	if id <= 0 {
		return nil, errors.New("connectionservice: id must be positive")
	}
	c, err := s.connectionRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	return ToConnectionBO(c), nil
}

// Create 创建连接
func (s *ConnectionServiceImpl) Create(cmd *CreateConnectionCmd) (*ConnectionBO, error) {
	if cmd == nil {
		return nil, errors.New("connectionservice: cmd cannot be nil")
	}

	// 唯一性校验：同名连接禁止重复
	if existing, err := s.connectionRepository.GetByName(cmd.Name); err == nil && existing != nil {
		return nil, fmt.Errorf("connectionservice: connection name %q already exists", cmd.Name)
	}

	c := ToConnectionFromCreateCmd(cmd)
	if c == nil {
		return nil, errors.New("connectionservice: connection cannot be nil")
	}

	if err := c.Save(s.connectionRepository); err != nil {
		return nil, err
	}
	return ToConnectionBO(c), nil
}

// Update 更新连接
func (s *ConnectionServiceImpl) Update(id int64, cmd *UpdateConnectionCmd) (*ConnectionBO, error) {
	if cmd == nil {
		return nil, errors.New("connectionservice: cmd cannot be nil")
	}
	if id <= 0 {
		return nil, errors.New("connectionservice: id must be positive")
	}

	c, err := s.connectionRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errors.New("connectionservice: connection not found")
	}

	ApplyUpdateCmd(c, cmd)

	if err := c.Update(s.connectionRepository); err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

// Delete 删除连接（同时关闭活跃 session 与 MCP Server）
func (s *ConnectionServiceImpl) Delete(cmd *DeleteConnectionCmd) error {
	if cmd == nil {
		return errors.New("connectionservice: cmd cannot be nil")
	}
	if cmd.ID <= 0 {
		return errors.New("connectionservice: id must be positive")
	}
	if s.sessionMgr != nil {
		_ = s.sessionMgr.Close(cmd.ID)
	}
	s.stopMcpServer(cmd.ID)
	return (&connection.Connection{ID: cmd.ID}).Delete(s.connectionRepository)
}

// Test 测试连接（短连接 Ping，不缓存到 SessionManager）
func (s *ConnectionServiceImpl) Test(cmd *TestConnectionCmd) (*TestConnectionResult, error) {
	if cmd == nil {
		return nil, errors.New("connectionservice: cmd cannot be nil")
	}

	d, err := driver.Get(cmd.Type)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := d.Connect(ctx, driver.ConnConfig{
		Type:           cmd.Type,
		Host:           cmd.Host,
		Port:           cmd.Port,
		User:           cmd.User,
		Password:       cmd.Password,
		Database:       cmd.Database,
		SSL:            cmd.SSL,
		ConnectTimeout: 5 * time.Second,
	})
	if err != nil {
		return &TestConnectionResult{Success: false, Message: err.Error()}, err
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		return &TestConnectionResult{Success: false, Message: err.Error()}, err
	}

	// 尝试拿数据库版本（失败不影响测试结果）
	version, _ := conn.ServerVersion(ctx)

	return &TestConnectionResult{
		Success:       true,
		Message:       "connection ok",
		ServerVersion: version,
	}, nil
}

// Open 打开已保存的连接：从配置建立长连接，存入 SessionManager
func (s *ConnectionServiceImpl) Open(id int64) (*ConnectionBO, error) {
	if id <= 0 {
		return nil, errors.New("connectionservice: id must be positive")
	}
	if s.sessionMgr == nil {
		return nil, errors.New("connectionservice: session manager not configured")
	}

	c, err := s.connectionRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errors.New("connectionservice: connection not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := driver.ConnConfig{
		Type:           c.Type,
		Host:           c.Host,
		Port:           c.Port,
		User:           c.User,
		Password:       c.Password,
		Database:       c.Database,
		SSL:            c.SSL,
		ConnectTimeout: 10 * time.Second,
	}

	if err := s.sessionMgr.Open(ctx, id, cfg); err != nil {
		return nil, err
	}

	// 更新最后连接时间（领域方法 + 持久化）
	_ = c.TouchConnected()
	_ = c.Update(s.connectionRepository)

	return ToConnectionBO(c), nil
}

// Close 关闭已打开的连接（同时停止对应 MCP Server）
func (s *ConnectionServiceImpl) Close(id int64) error {
	if id <= 0 {
		return errors.New("connectionservice: id must be positive")
	}
	if s.sessionMgr == nil {
		return errors.New("connectionservice: session manager not configured")
	}
	s.stopMcpServer(id)
	return s.sessionMgr.Close(id)
}
