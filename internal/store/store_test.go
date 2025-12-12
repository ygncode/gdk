package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ygncode/gdk/internal/keygen"
)

func TestStore_Save(t *testing.T) {
	t.Run("creates files with correct permissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := New(tmpDir)

		gen := keygen.NewEd25519Generator()
		keyPair, err := gen.Generate("test")
		require.NoError(t, err)

		key := &DeployKey{
			RepoOwner: "user",
			RepoName:  "repo",
			HostAlias: "github.com-user-repo",
		}

		err = store.Save(key, keyPair)
		require.NoError(t, err)

		// Check private key permissions (0600)
		privKeyPath := filepath.Join(tmpDir, "user-repo", "id_ed25519")
		info, err := os.Stat(privKeyPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

		// Check public key permissions (0644)
		pubKeyPath := filepath.Join(tmpDir, "user-repo", "id_ed25519.pub")
		info, err = os.Stat(pubKeyPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
	})

	t.Run("creates directory structure", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := New(tmpDir)

		gen := keygen.NewEd25519Generator()
		keyPair, _ := gen.Generate("test")

		key := &DeployKey{
			RepoOwner: "myorg",
			RepoName:  "myrepo",
			HostAlias: "github.com-myorg-myrepo",
		}

		err := store.Save(key, keyPair)
		require.NoError(t, err)

		// Check directory exists
		keyDir := filepath.Join(tmpDir, "myorg-myrepo")
		info, err := os.Stat(keyDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
	})

	t.Run("updates paths in key struct", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := New(tmpDir)

		gen := keygen.NewEd25519Generator()
		keyPair, _ := gen.Generate("test")

		key := &DeployKey{
			RepoOwner: "user",
			RepoName:  "repo",
			HostAlias: "github.com-user-repo",
		}

		err := store.Save(key, keyPair)
		require.NoError(t, err)

		assert.Equal(t, filepath.Join(tmpDir, "user-repo", "id_ed25519"), key.PrivateKeyPath)
		assert.Equal(t, filepath.Join(tmpDir, "user-repo", "id_ed25519.pub"), key.PublicKeyPath)
		assert.False(t, key.CreatedAt.IsZero())
	})
}

func TestStore_Load(t *testing.T) {
	t.Run("returns saved key metadata", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := New(tmpDir)

		gen := keygen.NewEd25519Generator()
		keyPair, _ := gen.Generate("test")

		key := &DeployKey{
			RepoOwner: "user",
			RepoName:  "repo",
			HostAlias: "github.com-user-repo",
		}

		err := store.Save(key, keyPair)
		require.NoError(t, err)

		// Load and verify
		loaded, err := store.Load("user", "repo")
		require.NoError(t, err)
		assert.Equal(t, "user", loaded.RepoOwner)
		assert.Equal(t, "repo", loaded.RepoName)
		assert.Equal(t, "github.com-user-repo", loaded.HostAlias)
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := New(tmpDir)

		_, err := store.Load("nonexistent", "repo")
		assert.Error(t, err)
	})
}

func TestStore_Delete(t *testing.T) {
	t.Run("removes key files", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := New(tmpDir)

		gen := keygen.NewEd25519Generator()
		keyPair, _ := gen.Generate("test")

		key := &DeployKey{
			RepoOwner: "user",
			RepoName:  "repo",
			HostAlias: "github.com-user-repo",
		}

		err := store.Save(key, keyPair)
		require.NoError(t, err)

		// Delete
		err = store.Delete("user", "repo")
		require.NoError(t, err)

		// Verify files are gone
		keyDir := filepath.Join(tmpDir, "user-repo")
		_, err = os.Stat(keyDir)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := New(tmpDir)

		err := store.Delete("nonexistent", "repo")
		assert.Error(t, err)
	})
}

func TestStore_List(t *testing.T) {
	t.Run("returns all keys", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := New(tmpDir)

		gen := keygen.NewEd25519Generator()

		// Create two keys
		keyPair1, _ := gen.Generate("test1")
		key1 := &DeployKey{
			RepoOwner: "user1",
			RepoName:  "repo1",
			HostAlias: "github.com-user1-repo1",
		}
		store.Save(key1, keyPair1)

		keyPair2, _ := gen.Generate("test2")
		key2 := &DeployKey{
			RepoOwner: "user2",
			RepoName:  "repo2",
			HostAlias: "github.com-user2-repo2",
		}
		store.Save(key2, keyPair2)

		// List
		keys, err := store.List()
		require.NoError(t, err)
		assert.Len(t, keys, 2)
	})

	t.Run("returns empty list when no keys", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := New(tmpDir)

		keys, err := store.List()
		require.NoError(t, err)
		assert.Empty(t, keys)
	})
}

func TestStore_Exists(t *testing.T) {
	t.Run("returns true for existing key", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := New(tmpDir)

		gen := keygen.NewEd25519Generator()
		keyPair, _ := gen.Generate("test")

		key := &DeployKey{
			RepoOwner: "user",
			RepoName:  "repo",
			HostAlias: "github.com-user-repo",
		}

		store.Save(key, keyPair)

		assert.True(t, store.Exists("user", "repo"))
	})

	t.Run("returns false for non-existent key", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := New(tmpDir)

		assert.False(t, store.Exists("nonexistent", "repo"))
	})
}
