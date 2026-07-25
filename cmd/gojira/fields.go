package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/grokify/go-atlassian/jira"
	"github.com/spf13/cobra"
)

var fieldsCmd = &cobra.Command{
	Use:   "fields [flags]",
	Short: "List and filter custom fields",
	Long: `List custom fields from Jira, with optional filtering.

Examples:
  # List all custom fields
  gojira fields

  # Filter by field ID
  gojira fields --id customfield_10001

  # Filter by multiple IDs
  gojira fields --id customfield_10001,customfield_10002

  # Filter by name (partial match)
  gojira fields --name "Epic"

  # Filter by exact name
  gojira fields --name-exact "Epic Link"

  # Show only custom fields (exclude system fields)
  gojira fields --custom-only

  # Output as JSON
  gojira fields --json

  # Get Epic Link field specifically
  gojira fields --epic-link

  # List fields available in project ABC
  gojira fields --project ABC

  # List fields for specific issue type in project
  gojira fields --project ABC --issue-type 10001

  # Show fields that have duplicate names
  gojira fields --show-duplicates`,
	RunE: runFields,
}

var (
	fieldsFilterIDs   string
	fieldsFilterName  string
	fieldsFilterExact string
	fieldsCustomOnly  bool
	fieldsEpicLink    bool
	fieldsOutputJSON  bool
	fieldsOutputTable bool
	fieldsProject     string
	fieldsIssueType   string
	fieldsShowDupes   bool
)

func init() {
	rootCmd.AddCommand(fieldsCmd)

	fieldsCmd.Flags().StringVar(&fieldsFilterIDs, "id", "", "Filter by field ID(s), comma-separated")
	fieldsCmd.Flags().StringVar(&fieldsFilterName, "name", "", "Filter by name (partial match)")
	fieldsCmd.Flags().StringVar(&fieldsFilterExact, "name-exact", "", "Filter by exact name")
	fieldsCmd.Flags().BoolVar(&fieldsCustomOnly, "custom-only", false, "Show only custom fields")
	fieldsCmd.Flags().BoolVar(&fieldsEpicLink, "epic-link", false, "Show Epic Link field")
	fieldsCmd.Flags().BoolVar(&fieldsOutputJSON, "json", false, "Output as JSON")
	fieldsCmd.Flags().BoolVar(&fieldsOutputTable, "table", true, "Output as table (default)")
	fieldsCmd.Flags().StringVar(&fieldsProject, "project", "", "Filter by project key (shows only fields available in project)")
	fieldsCmd.Flags().StringVar(&fieldsIssueType, "issue-type", "", "Filter by issue type ID (requires --project)")
	fieldsCmd.Flags().BoolVar(&fieldsShowDupes, "show-duplicates", false, "Show only fields with duplicate names")
}

func runFields(cmd *cobra.Command, args []string) error {
	client, err := NewClientFromOptions(getAuthOptions())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Validate flag combinations
	if fieldsIssueType != "" && fieldsProject == "" {
		return fmt.Errorf("--issue-type requires --project")
	}

	// Special case: get Epic Link field
	if fieldsEpicLink {
		return showEpicLinkField(client)
	}

	var fields jira.CustomFields

	// Get fields based on project filtering
	if fieldsProject != "" {
		ctx := context.Background()
		if fieldsIssueType != "" {
			// Get fields for specific issue type
			metaFields, err := client.CreateMetaAPI.GetFields(ctx, fieldsProject, fieldsIssueType)
			if err != nil {
				return fmt.Errorf("failed to get fields for project %q, issue type %q: %w", fieldsProject, fieldsIssueType, err)
			}
			// Get full metadata for the custom fields
			allFields, err := client.CustomFieldAPI.GetCustomFields()
			if err != nil {
				return fmt.Errorf("failed to get custom fields: %w", err)
			}
			fields = allFields.FilterByIDs(metaFields.CustomOnly().Keys()...)
		} else {
			// Get all fields for project
			fields, err = client.CustomFieldAPI.GetCustomFieldsForProject(ctx, fieldsProject)
			if err != nil {
				return fmt.Errorf("failed to get fields for project %q: %w", fieldsProject, err)
			}
		}
	} else {
		// Get all fields
		fields, err = client.CustomFieldAPI.GetCustomFields()
		if err != nil {
			return fmt.Errorf("failed to get fields: %w", err)
		}
	}

	// Apply filters
	fields = applyFieldFilters(fields)

	// Filter to only show duplicates if requested
	if fieldsShowDupes {
		fields = filterDuplicateNames(fields)
	}

	if len(fields) == 0 {
		fmt.Fprintln(os.Stderr, "No fields found matching criteria")
		return nil
	}

	// Output
	if fieldsOutputJSON {
		return outputFieldsJSON(fields)
	}

	return outputFieldsTable(fields)
}

func showEpicLinkField(client *jira.Client) error {
	field, err := client.CustomFieldAPI.GetCustomFieldEpicLink()
	if err != nil {
		return fmt.Errorf("failed to get Epic Link field: %w", err)
	}

	if field.ID == "" {
		fmt.Fprintln(os.Stderr, "Epic Link field not found")
		return nil
	}

	fields := jira.CustomFields{field}

	if fieldsOutputJSON {
		return outputFieldsJSON(fields)
	}

	return outputFieldsTable(fields)
}

func applyFieldFilters(fields jira.CustomFields) jira.CustomFields {
	// Filter by custom only
	if fieldsCustomOnly {
		var filtered jira.CustomFields
		for _, f := range fields {
			if f.Custom {
				filtered = append(filtered, f)
			}
		}
		fields = filtered
	}

	// Filter by IDs
	if fieldsFilterIDs != "" {
		ids := parseCommaSeparated(fieldsFilterIDs)
		fields = fields.FilterByIDs(ids...)
	}

	// Filter by exact name
	if fieldsFilterExact != "" {
		names := parseCommaSeparated(fieldsFilterExact)
		fields = fields.FilterByNames(names...)
	}

	// Filter by partial name match
	if fieldsFilterName != "" {
		var filtered jira.CustomFields
		searchLower := strings.ToLower(fieldsFilterName)
		for _, f := range fields {
			if strings.Contains(strings.ToLower(f.Name), searchLower) {
				filtered = append(filtered, f)
			}
		}
		fields = filtered
	}

	return fields
}

func parseCommaSeparated(s string) []string {
	var result []string
	for _, item := range strings.Split(s, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func outputFieldsJSON(fields jira.CustomFields) error {
	data, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputFieldsTable(fields jira.CustomFields) error {
	return fields.WriteTable(os.Stdout)
}

// filterDuplicateNames returns only fields whose names appear more than once.
func filterDuplicateNames(fields jira.CustomFields) jira.CustomFields {
	dupeNames := fields.DuplicateNames()
	if len(dupeNames) == 0 {
		return jira.CustomFields{}
	}
	return fields.FilterByNames(dupeNames...)
}
