package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHomeDir(t *testing.T) {
	home, err := HomeDir()
	require.NoError(t, err)
	assert.NotEmpty(t, home)

	// Should return same value as os.UserHomeDir
	osHome, _ := os.UserHomeDir()
	assert.Equal(t, osHome, home)
}

func TestSSHDir(t *testing.T) {
	sshDir, err := SSHDir()
	require.NoError(t, err)

	home, _ := HomeDir()
	expected := filepath.Join(home, ".ssh")
	assert.Equal(t, expected, sshDir)
}

func TestSSHConfigPath(t *testing.T) {
	configPath, err := SSHConfigPath()
	require.NoError(t, err)

	home, _ := HomeDir()
	expected := filepath.Join(home, ".ssh", "config")
	assert.Equal(t, expected, configPath)
}

func TestDeployKeysDir(t *testing.T) {
	keysDir, err := DeployKeysDir()
	require.NoError(t, err)

	home, _ := HomeDir()
	expected := filepath.Join(home, ".ssh", "deploy-keys")
	assert.Equal(t, expected, keysDir)
}

func TestExpandTilde(t *testing.T) {
	home, _ := HomeDir()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"tilde only", "~", home, false},
		{"tilde with path", "~/.ssh/config", filepath.Join(home, ".ssh", "config"), false},
		{"no tilde", "/etc/ssh/config", "/etc/ssh/config", false},
		{"empty", "", "", false},
		{"relative path", "foo/bar", "foo/bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandTilde(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
