package core

import (
	"context"
	"fmt"
	"os"
	"strings"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/go-atlassian/jira"
	"gopkg.in/yaml.v3"
)

// CreateIssueFromFile reads a YAML file and creates a Jira issue.
func CreateIssueFromFile(ctx context.Context, client *jira.Client, filename string) (*IssueResult, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return CreateIssueFromYAML(ctx, client, data)
}

// CreateIssueFromYAML parses YAML content and creates a Jira issue.
func CreateIssueFromYAML(ctx context.Context, client *jira.Client, data []byte) (*IssueResult, error) {
	input, err := ParseIssueYAML(data)
	if err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}

	return CreateIssue(ctx, client, input)
}

// ParseIssueYAML parses YAML data into IssueInput.
func ParseIssueYAML(data []byte) (*IssueInput, error) {
	var input IssueInput

	// First unmarshal to get standard fields
	if err := yaml.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	// Then unmarshal to raw map to capture custom fields
	var rawMap map[string]any
	if err := yaml.Unmarshal(data, &rawMap); err != nil {
		return nil, err
	}
	input.RawFields = rawMap

	return &input, nil
}

// CreateIssue creates a Jira issue from IssueInput.
func CreateIssue(ctx context.Context, client *jira.Client, input *IssueInput) (*IssueResult, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}

	// Build the Jira issue
	issue := &gojira.Issue{
		Fields: &gojira.IssueFields{
			Summary:     input.Summary,
			Description: input.Description,
			Project: gojira.Project{
				Key: input.Project,
			},
			Type: gojira.IssueType{
				Name: input.Type,
			},
			Labels: input.Labels,
		},
	}

	// Set parent if provided (for subtasks or stories under epics)
	if input.Parent != "" {
		issue.Fields.Parent = &gojira.Parent{
			Key: input.Parent,
		}
	}

	// Set priority if provided
	if input.Priority != "" {
		issue.Fields.Priority = &gojira.Priority{
			Name: input.Priority,
		}
	}

	// Set assignee if provided
	if input.Assignee != "" {
		issue.Fields.Assignee = &gojira.User{
			Name: input.Assignee,
		}
	}

	// Set reporter if provided
	if input.Reporter != "" {
		issue.Fields.Reporter = &gojira.User{
			Name: input.Reporter,
		}
	}

	// Set components if provided
	if len(input.Components) > 0 {
		for _, c := range input.Components {
			issue.Fields.Components = append(issue.Fields.Components, &gojira.Component{
				Name: c,
			})
		}
	}

	// Set fix versions if provided
	if len(input.FixVersions) > 0 {
		for _, v := range input.FixVersions {
			issue.Fields.FixVersions = append(issue.Fields.FixVersions, &gojira.FixVersion{
				Name: v,
			})
		}
	}

	// Handle custom fields
	customFields := input.GetCustomFields()
	if len(customFields) > 0 {
		if issue.Fields.Unknowns == nil {
			issue.Fields.Unknowns = make(map[string]any)
		}
		for k, v := range customFields {
			issue.Fields.Unknowns[k] = v
		}
	}

	// Create the issue
	created, resp, err := client.JiraClient.Issue.CreateWithContext(ctx, issue)
	if err != nil {
		if resp != nil && resp.StatusCode >= 400 {
			return nil, fmt.Errorf("create issue failed (status %d): %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("create issue: %w", err)
	}

	return &IssueResult{
		Key:     created.Key,
		ID:      created.ID,
		Self:    created.Self,
		Summary: input.Summary,
	}, nil
}

func validateInput(input *IssueInput) error {
	var missing []string

	if input.Project == "" {
		missing = append(missing, "project")
	}
	if input.Type == "" {
		missing = append(missing, "type")
	}
	if input.Summary == "" {
		missing = append(missing, "summary")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

// DryRunCreate validates input and returns what would be created without actually creating.
func DryRunCreate(input *IssueInput) (*DryRunResult, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}

	return &DryRunResult{
		Project:      input.Project,
		Type:         input.Type,
		Summary:      input.Summary,
		Description:  input.Description,
		Parent:       input.Parent,
		Labels:       input.Labels,
		Priority:     input.Priority,
		Assignee:     input.Assignee,
		CustomFields: input.GetCustomFields(),
		Valid:        true,
	}, nil
}

// DryRunResult shows what would be created.
type DryRunResult struct {
	Valid        bool           `json:"valid"`
	Project      string         `json:"project"`
	Type         string         `json:"type"`
	Summary      string         `json:"summary"`
	Description  string         `json:"description,omitempty"`
	Parent       string         `json:"parent,omitempty"`
	Labels       []string       `json:"labels,omitempty"`
	Priority     string         `json:"priority,omitempty"`
	Assignee     string         `json:"assignee,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
}

// ValidationResult contains the result of validating input against createmeta.
type ValidationResult struct {
	Valid             bool     `json:"valid"`
	MissingRequired   []string `json:"missing_required,omitempty"`
	UnknownFields     []string `json:"unknown_fields,omitempty"`
	AvailableFields   []string `json:"available_fields,omitempty"`
	IssueTypeID       string   `json:"issue_type_id,omitempty"`
	RequiredFieldsMsg string   `json:"required_fields_message,omitempty"`
}

// ValidateAgainstMeta validates IssueInput against the Jira createmeta API.
// It checks that all required fields are provided and that field names are valid.
func ValidateAgainstMeta(ctx context.Context, client *jira.Client, input *IssueInput) (*ValidationResult, error) {
	// Basic validation first
	if err := validateInput(input); err != nil {
		return nil, err
	}

	result := &ValidationResult{
		Valid: true,
	}

	// Get issue types for the project
	issueTypes, err := client.CreateMetaAPI.GetIssueTypes(ctx, input.Project)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue types for project %q: %w", input.Project, err)
	}

	// Find the matching issue type
	var issueTypeID string
	for _, it := range issueTypes {
		if strings.EqualFold(it.Name, input.Type) {
			issueTypeID = it.ID
			break
		}
	}

	if issueTypeID == "" {
		availableTypes := make([]string, len(issueTypes))
		for i, it := range issueTypes {
			availableTypes[i] = it.Name
		}
		return nil, fmt.Errorf("issue type %q not found in project %q; available: %s",
			input.Type, input.Project, strings.Join(availableTypes, ", "))
	}

	result.IssueTypeID = issueTypeID

	// Get fields for this project/issue type combination
	fields, err := client.CreateMetaAPI.GetFields(ctx, input.Project, issueTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fields for issue type: %w", err)
	}

	// Build map of available fields
	availableByKey := fields.ByKey()
	for _, f := range fields {
		result.AvailableFields = append(result.AvailableFields, f.Key)
	}

	// Check required fields
	providedFields := getProvidedFieldKeys(input)

	for _, f := range fields {
		if f.Required && !f.HasDefaultValue {
			if _, provided := providedFields[f.Key]; !provided {
				result.MissingRequired = append(result.MissingRequired, fmt.Sprintf("%s (%s)", f.Key, f.Name))
				result.Valid = false
			}
		}
	}

	// Check for unknown custom fields
	customFields := input.GetCustomFields()
	for key := range customFields {
		if _, exists := availableByKey[key]; !exists {
			result.UnknownFields = append(result.UnknownFields, key)
			result.Valid = false
		}
	}

	// Build summary message
	if len(result.MissingRequired) > 0 {
		result.RequiredFieldsMsg = fmt.Sprintf("Missing required fields: %s", strings.Join(result.MissingRequired, ", "))
	}

	return result, nil
}

// getProvidedFieldKeys returns a set of field keys that are provided in the input.
func getProvidedFieldKeys(input *IssueInput) map[string]bool {
	provided := make(map[string]bool)

	// Standard fields
	if input.Summary != "" {
		provided["summary"] = true
	}
	if input.Description != "" {
		provided["description"] = true
	}
	if input.Project != "" {
		provided["project"] = true
	}
	if input.Parent != "" {
		provided["parent"] = true
	}
	if len(input.Labels) > 0 {
		provided["labels"] = true
	}
	if input.Priority != "" {
		provided["priority"] = true
	}
	if input.Assignee != "" {
		provided["assignee"] = true
	}
	if input.Reporter != "" {
		provided["reporter"] = true
	}
	if len(input.Components) > 0 {
		provided["components"] = true
	}
	if len(input.FixVersions) > 0 {
		provided["fixVersions"] = true
	}
	// issuetype is always provided via input.Type
	provided["issuetype"] = true

	// Custom fields
	for key := range input.GetCustomFields() {
		provided[key] = true
	}

	return provided
}
