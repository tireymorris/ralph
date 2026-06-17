package handlers

import (
	"bufio"
	"net/http"
	"os"

	"ralph/internal/shared/runpaths"
	"ralph/internal/web/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func (a *API) RunEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, ok := a.registry.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "run not found")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if err := replayRunEvents(w, run.WorkDir, id); err != nil {
		return
	}

	if runs.IsTerminalStatus(run.Status) {
		return
	}

	if ctrl, ok := a.controller(id); !ok && runs.NeedsSessionReattach(run.Status) {
		var err error
		ctrl, err = a.ensureController(id)
		if err != nil {
			return
		}
		_ = ctrl
	}

	a.mu.Lock()
	ctrl := a.controllers[id]
	a.mu.Unlock()
	if ctrl == nil {
		return
	}

	sub, unsub := ctrl.Subscribe()
	defer unsub()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-sub:
			if !ok {
				return
			}
			if err := writeSSEEvent(w, ev); err != nil {
				return
			}
			if isTerminalEvent(ev) {
				return
			}
			if run, ok := a.registry.Get(id); ok && runs.IsTerminalStatus(run.Status) {
				return
			}
		}
	}
}

func replayRunEvents(w http.ResponseWriter, workDir, runID string) error {
	path := runpaths.EventsPath(workDir, runID)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		if err := writeSSEData(w, line); err != nil {
			return err
		}
	}
	return sc.Err()
}

func writeSSEEvent(w http.ResponseWriter, ev events.Event) error {
	data, err := runner.MarshalEventEnvelope(ev)
	if err != nil {
		return err
	}
	return writeSSEData(w, data)
}

func writeSSEData(w http.ResponseWriter, data []byte) error {
	if _, err := w.Write([]byte("data: ")); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if _, err := w.Write([]byte("\n\n")); err != nil {
		return err
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

func isTerminalEvent(ev events.Event) bool {
	switch ev.(type) {
	case events.EventCompleted, events.EventError:
		return true
	default:
		return false
	}
}
