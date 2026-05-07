package claude

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/suse/rancher/rancher-spur/internal/yaml"
)

// Client defines the interface for Claude API operations
type Client interface {
	GenerateYAML(ctx context.Context, input *PromptInput) (string, error)
}

// AnthropicClient implements the Claude Client interface using the official SDK
type AnthropicClient struct {
	client *anthropic.Client
	model  string
}

// NewClient creates a new Claude API client
func NewClient(apiKey, model string) Client {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicClient{
		client: client,
		model:  model,
	}
}

// GenerateYAML generates a Saddle YAML configuration from Jira issue data
// Includes automatic retry logic if YAML validation fails
func (c *AnthropicClient) GenerateYAML(ctx context.Context, input *PromptInput) (string, error) {
	maxRetries := input.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 2 // Default: allow 2 retries
	}

	userPrompt := buildUserPrompt(input)
	var lastError error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var promptText string
		if attempt == 0 {
			promptText = userPrompt
		} else {
			// Retry with error feedback
			promptText = buildRetryPrompt(userPrompt, lastError.Error())
		}

		yamlContent, err := c.callClaude(ctx, promptText, attempt == 0)
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

// callClaude makes the actual API call to Claude
func (c *AnthropicClient) callClaude(ctx context.Context, userPrompt string, enableCaching bool) (string, error) {
	params := anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.F(int64(4096)),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
		}),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		}),
	}

	message, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	// Extract text content from response
	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	var yamlContent strings.Builder
	for _, block := range message.Content {
		if block.Type == "text" {
			yamlContent.WriteString(block.Text)
		}
	}

	return yamlContent.String(), nil
}
