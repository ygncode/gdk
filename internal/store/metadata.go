package store

import "time"

// DeployKey represents a managed deploy key
type DeployKey struct {
	RepoOwner      string    `json:"repo_owner"`
	RepoName       string    `json:"repo_name"`
	HostAlias      string    `json:"host_alias"`
	PrivateKeyPath string    `json:"private_key_path"`
	PublicKeyPath  string    `json:"public_key_path"`
	CreatedAt      time.Time `json:"created_at"`
}

// Metadata stores all managed deploy keys
type Metadata struct {
	Version    string      `json:"version"`
	Keys       []DeployKey `json:"keys"`
	LastUpdate time.Time   `json:"last_update"`
}

// NewMetadata creates a new empty metadata structure
func NewMetadata() *Metadata {
	return &Metadata{
		Version: "1",
		Keys:    []DeployKey{},
	}
}
