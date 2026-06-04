package runs

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRegisterGet(t *testing.T) {
	reg := NewRegistry()
	workDir := t.TempDir()

	run := &Run{
		ID:        "run-1",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "running",
		Phase:     "clarify",
		CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		PRDPath:   "prd.json",
	}

	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	got, ok := reg.Get("run-1")
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if got.Prompt != run.Prompt {
		t.Errorf("Prompt = %q, want %q", got.Prompt, run.Prompt)
	}
}

func TestListSortedByCreatedAtDesc(t *testing.T) {
	reg := NewRegistry()
	workDir := t.TempDir()

	oldest := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	middle := time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)
	newest := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	for _, run := range []*Run{
		{ID: "run-old", WorkDir: workDir, CreatedAt: oldest, UpdatedAt: oldest},
		{ID: "run-mid", WorkDir: workDir, CreatedAt: middle, UpdatedAt: middle},
		{ID: "run-new", WorkDir: workDir, CreatedAt: newest, UpdatedAt: newest},
	} {
		if err := reg.Register(run); err != nil {
			t.Fatalf("Register(%s) error = %v", run.ID, err)
		}
	}

	list := reg.List()
	if len(list) != 3 {
		t.Fatalf("len(List()) = %d, want 3", len(list))
	}
	want := []string{"run-new", "run-mid", "run-old"}
	for i, id := range want {
		if list[i].ID != id {
			t.Errorf("List()[%d].ID = %q, want %q", i, list[i].ID, id)
		}
	}
}

func TestUpdateStatus(t *testing.T) {
	reg := NewRegistry()
	workDir := t.TempDir()

	run := &Run{
		ID:        "run-upd",
		WorkDir:   workDir,
		Status:    "running",
		Phase:     "clarify",
		CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := reg.UpdateStatus("run-upd", "completed", "done"); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, ok := reg.Get("run-upd")
	if !ok {
		t.Fatal("Get() ok = false")
	}
	if got.Status != "completed" || got.Phase != "done" {
		t.Errorf("status/phase = %q/%q, want completed/done", got.Status, got.Phase)
	}
	if !got.UpdatedAt.After(run.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want after %v", got.UpdatedAt, run.UpdatedAt)
	}
}

func TestLoadFromWorkDir(t *testing.T) {
	workDir := t.TempDir()
	reg := NewRegistry()

	run := &Run{
		ID:        "run-persisted",
		WorkDir:   workDir,
		Prompt:    "saved run",
		Status:    "completed",
		Phase:     "complete",
		CreatedAt: time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	reloaded := NewRegistry()
	if err := reloaded.LoadFromWorkDir(workDir); err != nil {
		t.Fatalf("LoadFromWorkDir() error = %v", err)
	}

	got, ok := reloaded.Get("run-persisted")
	if !ok {
		t.Fatal("Get() ok = false after reload")
	}
	if got.Prompt != run.Prompt || got.Status != run.Status {
		t.Errorf("reloaded run = %+v, want prompt/status from registered run", got)
	}
}

func TestRegistryClearEmpty(t *testing.T) {
	reg := NewRegistry()
	reg.Clear()
}

func TestConcurrentRegisterList(t *testing.T) {
	reg := NewRegistry()
	workDir := t.TempDir()

	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("run-%d", i)
			run := &Run{
				ID:        id,
				WorkDir:   workDir,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			if err := reg.Register(run); err != nil {
				t.Errorf("Register(%s) error = %v", id, err)
			}
		}(i)
	}

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				_ = reg.List()
			}
		}
	}()

	wg.Wait()
	close(done)

	if len(reg.List()) != n {
		t.Fatalf("len(List()) = %d, want %d", len(reg.List()), n)
	}
}
