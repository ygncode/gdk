package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ygncode/gdk/internal/config"
	"github.com/ygncode/gdk/internal/keygen"
	"github.com/ygncode/gdk/pkg/utils"
)

var (
	// ErrKeyNotFound indicates the deploy key was not found
	ErrKeyNotFound = errors.New("deploy key not found")

	// ErrKeyExists indicates the deploy key already exists
	ErrKeyExists = errors.New("deploy key already exists")
)

// Store manages deploy key storage
type Store struct {
	baseDir string
}

// New creates a new Store
func New(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

// keyDir returns the directory path for a key
func (s *Store) keyDir(owner, repo string) string {
	return filepath.Join(s.baseDir, utils.SanitizeForPath(owner, repo))
}

// metadataPath returns the path to the metadata file
func (s *Store) metadataPath() string {
	return filepath.Join(s.baseDir, config.MetadataFileName)
}

// Save saves a deploy key to disk
func (s *Store) Save(key *DeployKey, keyPair *keygen.KeyPair) error {
	keyDir := s.keyDir(key.RepoOwner, key.RepoName)

	// Create directory with secure permissions
	if err := os.MkdirAll(keyDir, config.DirPerm); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Set paths
	key.PrivateKeyPath = filepath.Join(keyDir, config.PrivateKeyName)
	key.PublicKeyPath = filepath.Join(keyDir, config.PublicKeyName)
	key.CreatedAt = time.Now()

	// Write private key
	if err := os.WriteFile(key.PrivateKeyPath, keyPair.PrivateKey, config.PrivateKeyPerm); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Write public key (with the string version including comment)
	publicKeyContent := keyPair.PublicKeyStr + "\n"
	if err := os.WriteFile(key.PublicKeyPath, []byte(publicKeyContent), config.PublicKeyPerm); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	// Update metadata
	return s.updateMetadata(func(m *Metadata) {
		// Remove existing entry if any
		var newKeys []DeployKey
		for _, k := range m.Keys {
			if k.RepoOwner != key.RepoOwner || k.RepoName != key.RepoName {
				newKeys = append(newKeys, k)
			}
		}
		m.Keys = append(newKeys, *key)
		m.LastUpdate = time.Now()
	})
}

// Load loads a deploy key's metadata
func (s *Store) Load(owner, repo string) (*DeployKey, error) {
	metadata, err := s.loadMetadata()
	if err != nil {
		return nil, err
	}

	for _, key := range metadata.Keys {
		if key.RepoOwner == owner && key.RepoName == repo {
			return &key, nil
		}
	}

	return nil, fmt.Errorf("%w: %s/%s", ErrKeyNotFound, owner, repo)
}

// Delete removes a deploy key
func (s *Store) Delete(owner, repo string) error {
	// Check if exists first
	if !s.Exists(owner, repo) {
		return fmt.Errorf("%w: %s/%s", ErrKeyNotFound, owner, repo)
	}

	// Remove directory
	keyDir := s.keyDir(owner, repo)
	if err := os.RemoveAll(keyDir); err != nil {
		return fmt.Errorf("failed to delete key directory: %w", err)
	}

	// Update metadata
	return s.updateMetadata(func(m *Metadata) {
		var newKeys []DeployKey
		for _, k := range m.Keys {
			if k.RepoOwner != owner || k.RepoName != repo {
				newKeys = append(newKeys, k)
			}
		}
		m.Keys = newKeys
		m.LastUpdate = time.Now()
	})
}

// List returns all managed deploy keys
func (s *Store) List() ([]DeployKey, error) {
	metadata, err := s.loadMetadata()
	if err != nil {
		// If no metadata file, return empty list
		if os.IsNotExist(err) {
			return []DeployKey{}, nil
		}
		return nil, err
	}

	return metadata.Keys, nil
}

// Exists checks if a deploy key exists
func (s *Store) Exists(owner, repo string) bool {
	keyDir := s.keyDir(owner, repo)
	privKeyPath := filepath.Join(keyDir, config.PrivateKeyName)
	_, err := os.Stat(privKeyPath)
	return err == nil
}

// loadMetadata loads the metadata file
func (s *Store) loadMetadata() (*Metadata, error) {
	data, err := os.ReadFile(s.metadataPath())
	if err != nil {
		return nil, err
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &metadata, nil
}

// updateMetadata updates the metadata file with the given function
func (s *Store) updateMetadata(fn func(*Metadata)) error {
	// Ensure base directory exists
	if err := os.MkdirAll(s.baseDir, config.DirPerm); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Load existing or create new
	metadata, err := s.loadMetadata()
	if err != nil {
		if os.IsNotExist(err) {
			metadata = NewMetadata()
		} else {
			return err
		}
	}

	// Apply update
	fn(metadata)

	// Write back
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(s.metadataPath(), data, config.MetadataPerm); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}
