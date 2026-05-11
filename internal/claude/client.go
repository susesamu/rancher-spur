package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/suse/rancher/rancher-spur/internal/config"
	"github.com/suse/rancher/rancher-spur/internal/yaml"
)

// Client defines the interface for Claude API operations
type Client interface {
	GenerateYAML(ctx context.Context, input *PromptInput, awsConfig *config.AWSConfig, sshConfig *config.SSHConfig) (string, error)
}

// GCloudClient implements the Claude Client interface using gcloud CLI
type GCloudClient struct {
	model string
}

// NewClient creates a new Claude API client using gcloud
func NewClient(model string) Client {
	return &GCloudClient{
		model: model,
	}
}

// GenerateYAML generates a Saddle YAML configuration from Jira issue data
// Includes automatic retry logic if YAML validation fails
func (c *GCloudClient) GenerateYAML(ctx context.Context, input *PromptInput, awsConfig *config.AWSConfig, sshConfig *config.SSHConfig) (string, error) {
	maxRetries := input.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 2 // Default: allow 2 retries
	}

	userPrompt := buildUserPrompt(input, awsConfig, sshConfig)
	var lastError error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var promptText string
		if attempt == 0 {
			promptText = userPrompt
		} else {
			// Retry with error feedback
			promptText = buildRetryPrompt(userPrompt, lastError.Error())
		}

		yamlContent, err := c.callClaude(ctx, promptText)
		if err != nil {
			return "", fmt.Errorf("Claude API call failed: %w", err)
		}

		// Validate the generated YAML
		if err := yaml.Validate(yamlContent); err != nil {
			lastError = err
			if attempt < maxRetries {
				// Retry with error feedback
				continue
			}
			// Max retries reached
			return yamlContent, fmt.Errorf("generated YAML is invalid after %d attempts: %w", maxRetries+1, err)
		}

		// Success - valid YAML
		return yamlContent, nil
	}

	return "", fmt.Errorf("failed to generate valid YAML: %w", lastError)
}

// gcloudResponse represents the JSON response from gcloud ai models generate-content
type gcloudResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// callClaude makes the actual API call using gcloud CLI
func (c *GCloudClient) callClaude(ctx context.Context, userPrompt string) (string, error) {
	// Check if gcloud is installed
	if _, err := exec.LookPath("gcloud"); err != nil {
		return "", fmt.Errorf("gcloud CLI not found in PATH - please install gcloud and run 'gcloud auth login'")
	}

	// Combine system and user prompts
	fullPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt)

	// Build gcloud command
	// gcloud ai models generate-content --model=<model> --prompt="<prompt>"
	cmd := exec.CommandContext(ctx, "gcloud", "ai", "models", "generate-content",
		fmt.Sprintf("--model=%s", c.model),
		fmt.Sprintf("--prompt=%s", fullPrompt),
		"--format=json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gcloud command failed: %w\nStderr: %s", err, stderr.String())
	}

	// Parse JSON response
	var response gcloudResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return "", fmt.Errorf("failed to parse gcloud response: %w\nOutput: %s", err, stdout.String())
	}

	// Extract text from response
	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in gcloud response")
	}

	if len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts in gcloud response")
	}

	var yamlContent strings.Builder
	for _, part := range response.Candidates[0].Content.Parts {
		yamlContent.WriteString(part.Text)
	}

	return yamlContent.String(), nil
}
