package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grokify/go-atlassian/jira"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [flags]",
	Short: "Export issues to JSON or XLSX",
	Long: `Export Jira issues to JSON or XLSX format.

Examples:
  # Export search results to JSON
  gojira export --jql "project = FOO" --json output.json

  # Export to Excel
  gojira export --jql "project = FOO AND status = Open" --xlsx output.xlsx

  # Export with parent issues included
  gojira export --jql "project = FOO" --include-parents --xlsx output.xlsx

  # Export from existing JSON file to XLSX
  gojira export --from-json issues.json --xlsx output.xlsx

  # Export specific issues by key
  gojira export --keys ISSUE-1,ISSUE-2,ISSUE-3 --json output.json`,
	RunE: runExport,
}

var (
	exportJQL            string
	exportKeys           string
	exportJSONOutput     string
	exportXLSXOutput     string
	exportFromJSON       string
	exportIncludeParents bool
	exportSheetName      string
)

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVar(&exportJQL, "jql", "", "JQL query to search issues")
	exportCmd.Flags().StringVar(&exportKeys, "keys", "", "Comma-separated issue keys to export")
	exportCmd.Flags().StringVar(&exportJSONOutput, "json", "", "Output JSON file path")
	exportCmd.Flags().StringVar(&exportXLSXOutput, "xlsx", "", "Output XLSX file path")
	exportCmd.Flags().StringVar(&exportFromJSON, "from-json", "", "Read issues from existing JSON file instead of querying")
	exportCmd.Flags().BoolVar(&exportIncludeParents, "include-parents", false, "Include parent issues in export")
	exportCmd.Flags().StringVar(&exportSheetName, "sheet", "issues", "Sheet name for XLSX export")
}

func runExport(cmd *cobra.Command, args []string) error {
	// Validate output flags
	if exportJSONOutput == "" && exportXLSXOutput == "" {
		return fmt.Errorf("at least one output format required: --json or --xlsx")
	}

	var issuesSet *jira.IssuesSet
	var err error

	// Get issues from source
	if exportFromJSON != "" {
		// Read from existing JSON file
		issuesSet, err = jira.IssuesSetReadFileJSON(exportFromJSON)
		if err != nil {
			return fmt.Errorf("failed to read JSON file: %w", err)
		}
		if !flagQuiet {
			fmt.Fprintf(os.Stderr, "Loaded %d issues from %s\n", issuesSet.Len(), exportFromJSON)
		}
	} else {
		// Query from Jira
		issuesSet, err = fetchIssuesForExport()
		if err != nil {
			return err
		}
	}

	// Include parents if requested
	if exportIncludeParents && exportFromJSON == "" {
		client, err := NewClientFromOptions(getAuthOptions())
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		if !flagQuiet {
			fmt.Fprintf(os.Stderr, "Fetching parent issues...\n")
		}
		if err := client.IssueAPI.IssuesSetAddParents(issuesSet); err != nil {
			return fmt.Errorf("failed to fetch parents: %w", err)
		}
		if !flagQuiet && issuesSet.Parents != nil {
			fmt.Fprintf(os.Stderr, "Added %d parent issues\n", len(issuesSet.Parents.Keys()))
		}
	}

	// Export to JSON
	if exportJSONOutput != "" {
		if err := writeJSONExport(issuesSet, exportJSONOutput); err != nil {
			return err
		}
	}

	// Export to XLSX
	if exportXLSXOutput != "" {
		if err := writeXLSXExport(issuesSet, exportXLSXOutput); err != nil {
			return err
		}
	}

	return nil
}

func fetchIssuesForExport() (*jira.IssuesSet, error) {
	if exportJQL == "" && exportKeys == "" {
		return nil, fmt.Errorf("query required: use --jql or --keys")
	}

	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if exportKeys != "" {
		// Fetch specific issues by key
		keys := parseKeys(exportKeys)
		if len(keys) == 0 {
			return nil, fmt.Errorf("no valid issue keys provided")
		}

		if !flagQuiet {
			fmt.Fprintf(os.Stderr, "Fetching %d issues...\n", len(keys))
		}

		issuesSet, err := client.IssueAPI.SearchIssuesSet(fmt.Sprintf("key in (%s)", strings.Join(keys, ",")))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues: %w", err)
		}
		return issuesSet, nil
	}

	// Search with JQL
	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "Searching issues with JQL: %s\n", exportJQL)
	}

	issuesSet, err := client.IssueAPI.SearchIssuesSet(exportJQL)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "Found %d issues\n", issuesSet.Len())
	}

	return issuesSet, nil
}

func parseKeys(keysStr string) []string {
	var keys []string
	for _, k := range strings.Split(keysStr, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}

func writeJSONExport(issuesSet *jira.IssuesSet, outputPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Write JSON
	data, err := json.MarshalIndent(issuesSet, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "Wrote %d issues to %s\n", issuesSet.Len(), outputPath)
	}

	return nil
}

func writeXLSXExport(issuesSet *jira.IssuesSet, outputPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Generate table
	tbl, err := issuesSet.TableDefault(nil, true, "Top-level initiative", []string{})
	if err != nil {
		return fmt.Errorf("failed to generate table: %w", err)
	}

	// Write XLSX
	if err := tbl.WriteXLSX(outputPath, exportSheetName); err != nil {
		return fmt.Errorf("failed to write XLSX: %w", err)
	}

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "Wrote %d issues to %s\n", issuesSet.Len(), outputPath)
	}

	return nil
}
