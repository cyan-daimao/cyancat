package postgres

import (
	"net/url"
	"testing"

	"cyancat/internal/infra/driver"
)

func TestBuildDSNDefaultsEmptyDatabaseToPostgres(t *testing.T) {
	dsn := buildDSN(driver.ConnConfig{
		Type:     driver.PostgreSQL,
		Host:     "db.example.com",
		Port:     5432,
		User:     "BASIC$dbuser_cdp_scheduler",
		Password: "secret",
	})

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	if got, want := u.Path, "/postgres"; got != want {
		t.Fatalf("database path = %q, want %q", got, want)
	}
	if got := u.User.Username(); got != "BASIC$dbuser_cdp_scheduler" {
		t.Fatalf("username = %q", got)
	}
}

func TestBuildDSNTrimsDatabase(t *testing.T) {
	dsn := buildDSN(driver.ConnConfig{
		Type:     driver.PostgreSQL,
		Host:     "db.example.com",
		Port:     5432,
		User:     "app",
		Password: "secret",
		Database: "  analytics  ",
	})

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	if got, want := u.Path, "/analytics"; got != want {
		t.Fatalf("database path = %q, want %q", got, want)
	}
}
