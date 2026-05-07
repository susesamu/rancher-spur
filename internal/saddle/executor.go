package saddle

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// Executor defines the interface for Saddle CLI operations
type Executor interface {
	Create(ctx context.Context, yamlFile string) error
}

// CLIExecutor implements the Executor interface by calling the saddle CLI
type CLIExecutor struct{}

// NewExecutor creates a new Saddle executor
func NewExecutor() Executor {
	return &CLIExecutor{}
}

// Create executes 'saddle create <yamlFile>' and streams output to the user
func (e *CLIExecutor) Create(ctx context.Context, yamlFile string) error {
	// Check if saddle binary exists in PATH
	if _, err := exec.LookPath("saddle"); err != nil {
		return fmt.Errorf("saddle CLI not found in PATH - please install saddle first")
	}

	// Check if YAML file exists
	if _, err := os.Stat(yamlFile); err != nil {
		return fmt.Errorf("YAML file not found: %s", yamlFile)
	}

	// Create command with context for cancellation support
	cmd := exec.CommandContext(ctx, "saddle", "create", yamlFile)

	// Stream stdout and stderr to the user in real-time
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Execute the command
	if err := cmd.Run(); err != nil {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return fmt.Errorf("saddle execution cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("saddle execution failed: %w", err)
	}

	return nil
}
