package version

import (
	"strings"
	"testing"
)

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

func TestInfo(t *testing.T) {
	info := Info()
	if info == "" {
		t.Fatal("Info() returned empty string")
	}
	if !strings.Contains(info, "commit=unknown") {
		t.Errorf("Info() = %q, want substring commit=unknown", info)
	}
}
