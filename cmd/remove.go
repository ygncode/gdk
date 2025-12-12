package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ygncode/gdk/internal/config"
	"github.com/ygncode/gdk/internal/sshconfig"
	"github.com/ygncode/gdk/internal/store"
	"github.com/ygncode/gdk/pkg/utils"
)

var removeCmd = &cobra.Command{
	Use:     "remove <owner/repo>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a deploy key",
	Long: `Remove a deploy key and its SSH config entry.

Example:
  gdk remove myorg/private-repo
  gdk remove myorg/private-repo --yes  # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

var skipConfirm bool

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")
}

func runRemove(cmd *cobra.Command, args []string) error {
	repoRef := args[0]

	// Parse and validate input
	owner, repo, err := utils.ParseRepoRef(repoRef)
	if err != nil {
		return fmt.Errorf("invalid repository format: %w", err)
	}

	// Get paths
	keysDir, err := config.DeployKeysDir()
	if err != nil {
		return fmt.Errorf("failed to get deploy keys directory: %w", err)
	}

	sshConfigPath, err := config.SSHConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get SSH config path: %w", err)
	}

	// Initialize store and SSH config manager
	keyStore := store.New(keysDir)
	sshConfigMgr := sshconfig.New(sshConfigPath)

	// Check if key exists
	deployKey, err := keyStore.Load(owner, repo)
	if err != nil {
		return fmt.Errorf("deploy key not found for %s/%s", owner, repo)
	}

	// Confirm deletion
	if !skipConfirm {
		fmt.Printf("Remove deploy key for %s/%s? [y/N] ", owner, repo)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Remove from SSH config first
	if err := sshConfigMgr.RemoveHost(deployKey.HostAlias); err != nil {
		// Ignore error if host not found in config (might have been manually removed)
		if !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("failed to update SSH config: %w", err)
		}
	}

	// Delete key files
	if err := keyStore.Delete(owner, repo); err != nil {
		return fmt.Errorf("failed to delete key files: %w", err)
	}

	fmt.Printf("Deploy key for %s/%s removed successfully.\n", owner, repo)
	return nil
}
