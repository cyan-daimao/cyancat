package driver

import (
	"math/big"
	"testing"
)

func newBigInt(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid big.Int string: " + s)
	}
	return n
}

func TestNormalizeValue(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want any
	}{
		{"small int64", int64(42), int64(42)},
		{"max safe int64", int64(maxSafeInteger), int64(maxSafeInteger)},
		{"unsafe int64", int64(9223372036854775807), "9223372036854775807"},
		{"negative unsafe int64", int64(-9223372036854775808), "-9223372036854775808"},
		{"safe uint64", uint64(maxSafeInteger), uint64(maxSafeInteger)},
		{"unsafe uint64", uint64(18446744073709551615), "18446744073709551615"},
		{"small int", int(42), int(42)},
		{"big.Int", newBigInt("12345678901234567890"), "12345678901234567890"},
		{"nil big.Int", (*big.Int)(nil), nil},
		{"string unchanged", "hello", "hello"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := NormalizeValue(c.in)
			if got != c.want {
				t.Fatalf("NormalizeValue(%#v) = %#v, want %#v", c.in, got, c.want)
			}
		})
	}
}

func TestNormalizeRows(t *testing.T) {
	rows := [][]any{
		{int64(42), uint64(18446744073709551615), "text"},
	}
	got := NormalizeRows(rows)
	if got[0][0] != int64(42) {
		t.Fatalf("expected small int64 unchanged, got %#v", got[0][0])
	}
	if got[0][1] != "18446744073709551615" {
		t.Fatalf("expected unsafe uint64 as string, got %#v", got[0][1])
	}
	if got[0][2] != "text" {
		t.Fatalf("expected text unchanged, got %#v", got[0][2])
	}
}
