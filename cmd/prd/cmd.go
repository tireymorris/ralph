// Package prd implements the "ralph prd" subcommand.
package prd

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
	"ralph/internal/shared/prd"
	"ralph/internal/prompt"
	"ralph/internal/workflow"
)

type workflowExecutor interface {
	RunClarify(ctx context.Context, userPrompt string) ([]prompt.QuestionAnswer, error)
	RunGenerateWithAnswers(ctx context.Context, userPrompt string, qas []prompt.QuestionAnswer) (*prd.PRD, error)
	RunImplementation(ctx context.Context, p *prd.PRD) error
}

// Cmd runs the prd subcommand workflow.
type Cmd struct {
	cfg      *config.Config
	prompt   string
	verbose  bool
	executor workflowExecutor
	eventsCh chan workflow.Event
}

// NewCmd creates a new Cmd instance.
func NewCmd(cfg *config.Config, userPrompt string, verbose bool) *Cmd {
	eventsCh := make(chan workflow.Event, constants.EventChannelBuffer)
	return &Cmd{
		cfg:      cfg,
		prompt:   userPrompt,
		verbose:  verbose,
		eventsCh: eventsCh,
		executor: workflow.NewExecutor(cfg, eventsCh),
	}
}

// Run executes the clarify and generate phases, writes prd.json, and exits.
func (c *Cmd) Run() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	doneCh := make(chan int, 1)
	go c.handleEvents(c.eventsCh, doneCh)

	qas, err := c.executor.RunClarify(ctx, c.prompt)
	if err != nil {
		logger.Error("clarification step failed", "error", err)
		close(c.eventsCh)
		<-doneCh
		return 1
	}

	p, err := c.executor.RunGenerateWithAnswers(ctx, c.prompt, qas)
	if err != nil {
		logger.Error("PRD generation failed", "error", err)
		close(c.eventsCh)
		<-doneCh
		return 1
	}

	fmt.Printf("PRD generated: %s (%d stories)\n", p.ProjectName, len(p.Stories))
	fmt.Printf("Saved to: %s\n", c.cfg.PRDFile)
	fmt.Println("Dry run complete - PRD saved, no implementation performed")

	close(c.eventsCh)
	return <-doneCh
}

func (c *Cmd) handleEvents(eventsCh <-chan workflow.Event, doneCh chan<- int) {
	exitCode := 0

	for event := range eventsCh {
		switch e := event.(type) {
		case workflow.EventClarifyingQuestions:
			answers := c.collectAnswers(e.Questions)
			e.AnswersCh <- answers

		case workflow.EventPRDGenerating:
			fmt.Println("Generating PRD...")

		case workflow.EventPRDGenerated:
			fmt.Printf("PRD generated: %s (%d stories)\n", e.PRD.ProjectName, len(e.PRD.Stories))

		case workflow.EventOutput:
			if e.Verbose && !c.verbose {
				continue
			}
			fmt.Printf("%s\n", e.Text)

		case workflow.EventError:
			fmt.Printf("Error: %v\n", e.Err)
			exitCode = 1
		}
	}

	doneCh <- exitCode
}

func (c *Cmd) collectAnswers(questions []string) []prompt.QuestionAnswer {
	// No questions to answer
	if len(questions) == 0 {
		return nil
	}

	fmt.Println()
	fmt.Println("Before generating your PRD, please answer a few clarifying questions.")
	fmt.Println("(Press Enter to skip a question, or Ctrl+C to abort.)")
	fmt.Println()

	// Read from stdin for interactive answers
	return readAnswersFromStdin(questions)
}

func readAnswersFromStdin(questions []string) []prompt.QuestionAnswer {
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

// Run is a convenience function that creates a Cmd and runs it.
func Run(cfg *config.Config, userPrompt string, verbose bool) int {
	return NewCmd(cfg, userPrompt, verbose).Run()
}
