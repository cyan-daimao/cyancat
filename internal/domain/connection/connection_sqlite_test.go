package connection

import (
	"testing"

	"cyancat/internal/infra/driver"
)

func TestSQLiteConnectionValidationAllowsFileOnly(t *testing.T) {
	conn := &Connection{
		Name:  "local sqlite",
		Type:  driver.SQLite,
		Host:  "/tmp/example.sqlite",
		Group: GroupDevelopment,
	}

	if err := conn.prepareSave(); err != nil {
		t.Fatalf("expected sqlite file-only config to be valid, got %v", err)
	}
	if conn.Port != 0 {
		t.Fatalf("expected sqlite port to stay 0, got %d", conn.Port)
	}
}

func TestSQLiteConnectionValidationRequiresFile(t *testing.T) {
	conn := &Connection{
		Name:  "local sqlite",
		Type:  driver.SQLite,
		Group: GroupDevelopment,
	}

	if err := conn.prepareSave(); err == nil {
		t.Fatalf("expected sqlite config without file path to fail")
	}
}
