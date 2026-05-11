package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Set up environment variables
	os.Setenv("SPUR_JIRA_URL", "https://test.atlassian.net")
	os.Setenv("SPUR_JIRA_BEARER_TOKEN", "test-bearer-token")
	defer cleanupEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Jira.URL != "https://test.atlassian.net" {
		t.Errorf("expected Jira URL 'https://test.atlassian.net', got '%s'", cfg.Jira.URL)
	}
	if cfg.Jira.BearerToken != "test-bearer-token" {
		t.Errorf("expected Jira bearer token 'test-bearer-token', got '%s'", cfg.Jira.BearerToken)
	}
	if cfg.Claude.Model != "claude-sonnet-4-6" {
		t.Errorf("expected default model 'claude-sonnet-4-6', got '%s'", cfg.Claude.Model)
	}
}

func TestLoad_CustomModel(t *testing.T) {
	os.Setenv("SPUR_JIRA_URL", "https://test.atlassian.net")
	os.Setenv("SPUR_JIRA_BEARER_TOKEN", "test-bearer-token")
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

func TestLoad_AWSConfig(t *testing.T) {
	os.Setenv("SPUR_JIRA_URL", "https://test.atlassian.net")
	os.Setenv("SPUR_JIRA_BEARER_TOKEN", "test-bearer-token")
	os.Setenv("SPUR_AWS_ACCESS_KEY", "AKIATEST")
	os.Setenv("SPUR_AWS_SECRET_KEY", "secret123")
	os.Setenv("SPUR_AWS_REGION", "us-east-1")
	defer cleanupEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.AWS.AccessKey != "AKIATEST" {
		t.Errorf("expected AWS access key 'AKIATEST', got '%s'", cfg.AWS.AccessKey)
	}
	if cfg.AWS.SecretKey != "secret123" {
		t.Errorf("expected AWS secret key 'secret123', got '%s'", cfg.AWS.SecretKey)
	}
	if cfg.AWS.Region != "us-east-1" {
		t.Errorf("expected AWS region 'us-east-1', got '%s'", cfg.AWS.Region)
	}
}

func TestLoad_SSHConfig(t *testing.T) {
	os.Setenv("SPUR_JIRA_URL", "https://test.atlassian.net")
	os.Setenv("SPUR_JIRA_BEARER_TOKEN", "test-bearer-token")
	os.Setenv("SPUR_SSH_KEY_NAME", "my-key")
	os.Setenv("SPUR_SSH_PRIVATE_KEY_PATH", "/home/user/.ssh/id_rsa")
	os.Setenv("SPUR_SSH_USER", "admin")
	defer cleanupEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.SSH.KeyName != "my-key" {
		t.Errorf("expected SSH key name 'my-key', got '%s'", cfg.SSH.KeyName)
	}
	if cfg.SSH.PrivateKeyPath != "/home/user/.ssh/id_rsa" {
		t.Errorf("expected SSH private key path '/home/user/.ssh/id_rsa', got '%s'", cfg.SSH.PrivateKeyPath)
	}
	if cfg.SSH.User != "admin" {
		t.Errorf("expected SSH user 'admin', got '%s'", cfg.SSH.User)
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Setenv("SPUR_JIRA_URL", "https://test.atlassian.net")
	os.Setenv("SPUR_JIRA_BEARER_TOKEN", "test-bearer-token")
	defer cleanupEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check defaults
	if cfg.AWS.Region != "us-west-2" {
		t.Errorf("expected default AWS region 'us-west-2', got '%s'", cfg.AWS.Region)
	}
	if cfg.AWS.InstanceType != "t3.xlarge" {
		t.Errorf("expected default instance type 't3.xlarge', got '%s'", cfg.AWS.InstanceType)
	}
	if cfg.SSH.User != "ubuntu" {
		t.Errorf("expected default SSH user 'ubuntu', got '%s'", cfg.SSH.User)
	}
}

func TestLoad_MissingJiraURL(t *testing.T) {
	os.Setenv("SPUR_JIRA_BEARER_TOKEN", "test-bearer-token")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing Jira URL, got none")
	}
}

func TestLoad_MissingJiraBearerToken(t *testing.T) {
	os.Setenv("SPUR_JIRA_URL", "https://test.atlassian.net")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing Jira bearer token, got none")
	}
}

func cleanupEnv() {
	os.Unsetenv("SPUR_JIRA_URL")
	os.Unsetenv("SPUR_JIRA_BEARER_TOKEN")
	os.Unsetenv("SPUR_CLAUDE_MODEL")
	os.Unsetenv("SPUR_AWS_ACCESS_KEY")
	os.Unsetenv("SPUR_AWS_SECRET_KEY")
	os.Unsetenv("SPUR_AWS_REGION")
	os.Unsetenv("SPUR_AWS_INSTANCE_TYPE")
	os.Unsetenv("SPUR_SSH_KEY_NAME")
	os.Unsetenv("SPUR_SSH_PRIVATE_KEY_PATH")
	os.Unsetenv("SPUR_SSH_USER")
}
