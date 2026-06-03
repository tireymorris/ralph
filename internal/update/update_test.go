package update

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ralph/internal/version"
)

type mockRunner struct {
	calls []call
	fn    func(call) ([]byte, error)
}

type call struct {
	dir  string
	name string
	args []string
}

func (m *mockRunner) Output(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	c := call{dir: dir, name: name, args: append([]string(nil), args...)}
	m.calls = append(m.calls, c)
	if m.fn != nil {
		return m.fn(c)
	}
	return nil, errors.New("unexpected command")
}

func TestRemoteHEAD(t *testing.T) {
	wantSHA := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	old := commandRunner
	defer func() { commandRunner = old }()

	commandRunner = &mockRunner{
		fn: func(c call) ([]byte, error) {
			if c.name != "git" || len(c.args) < 2 || c.args[0] != "ls-remote" {
				return nil, fmt.Errorf("unexpected: %v", c)
			}
			if c.args[len(c.args)-1] == "refs/heads/main" {
				return []byte(wantSHA + "\trefs/heads/main\n"), nil
			}
			return nil, errors.New("ls-remote failed")
		},
	}

	got, err := RemoteHEAD(context.Background(), DefaultRepo, "main")
	if err != nil {
		t.Fatalf("RemoteHEAD() error = %v", err)
	}
	if got != wantSHA {
		t.Errorf("RemoteHEAD() = %q, want %q", got, wantSHA)
	}
}

func TestRemoteHEADGitFailure(t *testing.T) {
	old := commandRunner
	defer func() { commandRunner = old }()

	commandRunner = &mockRunner{
		fn: func(call) ([]byte, error) {
			return nil, errors.New("git ls-remote: exit status 128")
		},
	}

	_, err := RemoteHEAD(context.Background(), DefaultRepo, "main")
	if err == nil {
		t.Fatal("RemoteHEAD() error = nil, want error")
	}
}

func TestInstallInvokesCloneAndGoInstall(t *testing.T) {
	const repo = "https://example.com/ralph.git"
	const ref = "main"

	old := commandRunner
	defer func() { commandRunner = old }()

	var gotClone, gotGoInstall bool
	gopath := t.TempDir()
	commandRunner = &mockRunner{
		fn: func(c call) ([]byte, error) {
			switch c.name {
			case "git":
				if len(c.args) >= 6 &&
					c.args[0] == "clone" && c.args[1] == "--depth" && c.args[2] == "1" &&
					c.args[3] == "--branch" && c.args[4] == ref && c.args[5] == repo {
					gotClone = true
					return nil, nil
				}
			case "bash":
				return []byte("-X ralph/internal/version.Version=test"), nil
			case "go":
				if c.dir != "" && len(c.args) >= 2 && c.args[0] == "install" && c.args[len(c.args)-1] == "." {
					gotGoInstall = true
					bin := filepath.Join(gopath, "bin", "ralph")
					if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
						return nil, err
					}
					if err := os.WriteFile(bin, []byte("x"), 0o755); err != nil {
						return nil, err
					}
					return nil, nil
				}
			}
			return nil, fmt.Errorf("unexpected %s %v dir=%q", c.name, c.args, c.dir)
		},
	}

	if err := Install(context.Background(), InstallOptions{Repo: repo, Ref: ref, GoPath: gopath}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !gotClone {
		t.Error("Install() did not invoke git clone --depth 1 --branch")
	}
	if !gotGoInstall {
		t.Error("Install() did not invoke go install . in clone dir")
	}
}

func TestInstallGoInstallFailure(t *testing.T) {
	old := commandRunner
	defer func() { commandRunner = old }()

	commandRunner = &mockRunner{
		fn: func(c call) ([]byte, error) {
			if c.name == "git" {
				return nil, nil
			}
			if c.name == "bash" {
				return []byte("-X test"), nil
			}
			if c.name == "go" {
				return nil, errors.New("go install .: exit status 1")
			}
			return nil, errors.New("unexpected")
		},
	}

	err := Install(context.Background(), InstallOptions{GoPath: t.TempDir()})
	if err == nil {
		t.Fatal("Install() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "go install") {
		t.Errorf("Install() error = %v, want substring go install", err)
	}
}

func TestCheckUpToDate(t *testing.T) {
	const sha = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	oldRunner := commandRunner
	oldCommit := version.Commit
	defer func() {
		commandRunner = oldRunner
		version.Commit = oldCommit
	}()
	version.Commit = sha
	commandRunner = &mockRunner{
		fn: func(call) ([]byte, error) {
			return []byte(sha + "\trefs/heads/main\n"), nil
		},
	}

	up, local, remote, err := Check(context.Background(), DefaultRepo, DefaultRef)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !up || local != sha || remote != sha {
		t.Errorf("Check() = (%v, %q, %q), want (true, sha, sha)", up, local, remote)
	}
}

func TestCheckUpdateAvailable(t *testing.T) {
	oldRunner := commandRunner
	oldCommit := version.Commit
	defer func() {
		commandRunner = oldRunner
		version.Commit = oldCommit
	}()
	version.Commit = "cccccccccccccccccccccccccccccccccccccccc"
	commandRunner = &mockRunner{
		fn: func(call) ([]byte, error) {
			return []byte("dddddddddddddddddddddddddddddddddddddddd\trefs/heads/main\n"), nil
		},
	}

	up, _, remote, err := Check(context.Background(), DefaultRepo, DefaultRef)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if up {
		t.Error("Check() upToDate = true, want false")
	}
	if remote == version.Commit {
		t.Error("remote should differ from local")
	}
}

func TestCheckUnknownCommit(t *testing.T) {
	old := version.Commit
	defer func() { version.Commit = old }()
	version.Commit = "unknown"

	_, _, _, err := Check(context.Background(), DefaultRepo, DefaultRef)
	if err == nil {
		t.Fatal("Check() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "build metadata") {
		t.Errorf("Check() error = %v, want build metadata message", err)
	}
}
