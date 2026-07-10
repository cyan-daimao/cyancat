package parser

import (
	"testing"
)

func TestMySQLParseTableAlias(t *testing.T) {
	p := NewMySQL()
	sql := `SELECT a.id, b.name FROM users a JOIN orders b ON a.id = b.user_id`
	res, err := p.Parse(sql, 1, 10) // 光标在 SELECT 后
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(res.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(res.Tables))
	}
	if res.Tables[0].Name != "users" || res.Tables[0].Alias != "a" {
		t.Errorf("unexpected first table: %+v", res.Tables[0])
	}
	if res.Tables[1].Name != "orders" || res.Tables[1].Alias != "b" {
		t.Errorf("unexpected second table: %+v", res.Tables[1])
	}
}

func TestMySQLParseMemberAccess(t *testing.T) {
	p := NewMySQL()
	sql := `SELECT a. FROM users a`
	res, err := p.Parse(sql, 1, 10) // 光标在 a. 后
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if !res.IsMemberAccess {
		t.Fatalf("expected member access")
	}
	if res.TablePrefix != "a" {
		t.Errorf("expected prefix a, got %s", res.TablePrefix)
	}
	if res.Context != CtxColumn {
		t.Errorf("expected column context, got %s", res.Context)
	}
}

func TestPostgresParseTableAlias(t *testing.T) {
	p := NewPostgres()
	sql := `SELECT a.id, b.name FROM users a JOIN orders b ON a.id = b.user_id`
	res, err := p.Parse(sql, 1, 10)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(res.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(res.Tables))
	}
	foundUsers := false
	foundOrders := false
	for _, table := range res.Tables {
		if table.Name == "users" && table.Alias == "a" {
			foundUsers = true
		}
		if table.Name == "orders" && table.Alias == "b" {
			foundOrders = true
		}
	}
	if !foundUsers {
		t.Errorf("users table not found: %+v", res.Tables)
	}
	if !foundOrders {
		t.Errorf("orders table not found: %+v", res.Tables)
	}
}

func TestPostgresParseMemberAccess(t *testing.T) {
	p := NewPostgres()
	sql := `SELECT a. FROM users a`
	res, err := p.Parse(sql, 1, 10)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if !res.IsMemberAccess {
		t.Fatalf("expected member access")
	}
	if res.TablePrefix != "a" {
		t.Errorf("expected prefix a, got %s", res.TablePrefix)
	}
}
