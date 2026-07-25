package jira

import (
	"context"
	"fmt"

	gojira "github.com/andygrunwald/go-jira"
)

// WorkflowTemplate defines a reusable sequence of operations to apply to issues.
type WorkflowTemplate struct {
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	Steps       []WorkflowStep `json:"steps" yaml:"steps"`
}

// WorkflowStep represents a single step in a workflow template.
type WorkflowStep struct {
	// Type is the type of operation: "transition", "update", "comment", "assign"
	Type string `json:"type" yaml:"type"`

	// Transition name or ID (for type="transition")
	Transition string `json:"transition,omitempty" yaml:"transition,omitempty"`

	// Comment body (for type="comment" or type="transition" with comment)
	Comment string `json:"comment,omitempty" yaml:"comment,omitempty"`

	// Assignee username or account ID (for type="assign")
	Assignee string `json:"assignee,omitempty" yaml:"assignee,omitempty"`

	// Labels to add (for type="update")
	AddLabels []string `json:"add_labels,omitempty" yaml:"add_labels,omitempty"`

	// Labels to remove (for type="update")
	RemoveLabels []string `json:"remove_labels,omitempty" yaml:"remove_labels,omitempty"`

	// SkipOnError allows the workflow to continue even if this step fails
	SkipOnError bool `json:"skip_on_error,omitempty" yaml:"skip_on_error,omitempty"`

	// Condition allows conditional execution based on issue state
	Condition *WorkflowCondition `json:"condition,omitempty" yaml:"condition,omitempty"`
}

// WorkflowCondition defines conditions for conditional workflow step execution.
type WorkflowCondition struct {
	// CurrentStatus - only execute if the issue is in one of these statuses
	CurrentStatus []string `json:"current_status,omitempty" yaml:"current_status,omitempty"`

	// HasLabel - only execute if the issue has one of these labels
	HasLabel []string `json:"has_label,omitempty" yaml:"has_label,omitempty"`

	// IssueType - only execute for these issue types
	IssueType []string `json:"issue_type,omitempty" yaml:"issue_type,omitempty"`
}

// WorkflowResult contains the result of executing a workflow template.
type WorkflowResult struct {
	IssueKey    string       `json:"issue_key"`
	Success     bool         `json:"success"`
	StepsRun    int          `json:"steps_run"`
	StepResults []StepResult `json:"step_results"`
	Error       string       `json:"error,omitempty"`
}

// StepResult contains the result of a single workflow step.
type StepResult struct {
	Step       int    `json:"step"`
	Type       string `json:"type"`
	Success    bool   `json:"success"`
	Skipped    bool   `json:"skipped,omitempty"`
	SkipReason string `json:"skip_reason,omitempty"`
	Error      string `json:"error,omitempty"`
}

// WorkflowService executes workflow templates on issues.
type WorkflowService struct {
	Client *Client
}

// NewWorkflowService creates a new WorkflowService.
func NewWorkflowService(client *Client) *WorkflowService {
	return &WorkflowService{Client: client}
}

// ExecuteWorkflow executes a workflow template on an issue.
func (svc *WorkflowService) ExecuteWorkflow(ctx context.Context, issueKey string, template *WorkflowTemplate) (*WorkflowResult, error) {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	if issueKey == "" {
		return nil, fmt.Errorf("issue key is required")
	}

	if template == nil || len(template.Steps) == 0 {
		return nil, fmt.Errorf("workflow template must have at least one step")
	}

	result := &WorkflowResult{
		IssueKey:    issueKey,
		Success:     true,
		StepResults: make([]StepResult, 0, len(template.Steps)),
	}

	// Get current issue state for condition evaluation
	issue, err := svc.Client.IssueAPI.Issue(ctx, issueKey, nil)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to get issue: %v", err)
		return result, err
	}

	for i, step := range template.Steps {
		stepResult := StepResult{
			Step:    i + 1,
			Type:    step.Type,
			Success: true,
		}

		// Check condition
		if step.Condition != nil {
			skip, reason := svc.evaluateCondition(issue, step.Condition)
			if skip {
				stepResult.Skipped = true
				stepResult.SkipReason = reason
				result.StepResults = append(result.StepResults, stepResult)
				continue
			}
		}

		// Execute the step
		err := svc.executeStep(ctx, issueKey, &step)
		if err != nil {
			stepResult.Success = false
			stepResult.Error = err.Error()

			if !step.SkipOnError {
				result.Success = false
				result.Error = fmt.Sprintf("step %d failed: %v", i+1, err)
				result.StepResults = append(result.StepResults, stepResult)
				result.StepsRun = i + 1
				return result, nil
			}
		}

		result.StepResults = append(result.StepResults, stepResult)

		// Refresh issue state after each step for accurate condition evaluation
		if i < len(template.Steps)-1 {
			issue, _ = svc.Client.IssueAPI.Issue(ctx, issueKey, nil)
		}
	}

	result.StepsRun = len(template.Steps)
	return result, nil
}

// evaluateCondition checks if a step's condition is met.
func (svc *WorkflowService) evaluateCondition(issue *gojira.Issue, condition *WorkflowCondition) (skip bool, reason string) {
	if issue == nil || issue.Fields == nil {
		return false, ""
	}

	// Check current status condition
	if len(condition.CurrentStatus) > 0 {
		currentStatus := ""
		if issue.Fields.Status != nil {
			currentStatus = issue.Fields.Status.Name
		}
		if !containsIgnoreCase(condition.CurrentStatus, currentStatus) {
			return true, fmt.Sprintf("current status %q not in %v", currentStatus, condition.CurrentStatus)
		}
	}

	// Check has label condition
	if len(condition.HasLabel) > 0 {
		hasLabel := false
		for _, label := range issue.Fields.Labels {
			if containsIgnoreCase(condition.HasLabel, label) {
				hasLabel = true
				break
			}
		}
		if !hasLabel {
			return true, fmt.Sprintf("issue does not have any of labels %v", condition.HasLabel)
		}
	}

	// Check issue type condition
	if len(condition.IssueType) > 0 {
		issueType := ""
		if issue.Fields.Type.Name != "" {
			issueType = issue.Fields.Type.Name
		}
		if !containsIgnoreCase(condition.IssueType, issueType) {
			return true, fmt.Sprintf("issue type %q not in %v", issueType, condition.IssueType)
		}
	}

	return false, ""
}

// containsIgnoreCase checks if a slice contains a string (case-insensitive).
func containsIgnoreCase(slice []string, item string) bool {
	for _, s := range slice {
		if equalIgnoreCase(s, item) {
			return true
		}
	}
	return false
}

// equalIgnoreCase compares two strings case-insensitively.
func equalIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// executeStep executes a single workflow step.
func (svc *WorkflowService) executeStep(ctx context.Context, issueKey string, step *WorkflowStep) error {
	switch step.Type {
	case "transition":
		return svc.executeTransition(ctx, issueKey, step)
	case "comment":
		return svc.executeComment(ctx, issueKey, step)
	case "assign":
		return svc.executeAssign(ctx, issueKey, step)
	case "update":
		return svc.executeUpdate(ctx, issueKey, step)
	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
}

func (svc *WorkflowService) executeTransition(ctx context.Context, issueKey string, step *WorkflowStep) error {
	if step.Transition == "" {
		return fmt.Errorf("transition name or ID is required")
	}

	opts := &TransitionOptions{}
	if step.Comment != "" {
		opts.Comment = step.Comment
	}

	_, err := svc.Client.IssueAPI.TransitionIssue(ctx, issueKey, step.Transition, opts)
	return err
}

func (svc *WorkflowService) executeComment(ctx context.Context, issueKey string, step *WorkflowStep) error {
	if step.Comment == "" {
		return fmt.Errorf("comment body is required")
	}

	comment := &gojira.Comment{Body: step.Comment}
	_, _, err := svc.Client.JiraClient.Issue.AddCommentWithContext(ctx, issueKey, comment)
	return err
}

func (svc *WorkflowService) executeAssign(ctx context.Context, issueKey string, step *WorkflowStep) error {
	if step.Assignee == "" {
		return fmt.Errorf("assignee is required")
	}

	// Use the Jira API to update assignee
	user := &gojira.User{Name: step.Assignee}
	_, err := svc.Client.JiraClient.Issue.UpdateAssigneeWithContext(ctx, issueKey, user)
	return err
}

func (svc *WorkflowService) executeUpdate(ctx context.Context, issueKey string, step *WorkflowStep) error {
	updateBody := IssuePatchRequestBody{}

	if len(step.AddLabels) > 0 || len(step.RemoveLabels) > 0 {
		updateBody.Update = &IssuePatchRequestBodyUpdate{}

		for _, label := range step.AddLabels {
			labelCopy := label
			updateBody.Update.Labels = append(updateBody.Update.Labels, IssuePatchRequestBodyUpdateLabel{
				Add: &labelCopy,
			})
		}

		for _, label := range step.RemoveLabels {
			labelCopy := label
			updateBody.Update.Labels = append(updateBody.Update.Labels, IssuePatchRequestBodyUpdateLabel{
				Remove: &labelCopy,
			})
		}
	}

	_, err := svc.Client.IssueAPI.IssuePatch(ctx, issueKey, updateBody)
	return err
}

// ExecuteWorkflowBulk executes a workflow template on multiple issues.
func (svc *WorkflowService) ExecuteWorkflowBulk(ctx context.Context, issueKeys []string, template *WorkflowTemplate) ([]*WorkflowResult, error) {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	if len(issueKeys) == 0 {
		return nil, fmt.Errorf("at least one issue key is required")
	}

	results := make([]*WorkflowResult, 0, len(issueKeys))
	for _, key := range issueKeys {
		result, _ := svc.ExecuteWorkflow(ctx, key, template)
		results = append(results, result)
	}

	return results, nil
}

// PredefinedTemplates contains commonly used workflow templates.
var PredefinedTemplates = map[string]*WorkflowTemplate{
	"start-work": {
		Name:        "Start Work",
		Description: "Transition issue to In Progress and assign to current user",
		Steps: []WorkflowStep{
			{
				Type:       "transition",
				Transition: "In Progress",
			},
		},
	},
	"complete": {
		Name:        "Complete Issue",
		Description: "Transition issue to Done",
		Steps: []WorkflowStep{
			{
				Type:       "transition",
				Transition: "Done",
			},
		},
	},
	"close-with-comment": {
		Name:        "Close With Comment",
		Description: "Add a closing comment and transition to Done",
		Steps: []WorkflowStep{
			{
				Type:    "comment",
				Comment: "Issue completed.",
			},
			{
				Type:       "transition",
				Transition: "Done",
			},
		},
	},
	"triage": {
		Name:        "Triage Bug",
		Description: "Add triage label and transition to backlog",
		Steps: []WorkflowStep{
			{
				Type:      "update",
				AddLabels: []string{"triaged"},
			},
			{
				Type:       "transition",
				Transition: "Backlog",
				Condition: &WorkflowCondition{
					IssueType: []string{"Bug"},
				},
			},
		},
	},
}

// GetPredefinedTemplate returns a predefined workflow template by name.
func GetPredefinedTemplate(name string) (*WorkflowTemplate, bool) {
	template, ok := PredefinedTemplates[name]
	return template, ok
}
