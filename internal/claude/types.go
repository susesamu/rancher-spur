package claude

// PromptInput contains the data needed to build a prompt for Claude
type PromptInput struct {
	IssueID     string
	Summary     string
	Description string
	Environment string
	Labels      []string
	Components  []string
	MaxRetries  int
}
