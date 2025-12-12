package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ygncode/gdk/internal/config"
	"github.com/ygncode/gdk/internal/keygen"
	"github.com/ygncode/gdk/internal/sshconfig"
	"github.com/ygncode/gdk/internal/store"
	"github.com/ygncode/gdk/pkg/utils"
)

// TestFullWorkflow tests the complete add-list-remove workflow
func TestFullWorkflow(t *testing.T) {
	// Create isolated test environment
	tmpDir := t.TempDir()
	keysDir := filepath.Join(tmpDir, ".ssh", "deploy-keys")
	configPath := filepath.Join(tmpDir, ".ssh", "config")

	// Initialize components
	keyStore := store.New(keysDir)
	sshConfigMgr := sshconfig.New(configPath)
	keyGen := keygen.NewEd25519Generator()

	// Test Add
	t.Run("add deploy key", func(t *testing.T) {
		owner := "testuser"
		repo := "testrepo"
		repoRef := "testuser/testrepo"

		// Generate key
		comment := "gdk:" + repoRef
		keyPair, err := keyGen.Generate(comment)
		require.NoError(t, err)

		// Create deploy key
		hostAlias := utils.BuildHostAlias(owner, repo)
		deployKey := &store.DeployKey{
			RepoOwner: owner,
			RepoName:  repo,
			HostAlias: hostAlias,
		}

		// Save to store
		err = keyStore.Save(deployKey, keyPair)
		require.NoError(t, err)

		// Add to SSH config
		identityFile := filepath.Join("~/.ssh/deploy-keys", utils.SanitizeForPath(owner, repo), "id_ed25519")
		entry := sshconfig.HostEntry{
			Host:           hostAlias,
			HostName:       "github.com",
			User:           "git",
			IdentityFile:   identityFile,
			IdentitiesOnly: true,
		}
		err = sshConfigMgr.AddHost(entry, repoRef)
		require.NoError(t, err)

		// Verify files exist
		privKeyPath := filepath.Join(keysDir, "testuser-testrepo", "id_ed25519")
		_, err = os.Stat(privKeyPath)
		assert.NoError(t, err)

		pubKeyPath := filepath.Join(keysDir, "testuser-testrepo", "id_ed25519.pub")
		_, err = os.Stat(pubKeyPath)
		assert.NoError(t, err)

		// Verify SSH config
		configContent, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(configContent), "Host github.com-testuser-testrepo")
		assert.Contains(t, string(configContent), "# gdk:testuser/testrepo")
	})

	// Test List
	t.Run("list deploy keys", func(t *testing.T) {
		keys, err := keyStore.List()
		require.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.Equal(t, "testuser", keys[0].RepoOwner)
		assert.Equal(t, "testrepo", keys[0].RepoName)
	})

	// Test Remove
	t.Run("remove deploy key", func(t *testing.T) {
		owner := "testuser"
		repo := "testrepo"

		// Load key to get host alias
		deployKey, err := keyStore.Load(owner, repo)
		require.NoError(t, err)

		// Remove from SSH config
		err = sshConfigMgr.RemoveHost(deployKey.HostAlias)
		require.NoError(t, err)

		// Delete from store
		err = keyStore.Delete(owner, repo)
		require.NoError(t, err)

		// Verify files are gone
		privKeyPath := filepath.Join(keysDir, "testuser-testrepo", "id_ed25519")
		_, err = os.Stat(privKeyPath)
		assert.True(t, os.IsNotExist(err))

		// Verify SSH config updated
		configContent, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.NotContains(t, string(configContent), "github.com-testuser-testrepo")
	})

	// Verify list is empty after removal
	t.Run("list is empty after removal", func(t *testing.T) {
		keys, err := keyStore.List()
		require.NoError(t, err)
		assert.Empty(t, keys)
	})
}

// TestMultipleKeys tests managing multiple deploy keys
func TestMultipleKeys(t *testing.T) {
	tmpDir := t.TempDir()
	keysDir := filepath.Join(tmpDir, ".ssh", "deploy-keys")
	configPath := filepath.Join(tmpDir, ".ssh", "config")

	keyStore := store.New(keysDir)
	sshConfigMgr := sshconfig.New(configPath)
	keyGen := keygen.NewEd25519Generator()

	// Add multiple keys
	repos := []struct {
		owner string
		repo  string
	}{
		{"org1", "repo1"},
		{"org1", "repo2"},
		{"org2", "repo1"},
	}

	for _, r := range repos {
		repoRef := r.owner + "/" + r.repo
		keyPair, err := keyGen.Generate("gdk:" + repoRef)
		require.NoError(t, err)

		hostAlias := utils.BuildHostAlias(r.owner, r.repo)
		deployKey := &store.DeployKey{
			RepoOwner: r.owner,
			RepoName:  r.repo,
			HostAlias: hostAlias,
		}

		err = keyStore.Save(deployKey, keyPair)
		require.NoError(t, err)

		identityFile := filepath.Join("~/.ssh/deploy-keys", utils.SanitizeForPath(r.owner, r.repo), "id_ed25519")
		entry := sshconfig.HostEntry{
			Host:           hostAlias,
			HostName:       "github.com",
			User:           "git",
			IdentityFile:   identityFile,
			IdentitiesOnly: true,
		}
		err = sshConfigMgr.AddHost(entry, repoRef)
		require.NoError(t, err)
	}

	// Verify all keys exist
	keys, err := keyStore.List()
	require.NoError(t, err)
	assert.Len(t, keys, 3)

	// Verify SSH config has all entries
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(configContent), "github.com-org1-repo1")
	assert.Contains(t, string(configContent), "github.com-org1-repo2")
	assert.Contains(t, string(configContent), "github.com-org2-repo1")

	// Remove middle key
	deployKey, _ := keyStore.Load("org1", "repo2")
	sshConfigMgr.RemoveHost(deployKey.HostAlias)
	keyStore.Delete("org1", "repo2")

	// Verify remaining keys
	keys, err = keyStore.List()
	require.NoError(t, err)
	assert.Len(t, keys, 2)

	configContent, err = os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(configContent), "github.com-org1-repo1")
	assert.NotContains(t, string(configContent), "github.com-org1-repo2")
	assert.Contains(t, string(configContent), "github.com-org2-repo1")
}

// TestKeyUniqueness ensures each key is unique
func TestKeyUniqueness(t *testing.T) {
	keyGen := keygen.NewEd25519Generator()

	var publicKeys []string
	for i := 0; i < 5; i++ {
		keyPair, err := keyGen.Generate("test")
		require.NoError(t, err)

		// Ensure this public key is unique
		for _, existingKey := range publicKeys {
			assert.NotEqual(t, existingKey, keyPair.PublicKeyStr)
		}
		publicKeys = append(publicKeys, keyPair.PublicKeyStr)
	}
}

// TestSSHConfigPreservation ensures existing SSH config is preserved
func TestSSHConfigPreservation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".ssh", "config")

	// Ensure .ssh directory exists
	os.MkdirAll(filepath.Dir(configPath), 0700)

	// Write existing config
	existingConfig := `Host personal
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_rsa

Host work
    HostName gitlab.com
    User git
    IdentityFile ~/.ssh/id_work
`
	err := os.WriteFile(configPath, []byte(existingConfig), 0600)
	require.NoError(t, err)

	sshConfigMgr := sshconfig.New(configPath)

	// Add gdk managed host
	entry := sshconfig.HostEntry{
		Host:           "github.com-test-repo",
		HostName:       "github.com",
		User:           "git",
		IdentityFile:   "~/.ssh/deploy-keys/test-repo/id_ed25519",
		IdentitiesOnly: true,
	}
	err = sshConfigMgr.AddHost(entry, "test/repo")
	require.NoError(t, err)

	// Verify existing config is preserved
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Host personal")
	assert.Contains(t, string(content), "Host work")
	assert.Contains(t, string(content), "Host github.com-test-repo")

	// Remove gdk managed host
	err = sshConfigMgr.RemoveHost("github.com-test-repo")
	require.NoError(t, err)

	// Verify existing config is still preserved
	content, err = os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Host personal")
	assert.Contains(t, string(content), "Host work")
	assert.NotContains(t, string(content), "github.com-test-repo")
}

// TestKeyFormat verifies generated keys are in correct format
func TestKeyFormat(t *testing.T) {
	keyGen := keygen.NewEd25519Generator()

	keyPair, err := keyGen.Generate("test@example.com")
	require.NoError(t, err)

	// Verify private key format
	privKeyStr := string(keyPair.PrivateKey)
	assert.True(t, strings.HasPrefix(privKeyStr, "-----BEGIN OPENSSH PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(privKeyStr), "-----END OPENSSH PRIVATE KEY-----"))

	// Verify public key format
	assert.True(t, strings.HasPrefix(keyPair.PublicKeyStr, "ssh-ed25519 "))
	assert.True(t, strings.HasSuffix(keyPair.PublicKeyStr, "test@example.com"))
}

// TestFilePermissions verifies files are created with correct permissions
func TestFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	keysDir := filepath.Join(tmpDir, "deploy-keys")

	keyStore := store.New(keysDir)
	keyGen := keygen.NewEd25519Generator()

	keyPair, err := keyGen.Generate("test")
	require.NoError(t, err)

	deployKey := &store.DeployKey{
		RepoOwner: "user",
		RepoName:  "repo",
		HostAlias: "github.com-user-repo",
	}

	err = keyStore.Save(deployKey, keyPair)
	require.NoError(t, err)

	// Check directory permissions
	keyDir := filepath.Join(keysDir, "user-repo")
	info, err := os.Stat(keyDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(config.DirPerm), info.Mode().Perm())

	// Check private key permissions
	privKeyPath := filepath.Join(keyDir, "id_ed25519")
	info, err = os.Stat(privKeyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(config.PrivateKeyPerm), info.Mode().Perm())

	// Check public key permissions
	pubKeyPath := filepath.Join(keyDir, "id_ed25519.pub")
	info, err = os.Stat(pubKeyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(config.PublicKeyPerm), info.Mode().Perm())
}
