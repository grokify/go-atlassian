package jira

import (
	"context"
	"testing"

	gojira "github.com/andygrunwald/go-jira"
)

func TestNewWorkflowService(t *testing.T) {
	client := &Client{}
	svc := NewWorkflowService(client)

	if svc == nil {
		t.Fatal("NewWorkflowService() returned nil")
	}
	if svc.Client != client {
		t.Error("NewWorkflowService() client mismatch")
	}
}

func TestWorkflowServiceNilClient(t *testing.T) {
	svc := &WorkflowService{Client: nil}

	template := &WorkflowTemplate{
		Name: "test",
		Steps: []WorkflowStep{
			{Type: "comment", Comment: "test"},
		},
	}

	_, err := svc.ExecuteWorkflow(context.Background(), "TEST-123", template)
	if err == nil {
		t.Error("ExecuteWorkflow() with nil client should return error")
	}
}

func TestWorkflowServiceEmptyIssueKey(t *testing.T) {
	client := &Client{}
	svc := NewWorkflowService(client)

	template := &WorkflowTemplate{
		Name: "test",
		Steps: []WorkflowStep{
			{Type: "comment", Comment: "test"},
		},
	}

	_, err := svc.ExecuteWorkflow(context.Background(), "", template)
	if err == nil {
		t.Error("ExecuteWorkflow() with empty issue key should return error")
	}
}

func TestWorkflowServiceEmptyTemplate(t *testing.T) {
	client := &Client{}
	svc := NewWorkflowService(client)

	_, err := svc.ExecuteWorkflow(context.Background(), "TEST-123", nil)
	if err == nil {
		t.Error("ExecuteWorkflow() with nil template should return error")
	}

	_, err = svc.ExecuteWorkflow(context.Background(), "TEST-123", &WorkflowTemplate{
		Name:  "empty",
		Steps: []WorkflowStep{},
	})
	if err == nil {
		t.Error("ExecuteWorkflow() with empty steps should return error")
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		slice []string
		item  string
		want  bool
	}{
		{[]string{"Open", "Closed"}, "Open", true},
		{[]string{"Open", "Closed"}, "open", true},
		{[]string{"Open", "Closed"}, "OPEN", true},
		{[]string{"Open", "Closed"}, "In Progress", false},
		{[]string{}, "Open", false},
	}

	for _, tt := range tests {
		got := containsIgnoreCase(tt.slice, tt.item)
		if got != tt.want {
			t.Errorf("containsIgnoreCase(%v, %q) = %v, want %v", tt.slice, tt.item, got, tt.want)
		}
	}
}

func TestEqualIgnoreCase(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"Open", "Open", true},
		{"Open", "open", true},
		{"Open", "OPEN", true},
		{"Open", "Closed", false},
		{"", "", true},
		{"a", "ab", false},
	}

	for _, tt := range tests {
		got := equalIgnoreCase(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("equalIgnoreCase(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestEvaluateCondition(t *testing.T) {
	svc := &WorkflowService{}

	// Test nil issue
	skip, _ := svc.evaluateCondition(nil, &WorkflowCondition{
		CurrentStatus: []string{"Open"},
	})
	if skip {
		t.Error("should not skip for nil issue")
	}

	// Test nil fields
	issue := &gojira.Issue{}
	skip, _ = svc.evaluateCondition(issue, &WorkflowCondition{
		CurrentStatus: []string{"Open"},
	})
	if skip {
		t.Error("should not skip for nil fields")
	}

	// Test status condition - match
	issue = &gojira.Issue{
		Fields: &gojira.IssueFields{
			Status: &gojira.Status{Name: "Open"},
		},
	}
	skip, _ = svc.evaluateCondition(issue, &WorkflowCondition{
		CurrentStatus: []string{"Open", "In Progress"},
	})
	if skip {
		t.Error("should not skip when status matches")
	}

	// Test status condition - no match
	skip, reason := svc.evaluateCondition(issue, &WorkflowCondition{
		CurrentStatus: []string{"Closed"},
	})
	if !skip {
		t.Error("should skip when status does not match")
	}
	if reason == "" {
		t.Error("skip reason should be provided")
	}

	// Test label condition - match
	issue = &gojira.Issue{
		Fields: &gojira.IssueFields{
			Labels: []string{"bug", "urgent"},
		},
	}
	skip, _ = svc.evaluateCondition(issue, &WorkflowCondition{
		HasLabel: []string{"bug"},
	})
	if skip {
		t.Error("should not skip when label matches")
	}

	// Test label condition - no match
	skip, _ = svc.evaluateCondition(issue, &WorkflowCondition{
		HasLabel: []string{"feature"},
	})
	if !skip {
		t.Error("should skip when label does not match")
	}

	// Test issue type condition - match
	issue = &gojira.Issue{
		Fields: &gojira.IssueFields{
			Type: gojira.IssueType{Name: "Bug"},
		},
	}
	skip, _ = svc.evaluateCondition(issue, &WorkflowCondition{
		IssueType: []string{"Bug", "Task"},
	})
	if skip {
		t.Error("should not skip when issue type matches")
	}

	// Test issue type condition - no match
	skip, _ = svc.evaluateCondition(issue, &WorkflowCondition{
		IssueType: []string{"Story"},
	})
	if !skip {
		t.Error("should skip when issue type does not match")
	}
}

func TestGetPredefinedTemplate(t *testing.T) {
	template, ok := GetPredefinedTemplate("start-work")
	if !ok {
		t.Fatal("start-work template should exist")
	}
	if template.Name != "Start Work" {
		t.Errorf("template name = %q, want 'Start Work'", template.Name)
	}

	template, ok = GetPredefinedTemplate("complete")
	if !ok {
		t.Fatal("complete template should exist")
	}
	if len(template.Steps) != 1 {
		t.Errorf("complete template should have 1 step, got %d", len(template.Steps))
	}

	template, ok = GetPredefinedTemplate("close-with-comment")
	if !ok {
		t.Fatal("close-with-comment template should exist")
	}
	if len(template.Steps) != 2 {
		t.Errorf("close-with-comment template should have 2 steps, got %d", len(template.Steps))
	}

	_, ok = GetPredefinedTemplate("nonexistent")
	if ok {
		t.Error("nonexistent template should not exist")
	}
}

func TestWorkflowStepTypes(t *testing.T) {
	// Verify step type constants are used correctly in predefined templates
	templates := []string{"start-work", "complete", "close-with-comment", "triage"}
	validTypes := map[string]bool{
		"transition": true,
		"comment":    true,
		"assign":     true,
		"update":     true,
	}

	for _, name := range templates {
		template, ok := GetPredefinedTemplate(name)
		if !ok {
			t.Errorf("template %q should exist", name)
			continue
		}

		for i, step := range template.Steps {
			if !validTypes[step.Type] {
				t.Errorf("template %q step %d has invalid type %q", name, i+1, step.Type)
			}
		}
	}
}

func TestWorkflowTemplateStructure(t *testing.T) {
	template := &WorkflowTemplate{
		Name:        "Test Template",
		Description: "A test workflow template",
		Steps: []WorkflowStep{
			{
				Type:       "transition",
				Transition: "In Progress",
			},
			{
				Type:    "comment",
				Comment: "Started working on this",
			},
			{
				Type:         "update",
				AddLabels:    []string{"in-progress"},
				RemoveLabels: []string{"backlog"},
			},
			{
				Type:     "assign",
				Assignee: "testuser",
			},
		},
	}

	if len(template.Steps) != 4 {
		t.Errorf("template should have 4 steps, got %d", len(template.Steps))
	}

	// Verify each step type
	expectedTypes := []string{"transition", "comment", "update", "assign"}
	for i, step := range template.Steps {
		if step.Type != expectedTypes[i] {
			t.Errorf("step %d type = %q, want %q", i+1, step.Type, expectedTypes[i])
		}
	}
}

func TestWorkflowConditionStructure(t *testing.T) {
	condition := &WorkflowCondition{
		CurrentStatus: []string{"Open", "Reopened"},
		HasLabel:      []string{"bug", "urgent"},
		IssueType:     []string{"Bug", "Task"},
	}

	if len(condition.CurrentStatus) != 2 {
		t.Errorf("CurrentStatus should have 2 items, got %d", len(condition.CurrentStatus))
	}
	if len(condition.HasLabel) != 2 {
		t.Errorf("HasLabel should have 2 items, got %d", len(condition.HasLabel))
	}
	if len(condition.IssueType) != 2 {
		t.Errorf("IssueType should have 2 items, got %d", len(condition.IssueType))
	}
}

func TestExecuteWorkflowBulkNilClient(t *testing.T) {
	svc := &WorkflowService{Client: nil}

	template := &WorkflowTemplate{
		Name: "test",
		Steps: []WorkflowStep{
			{Type: "comment", Comment: "test"},
		},
	}

	_, err := svc.ExecuteWorkflowBulk(context.Background(), []string{"TEST-1", "TEST-2"}, template)
	if err == nil {
		t.Error("ExecuteWorkflowBulk() with nil client should return error")
	}
}

func TestExecuteWorkflowBulkEmptyKeys(t *testing.T) {
	client := &Client{}
	svc := NewWorkflowService(client)

	template := &WorkflowTemplate{
		Name: "test",
		Steps: []WorkflowStep{
			{Type: "comment", Comment: "test"},
		},
	}

	_, err := svc.ExecuteWorkflowBulk(context.Background(), []string{}, template)
	if err == nil {
		t.Error("ExecuteWorkflowBulk() with empty keys should return error")
	}
}
