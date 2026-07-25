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

var (
	burndownStoryPtsField string
)

var burndownCmd = &cobra.Command{
	Use:   "burndown <sprint-id>",
	Short: "Show burndown report for a sprint",
	Long: `Calculate and display burndown data for a Jira sprint.

Use 'gojira sprints <board-id>' to find sprint IDs.

Examples:
  # Show burndown for a sprint
  gojira burndown 456

  # Specify custom story points field
  gojira burndown 456 --story-points-field customfield_10016`,
	Args: cobra.ExactArgs(1),
	RunE: runBurndown,
}

func init() {
	rootCmd.AddCommand(burndownCmd)

	burndownCmd.Flags().StringVar(&burndownStoryPtsField, "story-points-field", "", "Custom field ID for story points (default: customfield_10016)")
}

func runBurndown(cmd *cobra.Command, args []string) error {
	sprintID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid sprint ID %q: %w", args[0], err)
	}

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	opts := &jira.BurndownOptions{
		StoryPointsFieldID: burndownStoryPtsField,
	}

	report, err := client.BoardAPI.GetBurndownReport(ctx, sprintID, opts)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeBurndownTable(report)
	case OutputTOON:
		return writeBurndownToon(report)
	case OutputCSV:
		return writeBurndownCSV(report)
	case OutputMarkdown:
		return writeBurndownMarkdown(report)
	default:
		return writeBurndownJSON(report)
	}
}

func writeBurndownJSON(report *jira.BurndownReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func writeBurndownTable(report *jira.BurndownReport) error {
	// Print summary
	fmt.Printf("Sprint: %d | Total: %.1f pts (%d issues)\n\n",
		report.SprintID, report.TotalPoints, report.TotalIssues)

	tw := tablewriter.NewWriter(os.Stdout)
	tw.Header([]string{"Date", "Remaining Pts", "Remaining Issues", "Completed Pts", "Completed Issues"})

	var rows [][]string
	for _, d := range report.DailyData {
		rows = append(rows, []string{
			d.Date,
			fmt.Sprintf("%.1f", d.RemainingPoints),
			fmt.Sprintf("%d", d.RemainingIssues),
			fmt.Sprintf("%.1f", d.CompletedPoints),
			fmt.Sprintf("%d", d.CompletedIssues),
		})
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

func writeBurndownToon(report *jira.BurndownReport) error {
	fmt.Printf("Sprint:%d|TotPts:%.1f|TotIss:%d\n", report.SprintID, report.TotalPoints, report.TotalIssues)
	for _, d := range report.DailyData {
		fmt.Printf("D:%s|RemPts:%.1f|RemIss:%d|DonePts:%.1f|DoneIss:%d\n",
			d.Date, d.RemainingPoints, d.RemainingIssues, d.CompletedPoints, d.CompletedIssues)
	}
	return nil
}

func writeBurndownCSV(report *jira.BurndownReport) error {
	fmt.Println("Date,Remaining Points,Remaining Issues,Completed Points,Completed Issues")
	for _, d := range report.DailyData {
		fmt.Printf("%s,%.1f,%d,%.1f,%d\n",
			d.Date, d.RemainingPoints, d.RemainingIssues, d.CompletedPoints, d.CompletedIssues)
	}
	return nil
}

func writeBurndownMarkdown(report *jira.BurndownReport) error {
	fmt.Printf("**Sprint:** %d | **Total:** %.1f pts (%d issues)\n\n",
		report.SprintID, report.TotalPoints, report.TotalIssues)
	fmt.Println("| Date | Remaining Pts | Remaining Issues | Completed Pts | Completed Issues |")
	fmt.Println("|------|---------------|------------------|---------------|------------------|")
	for _, d := range report.DailyData {
		fmt.Printf("| %s | %.1f | %d | %.1f | %d |\n",
			d.Date, d.RemainingPoints, d.RemainingIssues, d.CompletedPoints, d.CompletedIssues)
	}
	return nil
}
