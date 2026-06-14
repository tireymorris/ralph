package runpaths

import (
	"fmt"
	"path/filepath"
)

func RunDir(workDir, runID string) string {
	return filepath.Join(workDir, ".ralph", "runs", runID)
}

func EventsPath(workDir, runID string) string {
	return filepath.Join(RunDir(workDir, runID), "events.ndjson")
}

func MetaPath(workDir, runID string) string {
	return filepath.Join(RunDir(workDir, runID), "meta.json")
}

func ReviewTranscriptPath(workDir, runID string, iteration int) string {
	return filepath.Join(RunDir(workDir, runID), fmt.Sprintf("review-%d.txt", iteration))
}
