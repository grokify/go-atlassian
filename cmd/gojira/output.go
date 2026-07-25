package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/go-atlassian/jira"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// OutputFormat represents the output format type.
type OutputFormat int

const (
	OutputJSON OutputFormat = iota
	OutputTable
	OutputTOON
	OutputCSV
	OutputMarkdown
)

// OutputConfig holds output configuration.
type OutputConfig struct {
	Format OutputFormat
	Writer io.Writer
}

// NewOutputConfig creates a new OutputConfig with default writer (stdout).
func NewOutputConfig(format OutputFormat) *OutputConfig {
	return &OutputConfig{
		Format: format,
		Writer: os.Stdout,
	}
}

// WriteIssues writes issues in the specified format.
func WriteIssues(issues jira.Issues, cfg *OutputConfig) error {
	if cfg == nil {
		cfg = NewOutputConfig(OutputJSON)
	}
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}

	switch cfg.Format {
	case OutputJSON:
		return writeIssuesJSON(issues, cfg.Writer)
	case OutputTable:
		return writeIssuesTable(issues, cfg.Writer)
	case OutputTOON:
		return writeIssuesToon(issues, cfg.Writer)
	case OutputCSV:
		return writeIssuesCSV(issues, cfg.Writer)
	case OutputMarkdown:
		return writeIssuesMarkdown(issues, cfg.Writer)
	default:
		return writeIssuesJSON(issues, cfg.Writer)
	}
}

// WriteIssue writes a single issue in the specified format.
func WriteIssue(issue *gojira.Issue, cfg *OutputConfig) error {
	if issue == nil {
		return nil
	}
	issues := jira.Issues{*issue}
	return WriteIssues(issues, cfg)
}

// writeIssuesJSON outputs issues as JSON using the shared IssueOutput type.
func writeIssuesJSON(issues jira.Issues, w io.Writer) error {
	outputs := jira.ToIssueOutputs(issues)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(outputs)
}

// WriteIssuesRaw outputs the full API JSON for issues (all fields from Jira API).
func WriteIssuesRaw(issues jira.Issues) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(issues)
}

// writeIssuesTable outputs issues as an ASCII table.
func writeIssuesTable(issues jira.Issues, w io.Writer) error {
	tw := tablewriter.NewWriter(w)
	tw.Header([]string{"Key", "Type", "Status", "Assignee", "Summary"})

	var rows [][]string
	for _, iss := range issues {
		im := jira.NewIssueMore(&iss)
		summary := truncateString(im.Summary(), 50)
		rows = append(rows, []string{
			im.Key(),
			im.Type(),
			im.Status(),
			im.AssigneeName(),
			summary,
		})
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

// writeIssuesToon outputs issues in TOON (Token-Optimized Object Notation) format.
// TOON is a compact key-value format designed for minimal token usage with LLMs.
// Format: K:KEY|T:Type|S:Status|A:Assignee|Su:Summary
func writeIssuesToon(issues jira.Issues, w io.Writer) error {
	for _, iss := range issues {
		im := jira.NewIssueMore(&iss)
		line := formatTOON(im)
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

// formatTOON formats a single issue in TOON format.
func formatTOON(im jira.IssueMore) string {
	parts := []string{
		"K:" + im.Key(),
		"T:" + im.Type(),
		"S:" + im.Status(),
	}

	if assignee := im.AssigneeName(); assignee != "" {
		parts = append(parts, "A:"+assignee)
	}

	if resolution := im.Resolution(); resolution != "" {
		parts = append(parts, "R:"+resolution)
	}

	if project := im.ProjectKey(); project != "" {
		parts = append(parts, "P:"+project)
	}

	// Summary - truncate for readability
	if summary := im.Summary(); summary != "" {
		parts = append(parts, "Su:"+truncateString(summary, 80))
	}

	return strings.Join(parts, "|")
}

// truncateString truncates a string to maxLen and adds ellipsis if needed.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// outputResult writes any JSON-serializable result to stdout.
// Used for commands that return non-Issue results like create, update, etc.
func outputResult(_ *cobra.Command, result any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// writeIssuesCSV outputs issues as CSV.
func writeIssuesCSV(issues jira.Issues, w io.Writer) error {
	// Write header
	if _, err := fmt.Fprintln(w, "Key,Type,Status,Assignee,Summary"); err != nil {
		return err
	}

	for _, iss := range issues {
		im := jira.NewIssueMore(&iss)
		// Escape double quotes in summary and wrap in quotes
		summary := strings.ReplaceAll(im.Summary(), "\"", "\"\"")
		line := fmt.Sprintf("%s,%s,%s,%s,\"%s\"",
			im.Key(),
			im.Type(),
			im.Status(),
			im.AssigneeName(),
			summary)
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

// writeIssuesMarkdown outputs issues as a Markdown table.
func writeIssuesMarkdown(issues jira.Issues, w io.Writer) error {
	// Write header
	if _, err := fmt.Fprintln(w, "| Key | Type | Status | Assignee | Summary |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "|-----|------|--------|----------|---------|"); err != nil {
		return err
	}

	for _, iss := range issues {
		im := jira.NewIssueMore(&iss)
		// Escape pipes in summary
		summary := strings.ReplaceAll(im.Summary(), "|", "\\|")
		summary = truncateString(summary, 50)
		line := fmt.Sprintf("| %s | %s | %s | %s | %s |",
			im.Key(),
			im.Type(),
			im.Status(),
			im.AssigneeName(),
			summary)
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}
