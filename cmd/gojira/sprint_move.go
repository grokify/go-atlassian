package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/grokify/go-atlassian/jira"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	sprintMoveJQL string
)

var sprintMoveCmd = &cobra.Command{
	Use:   "sprint-move <sprint-id> <issue-keys...>",
	Short: "Move issues to a sprint",
	Long: `Move issues to a sprint for sprint planning.

Use 'gojira sprints <board-id>' to find sprint IDs.

Examples:
  # Move specific issues to a sprint
  gojira sprint-move 456 PROJ-1 PROJ-2 PROJ-3

  # Move issues from JQL query to a sprint
  gojira sprint-move 456 --jql "project = PROJ AND labels = ready"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSprintMove,
}

func init() {
	rootCmd.AddCommand(sprintMoveCmd)

	sprintMoveCmd.Flags().StringVar(&sprintMoveJQL, "jql", "", "JQL query to select issues to move")
}

func runSprintMove(cmd *cobra.Command, args []string) error {
	sprintID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid sprint ID %q: %w", args[0], err)
	}

	issueKeys := args[1:]

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	// If JQL provided, get issue keys from search
	if sprintMoveJQL != "" {
		issues, err := client.IssueAPI.SearchIssuesAPIV3(ctx, sprintMoveJQL, false)
		if err != nil {
			return fmt.Errorf("JQL search failed: %w", err)
		}
		for _, iss := range issues {
			issueKeys = append(issueKeys, iss.Key)
		}
	}

	if len(issueKeys) == 0 {
		return fmt.Errorf("no issues to move (provide issue keys or use --jql)")
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

	result, err := client.BoardAPI.MoveIssuesToSprint(ctx, sprintID, uniqueKeys)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeSprintMoveTable(result)
	case OutputTOON:
		return writeSprintMoveToon(result)
	default:
		return writeSprintMoveJSON(result)
	}
}

func writeSprintMoveJSON(result *jira.MoveToSprintResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func writeSprintMoveTable(result *jira.MoveToSprintResult) error {
	fmt.Printf("Moved %d issues to sprint %d\n\n", result.IssueCount, result.SprintID)

	tw := tablewriter.NewWriter(os.Stdout)
	tw.Header([]string{"Issue"})

	var rows [][]string
	for _, key := range result.IssuesMoved {
		rows = append(rows, []string{key})
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

func writeSprintMoveToon(result *jira.MoveToSprintResult) error {
	fmt.Printf("Sprint:%d|Moved:%d\n", result.SprintID, result.IssueCount)
	for _, key := range result.IssuesMoved {
		fmt.Printf("K:%s\n", key)
	}
	return nil
}
