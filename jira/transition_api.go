package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/mogo/errors/errorsutil"
	"github.com/grokify/mogo/net/http/httpsimple"
)

type TransitionOptionSet struct {
	ContinueOnUnlistedStatus  bool                        `json:"continueOnUnlistedStatus"`
	StatusToTransitionOptions map[string]TransitionOption `json:"statusToTransitionOptions,omitempty"`
}

type TransitionOption struct {
	Name    string             `json:"name,omitempty"`
	Payload *TransitionPayload `json:"payload,omitempty"`
}

type TransitionPayload struct {
	Transition gojira.TransitionPayload `json:"transition,omitempty"`
	Fields     map[string]any           `json:"fields,omitempty"`
	Update     TransitionPayloadUpdate  `json:"update,omitempty"`
}

type TransitionPayloadUpdate struct {
	Worklog []WorklogOperation `json:"worklog,omitempty"`
}

type WorklogOperation struct {
	Add gojira.WorklogRecord `json:"add"`
}

type TranstionFieldIDOrValue struct {
	ID    any `json:"id,omitempty"`
	Value any `json:"value,omitempty"`
}

type TransitionsAPIResponse struct {
	Transitions []Transition `json:"transitions"`
}

func (svc *IssueService) GetTransitions(ctx context.Context, id string, expandTransitionsFields bool) (Transitions, *gojira.Response, error) {
	if expandTransitionsFields {
		if !KeyIsValid(id) {
			return nil, nil, fmt.Errorf("id is not a valid Jira key (%s)", id)
		}
		sr := httpsimple.Request{
			Method: http.MethodGet,
			URL:    fmt.Sprintf("/rest/api/2/issue/%s/transitions", id),
			Query:  map[string][]string{"expand": {"transitions.fields"}}}
		var result TransitionsAPIResponse
		if resp, err := svc.Client.simpleClient.Do(ctx, sr); err != nil {
			return nil, nil, err
		} else if resp.StatusCode > 299 {
			return nil, nil, fmt.Errorf("bad api response status code (%d)", resp.StatusCode)
		} else if b, err := io.ReadAll(resp.Body); err != nil {
			return nil, nil, err
		} else {
			err = json.Unmarshal(b, &result)
			return result.Transitions, &gojira.Response{Response: resp}, err
		}
	} else {
		txnsSDK, resp, err := svc.Client.JiraClient.Issue.GetTransitionsWithContext(ctx, id)
		if err != nil {
			return nil, resp, err
		}
		txns := Transitions{}
		txns.AddTransitionsSDK(txnsSDK)
		return txns, resp, nil
	}
}

type issueTxnInfo struct {
	Key                     string
	Status                  string
	PossibleTransitionNames []string
}

func (info issueTxnInfo) String() string {
	b, err := json.Marshal(info)
	if err != nil {
		panic(err)
	} else {
		return string(b)
	}
}

func (svc *IssueService) DoTransitions(ctx context.Context, issueIDs []string, opts TransitionOptionSet) error {
	for _, issueID := range issueIDs {
		issue, resp, err := svc.Client.JiraClient.Issue.Get(issueID, nil)
		if err != nil {
			return err
		} else if resp.StatusCode > 299 {
			return fmt.Errorf("bad api response status code (%d)", resp.StatusCode)
		}
		im := NewIssueMore(issue)
		status := im.Status()
		txnOpts, ok := opts.StatusToTransitionOptions[status]
		if !ok {
			if opts.ContinueOnUnlistedStatus {
				continue
			} else {
				return fmt.Errorf("no txnOptions for status (%s) id (%s)", status, issueID)
			}
		}
		if err := svc.DoTransitionWithNameAndPayload(
			ctx, issueID, issue, txnOpts.Name, txnOpts.Payload,
		); err != nil {
			return err
		}
	}
	return nil
}

func (svc *IssueService) DoTransitionWithNameAndPayload(ctx context.Context, issueID string, issue *gojira.Issue, updateTransitionName string, payload *TransitionPayload) error {
	if issue == nil {
		if issueTry, resp, err := svc.Client.JiraClient.Issue.Get(issueID, nil); err != nil {
			return err
		} else if resp.StatusCode > 299 {
			return fmt.Errorf("bad api response status code (%d)", resp.StatusCode)
		} else {
			issue = issueTry
		}
	}
	im := NewIssueMore(issue)
	status := im.Status()
	issTxnMeta := issueTxnInfo{
		Key:    issueID,
		Status: status}
	possibleTxns, resp, err := svc.GetTransitions(ctx, issueID, true)
	if err != nil {
		return err
	} else if resp.StatusCode > 299 {
		return fmt.Errorf("bad api response status code (%d)", resp.StatusCode)
	}
	issTxnMeta.PossibleTransitionNames = possibleTxns.Names()
	wantTxn, err := possibleTxns.GetByName(updateTransitionName)
	if err != nil {
		return errorsutil.Wrapf(err, "jiraTxnInfo (%s)", issTxnMeta.String())
	}
	if payload != nil {
		payload.Transition.ID = wantTxn.ID
		if resp, err = svc.Client.JiraClient.Issue.DoTransitionWithPayloadWithContext(ctx, issueID, *payload); err != nil {
			return errorsutil.Wrapf(err, "meta (%s)", issTxnMeta.String())
		} else if resp.StatusCode > 299 {
			return fmt.Errorf("bad api response status code (%d)", resp.StatusCode)
		}
	} else {
		if resp, err = svc.Client.JiraClient.Issue.DoTransitionWithContext(ctx, issueID, wantTxn.ID); err != nil {
			return err
		} else if resp.StatusCode > 299 {
			return fmt.Errorf("bad api response status code (%d)", resp.StatusCode)
		}
	}

	return nil
}

// TransitionOptions configures a transition operation.
type TransitionOptions struct {
	Comment string         // Optional comment to add with the transition
	Fields  map[string]any // Optional fields to set during transition
}

// TransitionResult contains the result of a transition operation.
type TransitionResult struct {
	Key        string `json:"key"`
	FromStatus string `json:"from_status"`
	ToStatus   string `json:"to_status"`
	Transition string `json:"transition"`
}

// BulkTransitionResult contains results for a bulk transition operation.
type BulkTransitionResult struct {
	Successful   []TransitionResult `json:"successful"`
	Failed       []TransitionError  `json:"failed"`
	Total        int                `json:"total"`
	SuccessCount int                `json:"success_count"`
	FailCount    int                `json:"fail_count"`
}

// TransitionError represents a failed transition attempt.
type TransitionError struct {
	Key   string `json:"key"`
	Error string `json:"error"`
}

// BulkTransitionIssues transitions multiple issues to a new status.
// It continues processing even if some transitions fail.
func (svc *IssueService) BulkTransitionIssues(ctx context.Context, issueKeys []string, transitionNameOrID string, opts *TransitionOptions) (*BulkTransitionResult, error) {
	if svc.Client == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	result := &BulkTransitionResult{
		Successful: []TransitionResult{},
		Failed:     []TransitionError{},
		Total:      len(issueKeys),
	}

	for _, key := range issueKeys {
		txnResult, err := svc.TransitionIssue(ctx, key, transitionNameOrID, opts)
		if err != nil {
			result.Failed = append(result.Failed, TransitionError{
				Key:   key,
				Error: err.Error(),
			})
			result.FailCount++
		} else {
			result.Successful = append(result.Successful, *txnResult)
			result.SuccessCount++
		}
	}

	return result, nil
}

// TransitionIssue transitions an issue to a new status using the transition name or ID.
// It validates the transition is available before executing and returns the result.
func (svc *IssueService) TransitionIssue(ctx context.Context, issueKey string, transitionNameOrID string, opts *TransitionOptions) (*TransitionResult, error) {
	if svc.Client == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	// Get current issue to capture from status
	issue, resp, err := svc.Client.JiraClient.Issue.GetWithContext(ctx, issueKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %q: %w", issueKey, err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to get issue %q: status %d", issueKey, resp.StatusCode)
	}

	im := NewIssueMore(issue)
	fromStatus := im.Status()

	// Get available transitions
	transitions, resp, err := svc.GetTransitions(ctx, issueKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get transitions for %q: %w", issueKey, err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to get transitions for %q: status %d", issueKey, resp.StatusCode)
	}

	// Find the transition by name or ID
	var targetTransition *Transition
	for i := range transitions {
		if transitions[i].Name == transitionNameOrID || transitions[i].ID == transitionNameOrID {
			targetTransition = &transitions[i]
			break
		}
	}

	if targetTransition == nil {
		availableNames := transitions.Names()
		return nil, fmt.Errorf("transition %q not available for %q (current status: %s); available: %v",
			transitionNameOrID, issueKey, fromStatus, availableNames)
	}

	// Build the transition payload
	payload := map[string]any{
		"transition": map[string]string{
			"id": targetTransition.ID,
		},
	}

	// Add comment if provided
	if opts != nil && opts.Comment != "" {
		payload["update"] = map[string]any{
			"comment": []map[string]any{
				{
					"add": map[string]string{
						"body": opts.Comment,
					},
				},
			},
		}
	}

	// Add fields if provided
	if opts != nil && len(opts.Fields) > 0 {
		payload["fields"] = opts.Fields
	}

	// Execute the transition
	resp, err = svc.Client.JiraClient.Issue.DoTransitionWithPayloadWithContext(ctx, issueKey, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to transition %q: %w", issueKey, err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to transition %q: status %d", issueKey, resp.StatusCode)
	}

	return &TransitionResult{
		Key:        issueKey,
		FromStatus: fromStatus,
		ToStatus:   targetTransition.To.Name,
		Transition: targetTransition.Name,
	}, nil
}
