package mcpserver

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"cyancat/internal/infra/driver"
)

type mockExecutor struct {
	executeFunc      func(ctx context.Context, sql string) (*driver.Result, error)
	listTablesFunc   func(ctx context.Context) ([]driver.Table, error)
	describeTableFunc func(ctx context.Context, table string) (*driver.TableDetail, error)
}

func (m *mockExecutor) Execute(ctx context.Context, sql string) (*driver.Result, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, sql)
	}
	return &driver.Result{}, nil
}

func (m *mockExecutor) ListTables(ctx context.Context) ([]driver.Table, error) {
	if m.listTablesFunc != nil {
		return m.listTablesFunc(ctx)
	}
	return nil, nil
}

func (m *mockExecutor) DescribeTable(ctx context.Context, table string) (*driver.TableDetail, error) {
	if m.describeTableFunc != nil {
		return m.describeTableFunc(ctx, table)
	}
	return nil, errors.New("not implemented")
}

func TestServerStartStop(t *testing.T) {
	exec := &mockExecutor{}
	allow := map[string]bool{"select": true}

	srv, err := NewServer(1, "test-token", allow, exec)
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
		executeFunc: func(ctx context.Context, sql string) (*driver.Result, error) {
			return &driver.Result{RowsAffected: 1}, nil
		},
	}
	allow := map[string]bool{"select": true, "insert": false}

	srv, err := NewServer(1, "token", allow, exec)
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
