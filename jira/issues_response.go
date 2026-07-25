package jira

import (
	"encoding/json"
	"io"
	"os"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/mogo/encoding/jsonutil"
)

// IssuesResponse is only a small wrapper around the Search (with JQL) method to be able to parse the results
type IssuesResponse struct {
	Issues     Issues `json:"issues" structs:"issues"`
	Expand     string `json:"expand"`
	StartAt    int    `json:"startAt" structs:"startAt"`
	MaxResults int    `json:"maxResults" structs:"maxResults"`
	Total      int    `json:"total" structs:"total"`
}

func IssuesResponseReadFile(filename string) (*IssuesResponse, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseIssuesResponseReader(f)
}

func ParseIssuesResponseReader(r io.Reader) (*IssuesResponse, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseIssuesResponseBytes(b)
}

func ParseIssuesResponseBytes(b []byte) (*IssuesResponse, error) {
	ir := IssuesResponse{}
	return &ir, json.Unmarshal(b, &ir)
}

func GetIssueCustomValueStruct(iss gojira.Issue) (*IssueCustomField, error) {
	if iss.Fields == nil {
		return nil, nil
	}
	unv, ok := iss.Fields.Unknowns["customfield_12461"]
	if !ok {
		return nil, nil
	}
	icf := &IssueCustomField{}
	err := jsonutil.UnmarshalAny(unv, icf)
	return icf, err
}
