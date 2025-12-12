package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ygncode/gdk/internal/config"
	"github.com/ygncode/gdk/internal/store"
	"github.com/ygncode/gdk/pkg/utils"
)

var showCmd = &cobra.Command{
	Use:   "show <owner/repo>",
	Short: "Show details of a deploy key",
	Long: `Display detailed information about a specific deploy key,
including file paths, public key content, and git commands.

Example:
  gdk show myorg/private-repo
  gdk show myorg/private-repo --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

var showFormat string

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().StringVarP(&showFormat, "format", "o", "text", "Output format: text, json")
}

func runShow(cmd *cobra.Command, args []string) error {
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

	// Load key metadata
	keyStore := store.New(keysDir)
	deployKey, err := keyStore.Load(owner, repo)
	if err != nil {
		return fmt.Errorf("deploy key not found for %s/%s", owner, repo)
	}

	// Read public key content
	publicKeyContent, err := os.ReadFile(deployKey.PublicKeyPath)
	if err != nil {
		publicKeyContent = []byte("(unable to read public key)")
	}

	switch showFormat {
	case "json":
		return outputShowJSON(deployKey, string(publicKeyContent))
	default:
		return outputShowText(deployKey, string(publicKeyContent))
	}
}

func outputShowText(key *store.DeployKey, publicKey string) error {
	repoRef := fmt.Sprintf("%s/%s", key.RepoOwner, key.RepoName)

	fmt.Printf("Repository:     %s\n", repoRef)
	fmt.Printf("Host Alias:     %s\n", key.HostAlias)
	fmt.Printf("Created:        %s\n", key.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println("SSH Keys:")
	fmt.Printf("  Private Key:  %s\n", key.PrivateKeyPath)
	fmt.Printf("  Public Key:   %s\n", key.PublicKeyPath)
	fmt.Println()
	fmt.Println("Public Key Content:")
	fmt.Println(strings.TrimSpace(publicKey))
	fmt.Println()
	fmt.Println("Git Commands:")
	fmt.Printf("  Clone:        git clone git@%s:%s.git\n", key.HostAlias, repoRef)
	fmt.Printf("  Set Remote:   git remote set-url origin git@%s:%s.git\n", key.HostAlias, repoRef)

	return nil
}

type showOutput struct {
	Repository   string `json:"repository"`
	RepoOwner    string `json:"repo_owner"`
	RepoName     string `json:"repo_name"`
	HostAlias    string `json:"host_alias"`
	PrivateKey   string `json:"private_key_path"`
	PublicKey    string `json:"public_key_path"`
	PublicKeyCnt string `json:"public_key_content"`
	CreatedAt    string `json:"created_at"`
	CloneCmd     string `json:"clone_command"`
	RemoteCmd    string `json:"remote_command"`
}

func outputShowJSON(key *store.DeployKey, publicKey string) error {
	repoRef := fmt.Sprintf("%s/%s", key.RepoOwner, key.RepoName)

	output := showOutput{
		Repository:   repoRef,
		RepoOwner:    key.RepoOwner,
		RepoName:     key.RepoName,
		HostAlias:    key.HostAlias,
		PrivateKey:   key.PrivateKeyPath,
		PublicKey:    key.PublicKeyPath,
		PublicKeyCnt: strings.TrimSpace(publicKey),
		CreatedAt:    key.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CloneCmd:     fmt.Sprintf("git clone git@%s:%s.git", key.HostAlias, repoRef),
		RemoteCmd:    fmt.Sprintf("git remote set-url origin git@%s:%s.git", key.HostAlias, repoRef),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
