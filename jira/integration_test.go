package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	gojira "github.com/andygrunwald/go-jira"
)

// Integration test environment variables.
// When these are set, tests run against a real Jira server.
// When not set, tests use a mock server.
const (
	envJiraURL   = "JIRA_URL"
	envJiraUser  = "JIRA_USER"
	envJiraToken = "JIRA_TOKEN"
	// Optional: specify a test issue key for get operations
	envJiraTestIssue = "JIRA_TEST_ISSUE"
	// Optional: specify a test project for search operations
	envJiraTestProject = "JIRA_TEST_PROJECT"
)

// testEnv holds the test environment configuration.
type testEnv struct {
	useLive     bool
	jiraURL     string
	jiraUser    string
	jiraToken   string
	testIssue   string
	testProject string
	mockServer  *httptest.Server
}

// newTestEnv creates a test environment. If live credentials are set,
// it configures for live testing. Otherwise, it starts a mock server.
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	env := &testEnv{
		jiraURL:     os.Getenv(envJiraURL),
		jiraUser:    os.Getenv(envJiraUser),
		jiraToken:   os.Getenv(envJiraToken),
		testIssue:   os.Getenv(envJiraTestIssue),
		testProject: os.Getenv(envJiraTestProject),
	}

	// Check if we have live credentials
	if env.jiraURL != "" && env.jiraUser != "" && env.jiraToken != "" {
		env.useLive = true
		t.Log("Using live Jira server for integration tests")
	} else {
		env.useLive = false
		env.mockServer = newMockJiraServer(t)
		env.jiraURL = env.mockServer.URL
		env.jiraUser = "testuser"
		env.jiraToken = "testtoken"
		t.Log("Using mock Jira server for integration tests")
	}

	// Set defaults for test data
	if env.testIssue == "" {
		env.testIssue = "TEST-123"
	}
	if env.testProject == "" {
		env.testProject = "TEST"
	}

	return env
}

// cleanup closes the mock server if one was created.
func (e *testEnv) cleanup() {
	if e.mockServer != nil {
		e.mockServer.Close()
	}
}

// newClient creates a Jira client for the test environment.
func (e *testEnv) newClient(t *testing.T) *Client {
	t.Helper()

	client, err := NewClientFromBasicAuth(e.jiraURL, e.jiraUser, e.jiraToken, false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client
}

// newMockJiraServer creates a mock Jira server for testing.
func newMockJiraServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Mock issue endpoint
	mux.HandleFunc("/rest/api/2/issue/", func(w http.ResponseWriter, r *http.Request) {
		// Extract issue key from path
		path := strings.TrimPrefix(r.URL.Path, "/rest/api/2/issue/")
		issueKey := strings.Split(path, "/")[0]

		issue := gojira.Issue{
			Key: issueKey,
			Fields: &gojira.IssueFields{
				Summary: "Mock Issue Summary for " + issueKey,
				Type:    gojira.IssueType{Name: "Bug"},
				Status:  &gojira.Status{Name: "Open"},
				Project: gojira.Project{Key: "TEST", Name: "Test Project"},
				Assignee: &gojira.User{
					DisplayName: "Test User",
					Name:        "testuser",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(issue); err != nil {
			t.Errorf("failed to encode mock issue: %v", err)
		}
	})

	// Mock search response helper
	mockSearchResponse := func(w http.ResponseWriter) {
		response := struct {
			StartAt       int            `json:"startAt"`
			MaxResults    int            `json:"maxResults"`
			Total         int            `json:"total"`
			IsLast        bool           `json:"isLast"`
			NextPageToken string         `json:"nextPageToken"`
			Issues        []gojira.Issue `json:"issues"`
		}{
			StartAt:       0,
			MaxResults:    50,
			Total:         2,
			IsLast:        true,
			NextPageToken: "",
			Issues: []gojira.Issue{
				{
					Key: "TEST-1",
					Fields: &gojira.IssueFields{
						Summary: "First test issue",
						Type:    gojira.IssueType{Name: "Bug"},
						Status:  &gojira.Status{Name: "Open"},
						Project: gojira.Project{Key: "TEST", Name: "Test Project"},
					},
				},
				{
					Key: "TEST-2",
					Fields: &gojira.IssueFields{
						Summary: "Second test issue",
						Type:    gojira.IssueType{Name: "Story"},
						Status:  &gojira.Status{Name: "In Progress"},
						Project: gojira.Project{Key: "TEST", Name: "Test Project"},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("failed to encode mock search response: %v", err)
		}
	}

	// Mock search endpoint (v2) - handles multiple paths
	mux.HandleFunc("/rest/api/2/search", func(w http.ResponseWriter, r *http.Request) {
		mockSearchResponse(w)
	})

	// Mock search endpoint for jql path
	mux.HandleFunc("/rest/api/2/search/jql", func(w http.ResponseWriter, r *http.Request) {
		mockSearchResponse(w)
	})

	// Mock search endpoint (v3)
	mux.HandleFunc("/rest/api/3/search/jql", func(w http.ResponseWriter, r *http.Request) {
		mockSearchResponse(w)
	})

	// Mock custom fields endpoint
	mux.HandleFunc("/rest/api/2/field", func(w http.ResponseWriter, r *http.Request) {
		fields := []CustomField{
			{
				ID:     "customfield_10001",
				Name:   "Epic Link",
				Custom: true,
			},
			{
				ID:     "customfield_10002",
				Name:   "Sprint",
				Custom: true,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fields); err != nil {
			t.Errorf("failed to encode mock fields response: %v", err)
		}
	})

	return httptest.NewServer(mux)
}

// TestIntegrationGetIssue tests fetching a single issue.
func TestIntegrationGetIssue(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()

	client := env.newClient(t)
	ctx := context.Background()

	issue, err := client.IssueAPI.Issue(ctx, env.testIssue, nil)
	if err != nil {
		if env.useLive {
			t.Fatalf("failed to get issue %s: %v", env.testIssue, err)
		} else {
			t.Fatalf("failed to get mock issue: %v", err)
		}
	}

	if issue.Key == "" {
		t.Error("issue key is empty")
	}

	t.Logf("Retrieved issue: %s - %s", issue.Key, issue.Fields.Summary)
}

// TestIntegrationGetIssueWithExpand tests fetching an issue with changelog expansion.
func TestIntegrationGetIssueWithExpand(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()

	client := env.newClient(t)
	ctx := context.Background()

	opts := &GetQueryOptions{
		ExpandChangelog: true,
	}

	issue, err := client.IssueAPI.Issue(ctx, env.testIssue, opts)
	if err != nil {
		t.Fatalf("failed to get issue with expand: %v", err)
	}

	if issue.Key == "" {
		t.Error("issue key is empty")
	}

	t.Logf("Retrieved issue with expand: %s", issue.Key)
}

// TestIntegrationSearchIssues tests searching for issues with JQL.
func TestIntegrationSearchIssues(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()

	client := env.newClient(t)

	jql := "project = " + env.testProject
	if env.useLive {
		// Limit results for live testing
		jql += " ORDER BY created DESC"
	}

	issues, err := client.IssueAPI.SearchIssues(jql, false)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	t.Logf("Found %d issues", len(issues))

	if len(issues) == 0 && !env.useLive {
		t.Error("expected mock server to return issues")
	}

	// Log first few issues
	for i, issue := range issues {
		if i >= 5 {
			break
		}
		im := NewIssueMore(&issue)
		t.Logf("  %s: %s [%s]", im.Key(), im.Summary(), im.Status())
	}
}

// TestIntegrationSearchIssuesOnPremise tests the on-premise search function.
func TestIntegrationSearchIssuesOnPremise(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()

	// Skip for live Jira Cloud - this function is for on-premise only
	if env.useLive && strings.Contains(env.jiraURL, "atlassian.net") {
		t.Skip("Skipping on-premise test for Jira Cloud")
	}

	client := env.newClient(t)

	jql := "project = " + env.testProject

	issues, err := client.IssueAPI.SearchIssuesOnPremise(jql, true)
	if err != nil {
		// This may fail on cloud, which is expected
		if env.useLive {
			t.Skipf("SearchIssuesOnPremise not supported on this server: %v", err)
		}
		t.Fatalf("search failed: %v", err)
	}

	t.Logf("Found %d issues using on-premise API", len(issues))
}

// TestIntegrationIssueMore tests the IssueMore wrapper with real/mock data.
func TestIntegrationIssueMore(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()

	client := env.newClient(t)
	ctx := context.Background()

	issue, err := client.IssueAPI.Issue(ctx, env.testIssue, nil)
	if err != nil {
		t.Fatalf("failed to get issue: %v", err)
	}

	im := NewIssueMore(issue)

	// Test various IssueMore methods
	if key := im.Key(); key == "" {
		t.Error("Key() returned empty")
	} else {
		t.Logf("Key: %s", key)
	}

	if typ := im.Type(); typ == "" {
		t.Error("Type() returned empty")
	} else {
		t.Logf("Type: %s", typ)
	}

	if status := im.Status(); status == "" {
		t.Error("Status() returned empty")
	} else {
		t.Logf("Status: %s", status)
	}

	t.Logf("Summary: %s", im.Summary())
	t.Logf("Project: %s", im.Project())
	t.Logf("Assignee: %s", im.AssigneeName())
}

// TestIntegrationIssuesSet tests the IssuesSet aggregation.
func TestIntegrationIssuesSet(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()

	client := env.newClient(t)

	jql := "project = " + env.testProject

	issues, err := client.IssueAPI.SearchIssues(jql, false)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(issues) == 0 {
		if env.useLive {
			t.Skip("No issues found in project")
		}
		t.Fatal("expected mock issues")
	}

	// Convert to IssuesSet
	issuesSet, err := issues.IssuesSet(nil)
	if err != nil {
		t.Fatalf("failed to create IssuesSet: %v", err)
	}

	t.Logf("IssuesSet contains %d issues", issuesSet.Len())

	// Test Keys
	keys := issuesSet.Keys()
	t.Logf("Keys: %v", keys)

	// Test counts by type
	countsByType := issues.CountsByType()
	t.Logf("Counts by type: %v", countsByType)
}
