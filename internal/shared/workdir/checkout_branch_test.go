package workdir_test

import (
	"testing"

	"ralph/internal/shared/testgit"
	"ralph/internal/shared/workdir"
)

func TestCheckoutBranch(t *testing.T) {
	dir := t.TempDir()
	testgit.InitRepo(t, dir)

	const branchName = "feature/my-branch"
	if err := workdir.CheckoutBranch(dir, branchName); err != nil {
		t.Fatalf("CheckoutBranch() error = %v", err)
	}

	got, err := workdir.CurrentBranchName(dir)
	if err != nil {
		t.Fatalf("CurrentBranchName() error = %v", err)
	}
	if got != branchName {
		t.Fatalf("CurrentBranchName() = %q, want %q", got, branchName)
	}
}
