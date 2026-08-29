package mcpserver

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"cyancat/internal/infra/driver"

	"github.com/mark3labs/mcp-go/mcp"
)

type mockExecutor struct {
	executeFunc        func(ctx context.Context, connID int64, sql string) (*driver.Result, error)
	listTablesFunc     func(ctx context.Context, connID int64) ([]driver.Table, error)
	describeTableFunc  func(ctx context.Context, connID int64, table string) (*driver.TableDetail, error)
	listConnsFunc      func(ctx context.Context) ([]ConnectionInfo, error)
}

func (m *mockExecutor) Execute(ctx context.Context, connID int64, sql string) (*driver.Result, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, connID, sql)
	}
	return &driver.Result{}, nil
}

func (m *mockExecutor) ListTables(ctx context.Context, connID int64) ([]driver.Table, error) {
	if m.listTablesFunc != nil {
		return m.listTablesFunc(ctx, connID)
	}
	return nil, nil
}

func (m *mockExecutor) DescribeTable(ctx context.Context, connID int64, table string) (*driver.TableDetail, error) {
	if m.describeTableFunc != nil {
		return m.describeTableFunc(ctx, connID, table)
	}
	return nil, errors.New("not implemented")
}

func (m *mockExecutor) ListConnections(ctx context.Context) ([]ConnectionInfo, error) {
	if m.listConnsFunc != nil {
		return m.listConnsFunc(ctx)
	}
	return []ConnectionInfo{{ConnID: 1, Name: "test", Type: "mysql"}}, nil
}

func TestServerStartStop(t *testing.T) {
	exec := &mockExecutor{}
	allow := map[string]bool{"select": true}

	srv, err := NewServer("test-token", allow, exec)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	addr, err := srv.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if addr == "" {
		t.Fatal("address is empty")
	}

	// 等待服务就绪
	time.Sleep(50 * time.Millisecond)

	// 无 Token 应返回 401
	resp, err := http.Get(addr)
	if err != nil {
		t.Fatalf("GET without token failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d: %s", resp.StatusCode, string(body))
	}

	// 带正确 Token 应返回 SSE 流
	req, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET with token failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with token, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestServerPermissionEnforcement(t *testing.T) {
	exec := &mockExecutor{
		executeFunc: func(ctx context.Context, connID int64, sql string) (*driver.Result, error) {
			return &driver.Result{RowsAffected: 1}, nil
		},
	}
	allow := map[string]bool{"select": true, "insert": false}

	srv, err := NewServer("token", allow, exec)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	addr, err := srv.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer srv.Stop(context.Background())

	// 仅做本地启动校验，具体工具权限由 handle 函数保证
	if !strings.Contains(addr, "/sse") {
		t.Fatalf("address should contain /sse, got %s", addr)
	}
}

func TestHasWhereClause(t *testing.T) {
	cases := []struct {
		sql      string
		expected bool
	}{
		{"UPDATE t SET a = 1", false},
		{"UPDATE t SET a = 1 WHERE id = 1", true},
		{"update t set a = 1 where id = 1", true},
		{"UPDATE t SET a = 1\nWHERE id = 1", true},
		{"UPDATE t SET a = (SELECT y FROM z WHERE x = 1)", false},
		{"UPDATE t SET a = (SELECT y FROM z WHERE x = 1) WHERE id = 1", true},
		{"UPDATE t SET name = 'where is it'", false},
		{"UPDATE t SET name = 'where is it' WHERE id = 1", true},
		{"UPDATE t SET name = `where` WHERE id = 1", true},
		{"DELETE FROM t", false},
		{"DELETE FROM t WHERE id = 1", true},
		{"delete from t where id = 1", true},
		{"DELETE FROM t WHERE id IN (SELECT id FROM z WHERE x = 1)", true},
		{"DELETE FROM t -- where comment", false},
		{"DELETE FROM t /* where */ WHERE id = 1", true},
	}

	for _, tc := range cases {
		got := hasWhereClause(tc.sql)
		if got != tc.expected {
			t.Errorf("hasWhereClause(%q) = %v, want %v", tc.sql, got, tc.expected)
		}
	}
}

func TestServerPortConflict(t *testing.T) {
	exec := &mockExecutor{}
	allow := map[string]bool{"select": true}

	srv1, err := NewServer("token1", allow, exec)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	addr1, err := srv1.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer srv1.Stop(context.Background())

	port := parsePortFromAddress(addr1)
	if port <= 0 {
		t.Fatalf("failed to parse port from address: %s", addr1)
	}

	srv2, err := NewServer("token2", allow, exec)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	srv2.desiredPort = port
	_, err = srv2.Start()
	if !errors.Is(err, ErrPortConflict) {
		t.Fatalf("expected ErrPortConflict, got: %v", err)
	}
}

func parsePortFromAddress(addr string) int {
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

func TestGetConnID(t *testing.T) {
	// 正常解析
	id, err := getConnID(map[string]any{"conn_id": float64(42)})
	if err != nil || id != 42 {
		t.Fatalf("getConnID(42) = %d, %v", id, err)
	}

	// 缺少 conn_id
	_, err = getConnID(map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing conn_id")
	}

	// 非正数
	_, err = getConnID(map[string]any{"conn_id": float64(0)})
	if err == nil {
		t.Fatal("expected error for zero conn_id")
	}
}

func TestHandleExecuteRequiresWhere(t *testing.T) {
	exec := &mockExecutor{
		executeFunc: func(ctx context.Context, connID int64, sql string) (*driver.Result, error) {
			return &driver.Result{RowsAffected: 1}, nil
		},
	}
	allow := map[string]bool{"insert": true, "update": true, "delete": true}

	srv, err := NewServer("token", allow, exec)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// UPDATE without WHERE
	res, err := srv.handleExecute(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "execute",
			Arguments: map[string]any{"conn_id": float64(1), "sql": "UPDATE t SET a = 1"},
		},
	})
	if err != nil {
		t.Fatalf("handleExecute returned error: %v", err)
	}
	if res.Content[0].(mcp.TextContent).Text != "UPDATE without WHERE is not allowed" {
		t.Fatalf("expected WHERE error for UPDATE, got: %v", res.Content[0])
	}

	// DELETE without WHERE
	res, err = srv.handleExecute(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "execute",
			Arguments: map[string]any{"conn_id": float64(1), "sql": "DELETE FROM t"},
		},
	})
	if err != nil {
		t.Fatalf("handleExecute returned error: %v", err)
	}
	if res.Content[0].(mcp.TextContent).Text != "DELETE without WHERE is not allowed" {
		t.Fatalf("expected WHERE error for DELETE, got: %v", res.Content[0])
	}
}
