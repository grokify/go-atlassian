package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/grokify/go-atlassian/jira"
	"github.com/spf13/cobra"
)

var patchCmd = &cobra.Command{
	Use:   "patch ISSUE-KEY [flags]",
	Short: "Update issue fields",
	Long: `Update one or more fields on a Jira issue.

Examples:
  # Set a simple field
  gojira patch ISSUE-123 --set summary="New summary"

  # Set a custom field by ID
  gojira patch ISSUE-123 --set customfield_10001="value"

  # Set multiple fields
  gojira patch ISSUE-123 --set summary="New title" --set priority=High

  # Add a label
  gojira patch ISSUE-123 --add-label bug

  # Remove a label
  gojira patch ISSUE-123 --remove-label obsolete

  # Use JSON for complex field values
  gojira patch ISSUE-123 --json '{"fields":{"summary":"New title"}}'

  # Dry run (show what would be sent)
  gojira patch ISSUE-123 --set summary="Test" --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runPatch,
}

var (
	patchSetFlags      []string
	patchAddLabels     []string
	patchRemoveLabels  []string
	patchJSONBody      string
	patchDryRun        bool
	patchExpandChanges bool
)

func init() {
	rootCmd.AddCommand(patchCmd)

	patchCmd.Flags().StringArrayVar(&patchSetFlags, "set", nil, "Set field value (format: field=value)")
	patchCmd.Flags().StringArrayVar(&patchAddLabels, "add-label", nil, "Add label to issue")
	patchCmd.Flags().StringArrayVar(&patchRemoveLabels, "remove-label", nil, "Remove label from issue")
	patchCmd.Flags().StringVar(&patchJSONBody, "json", "", "Raw JSON body for complex updates")
	patchCmd.Flags().BoolVar(&patchDryRun, "dry-run", false, "Show request body without executing")
	patchCmd.Flags().BoolVar(&patchExpandChanges, "show-after", false, "Show issue after update")
}

func runPatch(cmd *cobra.Command, args []string) error {
	issueKey := args[0]

	// Build request body
	reqBody, err := buildPatchRequestBody()
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	// Show request body for dry run
	if patchDryRun {
		jsonBytes, err := json.MarshalIndent(reqBody, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		fmt.Printf("Would PATCH %s with:\n%s\n", issueKey, string(jsonBytes))
		return nil
	}

	// Get client
	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Execute patch
	resp, err := client.IssueAPI.IssuePatch(context.Background(), issueKey, reqBody)
	if err != nil {
		return fmt.Errorf("patch failed: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if !flagQuiet {
			fmt.Printf("Successfully updated %s (status: %d)\n", issueKey, resp.StatusCode)
		}
	} else {
		return fmt.Errorf("patch returned status %d", resp.StatusCode)
	}

	// Optionally show the updated issue
	if patchExpandChanges {
		issue, err := client.IssueAPI.Issue(context.Background(), issueKey, nil)
		if err != nil {
			return fmt.Errorf("failed to fetch updated issue: %w", err)
		}
		cfg := NewOutputConfig(getOutputFormat())
		return WriteIssues(jira.Issues{*issue}, cfg)
	}

	return nil
}

func buildPatchRequestBody() (jira.IssuePatchRequestBody, error) {
	var reqBody jira.IssuePatchRequestBody

	// If raw JSON provided, use it directly
	if patchJSONBody != "" {
		if err := json.Unmarshal([]byte(patchJSONBody), &reqBody); err != nil {
			return reqBody, fmt.Errorf("invalid JSON: %w", err)
		}
		return reqBody, nil
	}

	// Build from flags
	if len(patchSetFlags) > 0 {
		reqBody.Fields = make(map[string]jira.IssuePatchRequestBodyField)
		for _, setFlag := range patchSetFlags {
			parts := strings.SplitN(setFlag, "=", 2)
			if len(parts) != 2 {
				return reqBody, fmt.Errorf("invalid --set format: %q (expected field=value)", setFlag)
			}
			fieldName := strings.TrimSpace(parts[0])
			fieldValue := strings.TrimSpace(parts[1])

			reqBody.Fields[fieldName] = jira.IssuePatchRequestBodyField{
				Value: fieldValue,
			}
		}
	}

	// Handle label updates
	if len(patchAddLabels) > 0 || len(patchRemoveLabels) > 0 {
		if reqBody.Update == nil {
			reqBody.Update = &jira.IssuePatchRequestBodyUpdate{}
		}
		for _, label := range patchAddLabels {
			reqBody.Update.Labels = append(reqBody.Update.Labels, jira.IssuePatchRequestBodyUpdateLabel{
				Add: &label,
			})
		}
		for _, label := range patchRemoveLabels {
			reqBody.Update.Labels = append(reqBody.Update.Labels, jira.IssuePatchRequestBodyUpdateLabel{
				Remove: &label,
			})
		}
	}

	// Validate we have something to update
	if reqBody.Fields == nil && reqBody.Update == nil {
		fmt.Fprintln(os.Stderr, "No updates specified. Use --set, --add-label, --remove-label, or --json")
		os.Exit(1)
	}

	return reqBody, nil
}
