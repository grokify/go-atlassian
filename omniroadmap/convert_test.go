package omniroadmap

import (
	"testing"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/omniroadmap-core/provider"
)

func TestStatusFromJiraStatus(t *testing.T) {
	tests := []struct {
		name string
		s    *gojira.Status
		want provider.StatusCategory
	}{
		{"nil", nil, ""},
		{"todo", &gojira.Status{Name: "Backlog", StatusCategory: gojira.StatusCategory{Key: gojira.StatusCategoryToDo}}, provider.StatusCategoryTodo},
		{"in progress", &gojira.Status{Name: "Evaluating", StatusCategory: gojira.StatusCategory{Key: gojira.StatusCategoryInProgress}}, provider.StatusCategoryInProgress},
		{"done", &gojira.Status{Name: "Shipped", StatusCategory: gojira.StatusCategory{Key: gojira.StatusCategoryComplete}}, provider.StatusCategoryDone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusFromJiraStatus(tt.s)
			if tt.s == nil {
				if got != nil {
					t.Fatalf("statusFromJiraStatus(nil) = %+v, want nil", got)
				}
				return
			}
			if got.Category != tt.want {
				t.Errorf("Category = %q, want %q", got.Category, tt.want)
			}
		})
	}
}

func TestItemFromIssue(t *testing.T) {
	issue := &gojira.Issue{
		Key:  "PROJ-1",
		Self: "https://test.atlassian.net/rest/api/2/issue/10001",
		Fields: &gojira.IssueFields{
			Summary:     "Dark mode",
			Description: "Users want a dark theme",
			Type:        gojira.IssueType{Name: "Idea"},
			Project:     gojira.Project{Key: "PROJ"},
			Status:      &gojira.Status{Name: "Backlog", StatusCategory: gojira.StatusCategory{Key: gojira.StatusCategoryToDo}},
			Labels:      []string{"ui"},
		},
	}

	item := itemFromIssue(issue)

	if item.ID != "jpd:PROJ-1" {
		t.Errorf("ID = %q, want %q", item.ID, "jpd:PROJ-1")
	}
	if item.Provider != "jpd" {
		t.Errorf("Provider = %q, want %q", item.Provider, "jpd")
	}
	if item.Kind != provider.ItemKindFeature {
		t.Errorf("Kind = %q, want %q", item.Kind, provider.ItemKindFeature)
	}
	if item.Name != "Dark mode" {
		t.Errorf("Name = %q, want %q", item.Name, "Dark mode")
	}
	if item.Status == nil || item.Status.Category != provider.StatusCategoryTodo {
		t.Errorf("Status = %+v, want Category=todo", item.Status)
	}
	want := "https://test.atlassian.net/browse/PROJ-1"
	if item.SourceURL != want {
		t.Errorf("SourceURL = %q, want %q", item.SourceURL, want)
	}
	if item.Metadata["jpd.issue_type"] != "Idea" {
		t.Errorf("Metadata[jpd.issue_type] = %v, want %q", item.Metadata["jpd.issue_type"], "Idea")
	}
	if len(item.Tags) != 1 || item.Tags[0] != "ui" {
		t.Errorf("Tags = %v, want [ui]", item.Tags)
	}
}

func TestItemFromIssue_NilFields(t *testing.T) {
	issue := &gojira.Issue{Key: "PROJ-2"}
	item := itemFromIssue(issue)
	if item.SourceID != "PROJ-2" {
		t.Errorf("SourceID = %q, want %q", item.SourceID, "PROJ-2")
	}
	if item.Name != "" {
		t.Errorf("Name = %q, want empty", item.Name)
	}
}
