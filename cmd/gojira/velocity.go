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
	velocitySprintCount   int
	velocityIncludeActive bool
	velocityStoryPtsField string
)

var velocityCmd = &cobra.Command{
	Use:   "velocity <board-id>",
	Short: "Show velocity report for a Jira agile board",
	Long: `Calculate and display velocity metrics for a Jira agile board.

Velocity is calculated based on story points completed in closed sprints.

Use 'gojira boards' to find board IDs.

Examples:
  # Show velocity for last 5 sprints
  gojira velocity 123 --sprints 5

  # Include active sprint
  gojira velocity 123 --include-active

  # Specify custom story points field
  gojira velocity 123 --story-points-field customfield_10016`,
	Args: cobra.ExactArgs(1),
	RunE: runVelocity,
}

func init() {
	rootCmd.AddCommand(velocityCmd)

	velocityCmd.Flags().IntVarP(&velocitySprintCount, "sprints", "s", 0, "Number of sprints to include (0 = all)")
	velocityCmd.Flags().BoolVar(&velocityIncludeActive, "include-active", false, "Include active sprint in calculations")
	velocityCmd.Flags().StringVar(&velocityStoryPtsField, "story-points-field", "", "Custom field ID for story points (default: customfield_10016)")
}

func runVelocity(cmd *cobra.Command, args []string) error {
	boardID, err := jira.ParseBoardID(args[0])
	if err != nil {
		return err
	}

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	opts := &jira.VelocityOptions{
		SprintCount:        velocitySprintCount,
		IncludeActive:      velocityIncludeActive,
		StoryPointsFieldID: velocityStoryPtsField,
	}

	report, err := client.BoardAPI.GetVelocityReport(ctx, boardID, opts)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeVelocityTable(report)
	case OutputTOON:
		return writeVelocityToon(report)
	case OutputCSV:
		return writeVelocityCSV(report)
	case OutputMarkdown:
		return writeVelocityMarkdown(report)
	default:
		return writeVelocityJSON(report)
	}
}

func writeVelocityJSON(report *jira.VelocityReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func writeVelocityTable(report *jira.VelocityReport) error {
	// Print summary
	fmt.Printf("Board: %d | Sprints: %d | Avg Velocity: %.1f pts\n\n",
		report.BoardID, report.SprintCount, report.AveragePoints)

	tw := tablewriter.NewWriter(os.Stdout)
	tw.Header([]string{"Sprint", "State", "Issues", "Points", "Dates"})

	var rows [][]string
	for _, s := range report.Sprints {
		dates := ""
		if s.StartDate != "" && s.EndDate != "" {
			dates = fmt.Sprintf("%s to %s", s.StartDate, s.EndDate)
		}
		rows = append(rows, []string{
			s.SprintName,
			s.State,
			fmt.Sprintf("%d", s.IssueCount),
			fmt.Sprintf("%.1f", s.CompletedPts),
			dates,
		})
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

func writeVelocityToon(report *jira.VelocityReport) error {
	fmt.Printf("Board:%d|Sprints:%d|AvgVel:%.1f\n", report.BoardID, report.SprintCount, report.AveragePoints)
	for _, s := range report.Sprints {
		fmt.Printf("Sprint:%s|St:%s|Iss:%d|Pts:%.1f\n", s.SprintName, s.State, s.IssueCount, s.CompletedPts)
	}
	return nil
}

func writeVelocityCSV(report *jira.VelocityReport) error {
	fmt.Println("Sprint,State,Issues,Points,Start Date,End Date")
	for _, s := range report.Sprints {
		fmt.Printf("\"%s\",%s,%d,%.1f,%s,%s\n",
			s.SprintName, s.State, s.IssueCount, s.CompletedPts, s.StartDate, s.EndDate)
	}
	return nil
}

func writeVelocityMarkdown(report *jira.VelocityReport) error {
	fmt.Printf("**Board:** %d | **Sprints:** %d | **Avg Velocity:** %.1f pts\n\n",
		report.BoardID, report.SprintCount, report.AveragePoints)
	fmt.Println("| Sprint | State | Issues | Points | Dates |")
	fmt.Println("|--------|-------|--------|--------|-------|")
	for _, s := range report.Sprints {
		dates := ""
		if s.StartDate != "" && s.EndDate != "" {
			dates = fmt.Sprintf("%s to %s", s.StartDate, s.EndDate)
		}
		fmt.Printf("| %s | %s | %d | %.1f | %s |\n",
			s.SprintName, s.State, s.IssueCount, s.CompletedPts, dates)
	}
	return nil
}
