package headless

import (
	"os/exec"
	"strings"
	"testing"
)

func TestDoesNotImportWebRunner(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", "github.com/tireymorris/ralph/internal/headless")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps: %v\n%s", err, out)
	}
	for dep := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if dep == "github.com/tireymorris/ralph/internal/web/runner" {
			t.Fatal("headless must not depend on internal/web/runner")
		}
	}
}
