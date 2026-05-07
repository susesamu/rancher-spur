package claude

import (
	"fmt"
	"strings"
)

const systemPrompt = `You are an infrastructure automation expert. Your role is to generate valid YAML configurations for the Saddle provisioning tool.

Output Requirements:
- ONLY output valid YAML, no explanations or markdown
- Follow this exact schema:

cluster:
  name: string (required)
  nodes:
    - role: control-plane|worker (required)
      instance_type: string (required)
      count: int (optional, default 1)
  networking:
    plugin: calico|flannel|cilium (required)
  applications:
    - name: string (required)
      version: string (optional)
      config: map (optional)

Rules:
- Use reasonable defaults for missing information
- Infer instance types from workload description
- Choose networking plugin based on requirements (default: calico)
- Extract application names and versions from issue description`

// buildUserPrompt constructs the user prompt from issue data
func buildUserPrompt(input *PromptInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate Saddle YAML for Jira issue %s:\n\n", input.IssueID))
	sb.WriteString(fmt.Sprintf("Summary: %s\n\n", input.Summary))

	if input.Description != "" {
		sb.WriteString(fmt.Sprintf("Description:\n%s\n\n", input.Description))
	}

	if input.Environment != "" {
		sb.WriteString(fmt.Sprintf("Environment: %s\n\n", input.Environment))
	}

	if len(input.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(input.Labels, ", ")))
	}

	if len(input.Components) > 0 {
		sb.WriteString(fmt.Sprintf("Components: %s\n", strings.Join(input.Components, ", ")))
	}

	return sb.String()
}

// buildRetryPrompt constructs a retry prompt with error feedback
func buildRetryPrompt(originalPrompt, validationError string) string {
	return fmt.Sprintf(`The previous YAML output was invalid:
%s

Please regenerate valid YAML following the schema exactly.

Original request:
%s`, validationError, originalPrompt)
}
