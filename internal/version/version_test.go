package version

import (
	"strings"
	"testing"
)

func TestDefaults(t *testing.T) {
	for name, val := range map[string]string{
		"Version": Version,
		"Commit":  Commit,
		"Ref":     Ref,
	} {
		if val == "" {
			t.Errorf("%s is empty", name)
		}
	}
}

func TestInfo(t *testing.T) {
	info := Info()
	if info == "" {
		t.Fatal("Info() returned empty string")
	}
	for _, want := range []string{
		"version=" + Version,
		"commit=" + Commit,
		"ref=" + Ref,
	} {
		if !strings.Contains(info, want) {
			t.Errorf("Info() = %q, want substring %q", info, want)
		}
	}
}
