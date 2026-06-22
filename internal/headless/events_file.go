package headless

import (
	"fmt"
	"os"

	"ralph/internal/shared/runpaths"
)

func writeRunEventFile(workDir, runID string, line []byte) error {
	path := runpaths.EventsPath(workDir, runID)
	if err := os.MkdirAll(runpaths.RunDir(workDir, runID), 0o750); err != nil {
		return fmt.Errorf("create run event dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open events log: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	return nil
}
