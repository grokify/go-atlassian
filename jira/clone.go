package jira

import (
	"context"
	"fmt"

	gojira "github.com/andygrunwald/go-jira"
)

// CloneOptions configures how an issue is cloned.
type CloneOptions struct {
	// TargetProject overrides the project for the cloned issue.
	// If empty, uses the source issue's project.
	TargetProject string

	// TargetType overrides the issue type for the cloned issue.
	// If empty, uses the source issue's type.
	TargetType string

	// SummaryPrefix is prepended to the cloned issue's summary.
	// Common prefixes: "[Clone] ", "Copy of: "
	SummaryPrefix string

	// SummarySuffix is appended to the cloned issue's summary.
	SummarySuffix string

	// ExcludeFields lists field names to exclude from cloning.
	// Standard fields: "description", "labels", "priority", "components", "fixVersions"
	ExcludeFields []string

	// IncludeOnlyFields, if non-empty, limits cloning to only these fields.
	// Takes precedence over ExcludeFields.
	IncludeOnlyFields []string

	// LinkToOriginal creates a "clones/is cloned by" link to the source issue.
	LinkToOriginal bool

	// LinkType specifies the link type name when LinkToOriginal is true.
	// Default: "Cloners" (creates "clones" relationship)
	LinkType string

	// CustomFieldValues allows overriding specific custom field values.
	CustomFieldValues map[string]any

	// Parent sets the parent issue key for the cloned issue.
	// Useful when cloning a subtask to a different parent.
	Parent string
}

// CloneResult contains the result of a clone operation.
type CloneResult struct {
	SourceKey string `json:"source_key"`
	ClonedKey string `json:"cloned_key"`
	ClonedID  string `json:"cloned_id"`
	Self      string `json:"self"`
	Linked    bool   `json:"linked,omitempty"`
}

// CloneIssue creates a copy of an existing issue with configurable field mapping.
func (svc *IssueService) CloneIssue(ctx context.Context, sourceKey string, opts *CloneOptions) (*CloneResult, error) {
	if svc.Client == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	if opts == nil {
		opts = &CloneOptions{}
	}

	// Get the source issue
	source, err := svc.Issue(ctx, sourceKey, nil)
	if err != nil {
		return nil, fmt.Errorf("get source issue %s: %w", sourceKey, err)
	}

	if source.Fields == nil {
		return nil, fmt.Errorf("source issue %s has no fields", sourceKey)
	}

	// Build the new issue
	newIssue := &gojira.Issue{
		Fields: &gojira.IssueFields{},
	}

	// Set project
	projectKey := source.Fields.Project.Key
	if opts.TargetProject != "" {
		projectKey = opts.TargetProject
	}
	newIssue.Fields.Project = gojira.Project{Key: projectKey}

	// Set issue type
	issueTypeName := source.Fields.Type.Name
	if opts.TargetType != "" {
		issueTypeName = opts.TargetType
	}
	newIssue.Fields.Type = gojira.IssueType{Name: issueTypeName}

	// Set summary with optional prefix/suffix
	summary := source.Fields.Summary
	if opts.SummaryPrefix != "" {
		summary = opts.SummaryPrefix + summary
	}
	if opts.SummarySuffix != "" {
		summary = summary + opts.SummarySuffix
	}
	newIssue.Fields.Summary = summary

	// Build field inclusion/exclusion sets
	excludeSet := make(map[string]bool)
	for _, f := range opts.ExcludeFields {
		excludeSet[f] = true
	}

	includeSet := make(map[string]bool)
	for _, f := range opts.IncludeOnlyFields {
		includeSet[f] = true
	}
	useIncludeList := len(includeSet) > 0

	// Helper to check if field should be included
	shouldInclude := func(field string) bool {
		if useIncludeList {
			return includeSet[field]
		}
		return !excludeSet[field]
	}

	// Copy standard fields
	if shouldInclude("description") && source.Fields.Description != "" {
		newIssue.Fields.Description = source.Fields.Description
	}

	if shouldInclude("labels") && len(source.Fields.Labels) > 0 {
		newIssue.Fields.Labels = source.Fields.Labels
	}

	if shouldInclude("priority") && source.Fields.Priority != nil {
		newIssue.Fields.Priority = &gojira.Priority{Name: source.Fields.Priority.Name}
	}

	if shouldInclude("components") && len(source.Fields.Components) > 0 {
		for _, c := range source.Fields.Components {
			newIssue.Fields.Components = append(newIssue.Fields.Components, &gojira.Component{Name: c.Name})
		}
	}

	if shouldInclude("fixVersions") && len(source.Fields.FixVersions) > 0 {
		for _, v := range source.Fields.FixVersions {
			newIssue.Fields.FixVersions = append(newIssue.Fields.FixVersions, &gojira.FixVersion{Name: v.Name})
		}
	}

	// Set parent
	if opts.Parent != "" {
		newIssue.Fields.Parent = &gojira.Parent{Key: opts.Parent}
	} else if shouldInclude("parent") && source.Fields.Parent != nil {
		newIssue.Fields.Parent = &gojira.Parent{Key: source.Fields.Parent.Key}
	}

	// Copy custom fields
	if source.Fields.Unknowns != nil {
		newIssue.Fields.Unknowns = make(map[string]any)
		for key, value := range source.Fields.Unknowns {
			if shouldInclude(key) {
				newIssue.Fields.Unknowns[key] = value
			}
		}
	}

	// Apply custom field overrides
	if len(opts.CustomFieldValues) > 0 {
		if newIssue.Fields.Unknowns == nil {
			newIssue.Fields.Unknowns = make(map[string]any)
		}
		for key, value := range opts.CustomFieldValues {
			newIssue.Fields.Unknowns[key] = value
		}
	}

	// Create the cloned issue
	created, resp, err := svc.Client.JiraClient.Issue.CreateWithContext(ctx, newIssue)
	if err != nil {
		return nil, fmt.Errorf("create cloned issue: %w", err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("create cloned issue: status %d", resp.StatusCode)
	}

	result := &CloneResult{
		SourceKey: sourceKey,
		ClonedKey: created.Key,
		ClonedID:  created.ID,
		Self:      created.Self,
	}

	// Create link to original if requested
	if opts.LinkToOriginal {
		linkType := opts.LinkType
		if linkType == "" {
			linkType = "Cloners"
		}

		link := &gojira.IssueLink{
			Type: gojira.IssueLinkType{Name: linkType},
			OutwardIssue: &gojira.Issue{
				Key: sourceKey,
			},
			InwardIssue: &gojira.Issue{
				Key: created.Key,
			},
		}

		_, err := svc.Client.JiraClient.Issue.AddLinkWithContext(ctx, link)
		if err != nil {
			// Don't fail the whole operation, just note the link failed
			result.Linked = false
		} else {
			result.Linked = true
		}
	}

	return result, nil
}
