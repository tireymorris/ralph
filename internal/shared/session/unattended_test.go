package session

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/workflow/events"
)

type recordingSink struct {
	events []events.Event
	stop   bool
	code   int
	err    error
}

func (s *recordingSink) OnEvent(ev events.Event) (bool, int, error) {
	s.events = append(s.events, ev)
	if s.stop {
		return true, s.code, s.err
	}
	switch ev.(type) {
	case events.EventCompleted:
		return true, 0, nil
	default:
		return false, 0, nil
	}
}

func TestStartUnattendedResumeStartsCheckpointResume(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.PRDFile = "prd.json"

	prdJSON := `{"version":1,"project_name":"Resume","stories":[{"id":"story-1","title":"S1","description":"d","slices":[{"id":"slice-1","behavior":"a","red_hint":"x","passes":false}],"priority":1,"passes":false}]}`
	if err := osWriteFile(cfg.WorkDir+"/"+cfg.PRDFile, prdJSON); err != nil {
		t.Fatal(err)
	}

	s := NewWithRunner(cfg, runner.NoopRunner{})
	if err := s.StartUnattended(context.Background(), cfg, UnattendedOptions{Resume: true}); err != nil {
		t.Fatalf("StartUnattended() error = %v", err)
	}

	deadline := time.After(2 * time.Second)
	for {
		select {
		case ev := <-s.EventsCh():
			switch ev.(type) {
			case events.EventPRDLoaded, events.EventStoryStarted, events.EventCompleted:
				s.Cancel()
				s.Wait()
				return
			case events.EventError:
				t.Fatalf("resume emitted error: %v", ev.(events.EventError).Err)
			}
		case <-deadline:
			t.Fatal("timed out waiting for resume workflow event")
		}
	}
}

func TestRunEventLoopStopsOnCompleted(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewWithRunner(cfg, runner.NoopRunner{})

	go func() {
		s.EmitEvent(events.EventOutput{Output: events.Output{Text: "hello"}})
		s.EmitEvent(events.EventCompleted{})
	}()

	sink := &recordingSink{}
	code := s.RunEventLoop(sink)
	if code != 0 {
		t.Fatalf("RunEventLoop() = %d, want 0", code)
	}
	if len(sink.events) != 2 {
		t.Fatalf("sink events = %d, want 2", len(sink.events))
	}
}

func TestRunEventLoopStopsOnSinkError(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewWithRunner(cfg, runner.NoopRunner{})

	go func() {
		s.EmitEvent(events.EventOutput{Output: events.Output{Text: "boom"}})
	}()

	sink := &recordingSink{stop: true, code: 1, err: errors.New("write failed")}
	code := s.RunEventLoop(sink)
	if code != 1 {
		t.Fatalf("RunEventLoop() = %d, want 1", code)
	}
}

func TestRunEventLoopNilSink(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewWithRunner(cfg, runner.NoopRunner{})
	if code := s.RunEventLoop(nil); code != 1 {
		t.Fatalf("RunEventLoop(nil) = %d, want 1", code)
	}
}

func osWriteFile(path, content string) error {
	return os.WriteFile(filepath.Clean(path), []byte(content), 0o644)
}
