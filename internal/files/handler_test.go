package files

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestIsReadableFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.txt", true},
		{"test.log", true},
		{"test.yaml", true},
		{"test.json", true},
		{"test.tar.gz", true},
		{"test.zip", true},
		{"test.png", false},
		{"test.jpg", false},
		{"test.gif", false},
		{"test.exe", false},
		{"test.bin", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := IsReadableFile(tt.filename)
			if result != tt.expected {
				t.Errorf("IsReadableFile(%s) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestIsArchive(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.tar.gz", true},
		{"test.tgz", true},
		{"test.zip", true},
		{"test.tar", true},
		{"test.txt", false},
		{"test.log", false},
		{"test.TAR.GZ", true}, // Case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := IsArchive(tt.filename)
			if result != tt.expected {
				t.Errorf("IsArchive(%s) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestExtractTarGz(t *testing.T) {
	// Create a test tar.gz file
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	extractDir := filepath.Join(tmpDir, "extracted")

	// Create test tar.gz with sample content
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	gzWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzWriter)

	// Add a test file to the archive
	testContent := []byte("test file content")
	header := &tar.Header{
		Name: "testfile.txt",
		Mode: 0644,
		Size: int64(len(testContent)),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatal(err)
	}

	if _, err := tarWriter.Write(testContent); err != nil {
		t.Fatal(err)
	}

	tarWriter.Close()
	gzWriter.Close()
	file.Close()

	// Extract the archive
	err = ExtractArchive(archivePath, extractDir)
	if err != nil {
		t.Fatalf("ExtractArchive failed: %v", err)
	}

	// Verify extracted file
	extractedFile := filepath.Join(extractDir, "testfile.txt")
	content, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}

	if string(content) != string(testContent) {
		t.Errorf("expected content '%s', got '%s'", string(testContent), string(content))
	}
}

func TestExtractZip(t *testing.T) {
	// Create a test zip file
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.zip")
	extractDir := filepath.Join(tmpDir, "extracted")

	// Create test zip with sample content
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	zipWriter := zip.NewWriter(file)

	// Add a test file to the archive
	testContent := []byte("test zip content")
	writer, err := zipWriter.Create("testfile.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := writer.Write(testContent); err != nil {
		t.Fatal(err)
	}

	zipWriter.Close()
	file.Close()

	// Extract the archive
	err = ExtractArchive(archivePath, extractDir)
	if err != nil {
		t.Fatalf("ExtractArchive failed: %v", err)
	}

	// Verify extracted file
	extractedFile := filepath.Join(extractDir, "testfile.txt")
	content, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}

	if string(content) != string(testContent) {
		t.Errorf("expected content '%s', got '%s'", string(testContent), string(content))
	}
}

func TestListFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test directory structure
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.log"), []byte("content2"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "file3.txt"), []byte("content3"), 0644)

	files, err := ListFiles(tmpDir)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	// Check that all files are included
	fileMap := make(map[string]bool)
	for _, f := range files {
		fileMap[f] = true
	}

	expectedFiles := []string{"file1.txt", "file2.log", filepath.Join("subdir", "file3.txt")}
	for _, expected := range expectedFiles {
		if !fileMap[expected] {
			t.Errorf("expected file '%s' not found in list", expected)
		}
	}
}

func TestExtractArchive_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.rar")

	// Create a dummy file
	os.WriteFile(archivePath, []byte("dummy"), 0644)

	err := ExtractArchive(archivePath, tmpDir)
	if err == nil {
		t.Fatal("expected error for unsupported format, got none")
	}
}
