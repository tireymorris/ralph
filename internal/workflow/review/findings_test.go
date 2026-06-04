package review

import (
	"regexp"
	"testing"
)

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
