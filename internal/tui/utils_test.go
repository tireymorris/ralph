package tui

import (
	"strings"
	"testing"
)

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

func TestWrapText(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		width        int
		want         string
		wantLines    int
		minLines     int
		wantNonEmpty bool
	}{
		{
			name:  "short text unchanged at width 80",
			s:     "hello",
			width: 80,
			want:  "hello",
		},
		{
			name:  "empty string",
			s:     "",
			width: 40,
			want:  "",
		},
		{
			name:  "width zero returns input unchanged",
			s:     "hello world",
			width: 0,
			want:  "hello world",
		},
		{
			name:     "long text wraps at width 40",
			s:        strings.Repeat("a", 100),
			width:    40,
			minLines: 2,
		},
		{
			name:         "unbroken token at width 20",
			s:            strings.Repeat("x", 50),
			width:        20,
			wantNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.s, tt.width)
			if tt.want != "" && got != tt.want {
				t.Errorf("wrapText(%q, %d) = %q, want %q", tt.s, tt.width, got, tt.want)
			}
			if tt.minLines > 0 {
				lines := strings.Split(got, "\n")
				if len(lines) < tt.minLines {
					t.Errorf("wrapText(%q, %d) produced %d lines, want at least %d", tt.s, tt.width, len(lines), tt.minLines)
				}
			}
			if tt.wantLines > 0 && strings.Count(got, "\n") != tt.wantLines-1 {
				t.Errorf("wrapText(%q, %d) line count = %d, want %d lines", tt.s, tt.width, strings.Count(got, "\n")+1, tt.wantLines)
			}
			if tt.wantNonEmpty && got == "" {
				t.Errorf("wrapText(%q, %d) = empty, want non-empty output", tt.s, tt.width)
			}
		})
	}
}
