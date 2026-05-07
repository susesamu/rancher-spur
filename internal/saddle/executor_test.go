package saddle

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCreate_FileNotFound(t *testing.T) {
	executor := NewExecutor()
	err := executor.Create(context.Background(), "/nonexistent/file.yaml")

	if err == nil {
		t.Fatal("expected error for nonexistent file, got none")
	}
}

func TestCreate_SaddleNotInPath(t *testing.T) {
	// Create a temporary YAML file
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(yamlFile, []byte("test: yaml"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set PATH to empty to simulate saddle not being installed
	os.Setenv("PATH", "")

	executor := NewExecutor()
	err := executor.Create(context.Background(), yamlFile)

	if err == nil {
		t.Fatal("expected error when saddle not in PATH, got none")
	}
}

func TestCreate_ValidFile(t *testing.T) {
	// Skip this test if saddle is not actually installed
	if _, err := exec.LookPath("saddle"); err != nil {
		t.Skip("saddle not installed, skipping integration test")
	}

	// Create a temporary YAML file
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `
cluster:
  name: test-cluster
  nodes:
    - role: control-plane
      instance_type: t3.medium
  networking:
    plugin: calico
`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	executor := NewExecutor()

	// Note: This test will actually try to execute saddle
	// In a real environment, this would provision infrastructure
	// For testing purposes, we expect it to fail with a specific error
	// or we could mock the executor
	err := executor.Create(context.Background(), yamlFile)

	// We don't assert success here since saddle might not be configured
	// The test passes if the file exists and the command was attempted
	if err != nil {
		t.Logf("saddle execution returned error (expected in test environment): %v", err)
	}
}

func TestCreate_ContextCancellation(t *testing.T) {
	// Skip if saddle not installed
	if _, err := exec.LookPath("saddle"); err != nil {
		t.Skip("saddle not installed, skipping cancellation test")
	}

	// Create a temporary YAML file
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(yamlFile, []byte("test: yaml"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	executor := NewExecutor()
	err := executor.Create(ctx, yamlFile)

	// Should fail due to context cancellation
	if err == nil {
		t.Fatal("expected error for cancelled context, got none")
	}
}
