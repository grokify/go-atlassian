package jira

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/mogo/type/slicesutil"
	"github.com/grokify/mogo/type/stringsutil"
)

type IssueService struct {
	Client *Client
}

func NewIssueService(client *Client) *IssueService {
	return &IssueService{Client: client}
}

type GetQueryOptions struct {
	ExpandChangelog    bool // sent to andygrunwald SDK
	XMultiSkipNotFound bool // not sent to andygrunwald SDK; used for getting multiple issues
	XIncludeParents    bool
	// XMultiRecursive    bool
}

// Build returns a `*gojira.GetQueryOptions` for the andygrunwald SDK.
func (opts GetQueryOptions) Build() *gojira.GetQueryOptions {
	out := &gojira.GetQueryOptions{}
	if opts.ExpandChangelog {
		out.Expand = "changelog"
	}
	return out
}

func (svc *IssueService) Issue(ctx context.Context, issueIDOrKey string, opts *GetQueryOptions) (*gojira.Issue, error) {
	issueIDOrKey = strings.TrimSpace(issueIDOrKey)

	var opts2 *gojira.GetQueryOptions
	if opts != nil {
		opts2 = opts.Build()
	}
	if issueIDOrKey == "" {
		return nil, errors.New("issue key cannot be empty")
	} else if svc.Client == nil {
		return nil, errors.New("Client cannot be nil")
	} else if svc.Client.JiraClient == nil {
		return nil, errors.New("Client.JiraClient cannot be nil")
	} else if svc.Client.JiraClient.Issue == nil {
		return nil, errors.New("Client.JiraClient.issue cannot be nil")
	} else if iss, resp, err := svc.Client.JiraClient.Issue.GetWithContext(ctx, issueIDOrKey, opts2); err != nil {
		return nil, err
	} else if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unsuccessful jira api http status code (%d)", resp.StatusCode)
	} else {
		return iss, nil
	}
}

// Issues returns a list of issues given a set of keys. If no keys are provided,
// an empty slice is returned. The opts parameter is reserved for future use and
// currently has no effect on the search query.
func (svc *IssueService) Issues(ctx context.Context, keys []string, _ *GetQueryOptions) (Issues, error) {
	keys = stringsutil.SliceCondenseSpace(keys, true, true)
	if len(keys) == 0 {
		return Issues{}, nil
	}
	j := JQL{
		IssuesIncl: [][]string{keys},
	}
	return svc.SearchIssuesAPIV3(ctx, j.String(), true)
}

// Issues returns an `IssuesSet{}` given a set of keys. If no keys are provided,
// any empty slice is returned.
func (svc *IssueService) GetIssuesSetForKeys(keys []string) (*IssuesSet, error) {
	is := NewIssuesSet(nil)

	keysSlice := slicesutil.SplitMaxLength(stringsutil.SliceCondenseSpace(keys, true, true), JQLMaxResults)

	for _, keysIter := range keysSlice {
		keysIter = stringsutil.SliceCondenseSpace(keysIter, true, true)
		if len(keysIter) == 0 {
			continue
		}
		jqlInfo := JQL{KeysIncl: [][]string{keysIter}}
		if jql := jqlInfo.String(); jql == "" {
			continue
		} else if ii, err := svc.SearchIssuesPages(jql, 0, 0, 0); err != nil {
			return nil, err
		} else if err = is.Add(ii...); err != nil {
			return nil, err
		}
	}

	return is, nil
}
