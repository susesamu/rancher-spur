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
	AWS    AWSConfig
	SSH    SSHConfig
}

// JiraConfig holds Jira-related configuration
type JiraConfig struct {
	URL         string
	BearerToken string
}

// ClaudeConfig holds Claude API configuration
type ClaudeConfig struct {
	Model string
}

// AWSConfig holds AWS-specific configuration
type AWSConfig struct {
	AccessKey       string
	SecretKey       string
	Region          string
	AMI             string
	InstanceType    string
	SecurityGroupID string
	SubnetID        string
}

// SSHConfig holds SSH configuration
type SSHConfig struct {
	KeyName        string
	PrivateKeyPath string
	User           string
}

// Load loads configuration from environment variables and optional config file
// Environment variables take precedence over config file values
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("claude.model", "claude-sonnet-4-6")
	v.SetDefault("aws.region", "us-west-2")
	v.SetDefault("aws.instance_type", "t3.xlarge")
	v.SetDefault("ssh.user", "ubuntu")

	// Configure environment variable binding
	v.SetEnvPrefix("SPUR")
	v.AutomaticEnv()

	// Bind specific environment variables
	v.BindEnv("jira.url", "SPUR_JIRA_URL")
	v.BindEnv("jira.bearer_token", "SPUR_JIRA_BEARER_TOKEN")
	v.BindEnv("claude.model", "SPUR_CLAUDE_MODEL")

	// AWS configuration (optional)
	v.BindEnv("aws.access_key", "SPUR_AWS_ACCESS_KEY")
	v.BindEnv("aws.secret_key", "SPUR_AWS_SECRET_KEY")
	v.BindEnv("aws.region", "SPUR_AWS_REGION")
	v.BindEnv("aws.ami", "SPUR_AWS_AMI")
	v.BindEnv("aws.instance_type", "SPUR_AWS_INSTANCE_TYPE")
	v.BindEnv("aws.security_group_id", "SPUR_AWS_SECURITY_GROUP_ID")
	v.BindEnv("aws.subnet_id", "SPUR_AWS_SUBNET_ID")

	// SSH configuration (optional)
	v.BindEnv("ssh.key_name", "SPUR_SSH_KEY_NAME")
	v.BindEnv("ssh.private_key_path", "SPUR_SSH_PRIVATE_KEY_PATH")
	v.BindEnv("ssh.user", "SPUR_SSH_USER")

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
			URL:         v.GetString("jira.url"),
			BearerToken: v.GetString("jira.bearer_token"),
		},
		Claude: ClaudeConfig{
			Model: v.GetString("claude.model"),
		},
		AWS: AWSConfig{
			AccessKey:       v.GetString("aws.access_key"),
			SecretKey:       v.GetString("aws.secret_key"),
			Region:          v.GetString("aws.region"),
			AMI:             v.GetString("aws.ami"),
			InstanceType:    v.GetString("aws.instance_type"),
			SecurityGroupID: v.GetString("aws.security_group_id"),
			SubnetID:        v.GetString("aws.subnet_id"),
		},
		SSH: SSHConfig{
			KeyName:        v.GetString("ssh.key_name"),
			PrivateKeyPath: v.GetString("ssh.private_key_path"),
			User:           v.GetString("ssh.user"),
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
	if c.Jira.BearerToken == "" {
		return fmt.Errorf("JIRA Bearer token is required (set SPUR_JIRA_BEARER_TOKEN)")
	}
	return nil
}
