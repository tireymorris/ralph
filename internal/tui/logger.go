package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"ralph/internal/shared/runner"
)

type Logger struct {
	logView viewport.Model
	logs    []string
	maxLogs int
	width   int
	height  int
	verbose bool
}

func NewLogger(verbose bool) *Logger {
	v := viewport.New(80, 10)

	// Border and padding are applied by the surrounding log panel.
	v.Style = lipgloss.NewStyle()

	return &Logger{
		logView: v,
		logs:    make([]string, 0),
		maxLogs: 500,
		verbose: verbose,
	}
}

func (l *Logger) AddLog(line string) {
	l.logs = append(l.logs, line)
	if len(l.logs) > l.maxLogs {
		l.logs = l.logs[1:]
	}
	l.refreshLogView()
}

func (l *Logger) AddOutputLine(line runner.OutputLine) {

	// Skip verbose output unless --verbose is enabled.
	if !line.Verbose || l.verbose {
		l.AddLog(line.Text)
	}
}

func (l *Logger) SetSize(width, logHeight int) {
	l.width = width
	l.height = logHeight
	l.logView.Width = max(20, width-6)
	l.logView.Height = max(4, logHeight)
	l.refreshLogView()
}

func (l *Logger) LogCount() int {
	return len(l.logs)
}

func (l *Logger) GetView() viewport.Model {
	return l.logView
}

func (l *Logger) Update(msg interface{}) (viewport.Model, interface{}) {
	var cmd interface{}
	l.logView, cmd = l.logView.Update(msg)
	return l.logView, cmd
}

func (l *Logger) refreshLogView() {
	wasAtBottom := l.logView.AtBottom()

	w := l.logView.Width
	if w <= 0 {

		// Fall back before the first window size message arrives.
		w = max(30, l.width-6)
	}

	lines := make([]string, 0, len(l.logs))
	for _, logText := range l.logs {

		// Truncate before styling so ANSI sequences are not cut.
		line := truncate(logText, max(10, w-2))
		style := logLineStyle

		lowerText := strings.ToLower(logText)
		if strings.Contains(lowerText, "error") || strings.Contains(lowerText, "failed") || strings.Contains(lowerText, "failure") {
			style = logErrorStyle
		} else if strings.Contains(lowerText, "completed") || strings.Contains(lowerText, "success") || strings.Contains(lowerText, "done") {
			style = logSuccessStyle
		} else if strings.Contains(lowerText, "starting") || strings.Contains(lowerText, "generating") || strings.Contains(lowerText, "running") {
			style = logInfoStyle
		} else if strings.Contains(lowerText, "warning") || strings.Contains(lowerText, "warn") {
			style = logLineStyle.Foreground(warningColor)
		}

		lines = append(lines, style.Render(line))
	}

	l.logView.SetContent(strings.Join(lines, "\n"))
	if wasAtBottom {
		l.logView.GotoBottom()
	}
}
