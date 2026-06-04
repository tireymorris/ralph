package workflow

import (
	"context"
	"testing"

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
