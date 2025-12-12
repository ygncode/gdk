package sshconfig

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	gdkMarkerPrefix = "# gdk:"
)

var (
	// ErrHostExists indicates the SSH host already exists
	ErrHostExists = errors.New("SSH host already exists")

	// ErrHostNotFound indicates the SSH host was not found
	ErrHostNotFound = errors.New("SSH host not found")
)

// HostEntry represents an SSH config Host block
type HostEntry struct {
	Host           string // The Host alias (e.g., github.com-user-repo)
	HostName       string // Actual hostname (github.com)
	User           string // SSH user (git)
	IdentityFile   string // Path to private key
	IdentitiesOnly bool   // Use only specified identity
}

// Manager manages SSH config files
type Manager struct {
	configPath string
}

// New creates a new SSH config Manager
func New(configPath string) *Manager {
	return &Manager{configPath: configPath}
}

// AddHost adds a new host entry to the SSH config
func (m *Manager) AddHost(entry HostEntry, repoRef string) error {
	// Check if host already exists
	if m.HostExists(entry.Host) {
		return fmt.Errorf("%w: %s", ErrHostExists, entry.Host)
	}

	// Format the entry
	newEntry := FormatEntry(entry, repoRef)

	// Read existing content
	content, err := m.readConfig()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read SSH config: %w", err)
	}

	// Append new entry
	var newContent string
	if len(content) > 0 {
		// Ensure there's a newline before our entry
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		newContent = content + "\n" + newEntry
	} else {
		newContent = newEntry
	}

	// Write back
	return m.writeConfig(newContent)
}

// RemoveHost removes a host entry from the SSH config
func (m *Manager) RemoveHost(alias string) error {
	content, err := m.readConfig()
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrHostNotFound, alias)
		}
		return fmt.Errorf("failed to read SSH config: %w", err)
	}

	newContent, found := m.removeHostBlock(content, alias)
	if !found {
		return fmt.Errorf("%w: %s", ErrHostNotFound, alias)
	}

	return m.writeConfig(newContent)
}

// HostExists checks if a host alias exists in the config
func (m *Manager) HostExists(alias string) bool {
	content, err := m.readConfig()
	if err != nil {
		return false
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(strings.ToLower(line), "host ") {
			hosts := strings.Fields(line)[1:]
			for _, h := range hosts {
				if h == alias {
					return true
				}
			}
		}
	}

	return false
}

// readConfig reads the SSH config file
func (m *Manager) readConfig() (string, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// writeConfig writes the SSH config file
func (m *Manager) writeConfig(content string) error {
	// Ensure parent directory exists
	dir := strings.TrimSuffix(m.configPath, "/config")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	return os.WriteFile(m.configPath, []byte(content), 0600)
}

// removeHostBlock removes a host block from the config content
func (m *Manager) removeHostBlock(content, alias string) (string, bool) {
	lines := strings.Split(content, "\n")
	var result []string
	var found bool
	var inTargetBlock bool
	var skipGdkMarker bool

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for gdk marker that precedes the target host
		if strings.HasPrefix(trimmed, gdkMarkerPrefix) {
			// Look ahead to see if next Host line is our target
			for j := i + 1; j < len(lines); j++ {
				nextTrimmed := strings.TrimSpace(lines[j])
				if nextTrimmed == "" {
					continue
				}
				if strings.HasPrefix(strings.ToLower(nextTrimmed), "host ") {
					hosts := strings.Fields(nextTrimmed)[1:]
					for _, h := range hosts {
						if h == alias {
							skipGdkMarker = true
							break
						}
					}
				}
				break
			}
			if skipGdkMarker {
				skipGdkMarker = false
				continue // Skip the gdk marker line
			}
		}

		// Check if this is a Host line
		if strings.HasPrefix(strings.ToLower(trimmed), "host ") {
			if inTargetBlock {
				// We've hit a new Host block, stop skipping
				inTargetBlock = false
			}

			hosts := strings.Fields(trimmed)[1:]
			for _, h := range hosts {
				if h == alias {
					inTargetBlock = true
					found = true
					break
				}
			}

			if inTargetBlock {
				continue // Skip this line
			}
		}

		// Skip lines in target block
		if inTargetBlock {
			// Skip indented lines (part of the host block)
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") || trimmed == "" {
				continue
			}
			// Non-indented, non-empty line means new section
			inTargetBlock = false
		}

		result = append(result, line)
	}

	// Clean up multiple consecutive empty lines
	resultStr := strings.Join(result, "\n")
	for strings.Contains(resultStr, "\n\n\n") {
		resultStr = strings.ReplaceAll(resultStr, "\n\n\n", "\n\n")
	}

	return strings.TrimSpace(resultStr) + "\n", found
}

// FormatEntry formats a host entry for the SSH config file
func FormatEntry(entry HostEntry, repoRef string) string {
	var sb strings.Builder

	// Write gdk marker with timestamp
	sb.WriteString(fmt.Sprintf("%s%s - Added by gdk on %s\n",
		gdkMarkerPrefix,
		repoRef,
		time.Now().Format("2006-01-02")))

	// Write Host line
	sb.WriteString(fmt.Sprintf("Host %s\n", entry.Host))

	// Write directives with 4-space indent
	sb.WriteString(fmt.Sprintf("    HostName %s\n", entry.HostName))
	sb.WriteString(fmt.Sprintf("    User %s\n", entry.User))
	sb.WriteString(fmt.Sprintf("    IdentityFile %s\n", entry.IdentityFile))

	if entry.IdentitiesOnly {
		sb.WriteString("    IdentitiesOnly yes\n")
	}

	return sb.String()
}
