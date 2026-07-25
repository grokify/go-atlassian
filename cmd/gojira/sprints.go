package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/go-atlassian/jira"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	sprintsState string
)

var sprintsCmd = &cobra.Command{
	Use:   "sprints <board-id>",
	Short: "List sprints for a Jira agile board",
	Long: `List sprints for a Jira agile board.

Use 'gojira boards' to find board IDs.

Examples:
  # List all sprints for a board
  gojira sprints 123

  # Filter by state (active, closed, future)
  gojira sprints 123 --state active`,
	Args: cobra.ExactArgs(1),
	RunE: runSprints,
}

func init() {
	rootCmd.AddCommand(sprintsCmd)

	sprintsCmd.Flags().StringVarP(&sprintsState, "state", "s", "", "Filter by sprint state (active, closed, future)")
}

func runSprints(cmd *cobra.Command, args []string) error {
	boardID, err := jira.ParseBoardID(args[0])
	if err != nil {
		return err
	}

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	opts := &gojira.GetAllSprintsOptions{}
	if sprintsState != "" {
		opts.State = sprintsState
	}

	sprints, err := client.BoardAPI.GetSprints(ctx, boardID, opts)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeSprintsTable(sprints)
	case OutputTOON:
		return writeSprintsToon(sprints)
	default:
		return writeSprintsJSON(sprints)
	}
}

func writeSprintsJSON(sprints []jira.SprintOutput) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(sprints)
}

func writeSprintsTable(sprints []jira.SprintOutput) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.Header([]string{"ID", "Name", "State", "Start", "End"})

	var rows [][]string
	for _, s := range sprints {
		rows = append(rows, []string{
			fmt.Sprintf("%d", s.ID),
			s.Name,
			s.State,
			s.StartDate,
			s.EndDate,
		})
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

func writeSprintsToon(sprints []jira.SprintOutput) error {
	for _, s := range sprints {
		fmt.Printf("ID:%d|N:%s|St:%s|Start:%s|End:%s\n", s.ID, s.Name, s.State, s.StartDate, s.EndDate)
	}
	return nil
}
