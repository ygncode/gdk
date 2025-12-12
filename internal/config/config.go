package config

import (
	"github.com/ygncode/gdk/pkg/utils"
)

const (
	// MetadataFileName is the name of the metadata file
	MetadataFileName = ".gdk-metadata.json"

	// PrivateKeyName is the name of the private key file
	PrivateKeyName = "id_ed25519"

	// PublicKeyName is the name of the public key file
	PublicKeyName = "id_ed25519.pub"

	// DirPerm is the permission for directories
	DirPerm = 0700

	// PrivateKeyPerm is the permission for private key files
	PrivateKeyPerm = 0600

	// PublicKeyPerm is the permission for public key files
	PublicKeyPerm = 0644

	// MetadataPerm is the permission for metadata files
	MetadataPerm = 0600
)

// SSHConfigPath returns the path to the SSH config file
func SSHConfigPath() (string, error) {
	return utils.SSHConfigPath()
}

// DeployKeysDir returns the path to the deploy keys directory
func DeployKeysDir() (string, error) {
	return utils.DeployKeysDir()
}
