package claude

import (
	"strings"
	"testing"
)

func TestBuildUserPrompt_FullData(t *testing.T) {
	input := &PromptInput{
		IssueID:     "TEST-123",
		Summary:     "Test issue",
		Description: "Detailed description",
		Environment: "Production",
		Labels:      []string{"bug", "critical"},
		Components:  []string{"backend", "api"},
	}

	prompt := buildUserPrompt(input)

	if !strings.Contains(prompt, "TEST-123") {
		t.Error("prompt should contain issue ID")
	}
	if !strings.Contains(prompt, "Test issue") {
		t.Error("prompt should contain summary")
	}
	if !strings.Contains(prompt, "Detailed description") {
		t.Error("prompt should contain description")
	}
	if !strings.Contains(prompt, "Production") {
		t.Error("prompt should contain environment")
	}
	if !strings.Contains(prompt, "bug, critical") {
		t.Error("prompt should contain labels")
	}
	if !strings.Contains(prompt, "backend, api") {
		t.Error("prompt should contain components")
	}
}

func TestBuildUserPrompt_MinimalData(t *testing.T) {
	input := &PromptInput{
		IssueID: "TEST-456",
		Summary: "Minimal issue",
	}

	prompt := buildUserPrompt(input)

	if !strings.Contains(prompt, "TEST-456") {
		t.Error("prompt should contain issue ID")
	}
	if !strings.Contains(prompt, "Minimal issue") {
		t.Error("prompt should contain summary")
	}
}

func TestBuildRetryPrompt(t *testing.T) {
	originalPrompt := "Generate YAML for TEST-123"
	validationError := "cluster.name is required"

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
	if !strings.Contains(systemPrompt, "cluster") {
		t.Error("system prompt should mention cluster schema")
	}
	if !strings.Contains(systemPrompt, "nodes") {
		t.Error("system prompt should mention nodes")
	}
	if !strings.Contains(systemPrompt, "networking") {
		t.Error("system prompt should mention networking")
	}
	if !strings.Contains(systemPrompt, "calico") {
		t.Error("system prompt should mention networking plugins")
	}
}
