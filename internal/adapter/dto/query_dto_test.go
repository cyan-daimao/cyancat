package dto

import (
	"math"
	"testing"
	"time"
)

func ptr(s string) *string {
	return &s
}

func TestFormatCell(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want *string
	}{
		{"nil", nil, nil},
		{"string", "hello", ptr("hello")},
		{"bytes", []byte("world"), ptr("world")},
		{"int", int(42), ptr("42")},
		{"int8", int8(8), ptr("8")},
		{"int16", int16(16), ptr("16")},
		{"int32", int32(32), ptr("32")},
		{"int64", int64(64), ptr("64")},
		{"int64 max", int64(math.MaxInt64), ptr("9223372036854775807")},
		{"int64 min", int64(math.MinInt64), ptr("-9223372036854775808")},
		{"uint", uint(42), ptr("42")},
		{"uint64 max", uint64(math.MaxUint64), ptr("18446744073709551615")},
		{"float32", float32(3.14), ptr("3.14")},
		{"float64", float64(2.718), ptr("2.718")},
		{"bool true", true, ptr("true")},
		{"bool false", false, ptr("false")},
		{"time", time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC), ptr("2024-01-02 03:04:05")},
		{"other", struct{ X int }{X: 9}, ptr("{9}")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := formatCell(c.in)
			if c.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %q", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected %q, got nil", *c.want)
			}
			if *got != *c.want {
				t.Fatalf("formatCell(%#v) = %q, want %q", c.in, *got, *c.want)
			}
		})
	}
}

func TestToStringRows(t *testing.T) {
	in := [][]any{
		{int64(9223372036854775807), "text", nil},
		{true, []byte("bytes"), 3.14},
	}
	got := ToStringRows(in)
	if len(got) != 2 || len(got[0]) != 3 || len(got[1]) != 3 {
		t.Fatalf("unexpected shape: %#v", got)
	}
	if *got[0][0] != "9223372036854775807" {
		t.Fatalf("expected int64 max as string, got %q", *got[0][0])
	}
	if *got[0][1] != "text" {
		t.Fatalf("expected text unchanged, got %q", *got[0][1])
	}
	if got[0][2] != nil {
		t.Fatalf("expected nil for null cell, got %v", got[0][2])
	}
	if *got[1][0] != "true" || *got[1][1] != "bytes" || *got[1][2] != "3.14" {
		t.Fatalf("unexpected second row: %#v", got[1])
	}
}

func TestToStringRowsEmpty(t *testing.T) {
	got := ToStringRows(nil)
	if got == nil || len(got) != 0 {
		t.Fatalf("expected empty non-nil slice, got %#v", got)
	}
	got = ToStringRows([][]any{})
	if got == nil || len(got) != 0 {
		t.Fatalf("expected empty non-nil slice, got %#v", got)
	}
}
