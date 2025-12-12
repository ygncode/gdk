package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ygncode/gdk/internal/updater"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update gdk to the latest version",
	Long: `Check for and install the latest version of gdk.

Example:
  gdk update          # Update to latest version
  gdk update --check  # Only check for updates`,
	RunE: runUpdate,
}

var checkOnly bool

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Only check for updates, don't install")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	u := updater.New(Version, "ygncode", "gdk")

	fmt.Printf("Current version: %s\n", Version)

	// Get latest version
	latestVersion, err := u.GetLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	fmt.Printf("Latest version:  %s\n", latestVersion)
	fmt.Println()

	if !u.NeedsUpdate(latestVersion) {
		fmt.Println("You are already running the latest version!")
		return nil
	}

	if checkOnly {
		fmt.Println("Update available! Run 'gdk update' to install.")
		return nil
	}

	// Perform update
	fmt.Printf("Downloading %s...\n", updater.GetBinaryName())

	// Download checksums
	checksums, err := u.DownloadChecksums(latestVersion)
	if err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}

	// Download binary
	binaryName := updater.GetBinaryName()
	data, err := u.DownloadBinary(latestVersion)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}

	fmt.Println("Verifying checksum...")

	// Verify checksum
	expectedChecksum, ok := checksums[binaryName]
	if !ok {
		return fmt.Errorf("no checksum found for %s", binaryName)
	}

	if !updater.VerifyChecksum(data, expectedChecksum) {
		return fmt.Errorf("checksum verification failed - download may be corrupted")
	}

	fmt.Println("Installing update...")

	// Install
	if err := u.Install(data); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	fmt.Println()
	fmt.Printf("Successfully updated to %s!\n", latestVersion)

	return nil
}
