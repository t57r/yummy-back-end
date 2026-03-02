package recipes

import "testing"

func TestParseInt(t *testing.T) {
	tests := []struct {
		name string
		in   string
		def  int
		want int
	}{
		{name: "valid", in: "15", def: 20, want: 15},
		{name: "invalid", in: "abc", def: 20, want: 20},
		{name: "empty", in: "", def: 20, want: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseInt(tt.in, tt.def); got != tt.want {
				t.Fatalf("unexpected result: got=%d want=%d", got, tt.want)
			}
		})
	}
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		name string
		in   string
		def  int64
		want int64
	}{
		{name: "valid", in: "1956720", def: 0, want: 1956720},
		{name: "invalid", in: "nope", def: 0, want: 0},
		{name: "empty", in: "", def: 7, want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseInt64(tt.in, tt.def); got != tt.want {
				t.Fatalf("unexpected result: got=%d want=%d", got, tt.want)
			}
		})
	}
}

func TestClampInt(t *testing.T) {
	tests := []struct {
		name string
		v    int
		min  int
		max  int
		want int
	}{
		{name: "below min", v: 0, min: 1, max: 100, want: 1},
		{name: "within range", v: 20, min: 1, max: 100, want: 20},
		{name: "above max", v: 999, min: 1, max: 100, want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampInt(tt.v, tt.min, tt.max); got != tt.want {
				t.Fatalf("unexpected result: got=%d want=%d", got, tt.want)
			}
		})
	}
}
