package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"ralph/internal/config"
	"ralph/internal/constants"
	"ralph/internal/logger"
	"ralph/internal/prd"
	"ralph/internal/prompt"
	"ralph/internal/workflow"
)

type Runner struct {
	cfg             *config.Config
	prompt          string
	dryRun          bool
	resume          bool
	verbose         bool
	executor        *workflow.Executor
	eventsCh        chan workflow.Event
	reviewResponseCh chan bool
	cancelFunc      context.CancelFunc
}

func NewRunner(cfg *config.Config, userPrompt string, dryRun, resume, verbose bool) *Runner {
	eventsCh := make(chan workflow.Event, constants.EventChannelBuffer)
	reviewResponseCh := make(chan bool, 1)
	return &Runner{
		cfg:             cfg,
		prompt:          userPrompt,
		dryRun:          dryRun,
		resume:          resume,
		verbose:         verbose,
		eventsCh:        eventsCh,
		executor:        workflow.NewExecutor(cfg, eventsCh),
		reviewResponseCh: reviewResponseCh,
	}
}

func (r *Runner) Run() int {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancelFunc = cancel
	defer cancel()

	logger.Debug("cli runner starting", "prompt", r.prompt, "dry_run", r.dryRun, "resume", r.resume)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Debug("received interrupt signal")
		fmt.Println("\nInterrupted, shutting down...")
		cancel()
	}()

	doneCh := make(chan int, 1)
	go r.handleEvents(r.eventsCh, doneCh)

	// Run the full workflow in the foreground. Clarifying questions (if any)
	// are handled inside handleEvents when EventClarifyingQuestions arrives —
	// it reads stdin and sends answers back via the channel embedded in the
	// event, unblocking RunClarify.
	var p *prd.PRD
	var err error

	if r.resume {
		p, err = r.executor.RunLoad(ctx)
	} else {
		qas, clarifyErr := r.executor.RunClarify(ctx, r.prompt)
		if clarifyErr != nil {
			logger.Error("clarification step failed", "error", clarifyErr)
			close(r.eventsCh)
			<-doneCh
			return 1
		}
		p, err = r.executor.RunGenerateWithAnswers(ctx, r.prompt, qas)
	}

	if err != nil {
		logger.Error("operation failed", "error", err)
		close(r.eventsCh)
		<-doneCh
		return 1
	}

	if r.dryRun {
		fmt.Println("Dry run complete - PRD saved, no implementation performed")
		close(r.eventsCh)
		<-doneCh
		return 0
	}

	err = r.executor.RunImplementation(ctx, p)
	close(r.eventsCh)
	return <-doneCh
}

func (r *Runner) handleEvents(eventsCh <-chan workflow.Event, doneCh chan<- int) {
	exitCode := 0

	for event := range eventsCh {
		switch e := event.(type) {
		case workflow.EventClarifyingQuestions:
			answers := r.collectAnswersCLI(e.Questions)
			e.AnswersCh <- answers

		case workflow.EventPRDGenerating:
			fmt.Println("Generating PRD...")

		case workflow.EventPRDGenerated:
			fmt.Printf("PRD generated: %s (%d stories)\n", e.PRD.ProjectName, len(e.PRD.Stories))
			fmt.Printf("Saved to: %s\n\n", r.cfg.PRDFile)
			r.printStories(e.PRD)

		case workflow.EventPRDLoaded:
			fmt.Printf("Loaded PRD: %s (%d stories, %d completed)\n\n",
				e.PRD.ProjectName, len(e.PRD.Stories), e.PRD.CompletedCount())
			r.printStories(e.PRD)

		case workflow.EventPRDReview:
			if r.dryRun {
				fmt.Println("PRD ready for review")
				fmt.Println("(Dry run - you can edit prd.json, then run with --resume to proceed)")
				continue
			}
			r.promptPRDReview(e.PRD)
			proceed := <-r.reviewResponseCh
			if !proceed {
				fmt.Println("Review not confirmed, exiting.")
				if r.cancelFunc != nil {
					r.cancelFunc()
				}
				return
			}

		case workflow.EventStoryStarted:
			fmt.Printf("Story: %s (attempt %d/%d)\n",
				e.Story.Title, e.Story.RetryCount+1, r.cfg.RetryAttempts)

		case workflow.EventStoryCompleted:
			if e.Success {
				fmt.Printf("  Completed\n\n")
			} else {
				fmt.Printf("  Failed (will retry)\n\n")
			}

		case workflow.EventOutput:
			if e.Verbose && !r.verbose {
				continue
			}
			prefix := "  "
			if e.IsErr {
				prefix = "  [!]"
			}
			fmt.Printf("%s %s\n", prefix, e.Text)
			os.Stdout.Sync()

		case workflow.EventError:
			fmt.Printf("Error: %v\n", e.Err)
			exitCode = 1

		case workflow.EventCompleted:
			fmt.Println("All stories completed successfully!")
			exitCode = 0

		case workflow.EventFailed:
			fmt.Println("Implementation failed")
			if len(e.FailedStories) > 0 {
				fmt.Printf("\nFailed stories (%d):\n", len(e.FailedStories))
				for _, s := range e.FailedStories {
					fmt.Printf("  - %s (%d attempts)\n", s.Title, s.RetryCount)
				}
				fmt.Println("\nRun with --resume to retry after fixing issues.")
			}
			exitCode = 1
		}
	}

	doneCh <- exitCode
}

// collectAnswersCLI prints questions to stdout and reads answers from stdin.
// Pressing Enter with no input leaves the answer empty.
func (r *Runner) collectAnswersCLI(questions []string) []prompt.QuestionAnswer {
	fmt.Println()
	fmt.Println("Before generating your PRD, please answer a few clarifying questions.")
	fmt.Println("(Press Enter to skip a question, or Ctrl+C to abort.)")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	qas := make([]prompt.QuestionAnswer, 0, len(questions))

	for i, q := range questions {
		fmt.Printf("Q%d: %s\n> ", i+1, q)
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())
		qas = append(qas, prompt.QuestionAnswer{Question: q, Answer: answer})
	}

	fmt.Println()
	return qas
}

func (r *Runner) printStories(p *prd.PRD) {
	fmt.Println("Stories:")
	for _, s := range p.Stories {
		status := "[ ]"
		if s.Passes {
			status = "[x]"
		}
		fmt.Printf("  %s [P%d] %s\n", status, s.Priority, s.Title)
	}
	fmt.Println()
}

func (r *Runner) promptPRDReview(p *prd.PRD) {
	fmt.Println("PRD ready for review")
	fmt.Println()
	r.printStoriesWithDetails(p)
	fmt.Println("Please choose an action:")
	fmt.Println("  1. Edit in editor ($EDITOR)")
	fmt.Println("  2. Proceed without changes")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter choice (1-2): ")
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		cmd := exec.Command(editor, r.cfg.PRDFile)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Editor exited with error: %v\n", err)
		}
		fmt.Println("\nEditor closed.")
		fmt.Print("Proceed to implementation? [Y/n]: ")
		scanner.Scan()
		confirm := strings.TrimSpace(scanner.Text())
		r.reviewResponseCh <- confirm == "" || confirm == "y" || confirm == "Y"
	case "2":
		r.reviewResponseCh <- true
	default:
		fmt.Println("Invalid choice, proceeding without changes.")
		r.reviewResponseCh <- true
	}
}

func (r *Runner) printStoriesWithDetails(p *prd.PRD) {
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
