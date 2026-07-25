package jira

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewWebhookService(t *testing.T) {
	svc := NewWebhookService()
	if svc == nil {
		t.Fatal("NewWebhookService() returned nil")
	}
	if svc.handlers == nil {
		t.Error("handlers map should be initialized")
	}
}

func TestWebhookServiceOn(t *testing.T) {
	svc := NewWebhookService()
	called := false

	svc.On(WebhookEventIssueCreated, func(event *WebhookEvent) error {
		called = true
		return nil
	})

	event := &WebhookEvent{
		WebhookEvent: WebhookEventIssueCreated,
	}

	if err := svc.HandleEvent(event); err != nil {
		t.Fatalf("HandleEvent() error = %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestWebhookServiceOnAll(t *testing.T) {
	svc := NewWebhookService()
	callCount := 0

	svc.OnAll(func(event *WebhookEvent) error {
		callCount++
		return nil
	})

	events := []*WebhookEvent{
		{WebhookEvent: WebhookEventIssueCreated},
		{WebhookEvent: WebhookEventIssueUpdated},
		{WebhookEvent: WebhookEventCommentCreated},
	}

	for _, event := range events {
		if err := svc.HandleEvent(event); err != nil {
			t.Fatalf("HandleEvent() error = %v", err)
		}
	}

	if callCount != 3 {
		t.Errorf("Global handler called %d times, want 3", callCount)
	}
}

func TestWebhookServiceOnIssue(t *testing.T) {
	svc := NewWebhookService()
	callCount := 0

	svc.OnIssue(func(event *WebhookEvent) error {
		callCount++
		return nil
	})

	events := []*WebhookEvent{
		{WebhookEvent: WebhookEventIssueCreated},
		{WebhookEvent: WebhookEventIssueUpdated},
		{WebhookEvent: WebhookEventIssueDeleted},
		{WebhookEvent: WebhookEventCommentCreated}, // Should not trigger
	}

	for _, event := range events {
		if err := svc.HandleEvent(event); err != nil {
			t.Fatalf("HandleEvent() error = %v", err)
		}
	}

	if callCount != 3 {
		t.Errorf("Issue handler called %d times, want 3", callCount)
	}
}

func TestParseWebhookEvent(t *testing.T) {
	payload := `{
		"webhookEvent": "jira:issue_created",
		"timestamp": 1699286400000,
		"issue": {
			"id": "12345",
			"key": "TEST-123",
			"fields": {
				"summary": "Test Issue",
				"status": {"name": "Open"}
			}
		},
		"user": {
			"name": "testuser",
			"displayName": "Test User"
		}
	}`

	event, err := ParseWebhookEvent([]byte(payload))
	if err != nil {
		t.Fatalf("ParseWebhookEvent() error = %v", err)
	}

	if event.WebhookEvent != WebhookEventIssueCreated {
		t.Errorf("WebhookEvent = %q, want %q", event.WebhookEvent, WebhookEventIssueCreated)
	}

	if event.Issue == nil {
		t.Fatal("Issue should not be nil")
	}

	if event.Issue.Key != "TEST-123" {
		t.Errorf("Issue.Key = %q, want %q", event.Issue.Key, "TEST-123")
	}

	if event.Issue.Fields.Summary != "Test Issue" {
		t.Errorf("Issue.Fields.Summary = %q, want %q", event.Issue.Fields.Summary, "Test Issue")
	}

	if event.User == nil || event.User.Name != "testuser" {
		t.Error("User not parsed correctly")
	}
}

func TestParseWebhookEventInvalid(t *testing.T) {
	_, err := ParseWebhookEvent([]byte("invalid json"))
	if err == nil {
		t.Error("ParseWebhookEvent() should return error for invalid JSON")
	}
}

func TestWebhookEventTime(t *testing.T) {
	event := &WebhookEvent{
		Timestamp: 1699286400000, // 2023-11-06 16:00:00 UTC
	}

	ts := event.Time()
	if ts.Year() != 2023 || ts.Month() != 11 || ts.Day() != 6 {
		t.Errorf("Time() = %v, want 2023-11-06", ts)
	}
}

func TestWebhookServiceFilter(t *testing.T) {
	svc := NewWebhookService()
	called := false

	svc.AddFilter(WebhookFilter{
		Projects: []string{"TEST"},
	})

	svc.On(WebhookEventIssueCreated, func(event *WebhookEvent) error {
		called = true
		return nil
	})

	// Event with matching project
	event1 := &WebhookEvent{
		WebhookEvent: WebhookEventIssueCreated,
		Issue: &WebhookIssue{
			Key: "TEST-123",
			Fields: &WebhookIssueFields{
				Project: &WebhookProject{Key: "TEST"},
			},
		},
	}

	if err := svc.HandleEvent(event1); err != nil {
		t.Fatalf("HandleEvent() error = %v", err)
	}

	if !called {
		t.Error("Handler should be called for matching project")
	}

	// Reset and test non-matching project
	called = false
	event2 := &WebhookEvent{
		WebhookEvent: WebhookEventIssueCreated,
		Issue: &WebhookIssue{
			Key: "OTHER-123",
			Fields: &WebhookIssueFields{
				Project: &WebhookProject{Key: "OTHER"},
			},
		},
	}

	if err := svc.HandleEvent(event2); err != nil {
		t.Fatalf("HandleEvent() error = %v", err)
	}

	if called {
		t.Error("Handler should NOT be called for non-matching project")
	}
}

func TestWebhookServiceHTTPHandler(t *testing.T) {
	svc := NewWebhookService()
	handlerCalled := false

	svc.On(WebhookEventIssueCreated, func(event *WebhookEvent) error {
		handlerCalled = true
		if event.Issue == nil || event.Issue.Key != "TEST-123" {
			t.Error("Event not parsed correctly")
		}
		return nil
	})

	payload := `{
		"webhookEvent": "jira:issue_created",
		"timestamp": 1699286400000,
		"issue": {
			"key": "TEST-123"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := svc.HTTPHandler()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("HTTP status = %d, want %d", rec.Code, http.StatusOK)
	}

	if !handlerCalled {
		t.Error("Event handler was not called")
	}
}

func TestWebhookServiceHTTPHandlerInvalidMethod(t *testing.T) {
	svc := NewWebhookService()
	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	rec := httptest.NewRecorder()

	handler := svc.HTTPHandler()
	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("HTTP status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestWebhookServiceHTTPHandlerInvalidJSON(t *testing.T) {
	svc := NewWebhookService()
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString("invalid"))
	rec := httptest.NewRecorder()

	handler := svc.HTTPHandler()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("HTTP status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestWebhookChangelogParsing(t *testing.T) {
	payload := `{
		"webhookEvent": "jira:issue_updated",
		"timestamp": 1699286400000,
		"issue": {"key": "TEST-123"},
		"changelog": {
			"id": "10001",
			"items": [
				{
					"field": "status",
					"fromString": "Open",
					"toString": "In Progress"
				},
				{
					"field": "assignee",
					"fromString": null,
					"toString": "testuser"
				}
			]
		}
	}`

	event, err := ParseWebhookEvent([]byte(payload))
	if err != nil {
		t.Fatalf("ParseWebhookEvent() error = %v", err)
	}

	if event.Changelog == nil {
		t.Fatal("Changelog should not be nil")
	}

	if len(event.Changelog.Items) != 2 {
		t.Errorf("Changelog.Items length = %d, want 2", len(event.Changelog.Items))
	}

	if event.Changelog.Items[0].Field != "status" {
		t.Errorf("First changelog item field = %q, want status", event.Changelog.Items[0].Field)
	}

	if event.Changelog.Items[0].FromString != "Open" {
		t.Errorf("First changelog item fromString = %q, want Open", event.Changelog.Items[0].FromString)
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "Banana", "CHERRY"}

	tests := []struct {
		item string
		want bool
	}{
		{"apple", true},
		{"APPLE", true},
		{"banana", true},
		{"cherry", true},
		{"grape", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.item, func(t *testing.T) {
			got := contains(slice, tt.item)
			if got != tt.want {
				t.Errorf("contains(%q) = %v, want %v", tt.item, got, tt.want)
			}
		})
	}
}

func TestWebhookCommentParsing(t *testing.T) {
	payload := `{
		"webhookEvent": "comment_created",
		"timestamp": 1699286400000,
		"issue": {"key": "TEST-123"},
		"comment": {
			"id": "10001",
			"body": "This is a test comment",
			"author": {
				"name": "testuser",
				"displayName": "Test User"
			}
		}
	}`

	event, err := ParseWebhookEvent([]byte(payload))
	if err != nil {
		t.Fatalf("ParseWebhookEvent() error = %v", err)
	}

	if event.Comment == nil {
		t.Fatal("Comment should not be nil")
	}

	if event.Comment.Body != "This is a test comment" {
		t.Errorf("Comment.Body = %q, want 'This is a test comment'", event.Comment.Body)
	}

	if event.Comment.Author == nil || event.Comment.Author.Name != "testuser" {
		t.Error("Comment author not parsed correctly")
	}
}

func TestRawPayloadPreserved(t *testing.T) {
	payload := `{"webhookEvent":"jira:issue_created","timestamp":1699286400000}`

	event, err := ParseWebhookEvent([]byte(payload))
	if err != nil {
		t.Fatalf("ParseWebhookEvent() error = %v", err)
	}

	if event.RawPayload == nil {
		t.Fatal("RawPayload should be preserved")
	}

	// Verify it can be re-parsed
	var decoded map[string]any
	if err := json.Unmarshal(event.RawPayload, &decoded); err != nil {
		t.Fatalf("Failed to decode RawPayload: %v", err)
	}

	if decoded["webhookEvent"] != "jira:issue_created" {
		t.Errorf("RawPayload webhookEvent = %v, want jira:issue_created", decoded["webhookEvent"])
	}
}
