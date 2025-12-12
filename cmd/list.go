package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ygncode/gdk/internal/config"
	"github.com/ygncode/gdk/internal/store"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed deploy keys",
	Long:  "Display all deploy keys managed by gdk with their details.",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

var listFormat string

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&listFormat, "format", "o", "table", "Output format: table, json")
}

func runList(cmd *cobra.Command, args []string) error {
	keysDir, err := config.DeployKeysDir()
	if err != nil {
		return fmt.Errorf("failed to get deploy keys directory: %w", err)
	}

	keyStore := store.New(keysDir)
	keys, err := keyStore.List()
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	if len(keys) == 0 {
		fmt.Println("No deploy keys found.")
		fmt.Println("Use 'gdk add <owner/repo>' to create one.")
		return nil
	}

	switch listFormat {
	case "json":
		return outputJSON(keys)
	default:
		return outputTable(keys)
	}
}

func outputTable(keys []store.DeployKey) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "REPOSITORY\tHOST ALIAS\tCREATED")
	fmt.Fprintln(w, "----------\t----------\t-------")
	for _, k := range keys {
		fmt.Fprintf(w, "%s/%s\t%s\t%s\n",
			k.RepoOwner, k.RepoName,
			k.HostAlias,
			k.CreatedAt.Format("2006-01-02"))
	}
	return w.Flush()
}

func outputJSON(keys []store.DeployKey) error {
	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
