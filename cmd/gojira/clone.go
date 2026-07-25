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
	cloneProject  string
	cloneType     string
	clonePrefix   string
	cloneSuffix   string
	cloneExclude  []string
	cloneInclude  []string
	cloneLink     bool
	cloneLinkType string
	cloneParent   string
)

var cloneCmd = &cobra.Command{
	Use:   "clone <source-issue-key>",
	Short: "Clone a Jira issue",
	Long: `Clone a Jira issue with configurable field mapping.

By default, all standard fields are copied. Use --exclude to skip specific
fields, or --include to copy only specified fields.

Examples:
  # Basic clone
  gojira clone PROJ-123

  # Clone with prefix
  gojira clone PROJ-123 --prefix "[Clone] "

  # Clone to a different project
  gojira clone PROJ-123 --project NEWPROJ

  # Clone excluding certain fields
  gojira clone PROJ-123 --exclude labels --exclude components

  # Clone only specific fields
  gojira clone PROJ-123 --include description --include priority

  # Clone and link to original
  gojira clone PROJ-123 --link

  # Clone a subtask to a different parent
  gojira clone PROJ-456 --parent PROJ-100`,
	Args: cobra.ExactArgs(1),
	RunE: runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().StringVarP(&cloneProject, "project", "p", "", "Target project (default: same as source)")
	cloneCmd.Flags().StringVarP(&cloneType, "type", "t", "", "Target issue type (default: same as source)")
	cloneCmd.Flags().StringVar(&clonePrefix, "prefix", "", "Prefix for cloned issue summary")
	cloneCmd.Flags().StringVar(&cloneSuffix, "suffix", "", "Suffix for cloned issue summary")
	cloneCmd.Flags().StringSliceVarP(&cloneExclude, "exclude", "x", nil, "Fields to exclude (e.g., labels, components)")
	cloneCmd.Flags().StringSliceVarP(&cloneInclude, "include", "i", nil, "Fields to include (overrides exclude)")
	cloneCmd.Flags().BoolVar(&cloneLink, "link", false, "Create link to original issue")
	cloneCmd.Flags().StringVar(&cloneLinkType, "link-type", "Cloners", "Link type name when --link is used")
	cloneCmd.Flags().StringVar(&cloneParent, "parent", "", "Parent issue key for subtasks")
}

func runClone(cmd *cobra.Command, args []string) error {
	sourceKey := strings.TrimSpace(args[0])

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	opts := &jira.CloneOptions{
		TargetProject:     cloneProject,
		TargetType:        cloneType,
		SummaryPrefix:     clonePrefix,
		SummarySuffix:     cloneSuffix,
		ExcludeFields:     cloneExclude,
		IncludeOnlyFields: cloneInclude,
		LinkToOriginal:    cloneLink,
		LinkType:          cloneLinkType,
		Parent:            cloneParent,
	}

	result, err := client.IssueAPI.CloneIssue(ctx, sourceKey, opts)
	if err != nil {
		return err
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeCloneResultTable(result)
	case OutputTOON:
		return writeCloneResultToon(result)
	default:
		return writeCloneResultJSON(result)
	}
}

func writeCloneResultJSON(result *jira.CloneResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func writeCloneResultTable(result *jira.CloneResult) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.Header([]string{"Source", "Cloned", "Linked"})

	linked := "No"
	if result.Linked {
		linked = "Yes"
	}

	rows := [][]string{
		{result.SourceKey, result.ClonedKey, linked},
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

func writeCloneResultToon(result *jira.CloneResult) error {
	linked := "N"
	if result.Linked {
		linked = "Y"
	}
	fmt.Printf("Src:%s|Clone:%s|Lnk:%s\n", result.SourceKey, result.ClonedKey, linked)
	return nil
}
