package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gojira "github.com/andygrunwald/go-jira"
)

func TestGetQueryOptionsBuild(t *testing.T) {
	tests := []struct {
		name    string
		opts    GetQueryOptions
		wantExp string
	}{
		{
			name:    "empty options",
			opts:    GetQueryOptions{},
			wantExp: "",
		},
		{
			name:    "expand changelog",
			opts:    GetQueryOptions{ExpandChangelog: true},
			wantExp: "changelog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.Build()
			if got.Expand != tt.wantExp {
				t.Errorf("GetQueryOptions.Build() Expand = %q, want %q", got.Expand, tt.wantExp)
			}
		})
	}
}

func TestIssueServiceIssue(t *testing.T) {
	// Create a mock server that returns a valid issue
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issue := gojira.Issue{
			Key: "TEST-123",
			Fields: &gojira.IssueFields{
				Summary: "Test Issue Summary",
				Type:    gojira.IssueType{Name: "Bug"},
				Status:  &gojira.Status{Name: "Open"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(issue); err != nil {
			t.Fatalf("failed to encode issue: %v", err)
		}
	}))
	defer server.Close()

	// Create a client pointing to our test server
	jiraClient, err := gojira.NewClient(nil, server.URL)
	if err != nil {
		t.Fatalf("failed to create jira client: %v", err)
	}

	client := &Client{
		JiraClient: jiraClient,
	}
	// Create IssueService directly without calling Inflate
	svc := NewIssueService(client)

	// Test getting an issue
	issue, err := svc.Issue(context.Background(), "TEST-123", nil)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if issue.Key != "TEST-123" {
		t.Errorf("Issue() Key = %q, want %q", issue.Key, "TEST-123")
	}
	if issue.Fields.Summary != "Test Issue Summary" {
		t.Errorf("Issue() Summary = %q, want %q", issue.Fields.Summary, "Test Issue Summary")
	}
}

func TestIssueServiceIssueEmptyKey(t *testing.T) {
	client := &Client{}
	svc := NewIssueService(client)

	_, err := svc.Issue(context.Background(), "", nil)
	if err == nil {
		t.Error("Issue() with empty key should return error")
	}
}

func TestIssueServiceIssueNilClient(t *testing.T) {
	svc := &IssueService{Client: nil}

	_, err := svc.Issue(context.Background(), "TEST-123", nil)
	if err == nil {
		t.Error("Issue() with nil client should return error")
	}
}

func TestIssueServiceIssuesEmptyKeys(t *testing.T) {
	client := &Client{}
	svc := NewIssueService(client)

	issues, err := svc.Issues(context.Background(), []string{}, nil)
	if err != nil {
		t.Fatalf("Issues() with empty keys error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Issues() with empty keys returned %d issues, want 0", len(issues))
	}
}

func TestIssueServiceIssuesWhitespaceKeys(t *testing.T) {
	client := &Client{}
	svc := NewIssueService(client)

	// Keys with only whitespace should be treated as empty
	issues, err := svc.Issues(context.Background(), []string{"  ", "\t", "\n"}, nil)
	if err != nil {
		t.Fatalf("Issues() with whitespace keys error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Issues() with whitespace keys returned %d issues, want 0", len(issues))
	}
}

func TestNewIssueService(t *testing.T) {
	client := &Client{}
	svc := NewIssueService(client)

	if svc == nil {
		t.Fatal("NewIssueService() returned nil")
	}
	if svc.Client != client {
		t.Error("NewIssueService() client mismatch")
	}
}
