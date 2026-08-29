package mcpservice

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"cyancat/internal/application/connectionservice"
	"cyancat/internal/application/queryservice"
	"cyancat/internal/infra/db/mcprepo"
	"cyancat/internal/infra/driver"
	"cyancat/internal/infra/mcpserver"
	"cyancat/internal/infra/session"
)

// McpServiceImpl MCP Server 应用服务实现（全局单例）
type McpServiceImpl struct {
	mcpRepo    *mcprepo.McpServerRepository
	connSvc    connectionservice.ConnectionService
	querySvc   queryservice.QueryService
	sessionMgr session.Manager
	mcpMgr     mcpserver.Manager
}

// NewMcpServiceImpl 创建 MCP Server 应用服务
func NewMcpServiceImpl(
	mcpRepo *mcprepo.McpServerRepository,
	connSvc connectionservice.ConnectionService,
	querySvc queryservice.QueryService,
	sessionMgr session.Manager,
	mcpMgr mcpserver.Manager,
) *McpServiceImpl {
	return &McpServiceImpl{
		mcpRepo:    mcpRepo,
		connSvc:    connSvc,
		querySvc:   querySvc,
		sessionMgr: sessionMgr,
		mcpMgr:     mcpMgr,
	}
}

// GetStatus 获取全局 MCP Server 状态
func (s *McpServiceImpl) GetStatus() (*McpServerStatusBO, error) {
	do, err := s.mcpRepo.GetGlobal()
	if err != nil {
		return nil, err
	}

	var bo *McpServerStatusBO
	if do != nil {
		bo = ToMcpServerStatusBO(do)
	} else {
		bo = &McpServerStatusBO{}
	}

	// Enabled/Address/Token 以内存实际运行状态为准：
	// 应用退出时仅停止内存服务、DB 中 enabled 仍为 true，
	// 若直接采信 DB，重启后会误显示"运行中"但服务并未运行
	if info := s.mcpMgr.GetStatus(); info != nil {
		bo.Enabled = true
		bo.Address = info.Address
		bo.Token = info.Token
	} else {
		bo.Enabled = false
		bo.Address = ""
		bo.Token = ""
	}

	return bo, nil
}

// Start 启动全局 MCP Server
func (s *McpServiceImpl) Start(cmd *StartMcpServerCmd) (*McpServerStatusBO, error) {
	if cmd == nil {
		return nil, errors.New("mcpservice: cmd cannot be nil")
	}

	// 查询历史配置
	existing, err := s.mcpRepo.GetGlobal()
	if err != nil {
		return nil, fmt.Errorf("mcpservice: get existing config failed: %w", err)
	}

	// 复用历史 token：保证停止再开启（含重启应用后再开启）后 SSE 地址与凭据不变，
	// agent 侧已安装的 claude mcp 配置无需重新安装
	token := ""
	if existing != nil {
		token = existing.Token
	}
	if token == "" {
		token, err = generateToken()
		if err != nil {
			return nil, fmt.Errorf("mcpservice: generate token failed: %w", err)
		}
	}

	bo := &McpServerStatusBO{
		Enabled:     true,
		Token:       token,
		AllowSelect: cmd.AllowSelect,
		AllowInsert: cmd.AllowInsert,
		AllowUpdate: cmd.AllowUpdate,
		AllowDelete: cmd.AllowDelete,
		AllowDDL:    cmd.AllowDDL,
	}

	do := ToMcpServerDO(bo)
	if err := s.mcpRepo.SaveOrUpdate(do); err != nil {
		return nil, fmt.Errorf("mcpservice: save config failed: %w", err)
	}

	executor := &globalQueryExecutor{
		connSvc:    s.connSvc,
		querySvc:   s.querySvc,
		sessionMgr: s.sessionMgr,
	}

	serverCfg := mcpserver.ServerConfig{
		AllowSelect: cmd.AllowSelect,
		AllowInsert: cmd.AllowInsert,
		AllowUpdate: cmd.AllowUpdate,
		AllowDelete: cmd.AllowDelete,
		AllowDDL:    cmd.AllowDDL,
		Token:       token,
		Executor:    executor,
	}
	// 未强制要求新端口时，尝试复用历史端口
	if !cmd.ForceNewPort && existing != nil && existing.Port > 0 {
		serverCfg.Port = existing.Port
	}

	info, err := s.mcpMgr.Start(serverCfg)
	if err != nil {
		if errors.Is(err, mcpserver.ErrPortConflict) {
			port := 0
			if existing != nil {
				port = existing.Port
			}
			return nil, &PortConflictError{Port: port}
		}
		return nil, fmt.Errorf("mcpservice: start mcp server failed: %w", err)
	}

	bo.Address = info.Address
	// 从地址解析端口并持久化
	bo.Port = parsePortFromAddress(info.Address)
	do = ToMcpServerDO(bo)
	if err := s.mcpRepo.SaveOrUpdate(do); err != nil {
		return nil, fmt.Errorf("mcpservice: save config with port failed: %w", err)
	}

	return bo, nil
}

// PortConflictError 历史端口被占用错误
type PortConflictError struct {
	Port int
}

func (e *PortConflictError) Error() string {
	return fmt.Sprintf("历史端口 %d 被占用，是否使用新端口？新端口需要让 agent 重新安装 mcp。", e.Port)
}

func parsePortFromAddress(addr string) int {
	// addr 形如 http://127.0.0.1:12345/sse
	if idx := strings.LastIndex(addr, ":"); idx >= 0 {
		portPart := addr[idx+1:]
		if slashIdx := strings.Index(portPart, "/"); slashIdx >= 0 {
			portPart = portPart[:slashIdx]
		}
		if port, err := strconv.Atoi(portPart); err == nil {
			return port
		}
	}
	return 0
}

// Stop 停止全局 MCP Server
func (s *McpServiceImpl) Stop() error {
	if err := s.mcpMgr.Stop(); err != nil {
		return err
	}

	do, err := s.mcpRepo.GetGlobal()
	if err != nil {
		return err
	}
	if do != nil {
		do.Enabled = false
		if err := s.mcpRepo.SaveOrUpdate(do); err != nil {
			return fmt.Errorf("mcpservice: update config failed: %w", err)
		}
	}
	return nil
}

func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// globalQueryExecutor 全局 SQL 执行器（按 connID 路由到对应连接）
type globalQueryExecutor struct {
	connSvc    connectionservice.ConnectionService
	querySvc   queryservice.QueryService
	sessionMgr session.Manager
}

func (e *globalQueryExecutor) Execute(ctx context.Context, connID int64, sql string) (*driver.Result, error) {
	if err := e.ensureOpen(connID); err != nil {
		return nil, err
	}
	bo, err := e.querySvc.Execute(&queryservice.ExecuteCmd{
		ConnID:  connID,
		SQL:     sql,
		MaxRows: 1000,
	})
	if err != nil {
		return nil, err
	}
	return queryResultBOToDriverResult(bo), nil
}

func (e *globalQueryExecutor) ListTables(ctx context.Context, connID int64) ([]driver.Table, error) {
	if err := e.ensureOpen(connID); err != nil {
		return nil, err
	}
	database, schema, err := e.currentDatabaseSchema(ctx, connID)
	if err != nil {
		return nil, err
	}

	conn, err := e.sessionMgr.Get(connID)
	if err != nil {
		return nil, err
	}
	return conn.Inspector().ListTables(ctx, database, schema, 0, 0)
}

func (e *globalQueryExecutor) DescribeTable(ctx context.Context, connID int64, table string) (*driver.TableDetail, error) {
	if strings.TrimSpace(table) == "" {
		return nil, errors.New("table name is required")
	}
	if err := e.ensureOpen(connID); err != nil {
		return nil, err
	}

	database, schema, err := e.currentDatabaseSchema(ctx, connID)
	if err != nil {
		return nil, err
	}

	conn, err := e.sessionMgr.Get(connID)
	if err != nil {
		return nil, err
	}
	return conn.Inspector().DescribeTable(ctx, database, schema, table)
}

func (e *globalQueryExecutor) ListConnections(ctx context.Context) ([]mcpserver.ConnectionInfo, error) {
	ids := e.sessionMgr.List()
	result := make([]mcpserver.ConnectionInfo, 0, len(ids))
	for _, id := range ids {
		bo, err := e.connSvc.GetByID(id)
		if err != nil || bo == nil {
			continue
		}
		result = append(result, mcpserver.ConnectionInfo{
			ConnID: id,
			Name:   bo.Name,
			Type:   string(bo.Type),
		})
	}
	return result, nil
}

// ensureOpen 确保连接已打开，未打开时自动打开
func (e *globalQueryExecutor) ensureOpen(connID int64) error {
	if connID <= 0 {
		return errors.New("mcpservice: connID must be positive")
	}
	if e.sessionMgr.IsOpen(connID) {
		return nil
	}
	if _, err := e.connSvc.Open(connID); err != nil {
		return fmt.Errorf("mcpservice: open connection %d failed: %w", connID, err)
	}
	return nil
}

func (e *globalQueryExecutor) currentDatabaseSchema(ctx context.Context, connID int64) (database, schema string, err error) {
	driverType, err := e.sessionMgr.DriverType(connID)
	if err != nil {
		return "", "", err
	}

	switch driverType {
	case driver.MySQL, driver.StarRocks:
		res, err := e.Execute(ctx, connID, "SELECT DATABASE()")
		if err != nil {
			return "", "", err
		}
		if len(res.Rows) > 0 && len(res.Rows[0]) > 0 {
			if db, ok := res.Rows[0][0].(string); ok {
				database = db
			}
		}
		return database, "", nil
	case driver.PostgreSQL:
		res, err := e.Execute(ctx, connID, "SELECT current_schema()")
		if err != nil {
			return "", "", err
		}
		if len(res.Rows) > 0 && len(res.Rows[0]) > 0 {
			if s, ok := res.Rows[0][0].(string); ok {
				schema = s
			}
		}
		return "", schema, nil
	case driver.SQLite:
		return "", "", nil
	default:
		return "", "", nil
	}
}

func queryResultBOToDriverResult(bo *queryservice.QueryResultBO) *driver.Result {
	if bo == nil {
		return &driver.Result{}
	}
	columns := make([]driver.Column, 0, len(bo.Columns))
	for _, c := range bo.Columns {
		columns = append(columns, driver.Column{
			Name:         c.Name,
			DatabaseType: c.DatabaseType,
			Nullable:     c.Nullable,
		})
	}
	return &driver.Result{
		Columns:      columns,
		Rows:         bo.Rows,
		RowsAffected: bo.RowsAffected,
		LastInsertID: bo.LastInsertID,
	}
}
