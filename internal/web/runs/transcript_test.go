package runs_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ralph/internal/web/runs"
)

func TestReadEventTranscript_last200Lines(t *testing.T) {
	workDir := t.TempDir()
	runID := "run-tx"
	dir := filepath.Join(workDir, ".ralph", "runs", runID)
	if err := os.MkdirAll(dir, 0750); err != nil {
		t.Fatal(err)
	}
	var lines []string
	for i := 1; i <= 250; i++ {
		lines = append(lines, fmt.Sprintf(`{"n":%d}`, i))
	}
	path := filepath.Join(dir, "events.ndjson")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := runs.ReadEventTranscript(workDir, runID, 200)
	if err != nil {
		t.Fatalf("ReadEventTranscript: %v", err)
	}
	if !strings.Contains(got, `{"n":51}`) {
		t.Fatal("transcript should include line 51 (first of last 200)")
	}
	if strings.Contains(got, `{"n":50}`) {
		t.Fatal("transcript should not include line 50")
	}
	if !strings.Contains(got, `{"n":250}`) {
		t.Fatal("transcript should include last line")
	}
}
