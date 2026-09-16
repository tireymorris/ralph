package workflow

import "github.com/tireymorris/ralph/internal/shared/workdir"

func workdirContainsSource(workDir string) bool {
	return workdir.ContainsSource(workDir)
}
