package app

import "github.com/tireymorris/ralph/internal/args"

func Run(argv []string) int {
	return newCoordinator().Run(args.Parse(argv))
}
