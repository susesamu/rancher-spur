package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/suse/rancher/rancher-spur/internal/claude"
	"github.com/suse/rancher/rancher-spur/internal/files"
	"github.com/suse/rancher/rancher-spur/internal/jira"
)

var (
	keepFiles bool
)

// sniffCmd represents the sniff command
var sniffCmd = &cobra.Command{
	Use:   "sniff <jira-issue-id>",
	Short: "Download Jira attachments and analyze logs for errors",
	Long: `Downloads all readable attachments from a Jira issue, extracts archives,
and uses Claude AI to identify relevant files and extract error logs.

Creates a directory named after the issue ID containing:
- All downloaded and extracted files
- FINDINGS.txt with compiled error logs sorted by timestamp

Example:
  spur sniff SURE-11483
  spur sniff SURE-11483 --keep-files
  spur sniff SURE-11483 --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: runSniff,
}

func init() {
	rootCmd.AddCommand(sniffCmd)
	sniffCmd.Flags().BoolVar(&keepFiles, "keep-files", true, "keep downloaded files after analysis")
}

func runSniff(cmd *cobra.Command, args []string) error {
	issueID := args[0]
	ctx := context.Background()

	workDir := filepath.Join(".", issueID)

	if verbose {
		fmt.Printf("Analyzing Jira issue: %s\n", issueID)
		fmt.Printf("Working directory: %s\n\n", workDir)
	}

	// Create working directory
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create working directory: %w", err)
	}

	// Step 1: Fetch issue details
	if verbose {
		fmt.Println("Step 1: Fetching Jira issue details...")
	}
	jiraClient := jira.NewClient(cfg.Jira.URL, cfg.Jira.BearerToken)

	issue, err := jiraClient.GetIssue(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to fetch Jira issue: %w", err)
	}

	if verbose {
		fmt.Printf("  Issue: %s - %s\n\n", issue.ID, issue.Summary)
	} else {
		fmt.Printf("Fetched issue: %s - %s\n", issue.ID, issue.Summary)
	}

	// Step 2: List attachments
	if verbose {
		fmt.Println("Step 2: Listing attachments...")
	}

	attachments, err := jiraClient.ListAttachments(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to list attachments: %w", err)
	}

	if len(attachments) == 0 {
		fmt.Println("No attachments found for this issue.")
		return nil
	}

	if verbose {
		fmt.Printf("  Found %d attachment(s)\n\n", len(attachments))
	}

	// Step 3: Download readable files
	if verbose {
		fmt.Println("Step 3: Downloading readable files...")
	}

	var downloadedFiles []string
	var archiveFiles []string

	for _, att := range attachments {
		// Skip non-readable files
		if !files.IsReadableFile(att.Filename) {
			if verbose {
				fmt.Printf("  Skipping (not readable): %s\n", att.Filename)
			}
			continue
		}

		destPath := filepath.Join(workDir, att.Filename)

		if verbose {
			fmt.Printf("  Downloading: %s (%d bytes)\n", att.Filename, att.Size)
		}

		if err := jiraClient.DownloadAttachment(ctx, att.URL, destPath); err != nil {
			fmt.Printf("  Warning: failed to download %s: %v\n", att.Filename, err)
			continue
		}

		downloadedFiles = append(downloadedFiles, att.Filename)

		// Check if it's an archive
		if files.IsArchive(att.Filename) {
			archiveFiles = append(archiveFiles, att.Filename)
		}
	}

	if len(downloadedFiles) == 0 {
		fmt.Println("No readable files to analyze.")
		return nil
	}

	if verbose {
		fmt.Printf("  Downloaded %d file(s)\n\n", len(downloadedFiles))
	} else {
		fmt.Printf("Downloaded %d file(s)\n", len(downloadedFiles))
	}

	// Step 4: Extract archives
	if len(archiveFiles) > 0 {
		if verbose {
			fmt.Println("Step 4: Extracting archives...")
		}

		for _, archiveFile := range archiveFiles {
			archivePath := filepath.Join(workDir, archiveFile)

			if verbose {
				fmt.Printf("  Extracting: %s\n", archiveFile)
			}

			if err := files.ExtractArchive(archivePath, workDir); err != nil {
				fmt.Printf("  Warning: failed to extract %s: %v\n", archiveFile, err)
				continue
			}
		}

		if verbose {
			fmt.Println()
		}
	}

	// Step 5: List all files in work directory
	if verbose {
		fmt.Println("Step 5: Listing all files...")
	}

	allFiles, err := files.ListFiles(workDir)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if verbose {
		fmt.Printf("  Total files: %d\n\n", len(allFiles))
	}

	// Step 6: Analyze with Claude
	if verbose {
		fmt.Println("Step 6: Analyzing files with Claude...")
		fmt.Println("  Phase 1: Identifying relevant files...")
	} else {
		fmt.Println("Analyzing files with Claude...")
	}

	claudeClient := claude.NewClient(cfg.Claude.Model)

	analysisInput := &claude.AnalysisInput{
		IssueID:     issue.ID,
		Summary:     issue.Summary,
		Description: issue.Description,
		FileList:    allFiles,
	}

	result, err := claudeClient.AnalyzeFiles(ctx, analysisInput, workDir)
	if err != nil {
		return fmt.Errorf("failed to analyze files: %w", err)
	}

	if verbose {
		fmt.Printf("  Identified %d relevant file(s)\n", len(result.RelevantFiles))
		for _, file := range result.RelevantFiles {
			fmt.Printf("    - %s\n", file)
		}
		fmt.Println("  Phase 2: Extracting error logs...")
		fmt.Println()
	}

	// Step 7: Write findings
	if verbose {
		fmt.Println("Step 7: Writing findings...")
	}

	findingsPath := filepath.Join(workDir, "FINDINGS.txt")

	// Build findings header
	findings := fmt.Sprintf("JIRA Issue: %s\n", issue.ID)
	findings += fmt.Sprintf("Summary: %s\n\n", issue.Summary)
	findings += "=== Relevant Files Analyzed ===\n"
	if len(result.RelevantFiles) == 0 {
		findings += "No relevant files identified\n\n"
	} else {
		for _, file := range result.RelevantFiles {
			findings += fmt.Sprintf("- %s\n", file)
		}
		findings += "\n"
	}
	findings += "=== Error Logs ===\n\n"
	findings += result.Findings
	findings += "\n"

	if err := os.WriteFile(findingsPath, []byte(findings), 0644); err != nil {
		return fmt.Errorf("failed to write findings: %w", err)
	}

	fmt.Printf("\nFindings saved to: %s\n", findingsPath)

	// Step 8: Cleanup (optional)
	if !keepFiles {
		if verbose {
			fmt.Println("\nCleaning up downloaded files...")
		}
		// Delete all files except FINDINGS.txt
		for _, file := range allFiles {
			if file != "FINDINGS.txt" {
				filePath := filepath.Join(workDir, file)
				os.Remove(filePath)
			}
		}
	}

	return nil
}
