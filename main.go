package main

import (
	"os"

	"ralph/internal/app"
	"ralph/internal/shared/config"
)

func main() { os.Exit(run()) }

func run() int { return app.Run(os.Args[1:]) }

func runTUI(cfg *config.Config, prompt string, dryRun, resume, verbose bool) int {
	return app.RunTUI(cfg, prompt, dryRun, resume, verbose)
}

func runStatus(cfg *config.Config) int { return app.RunStatus(cfg) }

func validateResume(cfg *config.Config, resume bool) error { return app.ValidateResume(cfg, resume) }
