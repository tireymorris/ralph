package workflow

import "testing"

func drainEvents(ch chan Event) []Event {
	var evts []Event
	for len(ch) > 0 {
		evts = append(evts, <-ch)
	}
	return evts
}

type cleanupEventCounts struct {
	started   int
	completed int
}

func countCleanupEvents(evts []Event) cleanupEventCounts {
	var counts cleanupEventCounts
	for _, e := range evts {
		switch e.(type) {
		case EventCleanupStarted:
			counts.started++
		case EventCleanupCompleted:
			counts.completed++
		}
	}
	return counts
}

func cleanupEventIndices(evts []Event) (startedIdx, completedIdx, workflowCompletedIdx int) {
	startedIdx, completedIdx, workflowCompletedIdx = -1, -1, -1
	for i, e := range evts {
		switch e.(type) {
		case EventCleanupStarted:
			startedIdx = i
		case EventCleanupCompleted:
			completedIdx = i
		case EventCompleted:
			workflowCompletedIdx = i
		}
	}
	return startedIdx, completedIdx, workflowCompletedIdx
}

func assertCleanupBeforeCompleted(t *testing.T, evts []Event) {
	t.Helper()
	startedIdx, completedIdx, workflowCompletedIdx := cleanupEventIndices(evts)
	if startedIdx < 0 {
		t.Fatal("expected EventCleanupStarted to be emitted")
	}
	if completedIdx < 0 {
		t.Fatal("expected EventCleanupCompleted to be emitted")
	}
	if workflowCompletedIdx < 0 {
		t.Fatal("expected EventCompleted to be emitted")
	}
	if startedIdx >= workflowCompletedIdx {
		t.Error("EventCleanupStarted must come before EventCompleted")
	}
	if completedIdx >= workflowCompletedIdx {
		t.Error("EventCleanupCompleted must come before EventCompleted")
	}
	if startedIdx >= completedIdx {
		t.Error("EventCleanupStarted must come before EventCleanupCompleted")
	}
}

func TestDrainEventsDrainsBufferedChannel(t *testing.T) {
	ch := make(chan Event, 2)
	ch <- EventCleanupStarted{}
	ch <- EventCleanupCompleted{}

	evts := drainEvents(ch)
	if len(evts) != 2 {
		t.Fatalf("drainEvents len = %d, want 2", len(evts))
	}
	if _, ok := evts[0].(EventCleanupStarted); !ok {
		t.Fatalf("evts[0] = %T, want EventCleanupStarted", evts[0])
	}
}

func TestCountCleanupEvents(t *testing.T) {
	counts := countCleanupEvents([]Event{
		EventCleanupStarted{},
		EventStoryStarted{},
		EventCleanupCompleted{},
		EventCleanupStarted{},
	})
	if counts.started != 2 || counts.completed != 1 {
		t.Fatalf("counts = %+v, want started=2 completed=1", counts)
	}
}

func TestAssertCleanupBeforeCompleted(t *testing.T) {
	t.Run("accepts cleanup events before workflow completed", func(t *testing.T) {
		assertCleanupBeforeCompleted(t, []Event{
			EventCleanupStarted{},
			EventCleanupCompleted{},
			EventCompleted{},
		})
	})

	t.Run("rejects EventCompleted before EventCleanupStarted", func(t *testing.T) {
		startedIdx, _, workflowCompletedIdx := cleanupEventIndices([]Event{
			EventCompleted{},
			EventCleanupStarted{},
			EventCleanupCompleted{},
		})
		if startedIdx < workflowCompletedIdx {
			t.Fatal("expected EventCompleted to index before EventCleanupStarted")
		}
	})
}
