package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ygncode/gdk/internal/config"
	"github.com/ygncode/gdk/internal/keygen"
	"github.com/ygncode/gdk/internal/sshconfig"
	"github.com/ygncode/gdk/internal/store"
	"github.com/ygncode/gdk/pkg/utils"
)

var addCmd = &cobra.Command{
	Use:   "add <owner/repo>",
	Short: "Add a deploy key for a GitHub repository",
	Long: `Generate an Ed25519 SSH key pair for a GitHub repository,
configure SSH settings, and output the public key for adding to GitHub.

Example:
  gdk add myorg/private-repo
  gdk add myorg/private-repo --force  # Overwrite existing key`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

var forceAdd bool

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolVarP(&forceAdd, "force", "f", false, "Overwrite existing key")
}

func runAdd(cmd *cobra.Command, args []string) error {
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

	// Check for existing key
	hostAlias := utils.BuildHostAlias(owner, repo)
	if keyStore.Exists(owner, repo) && !forceAdd {
		return fmt.Errorf("deploy key already exists for %s/%s\n\nUse --force to overwrite or 'gdk remove %s/%s' first",
			owner, repo, owner, repo)
	}

	// If force and exists, remove old config entry first
	if forceAdd && sshConfigMgr.HostExists(hostAlias) {
		_ = sshConfigMgr.RemoveHost(hostAlias)
	}

	// Generate key pair
	generator := keygen.NewEd25519Generator()
	comment := fmt.Sprintf("gdk:%s/%s", owner, repo)
	keyPair, err := generator.Generate(comment)
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	// Create deploy key metadata
	deployKey := &store.DeployKey{
		RepoOwner: owner,
		RepoName:  repo,
		HostAlias: hostAlias,
	}

	// Save keys to disk
	if err := keyStore.Save(deployKey, keyPair); err != nil {
		return fmt.Errorf("failed to save key: %w", err)
	}

	// Build identity file path for SSH config (use ~ for portability)
	identityFile := filepath.Join("~/.ssh/deploy-keys", utils.SanitizeForPath(owner, repo), "id_ed25519")

	// Create SSH config entry
	entry := sshconfig.HostEntry{
		Host:           hostAlias,
		HostName:       "github.com",
		User:           "git",
		IdentityFile:   identityFile,
		IdentitiesOnly: true,
	}

	// Update SSH config
	if err := sshConfigMgr.AddHost(entry, repoRef); err != nil {
		// Rollback: delete saved keys
		_ = keyStore.Delete(owner, repo)
		return fmt.Errorf("failed to update SSH config: %w", err)
	}

	// Output success message
	printSuccess(owner, repo, hostAlias, keyPair.PublicKeyStr)

	return nil
}

func printSuccess(owner, repo, hostAlias, publicKey string) {
	fmt.Println("Deploy key created successfully!")
	fmt.Println()
	fmt.Println("Public key (add this to GitHub deploy keys):")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println(publicKey)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	fmt.Println("Clone command:")
	fmt.Printf("  git clone git@%s:%s/%s.git\n", hostAlias, owner, repo)
	fmt.Println()
	fmt.Println("For existing repos, update remote:")
	fmt.Printf("  git remote set-url origin git@%s:%s/%s.git\n", hostAlias, owner, repo)
}
