// Package runner launches OpenCode or Claude Code as a subprocess and streams
// stdout/stderr into structured OutputLine values for the workflow layer.
// Line-oriented reading and pipe lifecycle are shared via stream.go helpers.
package runner
