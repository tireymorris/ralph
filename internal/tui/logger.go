package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"ralph/internal/runner"
)

// Logger handles logging and viewport management
type Logger struct {
	logView viewport.Model
	logs    []string
	maxLogs int
	width   int
	height  int
	verbose bool
}

// NewLogger creates a new logger instance
func NewLogger(verbose bool) *Logger {
	v := viewport.New(80, 10)
	// Border/padding are applied by the surrounding log panel.
	v.Style = lipgloss.NewStyle()

	return &Logger{
		logView: v,
		logs:    make([]string, 0),
		maxLogs: 500,
		verbose: verbose,
	}
}

// AddLog adds a log line to the logger
func (l *Logger) AddLog(line string) {
	l.logs = append(l.logs, line)
	if len(l.logs) > l.maxLogs {
		l.logs = l.logs[1:]
	}
	l.refreshLogView()
}

// AddOutputLine adds an output line from runner, respecting verbose flag
func (l *Logger) AddOutputLine(line runner.OutputLine) {
	// Skip verbose output unless --verbose is enabled
	if !line.Verbose || l.verbose {
		l.AddLog(line.Text)
	}
}

// SetSize updates the viewport size
func (l *Logger) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.logView.Width = max(20, width-6)
	// Favor showing logs while still leaving room for the main content.
	l.logView.Height = min(max(8, height/3), max(8, height-14))
	l.refreshLogView()
}

// GetView returns the viewport model for rendering
func (l *Logger) GetView() viewport.Model {
	return l.logView
}

// Update handles viewport updates
func (l *Logger) Update(msg interface{}) viewport.Model {
	var cmd interface{}
	l.logView, cmd = l.logView.Update(msg)
	_ = cmd // Ignore cmd for now
	return l.logView
}

// refreshLogView updates the viewport content with styled logs
func (l *Logger) refreshLogView() {
	wasAtBottom := l.logView.AtBottom()

	w := l.logView.Width
	if w <= 0 {
		// Fall back to a conservative width if we haven't received a window size yet.
		w = max(30, l.width-6)
	}

	lines := make([]string, 0, len(l.logs))
	for _, logText := range l.logs {
		// Truncate *before* styling so we don't cut ANSI sequences.
		line := truncate(logText, max(10, w-2))
		style := logLineStyle

		// Colorize logs based on content
		if strings.Contains(logText, "Error") || strings.Contains(logText, "error") {
			style = logErrorStyle
		} else if strings.Contains(logText, "completed") || strings.Contains(logText, "success") {
			style = logLineStyle.Copy().Foreground(successColor)
		} else if strings.Contains(logText, "Starting") || strings.Contains(logText, "starting") {
			style = logLineStyle.Copy().Foreground(highlightColor)
		}

		lines = append(lines, style.Render(line))
	}

	l.logView.SetContent(strings.Join(lines, "\n"))
	if wasAtBottom {
		l.logView.GotoBottom()
	}
}
