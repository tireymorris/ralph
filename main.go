package main

import (
	"os"

	"github.com/tireymorris/ralph/internal/app"
)

func main() { os.Exit(app.Run(os.Args[1:])) }
