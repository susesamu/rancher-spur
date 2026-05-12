# Spur Sniff Command - Implementation Plan

## Overview

Implement a new `spur sniff <JIRA_ID>` command that:
1. Fetches Jira issue attachments
2. Downloads readable files (excluding images)
3. Extracts compressed archives
4. Uses Claude to analyze file names and identify relevant files
5. Extracts error logs from relevant files
6. Compiles findings into `FINDINGS.txt` with timestamps

---

## Command Usage

```bash
spur sniff SURE-11483
```

**Flow:**
1. Create directory: `./<JIRA_ID>/`
2. Fetch issue summary and description
3. List attachments from Jira
4. Download readable files (skip .png, .jpg, .gif, etc.)
5. Extract .tar.gz, .zip archives
6. Call Claude to:
   - First: Read file names only → identify relevant files
   - Second: Read content of relevant files → extract error logs
7. Create `FINDINGS.txt` with compiled logs sorted by timestamp

---

## Questions to Answer

### 1. File Type Filtering

**Question:** Which file extensions should be considered "readable" and downloadable?
- Text files: `.txt`, `.log`, `.yaml`, `.yml`, `.json`
- Archives: `.tar.gz`, `.tgz`, `.zip`, `.tar`

**Skip:**
- Images: `.png`, `.jpg`, `.jpeg`, `.gif`, `.svg`, `.bmp`
- Binary: `.exe`, `.bin`, `.dll`, `.so`

**Your preference?**

### 2. Archive Extraction

**Question:** How should we handle nested archives?
- A. Extract only top-level archives (tar.gz → files)

**Your preference?**

### 3. Timestamp Format

**Question:** What format should timestamps have in FINDINGS.txt?

- C. Keep original format from logs

**Your preference?**

### 4. File Cleanup

**Question:** Should we delete downloaded files after creating FINDINGS.txt?

**Options:**
- A. Keep all downloaded files in `<JIRA_ID>/` directory

**Your preference?**

### 5. Error Log Detection

**Question:** What patterns should be considered "error logs"?

- Lines containing: `ERROR`, `FATAL`, `CRITICAL`, `PANIC`
- Lines containing: `Exception`, `Traceback`, `Stack trace`
- Lines matching timestamp + error pattern

**Your preference?**

### 6. Claude Relevance Criteria

**Question:** How should Claude determine which files are "relevant"?

- Files matching Jira issue context (e.g., if issue mentions "rancher", prioritize rancher*.log)

**Your preference?**

### 7. Maximum File Size

**Question:** Should we limit the size of files Claude reads?

- For large files, read first/last N lines only

**Your preference?**

---

## Proposed Implementation

### 1. Jira Package Changes

**New Function: `ListAttachments`**

File: `internal/jira/client.go`

```go
// Attachment represents a Jira attachment
type Attachment struct {
    ID       string
    Filename string
    URL      string
    MimeType string
    Size     int64
}

// ListAttachments fetches all attachments for a Jira issue
func (c *HTTPClient) ListAttachments(ctx context.Context, issueID string) ([]Attachment, error)
```

**Implementation:**
```go
func (c *HTTPClient) ListAttachments(ctx context.Context, issueID string) ([]Attachment, error) {
    url := fmt.Sprintf("%s/rest/api/2/issue/%s?fields=attachment", c.baseURL, issueID)
    
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch attachments: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
    }
    
    var response struct {
        Fields struct {
            Attachment []struct {
                ID       string `json:"id"`
                Filename string `json:"filename"`
                Content  string `json:"content"`
                MimeType string `json:"mimeType"`
                Size     int64  `json:"size"`
            } `json:"attachment"`
        } `json:"fields"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    attachments := make([]Attachment, 0, len(response.Fields.Attachment))
    for _, att := range response.Fields.Attachment {
        attachments = append(attachments, Attachment{
            ID:       att.ID,
            Filename: att.Filename,
            URL:      att.Content,
            MimeType: att.MimeType,
            Size:     att.Size,
        })
    }
    
    return attachments, nil
}
```

**New Function: `DownloadAttachment`**

```go
// DownloadAttachment downloads a Jira attachment to the specified path
func (c *HTTPClient) DownloadAttachment(ctx context.Context, url, destPath string) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
    
    resp, err := c.client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to download attachment: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status %d", resp.StatusCode)
    }
    
    file, err := os.Create(destPath)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()
    
    _, err = io.Copy(file, resp.Body)
    if err != nil {
        return fmt.Errorf("failed to write file: %w", err)
    }
    
    return nil
}
```

**Tests:**
```go
// internal/jira/client_test.go
func TestListAttachments_Success(t *testing.T)
func TestListAttachments_NoAttachments(t *testing.T)
func TestDownloadAttachment_Success(t *testing.T)
func TestDownloadAttachment_Unauthorized(t *testing.T)
```

---

### 2. File Handling Package

**New Package: `internal/files/`**

File: `internal/files/handler.go`

```go
package files

import (
    "archive/tar"
    "archive/zip"
    "compress/gzip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
)

// IsReadableFile checks if a file extension is readable (not an image/binary)
func IsReadableFile(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    
    // Skip image files
    imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".svg", ".bmp", ".ico"}
    for _, imgExt := range imageExts {
        if ext == imgExt {
            return false
        }
    }
    
    // Skip binary files
    binaryExts := []string{".exe", ".bin", ".dll", ".so"}
    for _, binExt := range binaryExts {
        if ext == binExt {
            return false
        }
    }
    
    // Skip office documents (unless you want to support them)
    // officeExts := []string{".pdf", ".doc", ".docx", ".xls", ".xlsx"}
    
    return true
}

// ExtractArchive extracts tar.gz or zip files to the destination directory
func ExtractArchive(archivePath, destDir string) error {
    ext := strings.ToLower(filepath.Ext(archivePath))
    
    switch {
    case strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz"):
        return extractTarGz(archivePath, destDir)
    case ext == ".zip":
        return extractZip(archivePath, destDir)
    default:
        return fmt.Errorf("unsupported archive format: %s", ext)
    }
}

func extractTarGz(archivePath, destDir string) error {
    file, err := os.Open(archivePath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    gzReader, err := gzip.NewReader(file)
    if err != nil {
        return err
    }
    defer gzReader.Close()
    
    tarReader := tar.NewReader(gzReader)
    
    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        targetPath := filepath.Join(destDir, header.Name)
        
        switch header.Typeflag {
        case tar.TypeDir:
            if err := os.MkdirAll(targetPath, 0755); err != nil {
                return err
            }
        case tar.TypeReg:
            // Create parent directories
            if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
                return err
            }
            
            outFile, err := os.Create(targetPath)
            if err != nil {
                return err
            }
            
            if _, err := io.Copy(outFile, tarReader); err != nil {
                outFile.Close()
                return err
            }
            outFile.Close()
        }
    }
    
    return nil
}

func extractZip(archivePath, destDir string) error {
    reader, err := zip.OpenReader(archivePath)
    if err != nil {
        return err
    }
    defer reader.Close()
    
    for _, file := range reader.File {
        targetPath := filepath.Join(destDir, file.Name)
        
        if file.FileInfo().IsDir() {
            if err := os.MkdirAll(targetPath, file.Mode()); err != nil {
                return err
            }
            continue
        }
        
        // Create parent directories
        if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
            return err
        }
        
        outFile, err := os.Create(targetPath)
        if err != nil {
            return err
        }
        
        rc, err := file.Open()
        if err != nil {
            outFile.Close()
            return err
        }
        
        _, err = io.Copy(outFile, rc)
        rc.Close()
        outFile.Close()
        
        if err != nil {
            return err
        }
    }
    
    return nil
}

// ListFiles recursively lists all files in a directory
func ListFiles(dir string) ([]string, error) {
    var files []string
    
    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            // Get relative path
            relPath, err := filepath.Rel(dir, path)
            if err != nil {
                return err
            }
            files = append(files, relPath)
        }
        return nil
    })
    
    return files, err
}
```

**Tests:**
```go
// internal/files/handler_test.go
func TestIsReadableFile(t *testing.T)
func TestExtractTarGz(t *testing.T)
func TestExtractZip(t *testing.T)
func TestListFiles(t *testing.T)
```

---

### 3. Claude Analysis Package

**New Functions in `internal/claude/`**

File: `internal/claude/analyzer.go`

```go
package claude

import (
    "context"
    "fmt"
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
    
    prompt := buildLogExtractionPrompt(input, fileContents)
    
    response, err := c.callClaude(ctx, prompt)
    if err != nil {
        return "", err
    }
    
    return response, nil
}
```

**Prompt Builders:**

```go
// buildFileIdentificationPrompt creates a prompt for Phase 1
func buildFileIdentificationPrompt(input *AnalysisInput) string {
    var sb strings.Builder
    
    sb.WriteString("You are analyzing Jira issue attachments to identify relevant log files.\n\n")
    sb.WriteString(fmt.Sprintf("Issue ID: %s\n", input.IssueID))
    sb.WriteString(fmt.Sprintf("Summary: %s\n\n", input.Summary))
    sb.WriteString(fmt.Sprintf("Description:\n%s\n\n", input.Description))
    
    sb.WriteString("Available files:\n")
    for _, file := range input.FileList {
        sb.WriteString(fmt.Sprintf("- %s\n", file))
    }
    
    sb.WriteString("\nTask: Identify which files are most likely to contain error logs or relevant diagnostic information.\n")
    sb.WriteString("Consider:\n")
    sb.WriteString("- Files with 'log', 'error', 'debug', 'trace' in the name\n")
    sb.WriteString("- Files related to components mentioned in the issue\n")
    sb.WriteString("- Recent timestamp in filename\n")
    sb.WriteString("- Exclude config files, test files, and documentation\n\n")
    sb.WriteString("Output ONLY a comma-separated list of relevant filenames, nothing else.\n")
    sb.WriteString("Example: file1.log, dir/file2.txt, error-dump.log\n")
    
    return sb.String()
}

// buildLogExtractionPrompt creates a prompt for Phase 2
func buildLogExtractionPrompt(input *AnalysisInput, fileContents map[string]string) string {
    var sb strings.Builder
    
    sb.WriteString("You are extracting error logs from Jira issue attachments.\n\n")
    sb.WriteString(fmt.Sprintf("Issue ID: %s\n", input.IssueID))
    sb.WriteString(fmt.Sprintf("Summary: %s\n\n", input.Summary))
    sb.WriteString(fmt.Sprintf("Description:\n%s\n\n", input.Description))
    
    sb.WriteString("File Contents:\n\n")
    for filename, content := range fileContents {
        sb.WriteString(fmt.Sprintf("=== %s ===\n", filename))
        sb.WriteString(content)
        sb.WriteString("\n\n")
    }
    
    sb.WriteString("Task: Extract all error logs, exceptions, stack traces, and critical warnings.\n")
    sb.WriteString("Requirements:\n")
    sb.WriteString("1. Include timestamp if available\n")
    sb.WriteString("2. Include error level (ERROR, FATAL, CRITICAL, etc.)\n")
    sb.WriteString("3. Include error message and stack trace if present\n")
    sb.WriteString("4. Group by file and sort by timestamp\n")
    sb.WriteString("5. Format each entry as:\n")
    sb.WriteString("   [TIMESTAMP] [LEVEL] [FILE] Message\n")
    sb.WriteString("   Stack trace (if available)\n\n")
    sb.WriteString("Output the compiled error logs:\n")
    
    return sb.String()
}

// parseFileList parses Claude's comma-separated file list response
func parseFileList(response string) []string {
    // Clean up response
    response = strings.TrimSpace(response)
    
    // Split by comma
    parts := strings.Split(response, ",")
    
    var files []string
    for _, part := range parts {
        file := strings.TrimSpace(part)
        if file != "" {
            files = append(files, file)
        }
    }
    
    return files
}

// readFileWithLimit reads a file with a size limit
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
    
    // If file is too large, read only first portion
    if info.Size() > maxBytes {
        buf := make([]byte, maxBytes)
        n, err := file.Read(buf)
        if err != nil && err != io.EOF {
            return "", err
        }
        return string(buf[:n]) + "\n\n[FILE TRUNCATED - TOO LARGE]", nil
    }
    
    // Read entire file
    content, err := io.ReadAll(file)
    if err != nil {
        return "", err
    }
    
    return string(content), nil
}
```

**Tests:**
```go
// internal/claude/analyzer_test.go
func TestIdentifyRelevantFiles(t *testing.T)
func TestParseFileList(t *testing.T)
func TestReadFileWithLimit(t *testing.T)
func TestBuildFileIdentificationPrompt(t *testing.T)
func TestBuildLogExtractionPrompt(t *testing.T)
```

---

### 4. New Command: `sniff`

File: `cmd/sniff.go`

```go
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
        fmt.Println()
    }

    // Step 7: Write findings
    if verbose {
        fmt.Println("Step 7: Writing findings...")
    }

    findingsPath := filepath.Join(workDir, "FINDINGS.txt")
    if err := os.WriteFile(findingsPath, []byte(result.Findings), 0644); err != nil {
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
                os.Remove(filepath.Join(workDir, file))
            }
        }
    }

    return nil
}
```

**Helper function to add to `internal/files/handler.go`:**

```go
// IsArchive checks if a file is an archive
func IsArchive(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    return ext == ".zip" || ext == ".tar" || 
           strings.HasSuffix(filename, ".tar.gz") || 
           strings.HasSuffix(filename, ".tgz")
}
```

---

## File Structure

New files to create:

```
spur/
├── cmd/
│   └── sniff.go                    # New sniff command
├── internal/
│   ├── files/                      # New package
│   │   ├── handler.go
│   │   └── handler_test.go
│   ├── claude/
│   │   ├── client.go             # Update with new functions
│   │   └── client_test.go        # Add new tests
│   └── jira/
│       ├── client.go               # Update with ListAttachments, DownloadAttachment
│       └── client_test.go          # Add new tests
```

---

## Testing Strategy

### Unit Tests

1. **Jira Tests:**
   - `TestListAttachments_Success`
   - `TestListAttachments_NoAttachments`
   - `TestListAttachments_Unauthorized`
   - `TestDownloadAttachment_Success`
   - `TestDownloadAttachment_FileCreationError`

2. **Files Tests:**
   - `TestIsReadableFile_TextFiles`
   - `TestIsReadableFile_ImageFiles`
   - `TestIsReadableFile_Archives`
   - `TestExtractTarGz_Success`
   - `TestExtractZip_Success`
   - `TestListFiles_Success`

3. **Claude Analyzer Tests:**
   - `TestParseFileList_CommaSeparated`
   - `TestParseFileList_WithSpaces`
   - `TestReadFileWithLimit_SmallFile`
   - `TestReadFileWithLimit_LargeFile`
   - `TestBuildFileIdentificationPrompt`
   - `TestBuildLogExtractionPrompt`

### Integration Test (Manual)

```bash
# Create a test Jira issue with attachments
# Run sniff command
spur sniff SURE-11483 --verbose

# Verify:
# 1. Directory SURE-11483/ created
# 2. Readable files downloaded
# 3. Archives extracted
# 4. FINDINGS.txt created with error logs
```

---

## Example FINDINGS.txt Output

```
JIRA Issue: SURE-11483
Summary: Rancher crashes during upgrade
Analysis Date: 2026-05-11

=== Relevant Files Analyzed ===
- suse-observability_logs_2026-03-24/rancher.log
- 2026-03-24_15-58-04.txt
- cluster-state.yaml

=== Error Logs (Sorted by Timestamp) ===

[2026-03-24 15:52:37] [ERROR] [rancher.log]
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x7f8b9c]

goroutine 1 [running]:
main.(*RancherServer).Start(0xc000124000)
        /app/server.go:123 +0x45
...

[2026-03-24 15:58:04] [FATAL] [2026-03-24_15-58-04.txt]
Failed to connect to database: connection timeout after 30s
Database host: postgres.svc.cluster.local:5432
...

[2026-03-24 16:00:00] [ERROR] [rancher.log]
Failed to reconcile cluster: admission webhook "validate.cluster" denied the request
...

=== Summary ===
Total errors found: 3
Most recent: 2026-03-24 16:00:00
Recommendation: Check database connectivity and webhook configuration
```

---

Once you answer these, I'll proceed with the implementation!
