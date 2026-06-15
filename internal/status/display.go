package status

import (
	"fmt"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

func Display(cfg *config.Config) error {

	exists, err := prd.Exists(cfg)
	if err != nil {
		return fmt.Errorf("checking PRD file: %w", err)
	}
	if !exists {
		fmt.Println("No PRD file found. Run ralph with a prompt to create one.")
		return nil
	}

	p, err := prd.Load(cfg)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	fmt.Printf("Project: %s", p.ProjectName)
	if p.BranchName != "" {
		fmt.Printf(" (Branch: %s)", p.BranchName)
	}
	fmt.Println()

	total := len(p.Stories)
	completed := p.CompletedCount()
	pending := total - completed

	fmt.Printf("Stories: %d total, %d completed, %d pending\n",
		total, completed, pending)

	for _, story := range p.Stories {
		status := "⏳"
		if story.Passes {
			status = "✓"
		}

		fmt.Printf("%s [%s] %s (priority: %d)\n",
			status, story.ID, story.Title, story.Priority)
		if len(story.Slices) == 0 {
			continue
		}
		fmt.Printf("  %d/%d slices complete\n", story.CompletedSliceCount(), len(story.Slices))
		for _, slice := range story.Slices {
			sliceStatus := "⏳"
			if slice.Passes {
				sliceStatus = "✓"
			}
			fmt.Printf("    %s [%s] %s\n", sliceStatus, slice.ID, slice.Behavior)
			fmt.Printf("      Red hint: %s\n", slice.RedHint)
			if slice.RefactorHint != "" {
				fmt.Printf("      Refactor hint: %s\n", slice.RefactorHint)
			}
		}
	}

	return nil
}
