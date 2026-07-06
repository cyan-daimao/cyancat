package kafka

import (
	"testing"

	"cyancat/internal/infra/driver"
)

func TestKafkaDriverType(t *testing.T) {
	d := New()
	if got := d.Type(); got != driver.Kafka {
		t.Fatalf("expected type %q, got %q", driver.Kafka, got)
	}
}

func TestKafkaDriverTypeIsValid(t *testing.T) {
	if !driver.Kafka.IsValid() {
		t.Fatalf("expected kafka driver type to be valid")
	}
}

func TestParseBrokers(t *testing.T) {
	cases := []struct {
		host string
		port int
		want []string
	}{
		{"127.0.0.1", 9092, []string{"127.0.0.1:9092"}},
		{"127.0.0.1:9093", 9092, []string{"127.0.0.1:9093"}},
		{"a:9092,b:9093", 0, []string{"a:9092", "b:9093"}},
	}
	for _, c := range cases {
		got := parseBrokers(c.host, c.port)
		if len(got) == 0 {
			t.Fatalf("parseBrokers(%q, %d) returned empty", c.host, c.port)
		}
		if got[0] != c.want[0] {
			t.Fatalf("parseBrokers(%q, %d)[0] = %q, want %q", c.host, c.port, got[0], c.want[0])
		}
		if len(got) != len(c.want) {
			t.Fatalf("parseBrokers(%q, %d) len = %d, want %d", c.host, c.port, len(got), len(c.want))
		}
	}
}
