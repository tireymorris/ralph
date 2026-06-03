package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/workflow/events"
)

func TestDriverStartNewEmitsClarifyThenGenerate(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, _ chan<- runner.OutputLine) error {
		prdPath := filepath.Join(workDir, "prd.json")
		data := `{"project_name":"Test","stories":[{"id":"1","title":"S1","description":"d","acceptance_criteria":["a"],"priority":1}]}`
		return os.WriteFile(prdPath, []byte(data), 0644)
	}

	d := NewDriverWithRunner(cfg, mock)
	t.Cleanup(d.Cancel)
	d.StartNew(context.Background(), "build something")

	deadline := time.Now().Add(3 * time.Second)
	var gotReview bool
	for time.Now().Before(deadline) {
		select {
		case ev := <-d.EventsCh():
			d.TrackEventState(ev)
			if _, ok := ev.(events.EventPRDReview); ok {
				gotReview = true
			}
		default:
		}
		if gotReview {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !gotReview {
		t.Fatal("expected EventPRDReview, never received")
	}
	if d.CurrentPRD() == nil {
		t.Fatal("CurrentPRD() should be non-nil after review event")
	}
	if d.Prompt() != "build something" {
		t.Fatalf("Prompt() = %q, want %q", d.Prompt(), "build something")
	}
}

func TestDriverStartResumeEmitsLoadedEvent(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	prdData := `{"project_name":"Resumed","stories":[{"id":"1","title":"S1","description":"d","acceptance_criteria":["a"],"priority":1}]}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(prdData), 0644); err != nil {
		t.Fatal(err)
	}

	mock := newMockRunner()
	d := NewDriverWithRunner(cfg, mock)
	t.Cleanup(d.Cancel)
	d.StartResume(context.Background())

	deadline := time.Now().Add(3 * time.Second)
	var gotLoaded bool
	for time.Now().Before(deadline) {
		select {
		case ev := <-d.EventsCh():
			d.TrackEventState(ev)
			if _, ok := ev.(events.EventPRDLoaded); ok {
				gotLoaded = true
			}
		default:
		}
		if gotLoaded {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !gotLoaded {
		t.Fatal("expected EventPRDLoaded, never received")
	}
	if d.CurrentPRD() == nil || d.CurrentPRD().ProjectName != "Resumed" {
		t.Fatalf("CurrentPRD() = %v, want project name Resumed", d.CurrentPRD())
	}
}

func TestDriverSubmitClarify(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	mock := newMockRunner()
	d := NewDriverWithRunner(cfg, mock)
	t.Cleanup(d.Cancel)

	if err := d.SubmitClarify(nil); err != ErrNotWaitingClarify {
		t.Fatalf("SubmitClarify without pending channel: err = %v, want ErrNotWaitingClarify", err)
	}

	answersCh := make(chan []prompt.QuestionAnswer, 1)
	d.TrackEventState(events.EventClarifyingQuestions{
		Questions: []string{"Q1"},
		AnswersCh: answersCh,
	})

	qas := []prompt.QuestionAnswer{{Question: "Q1", Answer: "A1"}}
	if err := d.SubmitClarify(qas); err != nil {
		t.Fatalf("SubmitClarify: %v", err)
	}

	got := <-answersCh
	if len(got) != 1 || got[0].Answer != "A1" {
		t.Fatalf("answers = %v, want [{Q1 A1}]", got)
	}

	if err := d.SubmitClarify(nil); err != ErrNotWaitingClarify {
		t.Fatalf("second SubmitClarify: err = %v, want ErrNotWaitingClarify", err)
	}
}

func TestDriverTrackEventState(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	mock := newMockRunner()
	d := NewDriverWithRunner(cfg, mock)
	t.Cleanup(d.Cancel)

	testPRD := &prd.PRD{ProjectName: "Tracked"}
	d.TrackEventState(events.EventPRDGenerated{PRD: testPRD})
	if d.CurrentPRD() != testPRD {
		t.Fatal("CurrentPRD should match after EventPRDGenerated")
	}

	testPRD2 := &prd.PRD{ProjectName: "Loaded"}
	d.TrackEventState(events.EventPRDLoaded{PRD: testPRD2})
	if d.CurrentPRD() != testPRD2 {
		t.Fatal("CurrentPRD should match after EventPRDLoaded")
	}

	testPRD3 := &prd.PRD{ProjectName: "Review"}
	d.TrackEventState(events.EventPRDReview{PRD: testPRD3})
	if d.CurrentPRD() != testPRD3 {
		t.Fatal("CurrentPRD should match after EventPRDReview")
	}
}

func TestDriverCancel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	mock := newMockRunner()
	d := NewDriverWithRunner(cfg, mock)

	d.Cancel()
	select {
	case <-d.Ctx().Done():
	default:
		t.Fatal("context should be done after Cancel()")
	}
}
