package tui

import (
	"strings"
	"testing"
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
