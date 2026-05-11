package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/suse/rancher/rancher-spur/internal/claude"
	"github.com/suse/rancher/rancher-spur/internal/jira"
	"github.com/suse/rancher/rancher-spur/internal/saddle"
)

var (
	outputFile string
	dryRun     bool
)

// reproduceCmd represents the reproduce command
var reproduceCmd = &cobra.Command{
	Use:   "reproduce <jira-issue-id>",
	Short: "Reproduce an environment from a Jira issue",
	Long: `Fetches a Jira issue, generates a Saddle YAML configuration using Claude AI,
and optionally provisions the environment using Saddle CLI.

Example:
  spur reproduce JIRA-123
  spur reproduce JIRA-123 --output my-env.yaml --dry-run
  spur reproduce JIRA-123 --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: runReproduce,
}

func init() {
	rootCmd.AddCommand(reproduceCmd)

	reproduceCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output YAML file name (default: <issue-id>.yaml)")
	reproduceCmd.Flags().BoolVar(&dryRun, "dry-run", false, "generate YAML but do not execute Saddle")
}

func runReproduce(cmd *cobra.Command, args []string) error {
	issueID := args[0]
	ctx := context.Background()

	// Set default output file if not specified
	if outputFile == "" {
		outputFile = fmt.Sprintf("%s.yaml", issueID)
	}

	if verbose {
		fmt.Printf("Reproducing environment for issue: %s\n", issueID)
		fmt.Printf("Configuration:\n")
		fmt.Printf("  Jira URL: %s\n", cfg.Jira.URL)
		fmt.Printf("  Claude Model: %s\n", cfg.Claude.Model)
		fmt.Printf("  Output File: %s\n", outputFile)
		fmt.Printf("  Dry Run: %v\n\n", dryRun)
	}

	// Step 1: Fetch Jira issue
	if verbose {
		fmt.Println("Step 1: Fetching Jira issue...")
	}
	jiraClient := jira.NewClient(cfg.Jira.URL, cfg.Jira.BearerToken)

	issue, err := jiraClient.GetIssue(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to fetch Jira issue: %w", err)
	}

	if verbose {
		fmt.Printf("  Issue: %s\n", issue.ID)
		fmt.Printf("  Summary: %s\n", issue.Summary)
		if issue.Environment != "" {
			fmt.Printf("  Environment: %s\n", issue.Environment)
		}
		if len(issue.Labels) > 0 {
			fmt.Printf("  Labels: %v\n", issue.Labels)
		}
		if len(issue.Components) > 0 {
			fmt.Printf("  Components: %v\n", issue.Components)
		}
		fmt.Println()
	} else {
		fmt.Printf("Fetched issue: %s - %s\n", issue.ID, issue.Summary)
	}

	// Step 2: Generate YAML using Claude
	if verbose {
		fmt.Println("Step 2: Generating YAML configuration with Claude...")
	} else {
		fmt.Println("Generating YAML configuration...")
	}

	claudeClient := claude.NewClient(cfg.Claude.Model)

	promptInput := &claude.PromptInput{
		IssueID:     issue.ID,
		Summary:     issue.Summary,
		Description: issue.Description,
		Environment: issue.Environment,
		Labels:      issue.Labels,
		Components:  issue.Components,
		MaxRetries:  2,
	}

	startTime := time.Now()
	yamlContent, err := claudeClient.GenerateYAML(ctx, promptInput, &cfg.AWS, &cfg.SSH)
	if err != nil {
		return fmt.Errorf("failed to generate YAML: %w", err)
	}

	if verbose {
		fmt.Printf("  Generated in %v\n\n", time.Since(startTime))
	}

	// Step 3: Save YAML to file
	if err := os.WriteFile(outputFile, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	fmt.Printf("YAML configuration saved to: %s\n", outputFile)

	// Step 4: Execute Saddle (unless --dry-run)
	if dryRun {
		fmt.Println("\nDry run mode: skipping Saddle execution")
		return nil
	}

	fmt.Println("\nExecuting Saddle to provision environment...")
	if verbose {
		fmt.Println("Step 3: Running 'saddle create'...")
		fmt.Println()
	}

	saddleExecutor := saddle.NewExecutor()
	if err := saddleExecutor.Create(ctx, outputFile); err != nil {
		return fmt.Errorf("saddle execution failed: %w", err)
	}

	fmt.Println("\nEnvironment provisioned successfully!")
	return nil
}
