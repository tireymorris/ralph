package workflow

import "ralph/internal/shared/workdir"

func workdirContainsSource(workDir string) bool {
	return workdir.ContainsSource(workDir)
}
