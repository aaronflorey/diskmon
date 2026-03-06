//go:build cgo

package storage

import "testing"

func TestClassifyAttribute(t *testing.T) {
	cases := []struct {
		name string
		in   AttributePoint
		want string
	}{
		{
			name: "no threshold green",
			in:   AttributePoint{Threshold: 0, Value: 1},
			want: "GREEN",
		},
		{
			name: "at threshold red",
			in:   AttributePoint{Threshold: 10, Value: 10},
			want: "RED",
		},
		{
			name: "below threshold red",
			in:   AttributePoint{Threshold: 10, Value: 9},
			want: "RED",
		},
		{
			name: "within warn margin yellow",
			in:   AttributePoint{Threshold: 100, Value: 108},
			want: "YELLOW",
		},
		{
			name: "small threshold min margin yellow",
			in:   AttributePoint{Threshold: 5, Value: 6},
			want: "YELLOW",
		},
		{
			name: "well above threshold green",
			in:   AttributePoint{Threshold: 100, Value: 130},
			want: "GREEN",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyAttribute(tc.in)
			if got != tc.want {
				t.Fatalf("classifyAttribute(%+v)=%s want %s", tc.in, got, tc.want)
			}
		})
	}
}

func TestNullHelpers(t *testing.T) {
	if got := nullInt(nil); got != nil {
		t.Fatalf("expected nil for nullInt(nil), got %#v", got)
	}
	if got := nullInt64(nil); got != nil {
		t.Fatalf("expected nil for nullInt64(nil), got %#v", got)
	}

	i := 7
	i64 := int64(9)
	if got := nullInt(&i); got != 7 {
		t.Fatalf("expected 7, got %#v", got)
	}
	if got := nullInt64(&i64); got != int64(9) {
		t.Fatalf("expected 9, got %#v", got)
	}
}

