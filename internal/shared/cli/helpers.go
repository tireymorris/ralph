// Package cli provides shared CLI helper functions used by multiple subcommands.
package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"ralph/internal/shared/prd"
)

// OutputPrefix returns the prefix for output lines based on whether it's an error.
func OutputPrefix(isErr bool) string {
	if isErr {
		return "  [!]"
	}
	return "  "
}

// StoryStatus returns a visual status indicator for a story.
func StoryStatus(passes bool) string {
	if passes {
		return "[x]"
	}
	return "[ ]"
}

// PrintStoryList prints a summary of all stories in the PRD.
func PrintStoryList(w io.Writer, p *prd.PRD) {
	fmt.Fprintln(w, "Stories:")
	for _, s := range p.Stories {
		fmt.Fprintf(w, "  %s [P%d] %s\n", StoryStatus(s.Passes), s.Priority, s.Title)
	}
	fmt.Fprintln(w)
}

// PrintStoryDetails prints detailed information about all stories in the PRD.
func PrintStoryDetails(w io.Writer, p *prd.PRD) {
	for _, s := range p.Stories {
		fmt.Fprintf(w, "Story: %s\n", s.Title)
		fmt.Fprintf(w, "  ID: %s\n", s.ID)
		fmt.Fprintf(w, "  Priority: %d\n", s.Priority)
		if len(s.DependsOn) > 0 {
			fmt.Fprintf(w, "  Depends on: %s\n", strings.Join(s.DependsOn, ", "))
		}
		fmt.Fprintf(w, "  Description: %s\n", s.Description)
		if len(s.AcceptanceCriteria) > 0 {
			fmt.Fprintln(w, "  Acceptance Criteria:")
			for _, ac := range s.AcceptanceCriteria {
				fmt.Fprintf(w, "    - %s\n", ac)
			}
		}
		fmt.Fprintln(w)
	}
}

// RunEditor opens the given file path in the user's preferred editor.
func RunEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
