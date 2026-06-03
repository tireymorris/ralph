package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ralph/internal/version"
)

const (
	DefaultRepo   = "https://github.com/tireymorris/ralph.git"
	DefaultRef    = "main"
	EnvRALPHRepo  = "RALPH_REPO"
	binaryName    = "ralph"
)

type InstallOptions struct {
	Repo string
	Ref  string
	// GoPath is the GOPATH root (from `go env GOPATH`). Empty uses `go env GOPATH`.
	GoPath string
}

func RepoFromEnv() string {
	if v := os.Getenv(EnvRALPHRepo); v != "" {
		return v
	}
	return DefaultRepo
}

func RemoteHEAD(ctx context.Context, repoURL, ref string) (string, error) {
	patterns := []string{"refs/heads/" + ref, "refs/tags/" + ref}
	for _, pattern := range patterns {
		out, err := commandRunner.Output(ctx, "", "git", "ls-remote", repoURL, pattern)
		if err != nil {
			if isGitMissing(err) {
				return "", err
			}
			continue
		}
		sha, ok := parseLsRemote(out)
		if ok {
			return sha, nil
		}
	}
	return "", fmt.Errorf("git ls-remote: no commit found for ref %q in %s", ref, repoURL)
}

func parseLsRemote(out []byte) (string, bool) {
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		sha := fields[0]
		if len(sha) == 40 && isHex(sha) {
			return sha, true
		}
	}
	return "", false
}

func isHex(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

func isGitMissing(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "executable file not found") ||
		strings.Contains(msg, "no such file or directory")
}

func gopath(ctx context.Context, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	out, err := commandRunner.Output(ctx, "", "go", "env", "GOPATH")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func installedBinaryPath(ctx context.Context, goPathOverride string) (string, error) {
	if goPathOverride != "" {
		return filepath.Join(goPathOverride, "bin", binaryName), nil
	}
	out, err := commandRunner.Output(ctx, "", "go", "env", "GOBIN")
	if err != nil {
		return "", err
	}
	if gobin := strings.TrimSpace(string(out)); gobin != "" {
		return filepath.Join(gobin, binaryName), nil
	}
	gopathDir, err := gopath(ctx, "")
	if err != nil {
		return "", err
	}
	return filepath.Join(gopathDir, "bin", binaryName), nil
}

func Install(ctx context.Context, opts InstallOptions) error {
	repo := opts.Repo
	if repo == "" {
		repo = DefaultRepo
	}
	ref := opts.Ref
	if ref == "" {
		ref = DefaultRef
	}

	tmp, err := os.MkdirTemp("", "ralph-install-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	cloneDir := filepath.Join(tmp, "ralph")
	if _, err := commandRunner.Output(ctx, "", "git", "clone", "--depth", "1", "--branch", ref, repo, cloneDir); err != nil {
		return err
	}
	ldflags, err := commandRunner.Output(ctx, cloneDir, "bash", "scripts/version-ldflags.sh")
	if err != nil {
		return fmt.Errorf("version ldflags: %w", err)
	}
	installArgs := []string{"install", "-ldflags", strings.TrimSpace(string(ldflags)), "."}
	if _, err := commandRunner.Output(ctx, cloneDir, "go", installArgs...); err != nil {
		if !strings.Contains(err.Error(), "go install") {
			return fmt.Errorf("go install: %w", err)
		}
		return err
	}

	bin, err := installedBinaryPath(ctx, opts.GoPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(bin); err != nil {
		return fmt.Errorf("install succeeded but %s missing: %w", bin, err)
	}
	return nil
}

func InstalledBinaryPath(ctx context.Context, goPath string) (string, error) {
	return installedBinaryPath(ctx, goPath)
}

func Check(ctx context.Context, repo, ref string) (upToDate bool, local, remote string, err error) {
	local = version.Commit
	if local == "unknown" {
		return false, local, "", fmt.Errorf("build metadata required: rebuild with scripts/build.sh or go build -ldflags")
	}
	remote, err = RemoteHEAD(ctx, repo, ref)
	if err != nil {
		return false, local, remote, err
	}
	return local == remote, local, remote, nil
}
