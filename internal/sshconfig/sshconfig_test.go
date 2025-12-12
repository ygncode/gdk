package sshconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_AddHost(t *testing.T) {
	t.Run("adds host to empty config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		mgr := New(configPath)
		entry := HostEntry{
			Host:           "github.com-user-repo",
			HostName:       "github.com",
			User:           "git",
			IdentityFile:   "~/.ssh/deploy-keys/user-repo/id_ed25519",
			IdentitiesOnly: true,
		}

		err := mgr.AddHost(entry, "user/repo")
		require.NoError(t, err)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)

		assert.Contains(t, string(content), "Host github.com-user-repo")
		assert.Contains(t, string(content), "HostName github.com")
		assert.Contains(t, string(content), "User git")
		assert.Contains(t, string(content), "IdentityFile ~/.ssh/deploy-keys/user-repo/id_ed25519")
		assert.Contains(t, string(content), "IdentitiesOnly yes")
		assert.Contains(t, string(content), "# gdk:user/repo")
	})

	t.Run("appends without corrupting existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		// Write existing config
		existingConfig := `Host existing
    HostName example.com
    User admin
`
		err := os.WriteFile(configPath, []byte(existingConfig), 0600)
		require.NoError(t, err)

		mgr := New(configPath)
		entry := HostEntry{
			Host:           "github.com-test-repo",
			HostName:       "github.com",
			User:           "git",
			IdentityFile:   "~/.ssh/deploy-keys/test-repo/id_ed25519",
			IdentitiesOnly: true,
		}

		err = mgr.AddHost(entry, "test/repo")
		require.NoError(t, err)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)

		// Verify existing content preserved
		assert.Contains(t, string(content), "Host existing")
		assert.Contains(t, string(content), "HostName example.com")
		// Verify new content added
		assert.Contains(t, string(content), "Host github.com-test-repo")
	})

	t.Run("fails if host already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		mgr := New(configPath)
		entry := HostEntry{
			Host:           "github.com-user-repo",
			HostName:       "github.com",
			User:           "git",
			IdentityFile:   "~/.ssh/deploy-keys/user-repo/id_ed25519",
			IdentitiesOnly: true,
		}

		err := mgr.AddHost(entry, "user/repo")
		require.NoError(t, err)

		// Try to add again
		err = mgr.AddHost(entry, "user/repo")
		assert.Error(t, err)
	})
}

func TestManager_RemoveHost(t *testing.T) {
	t.Run("removes gdk-managed host", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		// Create config with gdk-managed host
		config := `# gdk:user/repo - Added by gdk
Host github.com-user-repo
    HostName github.com
    User git
    IdentityFile ~/.ssh/deploy-keys/user-repo/id_ed25519
    IdentitiesOnly yes

Host other-host
    HostName other.com
`
		err := os.WriteFile(configPath, []byte(config), 0600)
		require.NoError(t, err)

		mgr := New(configPath)
		err = mgr.RemoveHost("github.com-user-repo")
		require.NoError(t, err)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)

		assert.NotContains(t, string(content), "github.com-user-repo")
		assert.NotContains(t, string(content), "# gdk:user/repo")
		assert.Contains(t, string(content), "Host other-host")
	})

	t.Run("preserves other hosts", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		config := `Host first-host
    HostName first.com

# gdk:user/repo - Added by gdk
Host github.com-user-repo
    HostName github.com
    User git

Host last-host
    HostName last.com
`
		err := os.WriteFile(configPath, []byte(config), 0600)
		require.NoError(t, err)

		mgr := New(configPath)
		err = mgr.RemoveHost("github.com-user-repo")
		require.NoError(t, err)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)

		assert.Contains(t, string(content), "Host first-host")
		assert.Contains(t, string(content), "Host last-host")
		assert.NotContains(t, string(content), "github.com-user-repo")
	})

	t.Run("returns error for non-existent host", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		mgr := New(configPath)
		err := mgr.RemoveHost("nonexistent-host")
		assert.Error(t, err)
	})
}

func TestManager_HostExists(t *testing.T) {
	t.Run("returns true for existing host", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		config := `Host github.com-user-repo
    HostName github.com
`
		err := os.WriteFile(configPath, []byte(config), 0600)
		require.NoError(t, err)

		mgr := New(configPath)
		assert.True(t, mgr.HostExists("github.com-user-repo"))
	})

	t.Run("returns false for non-existent host", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		mgr := New(configPath)
		assert.False(t, mgr.HostExists("nonexistent"))
	})
}

func TestFormatEntry(t *testing.T) {
	entry := HostEntry{
		Host:           "github.com-user-repo",
		HostName:       "github.com",
		User:           "git",
		IdentityFile:   "~/.ssh/deploy-keys/user-repo/id_ed25519",
		IdentitiesOnly: true,
	}

	result := FormatEntry(entry, "user/repo")

	assert.Contains(t, result, "# gdk:user/repo")
	assert.Contains(t, result, "Host github.com-user-repo")
	assert.Contains(t, result, "    HostName github.com")
	assert.Contains(t, result, "    User git")
	assert.Contains(t, result, "    IdentityFile ~/.ssh/deploy-keys/user-repo/id_ed25519")
	assert.Contains(t, result, "    IdentitiesOnly yes")

	// Verify proper indentation (4 spaces)
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "Host") && !strings.HasPrefix(line, "#") {
			continue
		}
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "Host") {
			assert.True(t, strings.HasPrefix(line, "    "), "Line should have 4-space indent: %q", line)
		}
	}
}
