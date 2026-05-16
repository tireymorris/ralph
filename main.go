package main

import (
	"os"

	"ralph/internal/app"
)

func main() { os.Exit(app.Run(os.Args[1:])) }
