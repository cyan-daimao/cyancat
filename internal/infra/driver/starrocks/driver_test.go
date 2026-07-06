package starrocks

import (
	"testing"

	"cyancat/internal/infra/driver"
)

func TestStarRocksDriverType(t *testing.T) {
	d := New()
	if got := d.Type(); got != driver.StarRocks {
		t.Fatalf("expected type %q, got %q", driver.StarRocks, got)
	}
}

func TestStarRocksDriverTypeIsValid(t *testing.T) {
	if !driver.StarRocks.IsValid() {
		t.Fatalf("expected starrocks driver type to be valid")
	}
}
