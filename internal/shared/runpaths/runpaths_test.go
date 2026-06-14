package runpaths

import (
	"path/filepath"
	"testing"

	"ralph/internal/shared/runstate"
)

func TestPaths(t *testing.T) {
	workDir := filepath.Join(t.TempDir(), "work dir")
	cases := []struct {
		name  string
		runID string
	}{
		{name: "regular run", runID: "run-123"},
		{name: "prd local run", runID: runstate.LocalRunID},
		{name: "nested id", runID: "team/run-xyz"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wantDir := filepath.Join(workDir, ".ralph", "runs", tc.runID)
			if got := RunDir(workDir, tc.runID); got != wantDir {
				t.Fatalf("RunDir() = %q, want %q", got, wantDir)
			}
			if got := EventsPath(workDir, tc.runID); got != filepath.Join(wantDir, "events.ndjson") {
				t.Fatalf("EventsPath() = %q, want %q", got, filepath.Join(wantDir, "events.ndjson"))
			}
			if got := MetaPath(workDir, tc.runID); got != filepath.Join(wantDir, "meta.json") {
				t.Fatalf("MetaPath() = %q, want %q", got, filepath.Join(wantDir, "meta.json"))
			}
			if got := ReviewTranscriptPath(workDir, tc.runID, 7); got != filepath.Join(wantDir, "review-7.txt") {
				t.Fatalf("ReviewTranscriptPath() = %q, want %q", got, filepath.Join(wantDir, "review-7.txt"))
			}
		})
	}
}
