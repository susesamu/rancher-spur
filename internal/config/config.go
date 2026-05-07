package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for Spur
type Config struct {
	Jira   JiraConfig
	Claude ClaudeConfig
}

// JiraConfig holds Jira-related configuration
type JiraConfig struct {
	URL      string
	Username string
	Token    string
}

// ClaudeConfig holds Claude API configuration
type ClaudeConfig struct {
	APIKey string
	Model  string
}

// Load loads configuration from environment variables and optional config file
// Environment variables take precedence over config file values
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("claude.model", "claude-sonnet-4-6")

	// Configure environment variable binding
	v.SetEnvPrefix("SPUR")
	v.AutomaticEnv()

	// Bind specific environment variables
	v.BindEnv("jira.url", "SPUR_JIRA_URL")
	v.BindEnv("jira.username", "SPUR_JIRA_USER")
	v.BindEnv("jira.token", "SPUR_JIRA_TOKEN")
	v.BindEnv("claude.api_key", "SPUR_CLAUDE_API_KEY")
	v.BindEnv("claude.model", "SPUR_CLAUDE_MODEL")

	// Try to read config file from ~/.spur/config.yaml (optional)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".spur")
		v.AddConfigPath(configPath)
		v.SetConfigName("config")
		v.SetConfigType("yaml")

		// Ignore error if config file doesn't exist
		_ = v.ReadInConfig()
	}

	cfg := &Config{
		Jira: JiraConfig{
			URL:      v.GetString("jira.url"),
			Username: v.GetString("jira.username"),
			Token:    v.GetString("jira.token"),
		},
		Claude: ClaudeConfig{
			APIKey: v.GetString("claude.api_key"),
			Model:  v.GetString("claude.model"),
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that all required configuration fields are set
func (c *Config) Validate() error {
	if c.Jira.URL == "" {
		return fmt.Errorf("JIRA URL is required (set SPUR_JIRA_URL)")
	}
	if c.Jira.Username == "" {
		return fmt.Errorf("JIRA username is required (set SPUR_JIRA_USER)")
	}
	if c.Jira.Token == "" {
		return fmt.Errorf("JIRA token is required (set SPUR_JIRA_TOKEN)")
	}
	if c.Claude.APIKey == "" {
		return fmt.Errorf("Claude API key is required (set SPUR_CLAUDE_API_KEY)")
	}
	return nil
}
