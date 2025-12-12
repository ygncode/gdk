package utils

import (
	"os"
	"path/filepath"
)

// HomeDir returns the user's home directory
func HomeDir() (string, error) {
	return os.UserHomeDir()
}

// SSHDir returns the .ssh directory path
func SSHDir() (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh"), nil
}

// SSHConfigPath returns the SSH config file path
func SSHConfigPath() (string, error) {
	sshDir, err := SSHDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(sshDir, "config"), nil
}

// DeployKeysDir returns the deploy keys directory
func DeployKeysDir() (string, error) {
	sshDir, err := SSHDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(sshDir, "deploy-keys"), nil
}

// ExpandTilde expands ~ to home directory
func ExpandTilde(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}
	if path[0] != '~' {
		return path, nil
	}
	home, err := HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[1:]), nil
}
