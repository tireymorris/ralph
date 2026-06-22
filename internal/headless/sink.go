package headless

import (
	"io"
	"os"

	webrunner "ralph/internal/web/runner"
	"ralph/internal/workflow/events"
)

type ndjsonSink struct {
	workDir string
	runID   string
	w       io.Writer
	refresh func()
}

func newNDJSONSink(workDir, runID string, w io.Writer, refresh func()) *ndjsonSink {
	if w == nil {
		w = os.Stderr
	}
	return &ndjsonSink{workDir: workDir, runID: runID, w: w, refresh: refresh}
}

func (s *ndjsonSink) OnEvent(ev events.Event) (stop bool, exitCode int, err error) {
	if err := s.writeEvent(ev); err != nil {
		return true, 1, err
	}
	if s.refresh != nil {
		s.refresh()
	}
	switch ev.(type) {
	case events.EventCompleted:
		return true, 0, nil
	case events.EventError:
		return true, 1, nil
	default:
		return false, 0, nil
	}
}

func (s *ndjsonSink) writeEvent(ev events.Event) error {
	data, err := webrunner.MarshalEventEnvelope(ev)
	if err != nil {
		return err
	}
	return appendRunEvent(s.workDir, s.runID, append(data, '\n'), s.w)
}

func appendRunEvent(workDir, runID string, line []byte, w io.Writer) error {
	if err := writeRunEventFile(workDir, runID, line); err != nil {
		return err
	}
	if w != nil {
		if _, err := w.Write(line); err != nil {
			return err
		}
	}
	return nil
}
