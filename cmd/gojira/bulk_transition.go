package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/grokify/go-atlassian/jira"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	bulkTransitionComment string
	bulkTransitionJQL     string
)

var bulkTransitionCmd = &cobra.Command{
	Use:   "bulk-transition <transition-name> <issue-keys...>",
	Short: "Transition multiple Jira issues at once",
	Long: `Transition multiple Jira issues to a new status.

You can provide issue keys as arguments or use --jql to select issues.

Examples:
  # Transition specific issues
  gojira bulk-transition "In Progress" PROJ-1 PROJ-2 PROJ-3

  # Transition with a comment
  gojira bulk-transition "Done" PROJ-1 PROJ-2 --comment "Batch close"

  # Transition issues from JQL query
  gojira bulk-transition "Done" --jql "project = PROJ AND status = 'In Review'"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runBulkTransition,
}

func init() {
	rootCmd.AddCommand(bulkTransitionCmd)

	bulkTransitionCmd.Flags().StringVarP(&bulkTransitionComment, "comment", "c", "", "Comment to add with each transition")
	bulkTransitionCmd.Flags().StringVar(&bulkTransitionJQL, "jql", "", "JQL query to select issues (alternative to listing keys)")
}

func runBulkTransition(cmd *cobra.Command, args []string) error {
	transitionName := strings.TrimSpace(args[0])
	issueKeys := args[1:]

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	// If JQL provided, get issue keys from search
	if bulkTransitionJQL != "" {
		issues, err := client.IssueAPI.SearchIssuesAPIV3(ctx, bulkTransitionJQL, false)
		if err != nil {
			return fmt.Errorf("JQL search failed: %w", err)
		}
		for _, iss := range issues {
			issueKeys = append(issueKeys, iss.Key)
		}
	}

	if len(issueKeys) == 0 {
		return fmt.Errorf("no issues to transition (provide issue keys or use --jql)")
	}

	// Deduplicate keys
	seen := make(map[string]bool)
	uniqueKeys := []string{}
	for _, key := range issueKeys {
		key = strings.TrimSpace(key)
		if key != "" && !seen[key] {
			seen[key] = true
			uniqueKeys = append(uniqueKeys, key)
		}
	}

	var opts *jira.TransitionOptions
	if bulkTransitionComment != "" {
		opts = &jira.TransitionOptions{
			Comment: bulkTransitionComment,
		}
	}

	result, err := client.IssueAPI.BulkTransitionIssues(ctx, uniqueKeys, transitionName, opts)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeBulkTransitionResultTable(result)
	case OutputTOON:
		return writeBulkTransitionResultToon(result)
	default:
		return writeBulkTransitionResultJSON(result)
	}
}

func writeBulkTransitionResultJSON(result *jira.BulkTransitionResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func writeBulkTransitionResultTable(result *jira.BulkTransitionResult) error {
	// Print summary
	fmt.Printf("Total: %d | Success: %d | Failed: %d\n\n", result.Total, result.SuccessCount, result.FailCount)

	if len(result.Successful) > 0 {
		fmt.Println("Successful:")
		tw := tablewriter.NewWriter(os.Stdout)
		tw.Header([]string{"Key", "From", "To"})

		var rows [][]string
		for _, r := range result.Successful {
			rows = append(rows, []string{r.Key, r.FromStatus, r.ToStatus})
		}

		if err := tw.Bulk(rows); err != nil {
			return err
		}
		if err := tw.Render(); err != nil {
			return err
		}
	}

	if len(result.Failed) > 0 {
		fmt.Println("\nFailed:")
		tw := tablewriter.NewWriter(os.Stdout)
		tw.Header([]string{"Key", "Error"})

		var rows [][]string
		for _, f := range result.Failed {
			rows = append(rows, []string{f.Key, truncateString(f.Error, 60)})
		}

		if err := tw.Bulk(rows); err != nil {
			return err
		}
		if err := tw.Render(); err != nil {
			return err
		}
	}

	return nil
}

func writeBulkTransitionResultToon(result *jira.BulkTransitionResult) error {
	fmt.Printf("Tot:%d|OK:%d|Fail:%d\n", result.Total, result.SuccessCount, result.FailCount)
	for _, r := range result.Successful {
		fmt.Printf("OK|K:%s|From:%s|To:%s\n", r.Key, r.FromStatus, r.ToStatus)
	}
	for _, f := range result.Failed {
		fmt.Printf("FAIL|K:%s|E:%s\n", f.Key, f.Error)
	}
	return nil
}
