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
	transitionComment string
)

var transitionCmd = &cobra.Command{
	Use:   "transition <issue-key> <transition-name>",
	Short: "Transition a Jira issue to a new status",
	Long: `Transition a Jira issue to a new status.

The transition name must match an available transition for the issue's
current status. Use 'gojira transitions <issue-key>' to see available
transitions.

Examples:
  # Transition an issue to "In Progress"
  gojira transition PROJ-123 "In Progress"

  # Transition with a comment
  gojira transition PROJ-123 "Done" --comment "Fixed in latest build"

  # View available transitions first
  gojira transitions PROJ-123`,
	Args: cobra.ExactArgs(2),
	RunE: runTransition,
}

func init() {
	rootCmd.AddCommand(transitionCmd)

	transitionCmd.Flags().StringVarP(&transitionComment, "comment", "c", "", "Comment to add with the transition")
}

func runTransition(cmd *cobra.Command, args []string) error {
	issueKey := strings.TrimSpace(args[0])
	transitionName := strings.TrimSpace(args[1])

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	var opts *jira.TransitionOptions
	if transitionComment != "" {
		opts = &jira.TransitionOptions{
			Comment: transitionComment,
		}
	}

	result, err := client.IssueAPI.TransitionIssue(ctx, issueKey, transitionName, opts)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeTransitionResultTable(result)
	case OutputTOON:
		return writeTransitionResultToon(result)
	default:
		return writeTransitionResultJSON(result)
	}
}

func writeTransitionResultJSON(result *jira.TransitionResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func writeTransitionResultTable(result *jira.TransitionResult) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.Header([]string{"Key", "From", "Transition", "To"})

	rows := [][]string{
		{result.Key, result.FromStatus, result.Transition, result.ToStatus},
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

func writeTransitionResultToon(result *jira.TransitionResult) error {
	fmt.Printf("K:%s|From:%s|Txn:%s|To:%s\n", result.Key, result.FromStatus, result.Transition, result.ToStatus)
	return nil
}
