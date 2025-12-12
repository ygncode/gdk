package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gdk",
	Short: "GitHub Deploy Key Manager",
	Long: `GDK (GitHub Deploy Key Manager) automates the creation of
repository-specific SSH keys and manages ~/.ssh/config to allow
multiple distinct GitHub Deploy Keys on a single machine.

Example:
  gdk add myorg/private-repo    # Create deploy key for a repository
  gdk list                      # List all managed deploy keys
  gdk remove myorg/private-repo # Remove a deploy key`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
