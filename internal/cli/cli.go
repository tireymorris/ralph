package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ralph/internal/config"
	"ralph/internal/git"
	"ralph/internal/prd"
	"ralph/internal/runner"
	"ralph/internal/story"
)

// Runner handles CLI (non-TUI) execution
type Runner struct {
	cfg     *config.Config
	prompt  string
	dryRun  bool
	resume  bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewRunner creates a new CLI runner
func NewRunner(cfg *config.Config, prompt string, dryRun, resume bool) *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{
		cfg:    cfg,
		prompt: prompt,
		dryRun: dryRun,
		resume: resume,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Run executes the CLI workflow and returns an exit code
func (r *Runner) Run() int {
	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted, shutting down...")
		r.cancel()
	}()

	fmt.Printf("🤖 Ralph - Autonomous Software Development Agent\n")
	fmt.Printf("   Model: %s\n\n", r.cfg.Model)

	var p *prd.PRD
	var err error

	if r.resume {
		fmt.Println("📂 Loading existing PRD...")
		p, err = prd.Load(r.cfg)
		if err != nil {
			fmt.Printf("❌ Error loading PRD: %v\n", err)
			return 1
		}
		fmt.Printf("   Loaded: %s (%d stories, %d completed)\n\n", p.ProjectName, len(p.Stories), p.CompletedCount())
	} else {
		fmt.Println("📝 Generating PRD from prompt...")
		fmt.Printf("   Prompt: %s\n\n", truncate(r.prompt, 60))

		p, err = r.generatePRD()
		if err != nil {
			fmt.Printf("❌ Error generating PRD: %v\n", err)
			return 1
		}

		if err := prd.Save(r.cfg, p); err != nil {
			fmt.Printf("❌ Error saving PRD: %v\n", err)
			return 1
		}
		fmt.Printf("✅ PRD generated: %s (%d stories)\n", p.ProjectName, len(p.Stories))
		fmt.Printf("   Saved to: %s\n\n", r.cfg.PRDFile)
	}

	// Print stories
	fmt.Println("📋 Stories:")
	for _, s := range p.Stories {
		status := "⬜"
		if s.Passes {
			status = "✅"
		}
		fmt.Printf("   %s [P%d] %s\n", status, s.Priority, s.Title)
	}
	fmt.Println()

	if r.dryRun {
		fmt.Println("🏁 Dry run complete - PRD saved, no implementation performed")
		return 0
	}

	// Setup branch
	if p.BranchName != "" {
		gitMgr := git.New()
		if err := gitMgr.CreateBranch(p.BranchName); err != nil {
			fmt.Printf("⚠️  Warning: failed to create branch: %v\n", err)
		} else {
			fmt.Printf("🌿 Branch: %s\n\n", p.BranchName)
		}
	}

	// Implement stories
	return r.implementStories(p)
}

func (r *Runner) generatePRD() (*prd.PRD, error) {
	gen := prd.NewGenerator(r.cfg)
	outputCh := make(chan runner.OutputLine, 100)

	// Print output in background
	go func() {
		for line := range outputCh {
			if line.IsErr {
				fmt.Printf("   [stderr] %s\n", line.Text)
			} else {
				fmt.Printf("   %s\n", line.Text)
			}
		}
	}()

	p, err := gen.Generate(r.ctx, r.prompt, outputCh)
	close(outputCh)
	return p, err
}

func (r *Runner) implementStories(p *prd.PRD) int {
	impl := story.NewImplementer(r.cfg)
	iteration := 0

	fmt.Println("🚀 Starting implementation...")
	fmt.Println()

	for {
		// Check context
		select {
		case <-r.ctx.Done():
			fmt.Println("\n⚠️  Cancelled")
			return 1
		default:
		}

		// Check if all done
		if p.AllCompleted() {
			prd.Delete(r.cfg)
			fmt.Println()
			fmt.Println("🎉 All stories completed successfully!")
			return 0
		}

		// Get next story
		next := p.NextPendingStory(r.cfg.RetryAttempts)
		if next == nil {
			// All remaining stories have failed
			fmt.Println()
			fmt.Println("❌ Implementation failed - some stories exceeded retry limit")
			r.printFailedStories(p)
			return 1
		}

		// Check max iterations
		iteration++
		if iteration > r.cfg.MaxIterations {
			fmt.Println()
			fmt.Printf("❌ Max iterations (%d) reached\n", r.cfg.MaxIterations)
			return 1
		}

		// Implement story
		fmt.Printf("▶️  Story: %s (attempt %d/%d)\n", next.Title, next.RetryCount+1, r.cfg.RetryAttempts)

		outputCh := make(chan runner.OutputLine, 100)
		doneCh := make(chan struct{})

		// Print output in background
		go func() {
			for line := range outputCh {
				prefix := "   "
				if line.IsErr {
					prefix = "   [!]"
				}
				fmt.Printf("%s %s\n", prefix, line.Text)
			}
			close(doneCh)
		}()

		success, err := impl.Implement(r.ctx, next, iteration, p, outputCh)
		close(outputCh)
		<-doneCh

		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			next.RetryCount++
		} else if success {
			next.Passes = true
			fmt.Printf("   ✅ Completed\n")
		} else {
			next.RetryCount++
			fmt.Printf("   ❌ Failed (will retry)\n")
		}

		// Save state
		if err := prd.Save(r.cfg, p); err != nil {
			fmt.Printf("   ⚠️  Warning: failed to save state: %v\n", err)
		}

		fmt.Println()
	}
}

func (r *Runner) printFailedStories(p *prd.PRD) {
	failed := p.FailedStories(r.cfg.RetryAttempts)
	if len(failed) > 0 {
		fmt.Printf("\nFailed stories (%d):\n", len(failed))
		for _, s := range failed {
			fmt.Printf("   • %s (%d attempts)\n", s.Title, s.RetryCount)
		}
		fmt.Println("\nRun with --resume to retry after fixing issues.")
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 4 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
