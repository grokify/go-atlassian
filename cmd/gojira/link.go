package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	jira "github.com/andygrunwald/go-jira"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	linkListTypes bool
	linkType      string
	linkComment   string
)

var linkCmd = &cobra.Command{
	Use:   "link <from-issue> <to-issue>",
	Short: "Create a link between two Jira issues",
	Long: `Create a link between two Jira issues.

The link type specifies the relationship direction. Common link types include:
  - Blocks/is blocked by
  - Clones/is cloned by
  - Duplicates/is duplicated by
  - Relates to

Use --list-types to see available link types for your Jira instance.

Examples:
  # List available link types
  gojira link --list-types

  # Create a "Blocks" link (ISSUE-1 blocks ISSUE-2)
  gojira link ISSUE-1 ISSUE-2 --type "Blocks"

  # Create a link with a comment
  gojira link ISSUE-1 ISSUE-2 --type "Relates" --comment "Related to the same feature"`,
	Args: func(cmd *cobra.Command, args []string) error {
		if linkListTypes {
			return nil
		}
		if len(args) != 2 {
			return fmt.Errorf("requires exactly 2 issue keys (from-issue and to-issue)")
		}
		return nil
	},
	RunE: runLink,
}

func init() {
	rootCmd.AddCommand(linkCmd)

	linkCmd.Flags().BoolVar(&linkListTypes, "list-types", false, "List available link types")
	linkCmd.Flags().StringVarP(&linkType, "type", "t", "", "Link type name (e.g., 'Blocks', 'Relates')")
	linkCmd.Flags().StringVarP(&linkComment, "comment", "c", "", "Optional comment to add with the link")
}

func runLink(cmd *cobra.Command, args []string) error {
	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	if linkListTypes {
		return listLinkTypes(ctx, client.JiraClient)
	}

	// Create link between issues
	if linkType == "" {
		return fmt.Errorf("--type is required when creating a link")
	}

	fromIssue := strings.TrimSpace(args[0])
	toIssue := strings.TrimSpace(args[1])

	link := &jira.IssueLink{
		Type: jira.IssueLinkType{
			Name: linkType,
		},
		OutwardIssue: &jira.Issue{
			Key: toIssue,
		},
		InwardIssue: &jira.Issue{
			Key: fromIssue,
		},
	}

	if linkComment != "" {
		link.Comment = &jira.Comment{
			Body: linkComment,
		}
	}

	_, err = client.JiraClient.Issue.AddLinkWithContext(ctx, link)
	if err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}

	if !flagQuiet {
		fmt.Printf("Created link: %s -[%s]-> %s\n", fromIssue, linkType, toIssue)
	}

	return nil
}

func listLinkTypes(ctx context.Context, jiraClient *jira.Client) error {
	linkTypes, _, err := jiraClient.IssueLinkType.GetListWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get link types: %w", err)
	}

	if len(linkTypes) == 0 {
		if !flagQuiet {
			fmt.Fprintln(os.Stderr, "No link types found")
		}
		return nil
	}

	format := getOutputFormat()
	switch format {
	case OutputTable:
		return writeLinkTypesTable(linkTypes)
	case OutputTOON:
		return writeLinkTypesToon(linkTypes)
	default:
		return writeLinkTypesJSON(linkTypes)
	}
}

// LinkTypeOutput is a simplified view of a link type for output.
type LinkTypeOutput struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

func linkTypesToOutputs(linkTypes []jira.IssueLinkType) []LinkTypeOutput {
	outputs := make([]LinkTypeOutput, len(linkTypes))
	for i, lt := range linkTypes {
		outputs[i] = LinkTypeOutput{
			ID:      lt.ID,
			Name:    lt.Name,
			Inward:  lt.Inward,
			Outward: lt.Outward,
		}
	}
	return outputs
}

func writeLinkTypesJSON(linkTypes []jira.IssueLinkType) error {
	outputs := linkTypesToOutputs(linkTypes)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(outputs)
}

func writeLinkTypesTable(linkTypes []jira.IssueLinkType) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.Header([]string{"ID", "Name", "Inward", "Outward"})

	var rows [][]string
	for _, lt := range linkTypes {
		rows = append(rows, []string{lt.ID, lt.Name, lt.Inward, lt.Outward})
	}

	if err := tw.Bulk(rows); err != nil {
		return err
	}
	return tw.Render()
}

func writeLinkTypesToon(linkTypes []jira.IssueLinkType) error {
	for _, lt := range linkTypes {
		fmt.Printf("ID:%s|N:%s|In:%s|Out:%s\n", lt.ID, lt.Name, lt.Inward, lt.Outward)
	}
	return nil
}
