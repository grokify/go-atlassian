package omniroadmap

import (
	"context"
	"fmt"

	jira "github.com/grokify/go-atlassian/jira"
	omniroadmap "github.com/grokify/omniroadmap-core"
	"github.com/grokify/omniroadmap-core/provider"
)

// ListItems searches for JPD Idea issues in the configured project.
// Requires WithProjectKey to have been set — without a project scope,
// searching would return every issue the credentials can see, not just
// Ideas, so this returns omniroadmap.ErrUnsupportedOperation instead.
//
// go-atlassian's SearchIssuesAPIV3 doesn't take page/perPage — it always
// retrieves all matching issues internally. req.Page/PerPage are applied
// as an in-memory slice afterward, which is fine for the thin v0.1 scope
// but not efficient for very large idea backlogs.
func (p *Provider) ListItems(ctx context.Context, req *provider.ListItemsRequest) (*provider.ListItemsResponse, error) {
	if p.projectKey == "" {
		return nil, omniroadmap.ErrUnsupportedOperation
	}
	if len(req.Kinds) > 0 && !containsKind(req.Kinds, provider.ItemKindFeature) {
		return &provider.ListItemsResponse{}, nil
	}

	jql := fmt.Sprintf(`project = %q AND issuetype = %q`, p.projectKey, p.ideaIssueType)
	issues, err := p.client.IssueAPI.SearchIssuesAPIV3(ctx, jql, true)
	if err != nil {
		return nil, wrapErr("SearchIssuesAPIV3", err)
	}

	items := make([]provider.Item, len(issues))
	for i := range issues {
		items[i] = *itemFromIssue(&issues[i])
	}
	items = paginate(items, req.Page, req.PerPage)

	return &provider.ListItemsResponse{
		Items:      items,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalCount: len(issues),
	}, nil
}

// GetItem fetches a single Idea issue by its Jira key (e.g. "PROJ-123").
func (p *Provider) GetItem(ctx context.Context, req *provider.GetItemRequest) (*provider.Item, error) {
	if req.Kind != provider.ItemKindFeature {
		return nil, omniroadmap.ErrUnsupportedOperation
	}
	issue, err := p.client.IssueAPI.Issue(ctx, req.ID, &jira.GetQueryOptions{})
	if err != nil {
		return nil, wrapErr("Issue", err)
	}
	return itemFromIssue(issue), nil
}

func containsKind(kinds []provider.ItemKind, kind provider.ItemKind) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}

// paginate applies a simple in-memory page/perPage slice. page is 1-indexed;
// page<=0 or perPage<=0 returns items unchanged (no pagination requested).
func paginate(items []provider.Item, page, perPage int) []provider.Item {
	if page <= 0 || perPage <= 0 {
		return items
	}
	start := (page - 1) * perPage
	if start >= len(items) {
		return nil
	}
	end := start + perPage
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func wrapErr(op string, err error) error {
	return &opError{op: op, err: err}
}

type opError struct {
	op  string
	err error
}

func (e *opError) Error() string { return "omniroadmap/jpd: " + e.op + ": " + e.err.Error() }
func (e *opError) Unwrap() error { return e.err }
