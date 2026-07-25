package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/grokify/go-atlassian/jira"
)

const (
	// Environment variable names for Jira authentication.
	EnvJiraURL   = "JIRA_URL"
	EnvJiraUser  = "JIRA_USER"
	EnvJiraToken = "JIRA_TOKEN"

	// Default goauth credentials file path (this is a path, not credentials).
	DefaultGoauthCredsFile = "~/.config/goauth/credentials.json" //nolint:gosec // This is a path constant, not credentials
)

// AuthOptions holds authentication configuration options.
type AuthOptions struct {
	CredsFile string // Path to goauth credentials file
	Account   string // Account key within the credentials file
}

// NewClientFromOptions creates a Jira client using the following priority:
// 1. CLI flags (credsFile, account)
// 2. Environment variables (JIRA_URL, JIRA_USER, JIRA_TOKEN)
// 3. Default goauth file (~/.config/goauth/credentials.json) with interactive selection
func NewClientFromOptions(opts *AuthOptions) (*jira.Client, error) {
	// 1. Check CLI flags for explicit credentials file
	if opts != nil && strings.TrimSpace(opts.CredsFile) != "" {
		credsFile := expandPath(opts.CredsFile)
		return jira.NewClientGoauthBasicAuthFile(credsFile, opts.Account, false)
	}

	// 2. Check environment variables
	url := strings.TrimSpace(os.Getenv(EnvJiraURL))
	user := strings.TrimSpace(os.Getenv(EnvJiraUser))
	token := strings.TrimSpace(os.Getenv(EnvJiraToken))

	if url != "" && user != "" && token != "" {
		return jira.NewClientFromBasicAuth(url, user, token, false)
	}

	// 3. Check if default goauth file exists
	defaultPath := expandPath(DefaultGoauthCredsFile)
	if _, err := os.Stat(defaultPath); err == nil {
		// File exists, try to use it
		if opts != nil && opts.Account != "" {
			return jira.NewClientGoauthBasicAuthFile(defaultPath, opts.Account, false)
		}
		// Fall through to interactive CLI selection
	}

	// 4. Fall back to goauth CLI (interactive selection)
	return jira.NewClientFromGoauthCLI(true, false)
}

// NewClientFromEnv creates a Jira client from environment variables only.
// Returns an error if required environment variables are not set.
func NewClientFromEnv() (*jira.Client, error) {
	url := strings.TrimSpace(os.Getenv(EnvJiraURL))
	user := strings.TrimSpace(os.Getenv(EnvJiraUser))
	token := strings.TrimSpace(os.Getenv(EnvJiraToken))

	if url == "" || user == "" || token == "" {
		return nil, errors.New("environment variables JIRA_URL, JIRA_USER, and JIRA_TOKEN must all be set")
	}

	return jira.NewClientFromBasicAuth(url, user, token, false)
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
