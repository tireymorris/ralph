package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"ralph/internal/args"
	"ralph/internal/update"
)

var (
	updateInstall       = update.Install
	updateCheck         = update.Check
	promptUpdateConfirm = defaultPromptUpdateConfirm
)

func defaultPromptUpdateConfirm(local, remote string, metadataUnavailable bool) bool {
	if metadataUnavailable {
		fmt.Fprint(os.Stdout, "Install ralph from remote? [y/N]: ")
	} else {
		fmt.Fprintf(os.Stdout, "Update available (local=%s remote=%s). Update now? [y/N]: ", local, remote)
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "y" || answer == "yes"
}

type updateAttemptResult struct {
	upToDate            bool
	local               string
	remote              string
	metadataUnavailable bool
}

func maybePromptedUpdate(ctx context.Context, repo, ref string, interactive bool) error {
	up, local, remote, checkErr := updateCheck(ctx, repo, ref)
	if checkErr == nil && up {
		return nil
	}
	if checkErr != nil && !strings.Contains(checkErr.Error(), "build metadata") {
		return checkErr
	}
	if !interactive {
		return nil
	}
	if !promptUpdateConfirm(local, remote, checkErr != nil) {
		return nil
	}
	return updateInstall(ctx, update.InstallOptions{Repo: repo, Ref: ref})
}

func attemptUpdate(ctx context.Context, repo, ref string) (updateAttemptResult, error) {
	up, local, remote, checkErr := updateCheck(ctx, repo, ref)
	if checkErr == nil && up {
		return updateAttemptResult{upToDate: true, local: local, remote: remote}, nil
	}
	if checkErr != nil && !strings.Contains(checkErr.Error(), "build metadata") {
		return updateAttemptResult{local: local, remote: remote}, checkErr
	}

	result := updateAttemptResult{local: local, remote: remote, metadataUnavailable: checkErr != nil}
	if err := updateInstall(ctx, update.InstallOptions{Repo: repo, Ref: ref}); err != nil {
		return result, err
	}
	return result, nil
}

func RunUpdate(opts *args.Options) int {
	ctx := context.Background()
	repo := update.RepoFromEnv()
	ref := opts.UpdateRef
	if ref == "" {
		ref = update.DefaultRef
	}

	result, err := attemptUpdate(ctx, repo, ref)
	if result.upToDate {
		fmt.Fprintf(os.Stdout, "ralph is already up to date (commit=%s)\n", result.local)
		return 0
	}
	if result.metadataUnavailable {
		fmt.Fprintf(os.Stdout, "installing ralph from %s@%s (version metadata unavailable)\n", repo, ref)
	} else {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stdout, "updating ralph from %s@%s (commit %s → %s)\n", repo, ref, result.local, result.remote)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	bin, err := update.InstalledBinaryPath(ctx, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if result.metadataUnavailable {
		fmt.Fprintf(os.Stdout, "installed %s\n", bin)
	} else {
		fmt.Fprintf(os.Stdout, "updated: installed %s (commit %s)\n", bin, result.remote)
	}
	return 0
}

func RunUpdateCheck(opts *args.Options) int {
	ctx := context.Background()
	repo := update.RepoFromEnv()
	ref := opts.UpdateRef
	if ref == "" {
		ref = update.DefaultRef
	}

	up, local, remote, err := update.Check(ctx, repo, ref)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if up {
		fmt.Fprintf(os.Stdout, "up to date (local=%s remote=%s)\n", local, remote)
		return 0
	}
	fmt.Fprintf(os.Stdout, "update available (local=%s remote=%s)\n", local, remote)
	return 2
}
