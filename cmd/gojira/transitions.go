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

var transitionsCmd = &cobra.Command{
	Use:   "transitions <issue-key>",
	Short: "Show available workflow transitions for an issue",
	Long: `Show available workflow transitions for a Jira issue.

Transitions represent the possible state changes for an issue based on
its current status and the project's workflow configuration.

Examples:
  # Show transitions for an issue
  gojira transitions ISSUE-123

  # Output as JSON
  gojira transitions ISSUE-123 --json

  # Output as table (default)
  gojira transitions ISSUE-123 --table`,
	Args: cobra.ExactArgs(1),
	RunE: runTransitions,
}

func init() {
	rootCmd.AddCommand(transitionsCmd)
}

// TransitionOutput is a simplified view of a transition for output.
type TransitionOutput struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ToStatus string `json:"to_status"`
	ToID     string `json:"to_id"`
}

func runTransitions(cmd *cobra.Command, args []string) error {
	issueKey := args[0]

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()
	transitions, _, err := client.IssueAPI.GetTransitions(ctx, issueKey, false)
	if err != nil {
		return fmt.Errorf("failed to get transitions for %q: %w", issueKey, err)
	}

	if len(transitions) == 0 {
		if !flagQuiet {
			fmt.Fprintln(os.Stderr, "No transitions available")
		}
		return nil
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeTransitionsTable(transitions)
	case OutputTOON:
		return writeTransitionsToon(transitions)
	default:
		return writeTransitionsJSON(transitions)
	}
}

func transitionsToOutputs(transitions jira.Transitions) []TransitionOutput {
	outputs := make([]TransitionOutput, len(transitions))
	for i, t := range transitions {
		outputs[i] = TransitionOutput{
			ID:       t.ID,
			Name:     t.Name,
			ToStatus: t.To.Name,
			ToID:     t.To.ID,
		}
	}
	return outputs
}

func writeTransitionsJSON(transitions jira.Transitions) error {
	outputs := transitionsToOutputs(transitions)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(outputs)
}

func writeTransitionsTable(transitions jira.Transitions) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.Header([]string{"ID", "Name", "To Status"})

	var rows [][]string
	for _, t := range transitions {
		rows = append(rows, []string{t.ID, t.Name, t.To.Name})
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

func writeTransitionsToon(transitions jira.Transitions) error {
	for _, t := range transitions {
		fmt.Printf("ID:%s|N:%s|To:%s\n", t.ID, t.Name, t.To.Name)
	}
	return nil
}
