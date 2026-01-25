package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_InvalidConfigPath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	os.Chdir(tmpDir)

	// Create an invalid JSON config file
	configPath := filepath.Join(tmpDir, "ralph.config.json")
	err := os.WriteFile(configPath, []byte(`{"model": "invalid", "max_iterations": -1}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := Load()
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}

	if cfg != nil && cfg.Model == "invalid" {
		t.Error("Should not have loaded invalid model")
	}

	// Check that error message contains context
	if err != nil && !containsString(err.Error(), "invalid model configuration") {
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
			if err != nil && tt.errMsg != "" && !containsString(err.Error(), tt.errMsg) {
				t.Errorf("Config.Validate() error = %v, expected to contain %q", err, tt.errMsg)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
