package app

import "ralph/internal/args"

func Run(argv []string) int {
	return newCoordinator().Run(args.Parse(argv))
}
