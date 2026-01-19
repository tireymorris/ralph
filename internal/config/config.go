package config

import (
	"encoding/json"
	"os"
)

var SupportedModels = []string{
	"opencode/big-pickle",
	"opencode/glm-4.7-free",
	"opencode/gpt-5-nano",
	"opencode/grok-code",
	"opencode/minimax-m2.1-free",
}

const DefaultModel = "opencode/grok-code"

type Config struct {
	Model         string `json:"model"`
	MaxIterations int    `json:"max_iterations"`
	RetryAttempts int    `json:"retry_attempts"`
	RetryDelay    int    `json:"retry_delay"`
	LogLevel      string `json:"log_level"`
	PRDFile       string `json:"prd_file"`
}

func DefaultConfig() *Config {
	return &Config{
		Model:         DefaultModel,
		MaxIterations: 50,
		RetryAttempts: 3,
		RetryDelay:    5,
		LogLevel:      "info",
		PRDFile:       "prd.json",
	}
}

func Load() *Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile("ralph.config.json")
	if err != nil {
		return cfg
	}

	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return cfg
	}

	if fileCfg.Model != "" {
		cfg.Model = fileCfg.Model
	}
	if fileCfg.MaxIterations > 0 {
		cfg.MaxIterations = fileCfg.MaxIterations
	}
	if fileCfg.RetryAttempts > 0 {
		cfg.RetryAttempts = fileCfg.RetryAttempts
	}
	if fileCfg.RetryDelay > 0 {
		cfg.RetryDelay = fileCfg.RetryDelay
	}
	if fileCfg.LogLevel != "" {
		cfg.LogLevel = fileCfg.LogLevel
	}
	if fileCfg.PRDFile != "" {
		cfg.PRDFile = fileCfg.PRDFile
	}

	return cfg
}
