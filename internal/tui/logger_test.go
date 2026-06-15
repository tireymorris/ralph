package tui

import (
	"strings"
	"testing"

	"ralph/internal/shared/runner"
)

func TestLoggerWrapsLongLine(t *testing.T) {
	logText := strings.Repeat("a", 150)
	first30 := logText[:30]
	last30 := logText[len(logText)-30:]

	l := NewLogger(false)
	l.SetSize(80, 10)
	l.AddLog(logText)

	view := l.GetView().View()
	if !strings.Contains(view, first30) {
		t.Errorf("view missing first 30 chars of log text")
	}
	if !strings.Contains(view, last30) {
		t.Errorf("view missing last 30 chars of log text")
	}
	if strings.Contains(view, "...") {
		t.Errorf("view should not truncate with ellipsis, got %q", view)
	}
}

func TestLoggerAppendsStreamingDeltas(t *testing.T) {
	l := NewLogger(false)
	l.SetSize(80, 10)

	for _, chunk := range []string{"I'm ", "fixing ", "that first."} {
		l.AddOutputLine(runner.OutputLine{Text: chunk, Append: true})
	}

	if len(l.logs) != 1 {
		t.Fatalf("logs len = %d, want 1 coalesced line", len(l.logs))
	}
	want := "I'm fixing that first."
	if l.logs[0] != want {
		t.Fatalf("logs[0] = %q, want %q", l.logs[0], want)
	}

	view := l.GetView().View()
	if !strings.Contains(view, want) {
		t.Fatalf("view = %q, want coalesced text %q", view, want)
	}
	if strings.Contains(view, "I'm \nfixing") || strings.Contains(view, "fixing \nthat") {
		t.Fatalf("view should not break streaming chunks onto separate lines, got %q", view)
	}
}

func TestLoggerAppendStartsNewLineAfterDiscreteOutput(t *testing.T) {
	l := NewLogger(false)
	l.SetSize(80, 10)

	l.AddOutputLine(runner.OutputLine{Text: "hello", Append: true})
	l.AddOutputLine(runner.OutputLine{Text: "Using tool: bash"})

	if len(l.logs) != 2 {
		t.Fatalf("logs len = %d, want 2", len(l.logs))
	}
	if l.logs[0] != "hello" {
		t.Fatalf("logs[0] = %q, want %q", l.logs[0], "hello")
	}
	if l.logs[1] != "Using tool: bash" {
		t.Fatalf("logs[1] = %q, want %q", l.logs[1], "Using tool: bash")
	}
}

func TestLoggerAppendStartsNewStreamAfterToolLine(t *testing.T) {
	l := NewLogger(false)
	l.SetSize(80, 10)

	l.AddOutputLine(runner.OutputLine{Text: "Using tool: bash"})
	l.AddOutputLine(runner.OutputLine{Text: "I'm ", Append: true})
	l.AddOutputLine(runner.OutputLine{Text: "back.", Append: true})

	if len(l.logs) != 2 {
		t.Fatalf("logs len = %d, want 2", len(l.logs))
	}
	if l.logs[0] != "Using tool: bash" {
		t.Fatalf("logs[0] = %q, want %q", l.logs[0], "Using tool: bash")
	}
	if l.logs[1] != "I'm back." {
		t.Fatalf("logs[1] = %q, want %q", l.logs[1], "I'm back.")
	}
}
