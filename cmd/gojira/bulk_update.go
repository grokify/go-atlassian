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
	bulkUpdateComment      string
	bulkUpdateJQL          string
	bulkUpdateAddLabels    []string
	bulkUpdateRemoveLabels []string
	bulkUpdateFields       []string
)

var bulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update <issue-keys...>",
	Short: "Update multiple Jira issues at once",
	Long: `Update multiple Jira issues with the same field changes.

You can provide issue keys as arguments or use --jql to select issues.

Examples:
  # Add a label to multiple issues
  gojira bulk-update PROJ-1 PROJ-2 PROJ-3 --add-label "reviewed"

  # Remove a label from issues
  gojira bulk-update PROJ-1 PROJ-2 --remove-label "stale"

  # Add a comment to multiple issues
  gojira bulk-update PROJ-1 PROJ-2 --comment "Batch update"

  # Set a custom field (field=value format)
  gojira bulk-update PROJ-1 --field "customfield_10016=5"

  # Update issues from JQL query
  gojira bulk-update --jql "project = PROJ AND labels = stale" --remove-label "stale"`,
	RunE: runBulkUpdate,
}

func init() {
	rootCmd.AddCommand(bulkUpdateCmd)

	bulkUpdateCmd.Flags().StringVarP(&bulkUpdateComment, "comment", "c", "", "Comment to add to each issue")
	bulkUpdateCmd.Flags().StringVar(&bulkUpdateJQL, "jql", "", "JQL query to select issues")
	bulkUpdateCmd.Flags().StringSliceVar(&bulkUpdateAddLabels, "add-label", nil, "Labels to add")
	bulkUpdateCmd.Flags().StringSliceVar(&bulkUpdateRemoveLabels, "remove-label", nil, "Labels to remove")
	bulkUpdateCmd.Flags().StringSliceVarP(&bulkUpdateFields, "field", "f", nil, "Field to set (format: field=value)")
}

func runBulkUpdate(cmd *cobra.Command, args []string) error {
	issueKeys := args

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	// If JQL provided, get issue keys from search
	if bulkUpdateJQL != "" {
		issues, err := client.IssueAPI.SearchIssuesAPIV3(ctx, bulkUpdateJQL, false)
		if err != nil {
			return fmt.Errorf("JQL search failed: %w", err)
		}
		for _, iss := range issues {
			issueKeys = append(issueKeys, iss.Key)
		}
	}

	if len(issueKeys) == 0 {
		return fmt.Errorf("no issues to update (provide issue keys or use --jql)")
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

	// Parse field updates
	fields := make(map[string]any)
	for _, f := range bulkUpdateFields {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid field format %q (expected field=value)", f)
		}
		fields[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	opts := &jira.BulkUpdateOptions{
		Fields:       fields,
		AddLabels:    bulkUpdateAddLabels,
		RemoveLabels: bulkUpdateRemoveLabels,
		Comment:      bulkUpdateComment,
	}

	// Validate that at least one update is requested
	if len(opts.Fields) == 0 && len(opts.AddLabels) == 0 && len(opts.RemoveLabels) == 0 && opts.Comment == "" {
		return fmt.Errorf("no updates specified (use --field, --add-label, --remove-label, or --comment)")
	}

	result, err := client.IssueAPI.BulkUpdateIssues(ctx, uniqueKeys, opts)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeBulkUpdateResultTable(result)
	case OutputTOON:
		return writeBulkUpdateResultToon(result)
	default:
		return writeBulkUpdateResultJSON(result)
	}
}

func writeBulkUpdateResultJSON(result *jira.BulkUpdateResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func writeBulkUpdateResultTable(result *jira.BulkUpdateResult) error {
	// Print summary
	fmt.Printf("Total: %d | Success: %d | Failed: %d\n\n", result.Total, result.SuccessCount, result.FailCount)

	if len(result.Successful) > 0 {
		fmt.Println("Successful:")
		tw := tablewriter.NewWriter(os.Stdout)
		tw.Header([]string{"Key"})

		var rows [][]string
		for _, r := range result.Successful {
			rows = append(rows, []string{r.Key})
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

func writeBulkUpdateResultToon(result *jira.BulkUpdateResult) error {
	fmt.Printf("Tot:%d|OK:%d|Fail:%d\n", result.Total, result.SuccessCount, result.FailCount)
	for _, r := range result.Successful {
		fmt.Printf("OK|K:%s\n", r.Key)
	}
	for _, f := range result.Failed {
		fmt.Printf("FAIL|K:%s|E:%s\n", f.Key, f.Error)
	}
	return nil
}
