package jira

import (
	"context"
	"fmt"

	gojira "github.com/andygrunwald/go-jira"
)

// GetComments fetches comments for an issue and returns them in a simplified format.
// maxResults limits the number of comments returned; if <= 0, all comments are returned.
func (c *Client) GetComments(ctx context.Context, issueKey string, maxResults int) (*CommentsResponse, error) {
	if c.JiraClient == nil {
		return nil, fmt.Errorf("jira client not initialized")
	}

	// Get issue with renderedFields expansion to include comments
	issue, _, err := c.JiraClient.Issue.GetWithContext(ctx, issueKey, &gojira.GetQueryOptions{
		Expand: "renderedFields",
	})
	if err != nil {
		return nil, fmt.Errorf("get issue %s: %w", issueKey, err)
	}

	response := &CommentsResponse{
		Key:      issueKey,
		Total:    0,
		Comments: []CommentResult{},
	}

	if issue.Fields.Comments == nil || len(issue.Fields.Comments.Comments) == 0 {
		return response, nil
	}

	comments := issue.Fields.Comments.Comments
	if maxResults > 0 && len(comments) > maxResults {
		comments = comments[:maxResults]
	}

	response.Total = len(comments)
	response.Comments = make([]CommentResult, 0, len(comments))

	for _, c := range comments {
		response.Comments = append(response.Comments, ToCommentResult(c))
	}

	return response, nil
}
