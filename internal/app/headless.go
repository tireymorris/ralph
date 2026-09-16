package app

import (
	"github.com/tireymorris/ralph/internal/headless"
	"github.com/tireymorris/ralph/internal/shared/config"
)

func runHeadless(cfg *config.Config, prompt string, resume bool) int {
	return headless.Run(cfg, prompt, resume)
}
