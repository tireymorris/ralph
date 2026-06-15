package app

import (
	"ralph/internal/headless"
	"ralph/internal/shared/config"
)

func runHeadless(cfg *config.Config, prompt string, resume bool) int {
	return headless.Run(cfg, prompt, resume)
}
