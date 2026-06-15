package cli

import (
	"fmt"
	"io"
	"strings"

	"ralph/internal/shared/prd"
)

func OutputPrefix(isErr bool) string {
	if isErr {
		return "  [!]"
	}
	return "  "
}

func StoryStatus(passes bool) string {
	if passes {
		return "[x]"
	}
	return "[ ]"
}

func PrintStoryList(w io.Writer, p *prd.PRD) {
	fmt.Fprintln(w, "Stories:")
	for _, s := range p.Stories {
		fmt.Fprintf(w, "  %s [P%d] %s\n", StoryStatus(s.Passes), s.Priority, s.Title)
	}
	fmt.Fprintln(w)
}

func PrintStoryDetails(w io.Writer, p *prd.PRD) {
	for _, s := range p.Stories {
		fmt.Fprintf(w, "Story: %s\n", s.Title)
		fmt.Fprintf(w, "  ID: %s\n", s.ID)
		fmt.Fprintf(w, "  Priority: %d\n", s.Priority)
		if len(s.DependsOn) > 0 {
			fmt.Fprintf(w, "  Depends on: %s\n", strings.Join(s.DependsOn, ", "))
		}
		fmt.Fprintf(w, "  Description: %s\n", s.Description)
		if len(s.Slices) > 0 {
			fmt.Fprintf(w, "  Slices: %d/%d complete\n", s.CompletedSliceCount(), len(s.Slices))
			for _, slice := range s.Slices {
				status := "[ ]"
				if slice.Passes {
					status = "[x]"
				}
				fmt.Fprintf(w, "    %s %s\n", status, slice.Behavior)
				fmt.Fprintf(w, "      Red hint: %s\n", slice.RedHint)
				if slice.RefactorHint != "" {
					fmt.Fprintf(w, "      Refactor hint: %s\n", slice.RefactorHint)
				}
			}
		}
		fmt.Fprintln(w)
	}
}
