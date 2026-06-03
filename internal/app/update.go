package app

import (
	"context"
	"fmt"
	"os"

	"ralph/internal/args"
	"ralph/internal/update"
)

var updateInstall = update.Install

func RunUpdate(opts *args.Options) int {
	ctx := context.Background()
	repo := update.RepoFromEnv()
	ref := opts.UpdateRef
	if ref == "" {
		ref = update.DefaultRef
	}

	fmt.Fprintf(os.Stdout, "updating ralph from %s@%s\n", repo, ref)
	if err := updateInstall(ctx, update.InstallOptions{Repo: repo, Ref: ref}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	bin, err := update.InstalledBinaryPath(ctx, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stdout, "installed %s\n", bin)
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
