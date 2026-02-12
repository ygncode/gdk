package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRepoRef(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		// Valid inputs
		{"valid simple", "user/repo", "user", "repo", false},
		{"valid with dashes", "my-org/my-repo", "my-org", "my-repo", false},
		{"valid with underscores", "user_name/repo_name", "user_name", "repo_name", false},
		{"valid with numbers", "user123/repo456", "user123", "repo456", false},
		{"valid mixed", "My-Org_123/Some-Repo_456", "My-Org_123", "Some-Repo_456", false},
		{"valid with dots", "setkyar/setkyar.com", "setkyar", "setkyar.com", false},
		{"valid dot in owner", "my.org/repo", "my.org", "repo", false},

		// Invalid inputs
		{"invalid no slash", "userrepo", "", "", true},
		{"invalid multiple slashes", "user/repo/extra", "", "", true},
		{"invalid empty owner", "/repo", "", "", true},
		{"invalid empty repo", "user/", "", "", true},
		{"invalid empty", "", "", "", true},
		{"invalid with spaces", "user /repo", "", "", true},
		{"invalid spaces in owner", "user name/repo", "", "", true},
		{"invalid spaces in repo", "user/repo name", "", "", true},
		{"invalid github URL", "https://github.com/user/repo", "", "", true},
		{"invalid git URL", "git@github.com:user/repo.git", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepoRef(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, owner)
				assert.Empty(t, repo)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantOwner, owner)
				assert.Equal(t, tt.wantRepo, repo)
			}
		})
	}
}

func TestSanitizeForPath(t *testing.T) {
	tests := []struct {
		name  string
		owner string
		repo  string
		want  string
	}{
		{"simple", "user", "repo", "user-repo"},
		{"with dashes", "my-org", "my-repo", "my-org-my-repo"},
		{"with underscores", "user_name", "repo_name", "user_name-repo_name"},
		{"with dots", "setkyar", "setkyar.com", "setkyar-setkyar.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeForPath(tt.owner, tt.repo)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildHostAlias(t *testing.T) {
	tests := []struct {
		name  string
		owner string
		repo  string
		want  string
	}{
		{"simple", "user", "repo", "github.com-user-repo"},
		{"with dashes", "my-org", "my-repo", "github.com-my-org-my-repo"},
		{"with dots", "setkyar", "setkyar.com", "github.com-setkyar-setkyar.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildHostAlias(tt.owner, tt.repo)
			assert.Equal(t, tt.want, got)
		})
	}
}
