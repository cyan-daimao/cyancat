// Package mcpserver 提供全局单例 MCP Server 及其管理器实现
package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"syscall"

	"cyancat/internal/infra/driver"

	"github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"
)

// ConnectionInfo 连接信息（用于 list_connections 工具）
type ConnectionInfo struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// Name 连接名称
	Name string `json:"name"`
	// Type 数据库类型
	Type string `json:"type"`
}

// SQLExecutor 执行 SQL 与查询元数据的接口（全局模式，按 connID 路由）
type SQLExecutor interface {
	// Execute 在指定连接上执行 SQL 并返回结果
	Execute(ctx context.Context, connID int64, sql string) (*driver.Result, error)
	// ListTables 列出指定连接默认数据库/schema 下的表
	ListTables(ctx context.Context, connID int64) ([]driver.Table, error)
	// DescribeTable 描述指定连接中指定表的结构
	DescribeTable(ctx context.Context, connID int64, table string) (*driver.TableDetail, error)
	// ListConnections 列出所有已打开的连接
	ListConnections(ctx context.Context) ([]ConnectionInfo, error)
}

// Server 全局 MCP Server 包装器
type Server struct {
	token       string
	allow       map[string]bool
	executor    SQLExecutor
	mcpServer   *mcpsrv.MCPServer
	sseServer   *mcpsrv.SSEServer
	httpServer  *http.Server
	listener    net.Listener
	address     string
	desiredPort int
}

// ErrPortConflict 指定端口被占用
var ErrPortConflict = errors.New("mcpserver: port conflict")

// NewServer 创建并配置全局 MCP Server
func NewServer(token string, allow map[string]bool, executor SQLExecutor) (*Server, error) {
	if token == "" {
		return nil, errors.New("mcpserver: token cannot be empty")
	}
	if executor == nil {
		return nil, errors.New("mcpserver: executor cannot be nil")
	}

	s := &Server{
		token:    token,
		allow:    allow,
		executor: executor,
	}

	mcpServer := mcpsrv.NewMCPServer(
		"dbstudio-mcp",
		"1.0.0",
		mcpsrv.WithToolCapabilities(false),
	)

	s.registerTools(mcpServer)

	s.mcpServer = mcpServer
	return s, nil
}

func (s *Server) registerTools(mcpServer *mcpsrv.MCPServer) {
	// list_connections: 列出所有已打开的数据库连接
	listConnsTool := mcp.NewTool("list_connections",
		mcp.WithDescription("List all opened database connections. Use the returned conn_id for other tools."),
	)
	mcpServer.AddTool(listConnsTool, s.handleListConnections)

	// query: SELECT-like SQL（需要 conn_id）
	queryTool := mcp.NewTool("query",
		mcp.WithDescription("Execute SELECT-like SQL (SELECT, WITH, SHOW, DESC, DESCRIBE, EXPLAIN) on a specific connection."),
		mcp.WithNumber("conn_id", mcp.Description("The connection ID (from list_connections)"), mcp.Required()),
		mcp.WithString("sql", mcp.Description("The SQL to execute"), mcp.Required()),
	)
	mcpServer.AddTool(queryTool, s.handleQuery)

	// execute: INSERT/UPDATE/DELETE（需要 conn_id）
	executeTool := mcp.NewTool("execute",
		mcp.WithDescription("Execute INSERT, UPDATE or DELETE SQL on a specific connection."),
		mcp.WithNumber("conn_id", mcp.Description("The connection ID (from list_connections)"), mcp.Required()),
		mcp.WithString("sql", mcp.Description("The SQL to execute"), mcp.Required()),
	)
	mcpServer.AddTool(executeTool, s.handleExecute)

	// execute_ddl: DDL（需要 conn_id）
	ddlTool := mcp.NewTool("execute_ddl",
		mcp.WithDescription("Execute DDL SQL such as CREATE, ALTER, DROP, TRUNCATE, RENAME on a specific connection."),
		mcp.WithNumber("conn_id", mcp.Description("The connection ID (from list_connections)"), mcp.Required()),
		mcp.WithString("sql", mcp.Description("The DDL SQL to execute"), mcp.Required()),
	)
	mcpServer.AddTool(ddlTool, s.handleDDL)

	// list_tables: list tables with optional pattern filter（需要 conn_id）
	listTablesTool := mcp.NewTool("list_tables",
		mcp.WithDescription("List tables in the default database/schema of a specific connection. Use 'pattern' to filter by table name and 'comment' to filter by table comment (both case-insensitive substring match)."),
		mcp.WithNumber("conn_id", mcp.Description("The connection ID (from list_connections)"), mcp.Required()),
		mcp.WithString("pattern",
			mcp.Description("Optional filter: only return tables whose names contain this substring (case-insensitive)"),
		),
		mcp.WithString("comment",
			mcp.Description("Optional filter: only return tables whose comments contain this substring (case-insensitive). SQLite has no table comments, so this always returns empty on SQLite."),
		),
	)
	mcpServer.AddTool(listTablesTool, s.handleListTables)

	// describe_table: describe table structure（需要 conn_id）
	describeTableTool := mcp.NewTool("describe_table",
		mcp.WithDescription("Describe the structure of a table (columns, indexes, foreign keys) on a specific connection."),
		mcp.WithNumber("conn_id", mcp.Description("The connection ID (from list_connections)"), mcp.Required()),
		mcp.WithString("table", mcp.Description("The table name"), mcp.Required()),
	)
	mcpServer.AddTool(describeTableTool, s.handleDescribeTable)
}

// Start 启动 MCP Server，返回访问地址
func (s *Server) Start() (string, error) {
	addr := "127.0.0.1:0"
	if s.desiredPort > 0 {
		addr = fmt.Sprintf("127.0.0.1:%d", s.desiredPort)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		if s.desiredPort > 0 && errors.Is(err, syscall.EADDRINUSE) {
			return "", fmt.Errorf("%w: port %d: %w", ErrPortConflict, s.desiredPort, err)
		}
		return "", fmt.Errorf("mcpserver: listen failed: %w", err)
	}
	s.listener = listener

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	sseServer := mcpsrv.NewSSEServer(
		s.mcpServer,
		mcpsrv.WithBaseURL(baseURL),
	)
	s.sseServer = sseServer

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer "+s.token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		sseServer.ServeHTTP(w, r)
	})

	s.httpServer = &http.Server{
		Handler: handler,
	}

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// 忽略正常关闭错误
		}
	}()

	endpoint, err := sseServer.CompleteSseEndpoint()
	if err != nil {
		return "", fmt.Errorf("mcpserver: complete sse endpoint failed: %w", err)
	}
	s.address = endpoint
	return s.address, nil
}

// Stop 停止 MCP Server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		_ = s.httpServer.Shutdown(ctx)
	}
	if s.sseServer != nil {
		_ = s.sseServer.Shutdown(ctx)
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
	return nil
}

// Address 返回访问地址
func (s *Server) Address() string {
	return s.address
}

// Token 返回访问令牌
func (s *Server) Token() string {
	return s.token
}

func (s *Server) handleListConnections(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	conns, err := s.executor.ListConnections(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return formatValue(conns), nil
}

func (s *Server) handleQuery(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.allow["select"] {
		return mcp.NewToolResultError("SELECT is not allowed"), nil
	}

	connID, sqlText, err := getConnIDAndSQL(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !isSelectLike(sqlText) {
		return mcp.NewToolResultError("only SELECT-like statements are allowed by query tool"), nil
	}

	result, err := s.executor.Execute(ctx, connID, sqlText)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return formatResult(result), nil
}

func (s *Server) handleExecute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	connID, sqlText, err := getConnIDAndSQL(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	kind := classifySQL(sqlText)
	switch kind {
	case "insert":
		if !s.allow["insert"] {
			return mcp.NewToolResultError("INSERT is not allowed"), nil
		}
	case "update":
		if !s.allow["update"] {
			return mcp.NewToolResultError("UPDATE is not allowed"), nil
		}
		if !hasWhereClause(sqlText) {
			return mcp.NewToolResultError("UPDATE without WHERE is not allowed"), nil
		}
	case "delete":
		if !s.allow["delete"] {
			return mcp.NewToolResultError("DELETE is not allowed"), nil
		}
		if !hasWhereClause(sqlText) {
			return mcp.NewToolResultError("DELETE without WHERE is not allowed"), nil
		}
	default:
		return mcp.NewToolResultError("only INSERT, UPDATE or DELETE are allowed by execute tool"), nil
	}

	result, err := s.executor.Execute(ctx, connID, sqlText)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("OK, rows affected: %d, last insert id: %d", result.RowsAffected, result.LastInsertID)), nil
}

func (s *Server) handleDDL(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.allow["ddl"] {
		return mcp.NewToolResultError("DDL is not allowed"), nil
	}

	connID, sqlText, err := getConnIDAndSQL(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if classifySQL(sqlText) != "ddl" {
		return mcp.NewToolResultError("only DDL statements are allowed by execute_ddl tool"), nil
	}

	result, err := s.executor.Execute(ctx, connID, sqlText)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("OK, rows affected: %d", result.RowsAffected)), nil
}

func (s *Server) handleListTables(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	connID, err := getConnID(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pattern, _ := args["pattern"].(string)
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	commentPattern, _ := args["comment"].(string)
	commentPattern = strings.ToLower(strings.TrimSpace(commentPattern))

	tables, err := s.executor.ListTables(ctx, connID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var names []string
	for _, t := range tables {
		if pattern != "" && !strings.Contains(strings.ToLower(t.Name), pattern) {
			continue
		}
		if commentPattern != "" && !strings.Contains(strings.ToLower(t.Comment), commentPattern) {
			continue
		}
		names = append(names, t.Name)
	}
	return formatValue(names), nil
}

func (s *Server) handleDescribeTable(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	connID, err := getConnID(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	table, _ := args["table"].(string)
	if table == "" {
		return mcp.NewToolResultError("table is required"), nil
	}

	detail, err := s.executor.DescribeTable(ctx, connID, table)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return formatValue(detail), nil
}

// getConnID 从工具参数中解析 conn_id（必填）
func getConnID(args map[string]any) (int64, error) {
	v, ok := args["conn_id"]
	if !ok || v == nil {
		return 0, errors.New("conn_id is required (use list_connections to get available connections)")
	}
	switch val := v.(type) {
	case float64:
		id := int64(val)
		if id <= 0 {
			return 0, errors.New("conn_id must be positive")
		}
		return id, nil
	case int64:
		if val <= 0 {
			return 0, errors.New("conn_id must be positive")
		}
		return val, nil
	case int:
		if val <= 0 {
			return 0, errors.New("conn_id must be positive")
		}
		return int64(val), nil
	default:
		return 0, fmt.Errorf("conn_id must be a number, got %T", v)
	}
}

func getConnIDAndSQL(request mcp.CallToolRequest) (int64, string, error) {
	args := request.GetArguments()
	connID, err := getConnID(args)
	if err != nil {
		return 0, "", err
	}
	sqlText, _ := args["sql"].(string)
	if sqlText == "" {
		return 0, "", errors.New("sql is required")
	}
	return connID, sqlText, nil
}

func isSelectLike(sql string) bool {
	lower := strings.ToLower(strings.TrimSpace(sql))
	prefixes := []string{"select", "with", "show", "desc", "describe", "explain"}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

func classifySQL(sql string) string {
	lower := strings.ToLower(strings.TrimSpace(sql))
	switch {
	case strings.HasPrefix(lower, "select"), strings.HasPrefix(lower, "with"), strings.HasPrefix(lower, "show"),
		strings.HasPrefix(lower, "desc"), strings.HasPrefix(lower, "describe"), strings.HasPrefix(lower, "explain"):
		return "select"
	case strings.HasPrefix(lower, "insert"):
		return "insert"
	case strings.HasPrefix(lower, "update"):
		return "update"
	case strings.HasPrefix(lower, "delete"):
		return "delete"
	case strings.HasPrefix(lower, "create"), strings.HasPrefix(lower, "alter"), strings.HasPrefix(lower, "drop"),
		strings.HasPrefix(lower, "truncate"), strings.HasPrefix(lower, "rename"):
		return "ddl"
	}
	return "other"
}

// hasWhereClause 检查 SQL 是否在顶层包含 WHERE 子句（忽略字符串字面量、注释和子查询中的 where）。
func hasWhereClause(sql string) bool {
	lower := strings.ToLower(sql)
	depth := 0
	inString := false
	var stringChar byte

	for i := 0; i < len(lower); {
		c := lower[i]

		// 字符串字面量（支持 ' " `，以及 MySQL 风格的 \ 转义）
		if inString {
			if c == '\\' && i+1 < len(lower) {
				i += 2
				continue
			}
			if c == stringChar {
				inString = false
			}
			i++
			continue
		}
		if c == '\'' || c == '"' || c == '`' {
			inString = true
			stringChar = c
			i++
			continue
		}

		// 行注释 --
		if c == '-' && i+1 < len(lower) && lower[i+1] == '-' {
			for i < len(lower) && lower[i] != '\n' {
				i++
			}
			continue
		}

		// 块注释 /* */
		if c == '/' && i+1 < len(lower) && lower[i+1] == '*' {
			i += 2
			for i < len(lower)-1 && !(lower[i] == '*' && lower[i+1] == '/') {
				i++
			}
			i += 2
			continue
		}

		// 括号深度
		if c == '(' {
			depth++
			i++
			continue
		}
		if c == ')' {
			if depth > 0 {
				depth--
			}
			i++
			continue
		}

		// 顶层 WHERE 关键字（要求不在字符串、注释、子查询内）
		if depth == 0 && i+5 <= len(lower) && lower[i:i+5] == "where" {
			beforeOK := i == 0 || !isIdentChar(lower[i-1])
			afterOK := i+5 == len(lower) || !isIdentChar(lower[i+5])
			if beforeOK && afterOK {
				return true
			}
		}

		i++
	}
	return false
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
}

func formatResult(result *driver.Result) *mcp.CallToolResult {
	if len(result.Rows) == 0 {
		return mcp.NewToolResultText("OK, 0 rows returned")
	}
	return mcp.NewToolResultText(formatRowsAsMarkdown(result.Rows, result.Columns))
}

func formatValue(v any) *mcp.CallToolResult {
	// 优先尝试 Markdown 表格（无列名）
	if rows, ok := v.([][]any); ok {
		return mcp.NewToolResultText(formatRowsAsMarkdown(rows, nil))
	}

	text, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format result: %v", err))
	}
	return mcp.NewToolResultText(string(text))
}

func formatRowsAsMarkdown(rows [][]any, columns []driver.Column) string {
	if len(rows) == 0 {
		return "OK, 0 rows returned"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d rows returned\n\n", len(rows)))

	if len(columns) > 0 {
		for i, col := range columns {
			if i > 0 {
				b.WriteString(" | ")
			}
			b.WriteString(col.Name)
		}
		b.WriteString("\n")
		for i := range columns {
			if i > 0 {
				b.WriteString(" | ")
			}
			b.WriteString("---")
		}
		b.WriteString("\n")
	}

	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				b.WriteString(" | ")
			}
			b.WriteString(fmt.Sprintf("%v", cell))
		}
		b.WriteString("\n")
	}
	return b.String()
}
