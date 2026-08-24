package omniroadmap

import (
	"fmt"
	"strings"
	"time"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/omniroadmap-core/provider"
)

func canonicalID(sourceID string) string {
	return fmt.Sprintf("%s:%s", providerName, sourceID)
}

// itemFromIssue converts a Jira issue (a JPD Idea) to a canonical Item.
// JPD Ideas map to ItemKindFeature — they're candidate features being
// evaluated, the closest fit among omniroadmap-core's ItemKinds. The
// original Jira issue type name is preserved in Metadata.
func itemFromIssue(issue *gojira.Issue) *provider.Item {
	item := &provider.Item{
		ID:        canonicalID(issue.Key),
		Provider:  providerName,
		SourceID:  issue.Key,
		SourceRef: issue.Key,
		Kind:      provider.ItemKindFeature,
	}
	if issue.Self != "" {
		item.SourceURL = browseURLFromSelf(issue.Self, issue.Key)
	}
	if issue.Fields == nil {
		return item
	}

	item.Name = issue.Fields.Summary
	item.Description = issue.Fields.Description
	item.Status = statusFromJiraStatus(issue.Fields.Status)

	if t := time.Time(issue.Fields.Created); !t.IsZero() {
		item.CreatedAt = &t
	}
	if t := time.Time(issue.Fields.Updated); !t.IsZero() {
		item.UpdatedAt = &t
	}
	if issue.Fields.Assignee != nil {
		item.Owner = &provider.Person{
			ID:    issue.Fields.Assignee.AccountID,
			Name:  issue.Fields.Assignee.DisplayName,
			Email: issue.Fields.Assignee.EmailAddress,
		}
	}
	item.Tags = issue.Fields.Labels

	item.Metadata = map[string]any{
		"jpd.issue_type": issue.Fields.Type.Name,
	}
	if issue.Fields.Project.Key != "" {
		item.Metadata["jpd.project_key"] = issue.Fields.Project.Key
	}

	return item
}

// browseURLFromSelf derives a human-facing browse URL from the issue's API
// "self" link (e.g. "https://x.atlassian.net/rest/api/2/issue/12345" ->
// "https://x.atlassian.net/browse/KEY-1"). Best-effort: falls back to
// empty if self doesn't have the expected REST API shape.
func browseURLFromSelf(self, key string) string {
	const marker = "/rest/api/"
	idx := strings.Index(self, marker)
	if idx < 0 {
		return ""
	}
	return self[:idx] + "/browse/" + key
}

// statusFromJiraStatus converts a Jira Status to a canonical Status,
// normalizing via Jira's own status category key ("new"/"indeterminate"/
// "done") — a clean, authoritative mapping (unlike Aha, which has no
// equivalent server-side category and needs a name/position heuristic).
func statusFromJiraStatus(s *gojira.Status) *provider.Status {
	if s == nil {
		return nil
	}
	category := provider.StatusCategoryInProgress
	switch s.StatusCategory.Key {
	case gojira.StatusCategoryToDo:
		category = provider.StatusCategoryTodo
	case gojira.StatusCategoryComplete:
		category = provider.StatusCategoryDone
	case gojira.StatusCategoryInProgress:
		category = provider.StatusCategoryInProgress
	}
	return &provider.Status{
		ID:       s.ID,
		Name:     s.Name,
		Category: category,
		Complete: s.StatusCategory.Key == gojira.StatusCategoryComplete,
	}
}
