package driver

import "testing"

func TestSQLiteDriverTypeIsValid(t *testing.T) {
	if !SQLite.IsValid() {
		t.Fatalf("expected sqlite driver type to be valid")
	}
}
