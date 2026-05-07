package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/suse/rancher/rancher-spur/internal/config"
)

var (
	verbose bool
	cfg     *config.Config
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "spur",
	Short: "Automate environment reproduction from Jira issues",
	Long: `Spur is a CLI tool that automates the creation of reproducible environments
based on Jira issue descriptions. It integrates with Jira API, Claude AI,
and Saddle CLI to streamline environment setup for debugging and testing.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}
