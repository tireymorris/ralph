package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestInitAndLog(t *testing.T) {
	var buf bytes.Buffer
	InitWithWriter(&buf, false)

	Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected 'test message' in output, got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected 'key=value' in output, got: %s", output)
	}
}

func TestDebugNotShownByDefault(t *testing.T) {
	var buf bytes.Buffer
	InitWithWriter(&buf, false)

	Debug("debug message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Errorf("debug should not be shown when verbose=false, got: %s", output)
	}
}

func TestDebugShownWhenVerbose(t *testing.T) {
	var buf bytes.Buffer
	InitWithWriter(&buf, true)

	Debug("debug message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("debug should be shown when verbose=true, got: %s", output)
	}
}

func TestSetVerbose(t *testing.T) {
	var buf bytes.Buffer
	InitWithWriter(&buf, false)

	Debug("hidden")
	if strings.Contains(buf.String(), "hidden") {
		t.Error("debug should be hidden initially")
	}

	SetVerbose(true)
	Debug("visible")
	if !strings.Contains(buf.String(), "visible") {
		t.Error("debug should be visible after SetVerbose(true)")
	}

	buf.Reset()
	SetVerbose(false)
	Debug("hidden again")
	if strings.Contains(buf.String(), "hidden again") {
		t.Error("debug should be hidden after SetVerbose(false)")
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	InitWithWriter(&buf, true)

	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")

	output := buf.String()
	for _, level := range []string{"debug", "info", "warn", "error"} {
		if !strings.Contains(output, level) {
			t.Errorf("expected '%s' in output", level)
		}
	}
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer
	InitWithWriter(&buf, false)

	contextLogger := With("component", "test")
	contextLogger.Info("contextual message")

	output := buf.String()
	if !strings.Contains(output, "component=test") {
		t.Errorf("expected context in output, got: %s", output)
	}
}

func TestGetInitializesDefault(t *testing.T) {
	// Reset the defaultLogger to test auto-initialization
	defaultLogger = nil

	// This should auto-initialize
	Info("auto init test")

	// Should not panic and defaultLogger should be set
	if defaultLogger == nil {
		t.Error("defaultLogger should be auto-initialized")
	}
}
