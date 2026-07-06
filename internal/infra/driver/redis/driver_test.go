package redis

import (
	"testing"

	"cyancat/internal/infra/driver"
)

func TestRedisDriverType(t *testing.T) {
	d := New()
	if got := d.Type(); got != driver.Redis {
		t.Fatalf("expected type %q, got %q", driver.Redis, got)
	}
}

func TestRedisDriverTypeIsValid(t *testing.T) {
	if !driver.Redis.IsValid() {
		t.Fatalf("expected redis driver type to be valid")
	}
}

func TestParseDatabase(t *testing.T) {
	if parseDatabase("") != 0 {
		t.Fatalf("empty database should be 0")
	}
	if parseDatabase("3") != 3 {
		t.Fatalf("database 3 expected")
	}
	if parseDatabase("abc") != 0 {
		t.Fatalf("invalid database should fallback to 0")
	}
}

func TestTokenize(t *testing.T) {
	toks := tokenize("SCAN 0 MATCH user:* COUNT 10")
	want := []string{"SCAN", "0", "MATCH", "USER:*", "COUNT", "10"}
	if len(toks) != len(want) {
		t.Fatalf("unexpected tokens: %v", toks)
	}
	for i := range want {
		if toks[i] != want[i] {
			t.Fatalf("token[%d] = %q, want %q", i, toks[i], want[i])
		}
	}
}
