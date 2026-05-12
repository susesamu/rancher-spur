package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFileList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "comma separated",
			input:    "file1.log, file2.txt, dir/file3.log",
			expected: []string{"file1.log", "file2.txt", "dir/file3.log"},
		},
		{
			name:     "no spaces",
			input:    "file1.log,file2.txt,file3.log",
			expected: []string{"file1.log", "file2.txt", "file3.log"},
		},
		{
			name:     "with extra spaces",
			input:    "  file1.log  ,  file2.txt  ",
			expected: []string{"file1.log", "file2.txt"},
		},
		{
			name:     "none response",
			input:    "NONE",
			expected: []string{},
		},
		{
			name:     "single file",
			input:    "file1.log",
			expected: []string{"file1.log"},
		},
		{
			name:     "with markdown code block",
			input:    "```file1.log, file2.txt```",
			expected: []string{"file1.log", "file2.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFileList(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d files, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expected file[%d] = '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestReadFileWithLimit_SmallFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "small file content"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	content, err := readFileWithLimit(testFile, 1024)
	if err != nil {
		t.Fatalf("readFileWithLimit failed: %v", err)
	}

	if content != testContent {
		t.Errorf("expected content '%s', got '%s'", testContent, content)
	}
}

func TestReadFileWithLimit_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create a large file (2000 bytes)
	largeContent := strings.Repeat("A", 1000) + strings.Repeat("B", 1000)
	err := os.WriteFile(testFile, []byte(largeContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Read with 1KB limit
	content, err := readFileWithLimit(testFile, 1000)
	if err != nil {
		t.Fatalf("readFileWithLimit failed: %v", err)
	}

	// Should contain truncation message
	if !strings.Contains(content, "FILE TRUNCATED") {
		t.Error("expected truncation message in content")
	}

	// Should contain parts from beginning and end
	if !strings.Contains(content, "AAA") {
		t.Error("expected content from beginning of file")
	}
	if !strings.Contains(content, "BBB") {
		t.Error("expected content from end of file")
	}
}

func TestBuildFileIdentificationPrompt(t *testing.T) {
	input := &AnalysisInput{
		IssueID:     "SURE-11483",
		Summary:     "Rancher crash",
		Description: "Rancher crashes during upgrade",
		FileList:    []string{"rancher.log", "config.yaml", "test.txt"},
	}

	prompt := buildFileIdentificationPrompt(input)

	// Check that prompt contains required elements
	requiredElements := []string{
		"SURE-11483",
		"Rancher crash",
		"Rancher crashes during upgrade",
		"rancher.log",
		"config.yaml",
		"test.txt",
		"comma-separated",
	}

	for _, element := range requiredElements {
		if !strings.Contains(prompt, element) {
			t.Errorf("prompt should contain '%s'", element)
		}
	}
}

func TestBuildLogExtractionPrompt(t *testing.T) {
	input := &AnalysisInput{
		IssueID:     "SURE-11483",
		Summary:     "Rancher crash",
		Description: "Rancher crashes during upgrade",
	}

	fileContents := map[string]string{
		"rancher.log": "ERROR: panic at line 123",
		"debug.txt":   "DEBUG: connection established",
	}

	prompt := buildLogExtractionPrompt(input, fileContents)

	// Check that prompt contains required elements
	requiredElements := []string{
		"SURE-11483",
		"Rancher crash",
		"rancher.log",
		"ERROR: panic at line 123",
		"debug.txt",
		"DEBUG: connection established",
		"ERROR",
		"FATAL",
		"timestamp",
	}

	for _, element := range requiredElements {
		if !strings.Contains(prompt, element) {
			t.Errorf("prompt should contain '%s'", element)
		}
	}
}

func TestReadFileWithLimit_NonexistentFile(t *testing.T) {
	_, err := readFileWithLimit("/nonexistent/file.txt", 1024)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got none")
	}
}

func TestBuildFileIdentificationPrompt_EmptyDescription(t *testing.T) {
	input := &AnalysisInput{
		IssueID:  "TEST-123",
		Summary:  "Test issue",
		FileList: []string{"file1.log"},
	}

	prompt := buildFileIdentificationPrompt(input)

	if !strings.Contains(prompt, "TEST-123") {
		t.Error("prompt should contain issue ID")
	}
	if !strings.Contains(prompt, "Test issue") {
		t.Error("prompt should contain summary")
	}
}

func TestBuildLogExtractionPrompt_MultipleFiles(t *testing.T) {
	input := &AnalysisInput{
		IssueID: "TEST-123",
		Summary: "Test issue",
	}

	fileContents := map[string]string{
		"file1.log": "ERROR: first error",
		"file2.log": "ERROR: second error",
		"file3.log": "ERROR: third error",
	}

	prompt := buildLogExtractionPrompt(input, fileContents)

	// All files should be included
	for filename := range fileContents {
		if !strings.Contains(prompt, filename) {
			t.Errorf("prompt should contain '%s'", filename)
		}
	}
}
