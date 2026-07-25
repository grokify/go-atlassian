package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// WebhookEventType represents the type of webhook event.
type WebhookEventType string

// Jira webhook event types.
const (
	WebhookEventIssueCreated          WebhookEventType = "jira:issue_created"
	WebhookEventIssueUpdated          WebhookEventType = "jira:issue_updated"
	WebhookEventIssueDeleted          WebhookEventType = "jira:issue_deleted"
	WebhookEventCommentCreated        WebhookEventType = "comment_created"
	WebhookEventCommentUpdated        WebhookEventType = "comment_updated"
	WebhookEventCommentDeleted        WebhookEventType = "comment_deleted"
	WebhookEventWorklogCreated        WebhookEventType = "worklog_created"
	WebhookEventWorklogUpdated        WebhookEventType = "worklog_updated"
	WebhookEventWorklogDeleted        WebhookEventType = "worklog_deleted"
	WebhookEventSprintCreated         WebhookEventType = "sprint_created"
	WebhookEventSprintUpdated         WebhookEventType = "sprint_updated"
	WebhookEventSprintDeleted         WebhookEventType = "sprint_deleted"
	WebhookEventSprintStarted         WebhookEventType = "sprint_started"
	WebhookEventSprintClosed          WebhookEventType = "sprint_closed"
	WebhookEventBoardCreated          WebhookEventType = "board_created"
	WebhookEventBoardUpdated          WebhookEventType = "board_updated"
	WebhookEventBoardDeleted          WebhookEventType = "board_deleted"
	WebhookEventProjectCreated        WebhookEventType = "project_created"
	WebhookEventProjectUpdated        WebhookEventType = "project_updated"
	WebhookEventProjectDeleted        WebhookEventType = "project_deleted"
	WebhookEventUserCreated           WebhookEventType = "user_created"
	WebhookEventUserUpdated           WebhookEventType = "user_updated"
	WebhookEventUserDeleted           WebhookEventType = "user_deleted"
	WebhookEventVersionCreated        WebhookEventType = "jira:version_created"
	WebhookEventVersionUpdated        WebhookEventType = "jira:version_updated"
	WebhookEventVersionDeleted        WebhookEventType = "jira:version_deleted"
	WebhookEventVersionReleased       WebhookEventType = "jira:version_released"
	WebhookEventVersionUnreleased     WebhookEventType = "jira:version_unreleased"
	WebhookEventIssueLinkCreated      WebhookEventType = "issuelink_created"
	WebhookEventIssueLinkDeleted      WebhookEventType = "issuelink_deleted"
	WebhookEventAttachmentCreated     WebhookEventType = "attachment_created"
	WebhookEventAttachmentDeleted     WebhookEventType = "attachment_deleted"
	WebhookEventOptionVotingChanged   WebhookEventType = "option_voting_changed"
	WebhookEventOptionWatchingChanged WebhookEventType = "option_watching_changed"
)

// WebhookEvent represents a Jira webhook event.
type WebhookEvent struct {
	WebhookEvent WebhookEventType   `json:"webhookEvent"`
	Timestamp    int64              `json:"timestamp"`
	User         *WebhookUser       `json:"user,omitempty"`
	Issue        *WebhookIssue      `json:"issue,omitempty"`
	Comment      *WebhookComment    `json:"comment,omitempty"`
	Changelog    *WebhookChangelog  `json:"changelog,omitempty"`
	Sprint       *WebhookSprint     `json:"sprint,omitempty"`
	Board        *WebhookBoard      `json:"board,omitempty"`
	Project      *WebhookProject    `json:"project,omitempty"`
	Version      *WebhookVersion    `json:"version,omitempty"`
	Worklog      *WebhookWorklog    `json:"worklog,omitempty"`
	IssueLink    *WebhookIssueLink  `json:"issueLink,omitempty"`
	Attachment   *WebhookAttachment `json:"attachment,omitempty"`
	RawPayload   json.RawMessage    `json:"-"`
}

// Time returns the timestamp as a time.Time.
func (e *WebhookEvent) Time() time.Time {
	return time.UnixMilli(e.Timestamp)
}

// WebhookUser represents a user in a webhook event.
type WebhookUser struct {
	Self         string `json:"self,omitempty"`
	AccountID    string `json:"accountId,omitempty"`
	Key          string `json:"key,omitempty"`
	Name         string `json:"name,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	Active       bool   `json:"active"`
}

// WebhookIssue represents an issue in a webhook event.
type WebhookIssue struct {
	ID     string              `json:"id"`
	Key    string              `json:"key"`
	Self   string              `json:"self,omitempty"`
	Fields *WebhookIssueFields `json:"fields,omitempty"`
}

// WebhookIssueFields represents issue fields in a webhook event.
type WebhookIssueFields struct {
	Summary     string           `json:"summary,omitempty"`
	Description string           `json:"description,omitempty"`
	Status      *WebhookStatus   `json:"status,omitempty"`
	IssueType   *WebhookType     `json:"issuetype,omitempty"`
	Priority    *WebhookPriority `json:"priority,omitempty"`
	Assignee    *WebhookUser     `json:"assignee,omitempty"`
	Reporter    *WebhookUser     `json:"reporter,omitempty"`
	Project     *WebhookProject  `json:"project,omitempty"`
	Created     string           `json:"created,omitempty"`
	Updated     string           `json:"updated,omitempty"`
	Labels      []string         `json:"labels,omitempty"`
}

// WebhookStatus represents a status in a webhook event.
type WebhookStatus struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// WebhookType represents an issue type in a webhook event.
type WebhookType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Subtask     bool   `json:"subtask"`
}

// WebhookPriority represents a priority in a webhook event.
type WebhookPriority struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// WebhookComment represents a comment in a webhook event.
type WebhookComment struct {
	ID      string       `json:"id"`
	Self    string       `json:"self,omitempty"`
	Body    string       `json:"body,omitempty"`
	Author  *WebhookUser `json:"author,omitempty"`
	Created string       `json:"created,omitempty"`
	Updated string       `json:"updated,omitempty"`
}

// WebhookChangelog represents a changelog in a webhook event.
type WebhookChangelog struct {
	ID    string                 `json:"id"`
	Items []WebhookChangelogItem `json:"items,omitempty"`
}

// WebhookChangelogItem represents a single change in a changelog.
type WebhookChangelogItem struct {
	Field      string `json:"field"`
	FieldType  string `json:"fieldtype,omitempty"`
	FieldID    string `json:"fieldId,omitempty"`
	From       string `json:"from,omitempty"`
	FromString string `json:"fromString,omitempty"`
	To         string `json:"to,omitempty"`
	ToString   string `json:"toString,omitempty"`
}

// WebhookSprint represents a sprint in a webhook event.
type WebhookSprint struct {
	ID            int    `json:"id"`
	Name          string `json:"name,omitempty"`
	State         string `json:"state,omitempty"`
	OriginBoardID int    `json:"originBoardId,omitempty"`
	StartDate     string `json:"startDate,omitempty"`
	EndDate       string `json:"endDate,omitempty"`
	CompleteDate  string `json:"completeDate,omitempty"`
}

// WebhookBoard represents a board in a webhook event.
type WebhookBoard struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// WebhookProject represents a project in a webhook event.
type WebhookProject struct {
	ID          string `json:"id,omitempty"`
	Key         string `json:"key,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ProjectType string `json:"projectTypeKey,omitempty"`
}

// WebhookVersion represents a version in a webhook event.
type WebhookVersion struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Released    bool   `json:"released"`
	Archived    bool   `json:"archived"`
	ReleaseDate string `json:"releaseDate,omitempty"`
}

// WebhookWorklog represents a worklog in a webhook event.
type WebhookWorklog struct {
	ID               string       `json:"id"`
	Self             string       `json:"self,omitempty"`
	Author           *WebhookUser `json:"author,omitempty"`
	UpdateAuthor     *WebhookUser `json:"updateAuthor,omitempty"`
	Comment          string       `json:"comment,omitempty"`
	Started          string       `json:"started,omitempty"`
	TimeSpent        string       `json:"timeSpent,omitempty"`
	TimeSpentSeconds int          `json:"timeSpentSeconds,omitempty"`
}

// WebhookIssueLink represents an issue link in a webhook event.
type WebhookIssueLink struct {
	ID            string           `json:"id,omitempty"`
	SourceIssueID string           `json:"sourceIssueId,omitempty"`
	DestIssueID   string           `json:"destinationIssueId,omitempty"`
	LinkType      *WebhookLinkType `json:"issueLinkType,omitempty"`
}

// WebhookLinkType represents a link type in a webhook event.
type WebhookLinkType struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Inward  string `json:"inward,omitempty"`
	Outward string `json:"outward,omitempty"`
}

// WebhookAttachment represents an attachment in a webhook event.
type WebhookAttachment struct {
	ID       string       `json:"id,omitempty"`
	Filename string       `json:"filename,omitempty"`
	Author   *WebhookUser `json:"author,omitempty"`
	Created  string       `json:"created,omitempty"`
	Size     int64        `json:"size,omitempty"`
	MimeType string       `json:"mimeType,omitempty"`
	Content  string       `json:"content,omitempty"`
}

// EventHandler is a function that handles a webhook event.
type EventHandler func(event *WebhookEvent) error

// WebhookService handles incoming Jira webhook events.
type WebhookService struct {
	handlers       map[WebhookEventType][]EventHandler
	globalHandlers []EventHandler
	mu             sync.RWMutex
	filters        []WebhookFilter
}

// WebhookFilter filters incoming webhook events.
type WebhookFilter struct {
	Projects   []string // Only process events for these projects
	IssueTypes []string // Only process events for these issue types
	Users      []string // Only process events from these users
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService() *WebhookService {
	return &WebhookService{
		handlers:       make(map[WebhookEventType][]EventHandler),
		globalHandlers: []EventHandler{},
	}
}

// On registers a handler for a specific event type.
func (s *WebhookService) On(eventType WebhookEventType, handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[eventType] = append(s.handlers[eventType], handler)
}

// OnAll registers a handler for all event types.
func (s *WebhookService) OnAll(handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalHandlers = append(s.globalHandlers, handler)
}

// OnIssue registers a handler for all issue-related events.
func (s *WebhookService) OnIssue(handler EventHandler) {
	s.On(WebhookEventIssueCreated, handler)
	s.On(WebhookEventIssueUpdated, handler)
	s.On(WebhookEventIssueDeleted, handler)
}

// OnComment registers a handler for all comment-related events.
func (s *WebhookService) OnComment(handler EventHandler) {
	s.On(WebhookEventCommentCreated, handler)
	s.On(WebhookEventCommentUpdated, handler)
	s.On(WebhookEventCommentDeleted, handler)
}

// OnSprint registers a handler for all sprint-related events.
func (s *WebhookService) OnSprint(handler EventHandler) {
	s.On(WebhookEventSprintCreated, handler)
	s.On(WebhookEventSprintUpdated, handler)
	s.On(WebhookEventSprintDeleted, handler)
	s.On(WebhookEventSprintStarted, handler)
	s.On(WebhookEventSprintClosed, handler)
}

// OnWorklog registers a handler for all worklog-related events.
func (s *WebhookService) OnWorklog(handler EventHandler) {
	s.On(WebhookEventWorklogCreated, handler)
	s.On(WebhookEventWorklogUpdated, handler)
	s.On(WebhookEventWorklogDeleted, handler)
}

// AddFilter adds a filter for webhook events.
func (s *WebhookService) AddFilter(filter WebhookFilter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filters = append(s.filters, filter)
}

// matchesFilters checks if an event matches any registered filter.
func (s *WebhookService) matchesFilters(event *WebhookEvent) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// If no filters, accept all events
	if len(s.filters) == 0 {
		return true
	}

	for _, filter := range s.filters {
		if s.matchesFilter(event, filter) {
			return true
		}
	}
	return false
}

func (s *WebhookService) matchesFilter(event *WebhookEvent, filter WebhookFilter) bool {
	// Check project filter
	if len(filter.Projects) > 0 {
		projectKey := ""
		if event.Issue != nil && event.Issue.Fields != nil && event.Issue.Fields.Project != nil {
			projectKey = event.Issue.Fields.Project.Key
		} else if event.Project != nil {
			projectKey = event.Project.Key
		}
		if projectKey == "" || !contains(filter.Projects, projectKey) {
			return false
		}
	}

	// Check issue type filter
	if len(filter.IssueTypes) > 0 {
		issueType := ""
		if event.Issue != nil && event.Issue.Fields != nil && event.Issue.Fields.IssueType != nil {
			issueType = event.Issue.Fields.IssueType.Name
		}
		if issueType == "" || !contains(filter.IssueTypes, issueType) {
			return false
		}
	}

	// Check user filter
	if len(filter.Users) > 0 {
		userName := ""
		if event.User != nil {
			userName = event.User.Name
			if userName == "" {
				userName = event.User.AccountID
			}
		}
		if userName == "" || !contains(filter.Users, userName) {
			return false
		}
	}

	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// HandleEvent processes a webhook event by dispatching to registered handlers.
func (s *WebhookService) HandleEvent(event *WebhookEvent) error {
	if !s.matchesFilters(event) {
		return nil // Event filtered out
	}

	s.mu.RLock()
	handlers := s.handlers[event.WebhookEvent]
	globalHandlers := s.globalHandlers
	s.mu.RUnlock()

	// Call global handlers first
	for _, handler := range globalHandlers {
		if err := handler(event); err != nil {
			return fmt.Errorf("global handler error: %w", err)
		}
	}

	// Call event-specific handlers
	for _, handler := range handlers {
		if err := handler(event); err != nil {
			return fmt.Errorf("handler error for %s: %w", event.WebhookEvent, err)
		}
	}

	return nil
}

// ParseWebhookEvent parses a webhook event from JSON.
func ParseWebhookEvent(data []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("parse webhook event: %w", err)
	}
	event.RawPayload = data
	return &event, nil
}

// HTTPHandler returns an http.HandlerFunc that processes incoming webhook requests.
func (s *WebhookService) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		event, err := ParseWebhookEvent(body)
		if err != nil {
			http.Error(w, "Failed to parse webhook event", http.StatusBadRequest)
			return
		}

		if err := s.HandleEvent(event); err != nil {
			http.Error(w, fmt.Sprintf("Failed to handle event: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
