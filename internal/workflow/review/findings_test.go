package review

import (
	"regexp"
	"testing"
)

func TestParseFindings(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		want    []Finding
	}{
		{
			name: "clean block",
			input: `Review complete.
===ralph-findings===
[]
===/ralph-findings===
`,
			wantLen: 0,
		},
		{
			name: "single finding",
			input: `===ralph-findings===
[{"category":"security","path":"internal/auth.go","line":10,"summary":"hardcoded secret"}]
===/ralph-findings===`,
			wantLen: 1,
			want: []Finding{{
				Category: "security",
				Path:     "internal/auth.go",
				Line:     10,
				Summary:  "hardcoded secret",
			}},
		},
		{
			name: "multiple findings",
			input: `===ralph-findings===
[
  {"category":"bug","path":"a.go","summary":"nil deref"},
  {"category":"style","path":"b.go","summary":"long function"}
]
===/ralph-findings===`,
			wantLen: 2,
			want: []Finding{
				{Category: "bug", Path: "a.go", Summary: "nil deref"},
				{Category: "style", Path: "b.go", Summary: "long function"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFindings(tt.input)
			if err != nil {
				t.Fatalf("ParseFindings() err = %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(got), tt.wantLen)
			}
			for i := range tt.want {
				if got[i].Category != tt.want[i].Category {
					t.Errorf("[%d] Category = %q, want %q", i, got[i].Category, tt.want[i].Category)
				}
				if got[i].Path != tt.want[i].Path {
					t.Errorf("[%d] Path = %q, want %q", i, got[i].Path, tt.want[i].Path)
				}
				if got[i].Summary != tt.want[i].Summary {
					t.Errorf("[%d] Summary = %q, want %q", i, got[i].Summary, tt.want[i].Summary)
				}
				if got[i].Line != tt.want[i].Line {
					t.Errorf("[%d] Line = %d, want %d", i, got[i].Line, tt.want[i].Line)
				}
				if got[i].ID == "" {
					t.Errorf("[%d] ID is empty, want deterministic slug", i)
				}
				if i > 0 && got[i].ID == got[i-1].ID {
					t.Errorf("[%d] duplicate ID %q", i, got[i].ID)
				}
			}
			if tt.wantLen > 0 {
				again, err := ParseFindings(tt.input)
				if err != nil {
					t.Fatalf("ParseFindings() second call err = %v", err)
				}
				for i := range got {
					if got[i].ID != again[i].ID {
						t.Errorf("[%d] ID not stable: %q vs %q", i, got[i].ID, again[i].ID)
					}
				}
			}
		})
	}
}

func TestParseFindingsEmptyTranscript(t *testing.T) {
	got, err := ParseFindings("")
	if err != nil {
		t.Fatalf("ParseFindings() err = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ParseFindings() = %v, want no findings", got)
	}
}

func TestFingerprintEmptyFindings(t *testing.T) {
	tests := []struct {
		name     string
		findings []Finding
	}{
		{"nil", nil},
		{"empty slice", []Finding{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Fingerprint(tt.findings); got != "" {
				t.Errorf("Fingerprint() = %q, want empty string", got)
			}
		})
	}
}

func TestFingerprintNonEmptyFindings(t *testing.T) {
	findings := []Finding{
		{ID: "bug-f-go-x", Category: "bug", Path: "f.go", Summary: "x"},
	}
	got := Fingerprint(findings)
	if len(got) != 64 {
		t.Fatalf("Fingerprint() len = %d, want 64", len(got))
	}
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(got) {
		t.Errorf("Fingerprint() = %q, want 64-char lowercase hex", got)
	}
}

func TestFingerprintStableAcrossFindingOrder(t *testing.T) {
	a := Finding{ID: "alpha", Category: "bug", Path: "a.go", Summary: "one"}
	b := Finding{ID: "beta", Category: "style", Path: "b.go", Summary: "two"}
	first := Fingerprint([]Finding{b, a})
	second := Fingerprint([]Finding{a, b})
	if first != second {
		t.Errorf("fingerprints differ: %q vs %q", first, second)
	}
}

func TestFingerprintChangesWhenSummaryChanges(t *testing.T) {
	base := []Finding{{ID: "same-id", Category: "bug", Path: "f.go", Summary: "before"}}
	changed := []Finding{{ID: "same-id", Category: "bug", Path: "f.go", Summary: "after"}}
	if Fingerprint(base) == Fingerprint(changed) {
		t.Fatal("expected different fingerprints when summary changes")
	}
}
