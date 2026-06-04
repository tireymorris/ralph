package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"ralph/internal/args"
	"ralph/internal/update"
)

var (
	updateInstall = update.Install
	updateCheck   = update.Check
)

func RunUpdate(opts *args.Options) int {
	ctx := context.Background()
	repo := update.RepoFromEnv()
	ref := opts.UpdateRef
	if ref == "" {
		ref = update.DefaultRef
	}

	up, local, remote, checkErr := updateCheck(ctx, repo, ref)
	if checkErr == nil && up {
		fmt.Fprintf(os.Stdout, "ralph is already up to date (commit=%s)\n", local)
		return 0
	}
	if checkErr != nil && !strings.Contains(checkErr.Error(), "build metadata") {
		fmt.Fprintf(os.Stderr, "Error: %v\n", checkErr)
		return 1
	}

	if checkErr != nil {
		fmt.Fprintf(os.Stdout, "installing ralph from %s@%s (version metadata unavailable)\n", repo, ref)
	} else {
		fmt.Fprintf(os.Stdout, "updating ralph from %s@%s (commit %s → %s)\n", repo, ref, local, remote)
	}

	if err := updateInstall(ctx, update.InstallOptions{Repo: repo, Ref: ref}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	bin, err := update.InstalledBinaryPath(ctx, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if checkErr != nil {
		fmt.Fprintf(os.Stdout, "installed %s\n", bin)
	} else {
		fmt.Fprintf(os.Stdout, "updated: installed %s (commit %s)\n", bin, remote)
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
