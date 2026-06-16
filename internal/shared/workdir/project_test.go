package workdir_test

import (
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/workdir"
)

func TestContainsSourceDetectsProjectManifests(t *testing.T) {
	cases := []struct {
		name  string
		setup func(dir string) error
		want  bool
	}{
		{
			name: "go.mod",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app\n"), 0644)
			},
			want: true,
		},
		{
			name: "package.json",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"app"}`), 0644)
			},
			want: true,
		},
		{
			name: "Cargo.toml",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"app\"\n"), 0644)
			},
			want: true,
		},
		{
			name: "pyproject.toml",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\nname = \"app\"\n"), 0644)
			},
			want: true,
		},
		{
			name: "README only",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "README.md"), []byte("# app"), 0644)
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := tc.setup(dir); err != nil {
				t.Fatal(err)
			}
			if got := workdir.ContainsSource(dir); got != tc.want {
				t.Fatalf("ContainsSource() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestContainsSourceStillDetectsSourceFiles(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "app.ts"), []byte("export {}"), 0644); err != nil {
		t.Fatal(err)
	}
	if !workdir.ContainsSource(dir) {
		t.Fatal("ContainsSource() = false, want true for nested source file")
	}
}

func TestDetectTestCommandFromManifests(t *testing.T) {
	cases := []struct {
		name  string
		setup func(dir string) error
		want  string
	}{
		{
			name: "go.mod",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app\n"), 0644)
			},
			want: "go test ./...",
		},
		{
			name: "package.json with test script",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"test":"vitest run"}}`), 0644)
			},
			want: "npm test",
		},
		{
			name: "Cargo.toml",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname=\"app\"\n"), 0644)
			},
			want: "cargo test",
		},
		{
			name:  "empty directory",
			setup: func(string) error { return nil },
			want:  "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := tc.setup(dir); err != nil {
				t.Fatal(err)
			}
			if got := workdir.DetectTestCommand(dir); got != tc.want {
				t.Fatalf("DetectTestCommand() = %q, want %q", got, tc.want)
			}
		})
	}
}
