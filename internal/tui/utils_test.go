package tui

import "testing"

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		max    int
		expect string
	}{
		{
			name:   "no truncation needed",
			s:      "hello",
			max:    10,
			expect: "hello",
		},
		{
			name:   "exact length",
			s:      "hello",
			max:    5,
			expect: "hello",
		},
		{
			name:   "truncation with ellipsis",
			s:      "hello world",
			max:    8,
			expect: "hello...",
		},
		{
			name:   "very short max",
			s:      "hello",
			max:    3,
			expect: "hel",
		},
		{
			name:   "max less than 4",
			s:      "hello",
			max:    2,
			expect: "he",
		},
		{
			name:   "empty string",
			s:      "",
			max:    10,
			expect: "",
		},
		{
			name:   "max equals 4",
			s:      "hello world",
			max:    4,
			expect: "h...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.s, tt.max)
			if got != tt.expect {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.expect)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{0, 0, 0},
		{-1, 1, -1},
		{100, 50, 50},
		{-5, -10, -10},
	}

	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
