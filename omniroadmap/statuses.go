package omniroadmap

import (
	"context"

	"github.com/grokify/omniroadmap-core/provider"
)

// ListStatuses lists all workflow statuses visible to the credentials, via
// go-jira's unscoped GetAllStatusesWithContext — a genuine "all statuses"
// endpoint, no project scoping required (unlike ListItems). req.Kind is
// ignored: Jira's statuses API has no per-issue-type filter.
func (p *Provider) ListStatuses(ctx context.Context, req *provider.ListStatusesRequest) (*provider.ListStatusesResponse, error) {
	jiraStatuses, _, err := p.client.JiraClient.Status.GetAllStatusesWithContext(ctx)
	if err != nil {
		return nil, wrapErr("GetAllStatuses", err)
	}

	statuses := make([]provider.Status, len(jiraStatuses))
	for i := range jiraStatuses {
		s := statusFromJiraStatus(&jiraStatuses[i])
		statuses[i] = *s
	}
	return &provider.ListStatusesResponse{Statuses: statuses}, nil
}
