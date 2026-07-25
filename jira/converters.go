package jira

import (
	"time"

	gojira "github.com/andygrunwald/go-jira"
)

const timeFormatRFC3339 = "2006-01-02T15:04:05Z"

// IssueOutput is a simplified issue representation for CLI and API output.
// It provides a consistent format for both CLI commands and MCP server responses.
type IssueOutput struct {
	Key          string         `json:"key"`
	Summary      string         `json:"summary"`
	Description  string         `json:"description,omitempty"`
	Status       string         `json:"status"`
	Type         string         `json:"type"`
	Priority     string         `json:"priority,omitempty"`
	Resolution   string         `json:"resolution,omitempty"`
	Assignee     string         `json:"assignee,omitempty"`
	Reporter     string         `json:"reporter,omitempty"`
	Creator      string         `json:"creator,omitempty"`
	Labels       []string       `json:"labels,omitempty"`
	Created      string         `json:"created,omitempty"`
	Updated      string         `json:"updated,omitempty"`
	Project      string         `json:"project,omitempty"`
	ProjectKey   string         `json:"projectKey,omitempty"`
	Parent       string         `json:"parent,omitempty"`
	EpicKey      string         `json:"epicKey,omitempty"`
	CustomFields map[string]any `json:"customFields,omitempty"`
}

// ToIssueOutput converts a Jira issue to a simplified output format.
func ToIssueOutput(issue *gojira.Issue) IssueOutput {
	if issue == nil {
		return IssueOutput{}
	}

	result := IssueOutput{
		Key: issue.Key,
	}

	if issue.Fields == nil {
		return result
	}

	result.Summary = issue.Fields.Summary
	result.Description = issue.Fields.Description
	result.Labels = issue.Fields.Labels

	if issue.Fields.Status != nil {
		result.Status = issue.Fields.Status.Name
	}
	if issue.Fields.Type.Name != "" {
		result.Type = issue.Fields.Type.Name
	}
	if issue.Fields.Priority != nil {
		result.Priority = issue.Fields.Priority.Name
	}
	if issue.Fields.Resolution != nil {
		result.Resolution = issue.Fields.Resolution.Name
	}
	if issue.Fields.Assignee != nil {
		result.Assignee = issue.Fields.Assignee.DisplayName
	}
	if issue.Fields.Reporter != nil {
		result.Reporter = issue.Fields.Reporter.DisplayName
	}
	if issue.Fields.Creator != nil {
		result.Creator = issue.Fields.Creator.DisplayName
	}
	if issue.Fields.Project.Key != "" {
		result.ProjectKey = issue.Fields.Project.Key
	}
	if issue.Fields.Project.Name != "" {
		result.Project = issue.Fields.Project.Name
	}
	if issue.Fields.Parent != nil {
		result.Parent = issue.Fields.Parent.Key
	}
	if issue.Fields.Epic != nil {
		result.EpicKey = issue.Fields.Epic.Key
	}
	if !time.Time(issue.Fields.Created).IsZero() {
		result.Created = time.Time(issue.Fields.Created).Format(timeFormatRFC3339)
	}
	if !time.Time(issue.Fields.Updated).IsZero() {
		result.Updated = time.Time(issue.Fields.Updated).Format(timeFormatRFC3339)
	}

	return result
}

// ToIssueOutputs converts multiple Jira issues to simplified output format.
func ToIssueOutputs(issues Issues) []IssueOutput {
	results := make([]IssueOutput, 0, len(issues))
	for i := range issues {
		results = append(results, ToIssueOutput(&issues[i]))
	}
	return results
}

// CommentResult represents a single comment for output.
type CommentResult struct {
	ID      string `json:"id"`
	Author  string `json:"author"`
	Body    string `json:"body"`
	Created string `json:"created"`
	Updated string `json:"updated,omitempty"`
}

// CommentsResponse represents the response for a comments request.
type CommentsResponse struct {
	Key      string          `json:"key"`
	Total    int             `json:"total"`
	Comments []CommentResult `json:"comments"`
}

// ToCommentResult converts a Jira comment to output format.
func ToCommentResult(comment *gojira.Comment) CommentResult {
	if comment == nil {
		return CommentResult{}
	}

	result := CommentResult{
		ID:      comment.ID,
		Body:    comment.Body,
		Created: comment.Created,
		Updated: comment.Updated,
	}

	if comment.Author.DisplayName != "" {
		result.Author = comment.Author.DisplayName
	}

	return result
}

// ToCommentResults converts Jira comments to output format with optional limit.
// If maxResults <= 0, all comments are returned.
func ToCommentResults(comments []*gojira.Comment, maxResults int) []CommentResult {
	if maxResults > 0 && len(comments) > maxResults {
		comments = comments[:maxResults]
	}

	results := make([]CommentResult, 0, len(comments))
	for _, c := range comments {
		results = append(results, ToCommentResult(c))
	}
	return results
}
