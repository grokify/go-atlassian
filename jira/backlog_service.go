package jira

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/grokify/mogo/net/http/httpsimple"
)

// project = foundation AND resolution = Unresolved AND status!=Closed AND (Sprint not in openSprints() OR Sprint is EMPTY) AND type not in (Epic, Sub-Task) ORDER BY Rank ASC
/*
func BacklogJQL(projectName string) string {
	parts := []string{
		"resolution = Unresolved",
		"status!=Closed",
		"(Sprint not in openSprints() OR Sprint is EMPTY)",
		"type not in (Epic, Sub-Task)",
	}
	projectName = strings.TrimSpace(projectName)
	if projectName != "" {
		// parts = []string{projectName, parts... }
	}
}
*/

const (
	ParamFields        = "fields"
	ParamJQL           = "jql"
	ParamMaxResults    = "maxResults"
	ParamStartAt       = "startAt"
	ParamValidateQuery = "validateQuery"
)

type BoardBacklogParams struct {
	StartAt       int    `url:"startAt"`
	MaxResults    int    `url:"maxResults"`
	JQL           string `url:"jql"`
	ValidateQuery bool   `url:"validateQuery"`
	Fields        string `url:"fields"`
	Expand        string `url:"expand"`
}

func (p BoardBacklogParams) URLValues() url.Values {
	u := url.Values{}
	if p.StartAt > 0 {
		u.Add(ParamStartAt, strconv.Itoa(p.StartAt))
	}
	if p.MaxResults > 0 {
		u.Add(ParamMaxResults, strconv.Itoa(p.MaxResults))
	}
	jql := strings.TrimSpace(p.JQL)
	if jql != "" {
		u.Add(ParamJQL, jql)
		if p.ValidateQuery {
			u.Add(ParamValidateQuery, "true")
		} else {
			u.Add(ParamValidateQuery, "false) ")
		}
	}
	fields := strings.TrimSpace(p.Fields)
	if fields != "" {
		u.Add(ParamFields, fields)
	}
	return u
}

// BacklogAPIURL returns a backlog issues API URL described at https://docs.atlassian.com/jira-software/REST/7.3.1/ .
// The description is here: Returns all issues from the board's backlog, for the given board Id. This only includes issues that the user has permission to view. The backlog contains incomplete issues that are not assigned to any future or active sprint. Note, if the user does not have permission to view the board, no issues will be returned at all. Issues returned from this resource include Agile fields, like sprint, closedSprints, flagged, and epic. By default, the returned issues are ordered by rank.
// Reference: https://docs.atlassian.com/jira-software/REST/7.3.1/#agile/1.0/board-getIssuesForBacklog
func BacklogAPIURL(baseURL string, boardID uint, qry *BoardBacklogParams) string {
	apiURL := baseURL + fmt.Sprintf(`/rest/agile/1.0/board/%d/backlog`, boardID)
	if qry != nil {
		v := qry.URLValues()
		q := strings.TrimSpace(v.Encode())
		if q != "" {
			apiURL += "?" + q
		}
	}
	return apiURL
}

type BacklogService struct {
	Client  *Client
	sclient httpsimple.Client
}

func NewBacklogService(client *Client) *BacklogService {
	return &BacklogService{
		Client: client,
		sclient: httpsimple.Client{
			HTTPClient: client.HTTPClient,
			BaseURL:    client.Config.ServerURL}}
}

func (svc *BacklogService) GetBacklogIssuesResponse(ctx context.Context, boardID uint, qry *BoardBacklogParams) (*IssuesResponse, []byte, error) {
	if svc.sclient.HTTPClient == nil {
		return nil, []byte{}, errors.New("client not set")
	}
	sreq := httpsimple.Request{
		Method: http.MethodGet,
		URL:    BacklogAPIURL(svc.Client.Config.ServerURL, boardID, nil),
		Query:  qry.URLValues(),
	}
	resp, err := svc.sclient.Do(ctx, sreq)
	if err != nil {
		return nil, []byte{}, err
	} else if resp.StatusCode >= 300 {
		return nil, []byte{}, fmt.Errorf("statusCode (%d)", resp.StatusCode)
	}
	if b, err := io.ReadAll(resp.Body); err != nil {
		return nil, []byte{}, err
	} else {
		ir, err := ParseIssuesResponseBytes(b)
		return ir, b, err
	}
}

func (svc *BacklogService) GetBacklogIssuesAll(ctx context.Context, boardID uint, jql string) (*IssuesResponse, [][]byte, error) {
	iragg := &IssuesResponse{}
	bb := [][]byte{}
	issues := Issues{}

	opts := &BoardBacklogParams{
		StartAt:    0,
		MaxResults: MaxResults,
		JQL:        jql,
	}

	for {
		ir, b, err := svc.GetBacklogIssuesResponse(ctx, boardID, opts)
		if err != nil {
			return iragg, bb, err
		} else {
			iragg.Total = ir.Total
			iragg.MaxResults = ir.Total
		}
		bb = append(bb, b)
		if len(ir.Issues) > 0 {
			issues = append(issues, ir.Issues...)
		} else {
			break
		}
		if opts.StartAt+opts.MaxResults > ir.Total {
			break
		} else {
			opts.StartAt += opts.MaxResults
		}
	}
	iragg.Issues = issues.AddRank()
	return iragg, bb, nil
}

func (svc *BacklogService) GetBacklogIssuesSetAll(ctx context.Context, boardID uint, jql string) (*IssuesSet, [][]byte, error) {
	iir, b, err := svc.GetBacklogIssuesAll(ctx, boardID, jql)
	if err != nil {
		return nil, b, err
	}
	is := NewIssuesSet(svc.Client.Config)
	err = is.Add(iir.Issues...)
	return is, b, err
}
