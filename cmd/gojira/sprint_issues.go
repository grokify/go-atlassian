package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/grokify/go-atlassian/jira"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var sprintIssuesCmd = &cobra.Command{
	Use:   "sprint-issues <sprint-id>",
	Short: "List issues in a sprint",
	Long: `List all issues in a Jira sprint.

Use 'gojira sprints <board-id>' to find sprint IDs.

Examples:
  # List all issues in sprint
  gojira sprint-issues 456`,
	Args: cobra.ExactArgs(1),
	RunE: runSprintIssues,
}

func init() {
	rootCmd.AddCommand(sprintIssuesCmd)
}

func runSprintIssues(cmd *cobra.Command, args []string) error {
	sprintID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid sprint ID %q: %w", args[0], err)
	}

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	issues, err := client.BoardAPI.GetSprintIssues(ctx, sprintID)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeSprintIssuesTable(issues)
	case OutputTOON:
		return writeSprintIssuesToon(issues)
	default:
		return writeSprintIssuesJSON(issues)
	}
}

func writeSprintIssuesJSON(issues jira.Issues) error {
	// Convert to simplified output format
	outputs := make([]sprintIssueOutput, 0, len(issues))
	for _, iss := range issues {
		outputs = append(outputs, sprintIssueOutput{
			Key:       iss.Key,
			IssueType: iss.Fields.Type.Name,
			Status:    iss.Fields.Status.Name,
			Summary:   iss.Fields.Summary,
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(outputs)
}

func writeSprintIssuesTable(issues jira.Issues) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.Header([]string{"Key", "Type", "Status", "Summary"})

	var rows [][]string
	for _, iss := range issues {
		rows = append(rows, []string{
			iss.Key,
			iss.Fields.Type.Name,
			iss.Fields.Status.Name,
			truncateString(iss.Fields.Summary, 50),
		})
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

func writeSprintIssuesToon(issues jira.Issues) error {
	for _, iss := range issues {
		fmt.Printf("K:%s|T:%s|S:%s|Sum:%s\n", iss.Key, iss.Fields.Type.Name, iss.Fields.Status.Name, iss.Fields.Summary)
	}
	return nil
}

type sprintIssueOutput struct {
	Key       string `json:"key"`
	IssueType string `json:"type"`
	Status    string `json:"status"`
	Summary   string `json:"summary"`
}
