// Package cli implements headless (stdout/stdin) Ralph execution.
package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/logger"
	"ralph/internal/prd"
	"ralph/internal/prompt"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

// HeadlessCommand wires parsed flags to a headless workflow run.
type HeadlessCommand struct {
	Cfg     *config.Config
	Prompt  string
	DryRun  bool
	Resume  bool
	Verbose bool
}

func (c HeadlessCommand) Run() int {
	return NewHeadless(c.Cfg, c.Prompt, c.DryRun, c.Resume, c.Verbose).Run()
}

// Headless runs the workflow in stdout mode (no Bubble Tea TUI).
type workflowExecutor interface {
	RunClarify(ctx context.Context, userPrompt string) ([]prompt.QuestionAnswer, error)
	RunLoad(ctx context.Context) (*prd.PRD, error)
	RunGenerateWithAnswers(ctx context.Context, userPrompt string, qas []prompt.QuestionAnswer) (*prd.PRD, error)
	RunImplementation(ctx context.Context, p *prd.PRD) error
}

type Headless struct {
	cfg              *config.Config
	prompt           string
	dryRun           bool
	resume           bool
	verbose          bool
	executor         workflowExecutor
	eventsCh         chan events.Event
	reviewResponseCh chan bool
	cancelFunc       context.CancelFunc
}

func NewHeadless(cfg *config.Config, userPrompt string, dryRun, resume, verbose bool) *Headless {
	eventsCh := make(chan events.Event, constants.EventChannelBuffer)
	reviewResponseCh := make(chan bool, 1)
	return &Headless{
		cfg:              cfg,
		prompt:           userPrompt,
		dryRun:           dryRun,
		resume:           resume,
		verbose:          verbose,
		eventsCh:         eventsCh,
		executor:         workflow.NewExecutor(cfg, eventsCh),
		reviewResponseCh: reviewResponseCh,
	}
}

func (r *Headless) Run() int {
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

	if implErr := r.executor.RunImplementation(ctx, p); implErr != nil {
		logger.Error("implementation failed", "error", implErr)
		close(r.eventsCh)
		<-doneCh
		return 1
	}
	close(r.eventsCh)
	return <-doneCh
}

func (r *Headless) handleEvents(eventsCh <-chan events.Event, doneCh chan<- int) {
	exitCode := 0

	for event := range eventsCh {
		switch e := event.(type) {
		case events.EventClarifyingQuestions:
			answers := r.collectAnswersCLI(e.Questions)
			e.AnswersCh <- answers

		case events.EventPRDGenerating:
			fmt.Println("Generating PRD...")

		case events.EventPRDGenerated:
			fmt.Printf("PRD generated: %s (%d stories)\n", e.PRD.ProjectName, len(e.PRD.Stories))
			fmt.Printf("Saved to: %s\n\n", r.cfg.PRDFile)
			r.printStoryList(e.PRD)

		case events.EventPRDLoaded:
			fmt.Printf("Loaded PRD: %s (%d stories, %d completed)\n\n",
				e.PRD.ProjectName, len(e.PRD.Stories), e.PRD.CompletedCount())
			r.printStoryList(e.PRD)

		case events.EventPRDReview:
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

		case events.EventStoryStarted:
			fmt.Printf("Story: %s\n", e.Story.Title)

		case events.EventStoryCompleted:
			fmt.Printf("  Completed\n\n")

		case events.EventOutput:
			if e.Verbose && !r.verbose {
				continue
			}
			fmt.Printf("%s %s\n", r.outputPrefix(e.IsErr), e.Text)
			os.Stdout.Sync()

		case events.EventError:
			fmt.Printf("Error: %v\n", e.Err)
			exitCode = 1

		case events.EventCompleted:
			fmt.Println("All stories completed successfully!")
			exitCode = 0
		}
	}

	doneCh <- exitCode
}

func (r *Headless) collectAnswersCLI(questions []string) []prompt.QuestionAnswer {
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

func (r *Headless) promptPRDReview(p *prd.PRD) {
	fmt.Println("PRD ready for review")
	fmt.Println()
	r.printStoryDetails(p)
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
		if err := r.runEditor(r.cfg.PRDFile); err != nil {
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
