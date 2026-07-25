package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/grokify/go-atlassian/jira"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	issueTypesProject string
)

var issueTypesCmd = &cobra.Command{
	Use:   "issue-types [flags]",
	Short: "List available issue types for a project",
	Long: `List available issue types for a project using the createmeta API.

Examples:
  # List issue types for a project
  gojira issue-types --project ABC

  # Output as JSON
  gojira issue-types --project ABC --json

  # Output as table (default)
  gojira issue-types --project ABC --table`,
	RunE: runIssueTypes,
}

func init() {
	rootCmd.AddCommand(issueTypesCmd)

	issueTypesCmd.Flags().StringVarP(&issueTypesProject, "project", "p", "", "Project key (required)")
	_ = issueTypesCmd.MarkFlagRequired("project")
}

func runIssueTypes(cmd *cobra.Command, args []string) error {
	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()
	issueTypes, err := client.CreateMetaAPI.GetIssueTypes(ctx, issueTypesProject)
	if err != nil {
		return fmt.Errorf("failed to get issue types for project %q: %w", issueTypesProject, err)
	}

	if len(issueTypes) == 0 {
		if !flagQuiet {
			fmt.Fprintln(os.Stderr, "No issue types found")
		}
		return nil
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeIssueTypesTable(issueTypes)
	case OutputTOON:
		return writeIssueTypesToon(issueTypes)
	default:
		return writeIssueTypesJSON(issueTypes)
	}
}

func writeIssueTypesJSON(issueTypes []jira.CreateMetaIssueType) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(issueTypes)
}

func writeIssueTypesTable(issueTypes []jira.CreateMetaIssueType) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.Header([]string{"ID", "Name", "Description", "Subtask"})

	var rows [][]string
	for _, it := range issueTypes {
		subtask := "false"
		if it.Subtask {
			subtask = "true"
		}
		desc := truncateString(it.Description, 40)
		rows = append(rows, []string{it.ID, it.Name, desc, subtask})
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

func writeIssueTypesToon(issueTypes []jira.CreateMetaIssueType) error {
	for _, it := range issueTypes {
		subtask := "N"
		if it.Subtask {
			subtask = "Y"
		}
		fmt.Printf("ID:%s|N:%s|Sub:%s\n", it.ID, it.Name, subtask)
	}
	return nil
}
