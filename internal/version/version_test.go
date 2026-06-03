package version

import "testing"

func TestDefaults(t *testing.T) {
	if Version != "dev" {
		t.Errorf("Version = %q, want dev", Version)
	}
	if Commit != "unknown" {
		t.Errorf("Commit = %q, want unknown", Commit)
	}
	if Ref != "unknown" {
		t.Errorf("Ref = %q, want unknown", Ref)
	}
}
