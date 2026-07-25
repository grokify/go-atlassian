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
	cycletimeJQL      string
	cycletimeSprintID int
)

var cycletimeCmd = &cobra.Command{
	Use:   "cycle-time",
	Short: "Show cycle time report for resolved issues",
	Long: `Calculate cycle time metrics for resolved issues.

Cycle time is measured from issue creation to resolution.

Examples:
  # Cycle time for a sprint
  gojira cycle-time --sprint 456

  # Cycle time for issues matching JQL
  gojira cycle-time --jql "project = PROJ AND resolved >= -30d"`,
	RunE: runCycleTime,
}

func init() {
	rootCmd.AddCommand(cycletimeCmd)

	cycletimeCmd.Flags().StringVar(&cycletimeJQL, "jql", "", "JQL query to filter issues")
	cycletimeCmd.Flags().IntVar(&cycletimeSprintID, "sprint", 0, "Sprint ID to filter issues")
}

func runCycleTime(cmd *cobra.Command, args []string) error {
	if cycletimeJQL == "" && cycletimeSprintID == 0 {
		return fmt.Errorf("either --jql or --sprint is required")
	}

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	opts := &jira.CycleTimeOptions{
		JQL:      cycletimeJQL,
		SprintID: cycletimeSprintID,
	}

	report, err := client.BoardAPI.GetCycleTimeReport(ctx, opts)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeCycleTimeTable(report)
	case OutputTOON:
		return writeCycleTimeToon(report)
	case OutputCSV:
		return writeCycleTimeCSV(report)
	case OutputMarkdown:
		return writeCycleTimeMarkdown(report)
	default:
		return writeCycleTimeJSON(report)
	}
}

func writeCycleTimeJSON(report *jira.CycleTimeReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func writeCycleTimeTable(report *jira.CycleTimeReport) error {
	// Print summary
	fmt.Printf("Issues: %d | Avg: %.1f days | Median: %.1f days | Min: %.1f days | Max: %.1f days\n\n",
		report.IssueCount, report.AverageCycleTime, report.MedianCycleTime,
		report.MinCycleTime, report.MaxCycleTime)

	if len(report.Issues) > 0 {
		tw := tablewriter.NewWriter(os.Stdout)
		tw.Header([]string{"Key", "Type", "Created", "Resolved", "Days", "Summary"})

		var rows [][]string
		for _, i := range report.Issues {
			rows = append(rows, []string{
				i.Key,
				i.IssueType,
				i.Created,
				i.Resolved,
				fmt.Sprintf("%.1f", i.CycleTimeDays),
				truncateString(i.Summary, 30),
			})
		}

		if err := tw.Bulk(rows); err != nil {
			return err
		}
		if err := tw.Render(); err != nil {
			return err
		}
	}

	return nil
}

func writeCycleTimeToon(report *jira.CycleTimeReport) error {
	fmt.Printf("Iss:%d|Avg:%.1fd|Med:%.1fd|Min:%.1fd|Max:%.1fd\n",
		report.IssueCount, report.AverageCycleTime, report.MedianCycleTime,
		report.MinCycleTime, report.MaxCycleTime)
	for _, i := range report.Issues {
		fmt.Printf("K:%s|T:%s|D:%.1f\n", i.Key, i.IssueType, i.CycleTimeDays)
	}
	return nil
}

func writeCycleTimeCSV(report *jira.CycleTimeReport) error {
	fmt.Println("Key,Type,Created,Resolved,Cycle Time (Days),Summary")
	for _, i := range report.Issues {
		summary := truncateString(i.Summary, 50)
		fmt.Printf("%s,%s,%s,%s,%.1f,\"%s\"\n",
			i.Key, i.IssueType, i.Created, i.Resolved, i.CycleTimeDays, summary)
	}
	return nil
}

func writeCycleTimeMarkdown(report *jira.CycleTimeReport) error {
	fmt.Printf("**Issues:** %d | **Avg:** %.1f days | **Median:** %.1f days | **Min:** %.1f days | **Max:** %.1f days\n\n",
		report.IssueCount, report.AverageCycleTime, report.MedianCycleTime,
		report.MinCycleTime, report.MaxCycleTime)
	fmt.Println("| Key | Type | Created | Resolved | Days | Summary |")
	fmt.Println("|-----|------|---------|----------|------|---------|")
	for _, i := range report.Issues {
		summary := truncateString(i.Summary, 30)
		fmt.Printf("| %s | %s | %s | %s | %.1f | %s |\n",
			i.Key, i.IssueType, i.Created, i.Resolved, i.CycleTimeDays, summary)
	}
	return nil
}
