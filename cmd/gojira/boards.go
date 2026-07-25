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
	boardsProject string
	boardsType    string
)

var boardsCmd = &cobra.Command{
	Use:   "boards",
	Short: "List Jira agile boards",
	Long: `List Jira agile boards with optional filtering.

Examples:
  # List all boards
  gojira boards

  # Filter by project
  gojira boards --project PROJ

  # Filter by type (scrum or kanban)
  gojira boards --type scrum`,
	RunE: runBoards,
}

func init() {
	rootCmd.AddCommand(boardsCmd)

	boardsCmd.Flags().StringVarP(&boardsProject, "project", "p", "", "Filter boards by project key")
	boardsCmd.Flags().StringVarP(&boardsType, "type", "t", "", "Filter boards by type (scrum, kanban)")
}

func runBoards(cmd *cobra.Command, args []string) error {
	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	opts := &gojira.BoardListOptions{}
	if boardsProject != "" {
		opts.ProjectKeyOrID = boardsProject
	}
	if boardsType != "" {
		opts.BoardType = boardsType
	}

	boards, err := client.BoardAPI.GetBoards(ctx, opts)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeBoardsTable(boards)
	case OutputTOON:
		return writeBoardsToon(boards)
	default:
		return writeBoardsJSON(boards)
	}
}

func writeBoardsJSON(boards []jira.BoardOutput) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(boards)
}

func writeBoardsTable(boards []jira.BoardOutput) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.Header([]string{"ID", "Name", "Type"})

	var rows [][]string
	for _, b := range boards {
		rows = append(rows, []string{
			fmt.Sprintf("%d", b.ID),
			b.Name,
			b.Type,
		})
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

func writeBoardsToon(boards []jira.BoardOutput) error {
	for _, b := range boards {
		fmt.Printf("ID:%d|N:%s|T:%s\n", b.ID, b.Name, b.Type)
	}
	return nil
}
