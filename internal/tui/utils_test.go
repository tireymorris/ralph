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

// TestMin removed - min() is now a built-in function in Go 1.21+
