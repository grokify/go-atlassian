package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	flagJSON      bool
	flagTable     bool
	flagTOON      bool
	flagCSV       bool
	flagMarkdown  bool
	flagCredsFile string
	flagAccount   string
	flagQuiet     bool
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gojira",
	Short: "A CLI for interacting with Jira",
	Long: `gojira is a command-line interface for the Jira REST API.

It is designed primarily for AI agents (with JSON output) but also
supports human-readable table output and token-optimized TOON format
for LLMs.

Authentication:
  1. CLI flags: --creds-file, --account
  2. Environment variables: JIRA_URL, JIRA_USER, JIRA_TOKEN
  3. Default goauth file: ~/.config/goauth/credentials.json

Examples:
  # Search with JQL
  gojira search --jql "project = FOO AND status = Open"

  # Get issue by key
  gojira get ISSUE-123

  # Human-readable table output
  gojira search --jql "assignee = currentUser()" --table

  # Token-optimized output for LLMs
  gojira search --jql "project = FOO" --toon`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Output format flags (mutually exclusive, JSON is default)
	rootCmd.PersistentFlags().BoolVarP(&flagJSON, "json", "j", false, "Output as JSON (default)")
	rootCmd.PersistentFlags().BoolVarP(&flagTable, "table", "t", false, "Output as human-readable table")
	rootCmd.PersistentFlags().BoolVar(&flagTOON, "toon", false, "Output as TOON (Token-Optimized Object Notation)")
	rootCmd.PersistentFlags().BoolVar(&flagCSV, "csv", false, "Output as CSV")
	rootCmd.PersistentFlags().BoolVar(&flagMarkdown, "markdown", false, "Output as Markdown table")

	// Authentication flags
	rootCmd.PersistentFlags().StringVar(&flagCredsFile, "creds-file", "", "Path to goauth credentials file")
	rootCmd.PersistentFlags().StringVar(&flagAccount, "account", "", "Account key in credentials file")

	// Other flags
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress non-essential output")

	// Mark output format flags as mutually exclusive
	rootCmd.MarkFlagsMutuallyExclusive("json", "table", "toon", "csv", "markdown")
}

// getOutputFormat returns the output format based on flags.
// JSON is the default if no format flag is specified.
func getOutputFormat() OutputFormat {
	if flagTable {
		return OutputTable
	}
	if flagTOON {
		return OutputTOON
	}
	if flagCSV {
		return OutputCSV
	}
	if flagMarkdown {
		return OutputMarkdown
	}
	return OutputJSON
}

// getAuthOptions returns authentication options from global flags.
func getAuthOptions() *AuthOptions {
	return &AuthOptions{
		CredsFile: flagCredsFile,
		Account:   flagAccount,
	}
}
