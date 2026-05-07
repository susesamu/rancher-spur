package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Set up environment variables
	os.Setenv("SPUR_JIRA_URL", "https://test.atlassian.net")
	os.Setenv("SPUR_JIRA_USER", "test@example.com")
	os.Setenv("SPUR_JIRA_TOKEN", "test-token")
	os.Setenv("SPUR_CLAUDE_API_KEY", "sk-test-key")
	defer cleanupEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Jira.URL != "https://test.atlassian.net" {
		t.Errorf("expected Jira URL 'https://test.atlassian.net', got '%s'", cfg.Jira.URL)
	}
	if cfg.Jira.Username != "test@example.com" {
		t.Errorf("expected Jira username 'test@example.com', got '%s'", cfg.Jira.Username)
	}
	if cfg.Jira.Token != "test-token" {
		t.Errorf("expected Jira token 'test-token', got '%s'", cfg.Jira.Token)
	}
	if cfg.Claude.APIKey != "sk-test-key" {
		t.Errorf("expected Claude API key 'sk-test-key', got '%s'", cfg.Claude.APIKey)
	}
	if cfg.Claude.Model != "claude-sonnet-4-6" {
		t.Errorf("expected default model 'claude-sonnet-4-6', got '%s'", cfg.Claude.Model)
	}
}

func TestLoad_CustomModel(t *testing.T) {
	os.Setenv("SPUR_JIRA_URL", "https://test.atlassian.net")
	os.Setenv("SPUR_JIRA_USER", "test@example.com")
	os.Setenv("SPUR_JIRA_TOKEN", "test-token")
	os.Setenv("SPUR_CLAUDE_API_KEY", "sk-test-key")
	os.Setenv("SPUR_CLAUDE_MODEL", "claude-opus-4-7")
	defer cleanupEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Claude.Model != "claude-opus-4-7" {
		t.Errorf("expected model 'claude-opus-4-7', got '%s'", cfg.Claude.Model)
	}
}

func TestLoad_MissingJiraURL(t *testing.T) {
	os.Setenv("SPUR_JIRA_USER", "test@example.com")
	os.Setenv("SPUR_JIRA_TOKEN", "test-token")
	os.Setenv("SPUR_CLAUDE_API_KEY", "sk-test-key")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing Jira URL, got none")
	}
}

func TestLoad_MissingJiraUsername(t *testing.T) {
	os.Setenv("SPUR_JIRA_URL", "https://test.atlassian.net")
	os.Setenv("SPUR_JIRA_TOKEN", "test-token")
	os.Setenv("SPUR_CLAUDE_API_KEY", "sk-test-key")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing Jira username, got none")
	}
}

func TestLoad_MissingJiraToken(t *testing.T) {
	os.Setenv("SPUR_JIRA_URL", "https://test.atlassian.net")
	os.Setenv("SPUR_JIRA_USER", "test@example.com")
	os.Setenv("SPUR_CLAUDE_API_KEY", "sk-test-key")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing Jira token, got none")
	}
}

func TestLoad_MissingClaudeAPIKey(t *testing.T) {
	os.Setenv("SPUR_JIRA_URL", "https://test.atlassian.net")
	os.Setenv("SPUR_JIRA_USER", "test@example.com")
	os.Setenv("SPUR_JIRA_TOKEN", "test-token")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing Claude API key, got none")
	}
}

func cleanupEnv() {
	os.Unsetenv("SPUR_JIRA_URL")
	os.Unsetenv("SPUR_JIRA_USER")
	os.Unsetenv("SPUR_JIRA_TOKEN")
	os.Unsetenv("SPUR_CLAUDE_API_KEY")
	os.Unsetenv("SPUR_CLAUDE_MODEL")
}
