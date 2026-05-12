package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"ralph/internal/prd"
)

func (r *Headless) printStories(p *prd.PRD) {
	r.printStoryList(p)
}

func (r *Headless) printStoryList(p *prd.PRD) {
	fmt.Println("Stories:")
	for _, s := range p.Stories {
		fmt.Printf("  %s [P%d] %s\n", storyStatus(s.Passes), s.Priority, s.Title)
	}
	fmt.Println()
}

func (r *Headless) outputPrefix(isErr bool) string {
	if isErr {
		return "  [!]"
	}
	return "  "
}

func (r *Headless) printStoryDetails(p *prd.PRD) {
	for _, s := range p.Stories {
		fmt.Printf("Story: %s\n", s.Title)
		fmt.Printf("  ID: %s\n", s.ID)
		fmt.Printf("  Priority: %d\n", s.Priority)
		if len(s.DependsOn) > 0 {
			fmt.Printf("  Depends on: %s\n", strings.Join(s.DependsOn, ", "))
		}
		fmt.Printf("  Description: %s\n", s.Description)
		if len(s.AcceptanceCriteria) > 0 {
			fmt.Println("  Acceptance Criteria:")
			for _, ac := range s.AcceptanceCriteria {
				fmt.Printf("    - %s\n", ac)
			}
		}
		fmt.Println()
	}
}

func (r *Headless) runEditor(path string) error {
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

func storyStatus(passes bool) string {
	if passes {
		return "[x]"
	}
	return "[ ]"
}
