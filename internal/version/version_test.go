package version

import (
	"strings"
	"testing"
)

func TestDefaults(t *testing.T) {
	if Version != "dev" {
		t.Skip("version vars set via -ldflags")
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
	wantSub := "commit=unknown"
	if Version != "dev" {
		wantSub = "commit=" + Commit
	}
	if !strings.Contains(info, wantSub) {
		t.Errorf("Info() = %q, want substring %q", info, wantSub)
	}
}
