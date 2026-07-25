package main

import (
	"context"
	"fmt"
	"os"

	"github.com/grokify/go-atlassian/jira"
	"github.com/spf13/cobra"
)

var (
	flagGetExpand bool
	flagGetFields string
	flagGetRaw    bool
)

var getCmd = &cobra.Command{
	Use:   "get <issue-key> [issue-key...]",
	Short: "Get one or more issues by key",
	Long: `Get retrieves one or more Jira issues by their keys.

Examples:
  # Get a single issue
  gojira get ISSUE-123

  # Get multiple issues
  gojira get ISSUE-123 ISSUE-456 ISSUE-789

  # Get with changelog expansion
  gojira get ISSUE-123 --expand

  # Output as table
  gojira get ISSUE-123 --table

  # Output full API JSON (all fields)
  gojira get ISSUE-123 --raw`,
	Args: cobra.MinimumNArgs(1),
	RunE: runGet,
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().BoolVar(&flagGetExpand, "expand", false, "Expand changelog and other fields")
	getCmd.Flags().StringVarP(&flagGetFields, "fields", "f", "", "Comma-separated list of fields to include")
	getCmd.Flags().BoolVar(&flagGetRaw, "raw", false, "Output full API JSON instead of simplified metadata")
}

func runGet(cmd *cobra.Command, args []string) error {
	// Create client
	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("failed to create Jira client: %w", err)
	}

	// Set up query options
	opts := &jira.GetQueryOptions{
		ExpandChangelog: flagGetExpand,
	}

	ctx := context.Background()

	// Fetch issues
	var issues jira.Issues
	if len(args) == 1 {
		// Single issue
		issue, err := client.IssueAPI.Issue(ctx, args[0], opts)
		if err != nil {
			return fmt.Errorf("failed to get issue %s: %w", args[0], err)
		}
		issues = jira.Issues{*issue}
	} else {
		// Multiple issues
		issues, err = client.IssueAPI.Issues(ctx, args, opts)
		if err != nil {
			return fmt.Errorf("failed to get issues: %w", err)
		}
	}

	if len(issues) == 0 {
		if !flagQuiet {
			fmt.Fprintln(os.Stderr, "No issues found")
		}
		return nil
	}

	// Output results
	if flagGetRaw {
		return WriteIssuesRaw(issues)
	}
	cfg := NewOutputConfig(getOutputFormat())
	return WriteIssues(issues, cfg)
}
