package review

import (
	"regexp"
	"testing"
)

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
