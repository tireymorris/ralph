package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad_InvalidConfigPath(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_MODEL", "invalid")
	defer os.Unsetenv("RALPH_MODEL")

	cfg, err := Load()
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}

	if cfg != nil && cfg.Model == "invalid" {
		t.Error("Should not have loaded invalid model")
	}

	if err != nil && !strings.Contains(err.Error(), "invalid model configuration") {
		t.Errorf("Error message should contain 'invalid model configuration', got: %v", err)
	}
}

func TestConfig_Validate_PathTraversal(t *testing.T) {
	tests := []struct {
		name    string
		prdFile string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid filename",
			prdFile: "prd.json",
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			prdFile: "../../../etc/passwd",
			wantErr: true,
			errMsg:  "simple filename",
		},
		{
			name:    "absolute path",
			prdFile: "/etc/passwd",
			wantErr: true,
			errMsg:  "simple filename",
		},
		{
			name:    "simple path with dots",
			prdFile: "./prd.json",
			wantErr: true,
			errMsg:  "simple filename",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Model:         DefaultModel,
				MaxIterations: 50,
				RetryAttempts: 3,
				PRDFile:       tt.prdFile,
			}
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Config.Validate() error = %v, expected to contain %q", err, tt.errMsg)
			}
		})
	}
}
