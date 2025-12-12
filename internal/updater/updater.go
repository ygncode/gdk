package updater

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Updater handles self-update functionality
type Updater struct {
	CurrentVersion string
	RepoOwner      string
	RepoName       string
}

// GitHubRelease represents the GitHub API response for a release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

// New creates a new Updater
func New(currentVersion, repoOwner, repoName string) *Updater {
	return &Updater{
		CurrentVersion: currentVersion,
		RepoOwner:      repoOwner,
		RepoName:       repoName,
	}
}

// GetBinaryName returns the binary name for the current platform
func GetBinaryName() string {
	return GetBinaryNameForPlatform(runtime.GOOS, runtime.GOARCH)
}

// GetBinaryNameForPlatform returns the binary name for a specific platform
func GetBinaryNameForPlatform(goos, goarch string) string {
	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("gdk-%s-%s%s", goos, goarch, ext)
}

// CompareVersions compares two version strings
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Strip 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Pad to same length
	for len(parts1) < 3 {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < 3 {
		parts2 = append(parts2, "0")
	}

	for i := 0; i < 3; i++ {
		n1, _ := strconv.Atoi(parts1[i])
		n2, _ := strconv.Atoi(parts2[i])

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

// NeedsUpdate checks if an update is needed
func (u *Updater) NeedsUpdate(latestVersion string) bool {
	return CompareVersions(u.CurrentVersion, latestVersion) < 0
}

// VerifyChecksum verifies SHA256 checksum of data
func VerifyChecksum(data []byte, expectedChecksum string) bool {
	hash := sha256.Sum256(data)
	actualChecksum := fmt.Sprintf("%x", hash)
	return actualChecksum == expectedChecksum
}

// ParseChecksumsFile parses a checksums.txt file content
// Format: <checksum>  <filename>
func ParseChecksumsFile(content string) map[string]string {
	checksums := make(map[string]string)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split by two spaces (standard checksum format)
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			// Try single space
			parts = strings.Fields(line)
			if len(parts) != 2 {
				continue
			}
		}

		checksum := strings.TrimSpace(parts[0])
		filename := strings.TrimSpace(parts[1])
		checksums[filename] = checksum
	}

	return checksums
}

// BuildAPIURL returns the GitHub API URL for latest release
func (u *Updater) BuildAPIURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.RepoOwner, u.RepoName)
}

// BuildDownloadURL returns the download URL for a specific version and binary
func (u *Updater) BuildDownloadURL(version, binaryName string) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
		u.RepoOwner, u.RepoName, version, binaryName)
}

// BuildChecksumsURL returns the checksums.txt URL for a specific version
func (u *Updater) BuildChecksumsURL(version string) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/checksums.txt",
		u.RepoOwner, u.RepoName, version)
}

// GetLatestVersion fetches the latest release version from GitHub
func (u *Updater) GetLatestVersion() (string, error) {
	resp, err := http.Get(u.BuildAPIURL())
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	return release.TagName, nil
}

// DownloadBinary downloads the binary for the current platform
func (u *Updater) DownloadBinary(version string) ([]byte, error) {
	binaryName := GetBinaryName()
	url := u.BuildDownloadURL(version, binaryName)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read binary: %w", err)
	}

	return data, nil
}

// DownloadChecksums downloads and parses the checksums file
func (u *Updater) DownloadChecksums(version string) (map[string]string, error) {
	url := u.BuildChecksumsURL(version)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checksums download failed with status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read checksums: %w", err)
	}

	return ParseChecksumsFile(string(content)), nil
}

// Install replaces the current binary with new data
func (u *Updater) Install(data []byte) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create temp file in same directory for atomic rename
	tmpPath := execPath + ".new"
	if err := os.WriteFile(tmpPath, data, 0755); err != nil {
		return fmt.Errorf("failed to write new binary: %w", err)
	}

	// Backup old binary
	backupPath := execPath + ".old"
	if err := os.Rename(execPath, backupPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to backup old binary: %w", err)
	}

	// Move new binary into place
	if err := os.Rename(tmpPath, execPath); err != nil {
		// Try to restore backup
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Remove backup
	os.Remove(backupPath)

	return nil
}

// Update performs the full update process
func (u *Updater) Update() (string, error) {
	// Get latest version
	latestVersion, err := u.GetLatestVersion()
	if err != nil {
		return "", err
	}

	// Check if update needed
	if !u.NeedsUpdate(latestVersion) {
		return latestVersion, nil
	}

	// Download checksums
	checksums, err := u.DownloadChecksums(latestVersion)
	if err != nil {
		return "", err
	}

	// Download binary
	binaryName := GetBinaryName()
	data, err := u.DownloadBinary(latestVersion)
	if err != nil {
		return "", err
	}

	// Verify checksum
	expectedChecksum, ok := checksums[binaryName]
	if !ok {
		return "", fmt.Errorf("no checksum found for %s", binaryName)
	}

	if !VerifyChecksum(data, expectedChecksum) {
		return "", fmt.Errorf("checksum verification failed")
	}

	// Install
	if err := u.Install(data); err != nil {
		return "", err
	}

	return latestVersion, nil
}
