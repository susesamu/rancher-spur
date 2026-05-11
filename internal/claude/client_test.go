package claude

import (
	"strings"
	"testing"

	"github.com/suse/rancher/rancher-spur/internal/config"
)

func TestBuildUserPrompt_FullData(t *testing.T) {
	input := &PromptInput{
		IssueID:     "SURE-11610",
		Summary:     "Rancher removes chartValues in YAML editor",
		Description: "Customer found a bug in Rancher UI",
		Environment: "Rancher version: 2.13.5/2.14.1",
		Labels:      []string{},
		Components:  []string{},
	}

	awsConfig := &config.AWSConfig{
		AccessKey:  "AKIATEST",
		SecretKey:  "secret123",
		Region:     "us-west-2",
	}

	sshConfig := &config.SSHConfig{
		KeyName:        "my-key",
		PrivateKeyPath: "/home/user/.ssh/id_rsa",
		User:           "ubuntu",
	}

	prompt := buildUserPrompt(input, awsConfig, sshConfig)

	if !strings.Contains(prompt, "SURE-11610") {
		t.Error("prompt should contain issue ID")
	}
	if !strings.Contains(prompt, "Rancher removes chartValues") {
		t.Error("prompt should contain summary")
	}
	if !strings.Contains(prompt, "Customer found a bug") {
		t.Error("prompt should contain description")
	}
	if !strings.Contains(prompt, "Rancher version: 2.13.5/2.14.1") {
		t.Error("prompt should contain environment")
	}
	if !strings.Contains(prompt, "AKIATEST") {
		t.Error("prompt should contain AWS access key")
	}
	if !strings.Contains(prompt, "my-key") {
		t.Error("prompt should contain SSH key name")
	}
}

func TestBuildUserPrompt_MinimalData(t *testing.T) {
	input := &PromptInput{
		IssueID: "TEST-456",
		Summary: "Minimal issue",
	}

	prompt := buildUserPrompt(input, nil, nil)

	if !strings.Contains(prompt, "TEST-456") {
		t.Error("prompt should contain issue ID")
	}
	if !strings.Contains(prompt, "Minimal issue") {
		t.Error("prompt should contain summary")
	}
}

func TestBuildUserPrompt_WithAWSConfigOnly(t *testing.T) {
	input := &PromptInput{
		IssueID: "TEST-789",
		Summary: "Test with AWS config",
	}

	awsConfig := &config.AWSConfig{
		AccessKey:       "AKIATEST",
		SecretKey:       "secret123",
		Region:          "us-east-1",
		InstanceType:    "t3.large",
		SecurityGroupID: "sg-123456",
		SubnetID:        "subnet-789",
		AMI:             "ami-12345",
	}

	prompt := buildUserPrompt(input, awsConfig, nil)

	if !strings.Contains(prompt, "AWS Configuration") {
		t.Error("prompt should contain AWS configuration section")
	}
	if !strings.Contains(prompt, "AKIATEST") {
		t.Error("prompt should contain access key")
	}
	if !strings.Contains(prompt, "us-east-1") {
		t.Error("prompt should contain region")
	}
	if !strings.Contains(prompt, "t3.large") {
		t.Error("prompt should contain instance type")
	}
}

func TestBuildUserPrompt_WithSSHConfigOnly(t *testing.T) {
	input := &PromptInput{
		IssueID: "TEST-999",
		Summary: "Test with SSH config",
	}

	sshConfig := &config.SSHConfig{
		KeyName:        "production-key",
		PrivateKeyPath: "~/.ssh/prod_rsa",
		User:           "admin",
	}

	prompt := buildUserPrompt(input, nil, sshConfig)

	if !strings.Contains(prompt, "SSH Configuration") {
		t.Error("prompt should contain SSH configuration section")
	}
	if !strings.Contains(prompt, "production-key") {
		t.Error("prompt should contain key name")
	}
	if !strings.Contains(prompt, "~/.ssh/prod_rsa") {
		t.Error("prompt should contain private key path")
	}
	if !strings.Contains(prompt, "admin") {
		t.Error("prompt should contain SSH user")
	}
}

func TestBuildRetryPrompt(t *testing.T) {
	originalPrompt := "Generate YAML for TEST-123"
	validationError := "cluster 'test': provider.type is required"

	retryPrompt := buildRetryPrompt(originalPrompt, validationError)

	if !strings.Contains(retryPrompt, "invalid") {
		t.Error("retry prompt should mention invalid YAML")
	}
	if !strings.Contains(retryPrompt, validationError) {
		t.Error("retry prompt should contain validation error")
	}
	if !strings.Contains(retryPrompt, originalPrompt) {
		t.Error("retry prompt should contain original prompt")
	}
}

func TestSystemPrompt_HasRequiredElements(t *testing.T) {
	if !strings.Contains(systemPrompt, "YAML") {
		t.Error("system prompt should mention YAML")
	}
	if !strings.Contains(systemPrompt, "clusters") {
		t.Error("system prompt should mention clusters schema")
	}
	if !strings.Contains(systemPrompt, "provider") {
		t.Error("system prompt should mention provider")
	}
	if !strings.Contains(systemPrompt, "kubernetes") {
		t.Error("system prompt should mention kubernetes")
	}
	if !strings.Contains(systemPrompt, "rke2") {
		t.Error("system prompt should mention rke2")
	}
	if !strings.Contains(systemPrompt, "aws") {
		t.Error("system prompt should mention aws")
	}
	if !strings.Contains(systemPrompt, "repro-") {
		t.Error("system prompt should mention repro- prefix for cluster naming")
	}
	if !strings.Contains(systemPrompt, "PLACEHOLDER_ACCESS_KEY") {
		t.Error("system prompt should mention placeholder credentials")
	}
}

func TestSystemPrompt_ContainsSaddleSchemaRules(t *testing.T) {
	requiredElements := []string{
		"type: aws",
		"distribution: rke2",
		"key_name:",
		"instance_count:",
		"node_prefix:",
		"t3.xlarge",
		"us-west-2",
	}

	for _, element := range requiredElements {
		if !strings.Contains(systemPrompt, element) {
			t.Errorf("system prompt should contain '%s'", element)
		}
	}
}
