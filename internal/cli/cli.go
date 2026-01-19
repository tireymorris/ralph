package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/workflow"
)

// Runner handles CLI (non-TUI) execution
type Runner struct {
	cfg    *config.Config
	prompt string
	dryRun bool
	resume bool
}

// NewRunner creates a new CLI runner
func NewRunner(cfg *config.Config, prompt string, dryRun, resume bool) *Runner {
	return &Runner{
		cfg:    cfg,
		prompt: prompt,
		dryRun: dryRun,
		resume: resume,
	}
}

// Run executes the CLI workflow and returns an exit code
func (r *Runner) Run() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted, shutting down...")
		cancel()
	}()

	r.printHeader()

	eventsCh := make(chan workflow.Event, 100)
	exec := workflow.NewExecutor(r.cfg, eventsCh)

	// Start event handler
	doneCh := make(chan int, 1)
	go r.handleEvents(eventsCh, doneCh)

	// Run workflow
	var p *prd.PRD
	var err error

	if r.resume {
		p, err = exec.RunLoad(ctx)
	} else {
		p, err = exec.RunGenerate(ctx, r.prompt)
	}

	if err != nil {
		close(eventsCh)
		<-doneCh
		return 1
	}

	if r.dryRun {
		fmt.Println("🏁 Dry run complete - PRD saved, no implementation performed")
		close(eventsCh)
		return 0
	}

	err = exec.RunImplementation(ctx, p)
	close(eventsCh)
	return <-doneCh
}

func (r *Runner) printHeader() {
	fmt.Printf("🤖 Ralph - Autonomous Software Development Agent\n")
	fmt.Printf("   Model: %s\n\n", r.cfg.Model)
}

func (r *Runner) handleEvents(eventsCh <-chan workflow.Event, doneCh chan<- int) {
	exitCode := 0

	for event := range eventsCh {
		switch e := event.(type) {
		case workflow.EventPRDGenerating:
			fmt.Println("📝 Generating PRD...")

		case workflow.EventPRDGenerated:
			fmt.Printf("✅ PRD generated: %s (%d stories)\n", e.PRD.ProjectName, len(e.PRD.Stories))
			fmt.Printf("   Saved to: %s\n\n", r.cfg.PRDFile)
			r.printStories(e.PRD)

		case workflow.EventPRDLoaded:
			fmt.Printf("📂 Loaded PRD: %s (%d stories, %d completed)\n\n",
				e.PRD.ProjectName, len(e.PRD.Stories), e.PRD.CompletedCount())
			r.printStories(e.PRD)

		case workflow.EventStoryStarted:
			fmt.Printf("▶️  Story: %s (attempt %d/%d)\n",
				e.Story.Title, e.Story.RetryCount+1, r.cfg.RetryAttempts)

		case workflow.EventStoryCompleted:
			if e.Success {
				fmt.Printf("   ✅ Completed\n\n")
			} else {
				fmt.Printf("   ❌ Failed (will retry)\n\n")
			}

		case workflow.EventOutput:
			prefix := "   "
			if e.IsErr {
				prefix = "   [!]"
			}
			fmt.Printf("%s %s\n", prefix, e.Text)

		case workflow.EventError:
			fmt.Printf("❌ Error: %v\n", e.Err)
			exitCode = 1

		case workflow.EventCompleted:
			fmt.Println("🎉 All stories completed successfully!")
			exitCode = 0

		case workflow.EventFailed:
			fmt.Println("❌ Implementation failed")
			if len(e.FailedStories) > 0 {
				fmt.Printf("\nFailed stories (%d):\n", len(e.FailedStories))
				for _, s := range e.FailedStories {
					fmt.Printf("   • %s (%d attempts)\n", s.Title, s.RetryCount)
				}
				fmt.Println("\nRun with --resume to retry after fixing issues.")
			}
			exitCode = 1
		}
	}

	doneCh <- exitCode
}

func (r *Runner) printStories(p *prd.PRD) {
	fmt.Println("📋 Stories:")
	for _, s := range p.Stories {
		status := "⬜"
		if s.Passes {
			status = "✅"
		}
		fmt.Printf("   %s [P%d] %s\n", status, s.Priority, s.Title)
	}
	fmt.Println()
}
