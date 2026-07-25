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
	worklogJQL      string
	worklogSprintID int
)

var worklogCmd = &cobra.Command{
	Use:   "worklog",
	Short: "Show worklog report for issues",
	Long: `Generate a worklog summary report showing time spent by author and issue.

Examples:
  # Worklog for a sprint
  gojira worklog --sprint 456

  # Worklog for issues matching JQL
  gojira worklog --jql "project = PROJ AND updated >= -7d"`,
	RunE: runWorklog,
}

func init() {
	rootCmd.AddCommand(worklogCmd)

	worklogCmd.Flags().StringVar(&worklogJQL, "jql", "", "JQL query to filter issues")
	worklogCmd.Flags().IntVar(&worklogSprintID, "sprint", 0, "Sprint ID to filter issues")
}

func runWorklog(cmd *cobra.Command, args []string) error {
	if worklogJQL == "" && worklogSprintID == 0 {
		return fmt.Errorf("either --jql or --sprint is required")
	}

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	opts := &jira.WorklogOptions{
		JQL:      worklogJQL,
		SprintID: worklogSprintID,
	}

	report, err := client.BoardAPI.GetWorklogReport(ctx, opts)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeWorklogTable(report)
	case OutputTOON:
		return writeWorklogToon(report)
	case OutputCSV:
		return writeWorklogCSV(report)
	case OutputMarkdown:
		return writeWorklogMarkdown(report)
	default:
		return writeWorklogJSON(report)
	}
}

func writeWorklogJSON(report *jira.WorklogReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func writeWorklogTable(report *jira.WorklogReport) error {
	// Print summary
	fmt.Printf("Total: %s | Issues: %d | Worklogs: %d\n\n",
		report.TotalTimeSpentStr, report.IssueCount, report.WorklogCount)

	if len(report.ByAuthor) > 0 {
		fmt.Println("By Author:")
		tw := tablewriter.NewWriter(os.Stdout)
		tw.Header([]string{"Author", "Time Spent", "Worklogs"})

		var rows [][]string
		for _, a := range report.ByAuthor {
			rows = append(rows, []string{
				a.Author,
				a.TimeSpentStr,
				fmt.Sprintf("%d", a.WorklogCount),
			})
		}

		if err := tw.Bulk(rows); err != nil {
			return err
		}
		if err := tw.Render(); err != nil {
			return err
		}
	}

	if len(report.ByIssue) > 0 {
		fmt.Println("\nBy Issue:")
		tw := tablewriter.NewWriter(os.Stdout)
		tw.Header([]string{"Key", "Summary", "Time Spent", "Worklogs"})

		var rows [][]string
		for _, i := range report.ByIssue {
			rows = append(rows, []string{
				i.Key,
				truncateString(i.Summary, 40),
				i.TimeSpentStr,
				fmt.Sprintf("%d", i.WorklogCount),
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

func writeWorklogToon(report *jira.WorklogReport) error {
	fmt.Printf("Tot:%s|Iss:%d|WL:%d\n", report.TotalTimeSpentStr, report.IssueCount, report.WorklogCount)
	for _, a := range report.ByAuthor {
		fmt.Printf("Auth:%s|T:%s|WL:%d\n", a.Author, a.TimeSpentStr, a.WorklogCount)
	}
	for _, i := range report.ByIssue {
		fmt.Printf("K:%s|T:%s|WL:%d\n", i.Key, i.TimeSpentStr, i.WorklogCount)
	}
	return nil
}

func writeWorklogCSV(report *jira.WorklogReport) error {
	fmt.Println("# By Author")
	fmt.Println("Author,Time Spent,Worklogs")
	for _, a := range report.ByAuthor {
		fmt.Printf("\"%s\",%s,%d\n", a.Author, a.TimeSpentStr, a.WorklogCount)
	}
	fmt.Println()
	fmt.Println("# By Issue")
	fmt.Println("Key,Summary,Time Spent,Worklogs")
	for _, i := range report.ByIssue {
		summary := truncateString(i.Summary, 50)
		fmt.Printf("%s,\"%s\",%s,%d\n", i.Key, summary, i.TimeSpentStr, i.WorklogCount)
	}
	return nil
}

func writeWorklogMarkdown(report *jira.WorklogReport) error {
	fmt.Printf("**Total:** %s | **Issues:** %d | **Worklogs:** %d\n\n",
		report.TotalTimeSpentStr, report.IssueCount, report.WorklogCount)

	if len(report.ByAuthor) > 0 {
		fmt.Println("### By Author")
		fmt.Println()
		fmt.Println("| Author | Time Spent | Worklogs |")
		fmt.Println("|--------|------------|----------|")
		for _, a := range report.ByAuthor {
			fmt.Printf("| %s | %s | %d |\n", a.Author, a.TimeSpentStr, a.WorklogCount)
		}
		fmt.Println()
	}

	if len(report.ByIssue) > 0 {
		fmt.Println("### By Issue")
		fmt.Println()
		fmt.Println("| Key | Summary | Time Spent | Worklogs |")
		fmt.Println("|-----|---------|------------|----------|")
		for _, i := range report.ByIssue {
			summary := truncateString(i.Summary, 40)
			fmt.Printf("| %s | %s | %s | %d |\n", i.Key, summary, i.TimeSpentStr, i.WorklogCount)
		}
	}
	return nil
}
