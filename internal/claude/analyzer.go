package claude

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AnalysisInput contains data for log analysis
type AnalysisInput struct {
	IssueID     string
	Summary     string
	Description string
	FileList    []string
}

// AnalysisResult contains the findings from log analysis
type AnalysisResult struct {
	RelevantFiles []string
	Findings      string
}

// AnalyzeFiles performs a two-phase analysis:
// Phase 1: Identify relevant files by name
// Phase 2: Extract error logs from relevant files
func (c *GCloudClient) AnalyzeFiles(ctx context.Context, input *AnalysisInput, workDir string) (*AnalysisResult, error) {
	// Phase 1: Identify relevant files
	relevantFiles, err := c.identifyRelevantFiles(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to identify relevant files: %w", err)
	}

	if len(relevantFiles) == 0 {
		return &AnalysisResult{
			RelevantFiles: []string{},
			Findings:      "No relevant files found based on issue context.",
		}, nil
	}

	// Phase 2: Extract error logs from relevant files
	findings, err := c.extractErrorLogs(ctx, input, relevantFiles, workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to extract error logs: %w", err)
	}

	return &AnalysisResult{
		RelevantFiles: relevantFiles,
		Findings:      findings,
	}, nil
}

// identifyRelevantFiles asks Claude to identify which files are likely relevant
func (c *GCloudClient) identifyRelevantFiles(ctx context.Context, input *AnalysisInput) ([]string, error) {
	prompt := buildFileIdentificationPrompt(input)

	response, err := c.callClaude(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse response - expecting comma-separated list of filenames
	files := parseFileList(response)
	return files, nil
}

// extractErrorLogs asks Claude to read relevant files and extract error logs
func (c *GCloudClient) extractErrorLogs(ctx context.Context, input *AnalysisInput, files []string, workDir string) (string, error) {
	// Read file contents
	fileContents := make(map[string]string)
	for _, file := range files {
		// Read file content (with size limit)
		content, err := readFileWithLimit(filepath.Join(workDir, file), 10*1024*1024) // 10MB limit
		if err != nil {
			// Skip files that can't be read
			continue
		}
		fileContents[file] = content
	}

	if len(fileContents) == 0 {
		return "No readable files found among identified relevant files.", nil
	}

	prompt := buildLogExtractionPrompt(input, fileContents)

	response, err := c.callClaude(ctx, prompt)
	if err != nil {
		return "", err
	}

	return response, nil
}

// buildFileIdentificationPrompt creates a prompt for Phase 1
func buildFileIdentificationPrompt(input *AnalysisInput) string {
	var sb strings.Builder

	sb.WriteString("You are analyzing Jira issue attachments to identify relevant log files.\n\n")
	sb.WriteString(fmt.Sprintf("Issue ID: %s\n", input.IssueID))
	sb.WriteString(fmt.Sprintf("Summary: %s\n\n", input.Summary))

	if input.Description != "" {
		sb.WriteString(fmt.Sprintf("Description:\n%s\n\n", input.Description))
	}

	sb.WriteString("Available files:\n")
	for _, file := range input.FileList {
		sb.WriteString(fmt.Sprintf("- %s\n", file))
	}

	sb.WriteString("\nTask: Identify which files are most likely to contain error logs or relevant diagnostic information.\n")
	sb.WriteString("Consider:\n")
	sb.WriteString("- Files with 'log', 'error', 'debug', 'trace', 'crash', 'dump' in the name\n")
	sb.WriteString("- Files related to components mentioned in the issue (e.g., if issue mentions 'rancher', prioritize rancher*.log)\n")
	sb.WriteString("- Recent timestamp in filename\n")
	sb.WriteString("- Exclude config files (.yaml, .yml unless they're clearly error-related), test files, and documentation\n\n")
	sb.WriteString("Output ONLY a comma-separated list of relevant filenames, nothing else.\n")
	sb.WriteString("Example: file1.log, dir/file2.txt, error-dump.log\n")
	sb.WriteString("If no relevant files exist, output: NONE\n")

	return sb.String()
}

// buildLogExtractionPrompt creates a prompt for Phase 2
func buildLogExtractionPrompt(input *AnalysisInput, fileContents map[string]string) string {
	var sb strings.Builder

	sb.WriteString("You are extracting error logs from Jira issue attachments.\n\n")
	sb.WriteString(fmt.Sprintf("Issue ID: %s\n", input.IssueID))
	sb.WriteString(fmt.Sprintf("Summary: %s\n\n", input.Summary))

	if input.Description != "" {
		sb.WriteString(fmt.Sprintf("Description:\n%s\n\n", input.Description))
	}

	sb.WriteString("File Contents:\n\n")
	for filename, content := range fileContents {
		sb.WriteString(fmt.Sprintf("=== %s ===\n", filename))
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("Task: Extract all error logs, exceptions, stack traces, and critical warnings.\n")
	sb.WriteString("Requirements:\n")
	sb.WriteString("1. Include timestamp if available (keep original format)\n")
	sb.WriteString("2. Include error level (ERROR, FATAL, CRITICAL, PANIC, etc.)\n")
	sb.WriteString("3. Include error message and stack trace if present\n")
	sb.WriteString("4. Look for: ERROR, FATAL, CRITICAL, PANIC, Exception, Traceback, Stack trace\n")
	sb.WriteString("5. Group by file and sort by timestamp where possible\n")
	sb.WriteString("6. Format each entry as:\n")
	sb.WriteString("   [TIMESTAMP] [LEVEL] [FILE] Message\n")
	sb.WriteString("   Stack trace (if available)\n\n")
	sb.WriteString("Output the compiled error logs:\n")

	return sb.String()
}

// parseFileList parses Claude's comma-separated file list response
func parseFileList(response string) []string {
	// Clean up response
	response = strings.TrimSpace(response)

	// Check for "NONE" response
	if strings.ToUpper(response) == "NONE" {
		return []string{}
	}

	// Remove markdown code blocks if present
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Split by comma
	parts := strings.Split(response, ",")

	var files []string
	for _, part := range parts {
		file := strings.TrimSpace(part)
		if file != "" && file != "NONE" {
			files = append(files, file)
		}
	}

	return files
}

// readFileWithLimit reads a file with a size limit
// For large files, reads first and last portions
func readFileWithLimit(path string, maxBytes int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	// If file is too large, read first and last portions
	if info.Size() > maxBytes {
		// Read first half of max bytes
		firstHalf := maxBytes / 2
		buf := make([]byte, firstHalf)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}

		var sb strings.Builder
		sb.WriteString(string(buf[:n]))
		sb.WriteString(fmt.Sprintf("\n\n[... FILE TRUNCATED - %d bytes omitted ...]\n\n", info.Size()-maxBytes))

		// Read last half of max bytes
		lastHalf := maxBytes - firstHalf
		_, err = file.Seek(-lastHalf, io.SeekEnd)
		if err != nil {
			// If seek fails, just return the first portion
			return sb.String(), nil
		}

		buf = make([]byte, lastHalf)
		n, err = file.Read(buf)
		if err != nil && err != io.EOF {
			return sb.String(), nil
		}

		sb.WriteString(string(buf[:n]))
		return sb.String(), nil
	}

	// Read entire file
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
