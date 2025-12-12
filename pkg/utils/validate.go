package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	// ErrInvalidRepoRef indicates the repository reference format is invalid
	ErrInvalidRepoRef = errors.New("invalid repository reference")

	// validRepoRefPattern matches owner/repo format
	// Allows alphanumeric, dashes, and underscores
	validRepoRefPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ParseRepoRef parses a repository reference in "owner/repo" format
// Returns owner, repo, and any error
func ParseRepoRef(ref string) (owner, repo string, err error) {
	if ref == "" {
		return "", "", fmt.Errorf("%w: empty reference", ErrInvalidRepoRef)
	}

	// Check for URL patterns (not supported)
	if strings.Contains(ref, "://") || strings.Contains(ref, "@") {
		return "", "", fmt.Errorf("%w: URLs are not supported, use owner/repo format", ErrInvalidRepoRef)
	}

	parts := strings.Split(ref, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("%w: expected format owner/repo, got %q", ErrInvalidRepoRef, ref)
	}

	owner = parts[0]
	repo = parts[1]

	// Check for spaces (don't allow them even at boundaries)
	if strings.ContainsAny(owner, " \t") || strings.ContainsAny(repo, " \t") {
		return "", "", fmt.Errorf("%w: spaces are not allowed", ErrInvalidRepoRef)
	}

	if owner == "" {
		return "", "", fmt.Errorf("%w: owner cannot be empty", ErrInvalidRepoRef)
	}
	if repo == "" {
		return "", "", fmt.Errorf("%w: repo cannot be empty", ErrInvalidRepoRef)
	}

	// Validate characters
	if !validRepoRefPattern.MatchString(owner) {
		return "", "", fmt.Errorf("%w: owner contains invalid characters", ErrInvalidRepoRef)
	}
	if !validRepoRefPattern.MatchString(repo) {
		return "", "", fmt.Errorf("%w: repo contains invalid characters", ErrInvalidRepoRef)
	}

	return owner, repo, nil
}

// SanitizeForPath creates a filesystem-safe directory name from owner and repo
func SanitizeForPath(owner, repo string) string {
	return fmt.Sprintf("%s-%s", owner, repo)
}

// BuildHostAlias creates an SSH host alias from owner and repo
func BuildHostAlias(owner, repo string) string {
	return fmt.Sprintf("github.com-%s-%s", owner, repo)
}
