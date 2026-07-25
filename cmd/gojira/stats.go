package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/grokify/go-atlassian/jira"
	"github.com/spf13/cobra"
	toon "github.com/toon-format/toon-go"
)

var statsCmd = &cobra.Command{
	Use:   "stats [flags]",
	Short: "Show issue statistics grouped by field",
	Long: `Show aggregate statistics for issues grouped by a field.

Examples:
  # Count by status
  gojira stats --jql "project = FOO" --by status

  # Count by type
  gojira stats --jql "project = FOO" --by type

  # Count by priority
  gojira stats --jql "project = FOO" --by priority

  # Count by custom field
  gojira stats --jql "project = FOO" --by customfield_12345

  # Count by assignee
  gojira stats --jql "project = FOO" --by assignee

  # Output formats (default: toon)
  gojira stats --jql "..." --by status --format toon   # Token-optimized (default)
  gojira stats --jql "..." --by status --format json   # JSON
  gojira stats --jql "..." --by status --format table  # Human-readable`,
	RunE: runStats,
}

var (
	statsJQL    string
	statsBy     string
	statsFormat string
)

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().StringVar(&statsJQL, "jql", "", "JQL query to search issues (required)")
	statsCmd.Flags().StringVar(&statsBy, "by", "", "Field to group by: status, type, priority, assignee, project, or customfield_XXXXX (required)")
	statsCmd.Flags().StringVar(&statsFormat, "format", "toon", "Output format: toon (default), json, table")

	_ = statsCmd.MarkFlagRequired("jql")
	_ = statsCmd.MarkFlagRequired("by")
}

// StatResult represents a single count result.
type StatResult struct {
	Value string `json:"value" toon:"v"`
	Count uint   `json:"count" toon:"n"`
}

// StatsOutput represents the full stats output.
type StatsOutput struct {
	Field   string       `json:"field" toon:"f"`
	Total   uint         `json:"total" toon:"t"`
	Results []StatResult `json:"results" toon:"r"`
}

func runStats(cmd *cobra.Command, args []string) error {
	// Validate format
	format := strings.ToLower(statsFormat)
	if format != "toon" && format != "json" && format != "table" {
		return fmt.Errorf("invalid format %q: use toon, json, or table", statsFormat)
	}

	// Get client
	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Search issues
	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "Searching issues...\n")
	}

	issues, err := client.IssueAPI.SearchIssues(statsJQL, false)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(issues) == 0 {
		if !flagQuiet {
			fmt.Fprintln(os.Stderr, "No issues found")
		}
		return nil
	}

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "Found %d issues\n", len(issues))
	}

	// Compute counts
	counts, err := computeCounts(issues, statsBy)
	if err != nil {
		return err
	}

	// Build output
	output := StatsOutput{
		Field: statsBy,
		Total: uint(len(issues)),
	}

	// Sort by count descending
	type kv struct {
		Key   string
		Value uint
	}
	var sorted []kv
	for k, v := range counts {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	for _, item := range sorted {
		output.Results = append(output.Results, StatResult{
			Value: item.Key,
			Count: item.Value,
		})
	}

	// Output
	return outputStats(output, format)
}

func computeCounts(issues jira.Issues, field string) (map[string]uint, error) {
	field = strings.ToLower(field)

	// Convert to IssuesSet for most operations
	issuesSet, err := issues.IssuesSet(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create issues set: %w", err)
	}

	// Handle built-in fields
	switch field {
	case "status":
		return issuesSet.CountsByStatus(), nil
	case "type":
		return issuesSet.CountsByType(true, true), nil
	case "project":
		return issuesSet.CountsByProject(), nil
	case "priority":
		return countsByPriority(issues), nil
	case "assignee":
		return countsByAssignee(issues), nil
	case "resolution":
		return countsByResolution(issues), nil
	}

	// Handle custom fields
	if strings.HasPrefix(field, "customfield_") {
		return issuesSet.CountsByCustomFieldValues(field)
	}

	return nil, fmt.Errorf("unknown field %q: use status, type, priority, assignee, project, resolution, or customfield_XXXXX", field)
}

func countsByPriority(issues jira.Issues) map[string]uint {
	counts := make(map[string]uint)
	for _, iss := range issues {
		priority := "(none)"
		if iss.Fields != nil && iss.Fields.Priority != nil {
			priority = iss.Fields.Priority.Name
		}
		counts[priority]++
	}
	return counts
}

func countsByAssignee(issues jira.Issues) map[string]uint {
	counts := make(map[string]uint)
	for _, iss := range issues {
		im := jira.NewIssueMore(&iss)
		assignee := im.AssigneeName()
		if assignee == "" {
			assignee = "(unassigned)"
		}
		counts[assignee]++
	}
	return counts
}

func countsByResolution(issues jira.Issues) map[string]uint {
	counts := make(map[string]uint)
	for _, iss := range issues {
		im := jira.NewIssueMore(&iss)
		resolution := im.Resolution()
		if resolution == "" {
			resolution = "(unresolved)"
		}
		counts[resolution]++
	}
	return counts
}

func outputStats(output StatsOutput, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table":
		return outputStatsTable(output)
	default: // toon
		data, err := toon.Marshal(output)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	}
	return nil
}

func outputStatsTable(output StatsOutput) error {
	// Calculate max width for value column
	maxWidth := 5 // minimum "VALUE"
	for _, r := range output.Results {
		if len(r.Value) > maxWidth {
			maxWidth = len(r.Value)
		}
	}
	if maxWidth > 40 {
		maxWidth = 40
	}

	// Print header
	fmt.Printf("%-*s  %6s  %6s\n", maxWidth, "VALUE", "COUNT", "%")
	fmt.Printf("%s  %s  %s\n", strings.Repeat("-", maxWidth), "------", "------")

	// Print rows
	for _, r := range output.Results {
		value := r.Value
		if len(value) > maxWidth {
			value = value[:maxWidth-3] + "..."
		}
		var pct float64
		if output.Total > 0 {
			pct = float64(r.Count) / float64(output.Total) * 100
		}
		fmt.Printf("%-*s  %6d  %5.1f%%\n", maxWidth, value, r.Count, pct)
	}

	// Print total
	fmt.Printf("%s  %s  %s\n", strings.Repeat("-", maxWidth), "------", "------")
	fmt.Printf("%-*s  %6d  %5.1f%%\n", maxWidth, "TOTAL", output.Total, 100.0)

	return nil
}
