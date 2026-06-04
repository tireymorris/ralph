package workflow

import (
	"context"
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
)

func TestRunImplementationReviewEventsAfterStoryCompleted(t *testing.T) {
	cfg, testPRD := saveSingleStoryPRDInGitRepo(t, false)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	evts := drainEvents(ch)
	if err := assertReviewAfterStoryCompleted(evts); err != nil {
		t.Fatal(err)
	}
}

func assertReviewAfterStoryCompleted(evts []Event) error {
	seenStoryCompleted := false
	seenReviewStarted := false
	seenReviewCompleted := false
	for _, e := range evts {
		switch e.(type) {
		case EventStoryCompleted:
			seenStoryCompleted = true
		case EventImplementationReviewStarted:
			if !seenStoryCompleted {
				return errEventOrder{"implementation review started before story completed"}
			}
			seenReviewStarted = true
		case EventImplementationReviewCompleted:
			if !seenReviewStarted {
				return errEventOrder{"implementation review completed before started"}
			}
			seenReviewCompleted = true
		}
	}
	if !seenStoryCompleted {
		return errEventOrder{"missing EventStoryCompleted"}
	}
	if !seenReviewStarted {
		return errEventOrder{"missing EventImplementationReviewStarted"}
	}
	if !seenReviewCompleted {
		return errEventOrder{"missing EventImplementationReviewCompleted"}
	}
	return nil
}

type errEventOrder struct{ msg string }

func (e errEventOrder) Error() string { return e.msg }

func TestRunImplementationReviewBeforeNextStory(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepoInDir(t, tmpDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = true

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{
			{ID: "story-1", Title: "One", Description: "d", AcceptanceCriteria: []string{"a"}, Priority: 1, Passes: false},
			{ID: "story-2", Title: "Two", Description: "d", AcceptanceCriteria: []string{"a"}, Priority: 2, Passes: false},
		},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("save PRD: %v", err)
	}
	commitPRDFile(t, tmpDir, cfg.PRDFile)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	if err := assertReviewBetweenStories(drainEvents(ch)); err != nil {
		t.Fatal(err)
	}
}

func assertReviewBetweenStories(evts []Event) error {
	var last string
	var storyStarts int
	for _, e := range evts {
		switch e.(type) {
		case EventStoryCompleted:
			last = "story_completed"
		case EventImplementationReviewStarted:
			if last != "story_completed" {
				return errEventOrder{"review started before story completed"}
			}
			last = "review_started"
		case EventImplementationReviewCompleted:
			if last != "review_started" {
				return errEventOrder{"review completed before started"}
			}
			last = "review_completed"
		case EventStoryStarted:
			storyStarts++
			if storyStarts > 1 && last != "review_completed" {
				return errEventOrder{"next story started before review completed"}
			}
			last = "story_started"
		}
	}
	if storyStarts < 2 {
		return errEventOrder{"expected two story starts"}
	}
	return nil
}

func TestRunImplementationReviewBeforeCleanup(t *testing.T) {
	cfg, testPRD := saveSingleStoryPRDInGitRepo(t, false)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	if err := assertReviewBeforeCleanup(drainEvents(ch)); err != nil {
		t.Fatal(err)
	}
}

func assertReviewBeforeCleanup(evts []Event) error {
	lastStoryCompleted := -1
	reviewCompleted := -1
	cleanupStarted := -1
	for i, e := range evts {
		switch e.(type) {
		case EventStoryCompleted:
			lastStoryCompleted = i
		case EventImplementationReviewCompleted:
			reviewCompleted = i
		case EventCleanupStarted:
			cleanupStarted = i
		}
	}
	if lastStoryCompleted < 0 {
		return errEventOrder{"missing EventStoryCompleted"}
	}
	if reviewCompleted < 0 {
		return errEventOrder{"missing EventImplementationReviewCompleted"}
	}
	if cleanupStarted < 0 {
		return errEventOrder{"missing EventCleanupStarted"}
	}
	if reviewCompleted <= lastStoryCompleted {
		return errEventOrder{"review must follow last story completed"}
	}
	if cleanupStarted <= reviewCompleted {
		return errEventOrder{"cleanup must follow implementation review"}
	}
	return nil
}

func TestRunImplementationNoReviewBetweenStoryStartAndComplete(t *testing.T) {
	cfg, testPRD := saveSingleStoryPRDInGitRepo(t, true)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	inStory := false
	mock.runFunc = func(_ context.Context, p string, _ chan<- runner.OutputLine) error {
		if strings.Contains(p, "critical diff review") && inStory {
			t.Fatal("implementation review runner invoked between story start and complete")
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	eventsDone := make(chan struct{})
	go func() {
		defer close(eventsDone)
		for ev := range ch {
			switch ev.(type) {
			case EventStoryStarted:
				inStory = true
			case EventStoryCompleted:
				inStory = false
			}
		}
	}()

	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}
	close(ch)
	<-eventsDone
}
