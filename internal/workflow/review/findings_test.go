package review

import "testing"

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
