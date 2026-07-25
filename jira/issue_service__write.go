package jira

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/mogo/net/http/httpsimple"
	"github.com/grokify/mogo/net/urlutil"
	"github.com/grokify/mogo/pointer"
	"github.com/grokify/mogo/type/stringsutil"
)

// IssuePatchRequestBody represents a API request body to patch an issue. The
// Jira API uses `PUT` however this struct and associated method use `Patch` to
// better align with API best practices for a partial update.
type IssuePatchRequestBody struct {
	Update *IssuePatchRequestBodyUpdate          `json:"update,omitempty"`
	Fields map[string]IssuePatchRequestBodyField `json:"fields,omitempty"`
}

// NewIssuePatchRequestBodyLabelAddRemove returns a body for patching the Jira issue
// by adding or removing a label.
func NewIssuePatchRequestBodyLabelAddRemove(label string, remove bool) IssuePatchRequestBody {
	labelUpdate := IssuePatchRequestBodyUpdateLabel{}
	if remove {
		labelUpdate.Remove = pointer.Pointer(label)
	} else {
		labelUpdate.Add = pointer.Pointer(label)
	}
	return IssuePatchRequestBody{
		Update: &IssuePatchRequestBodyUpdate{
			Labels: []IssuePatchRequestBodyUpdateLabel{
				labelUpdate,
			},
		},
	}
}

// NewIssuePatchRequestBodyCustomField returns a body for patching the Jira issue
// with a custom field value.
func NewIssuePatchRequestBodyCustomField(customFieldLabel, customFieldValue string) IssuePatchRequestBody {
	return IssuePatchRequestBody{
		Fields: map[string]IssuePatchRequestBodyField{
			customFieldLabel: {
				Value: customFieldValue,
			},
		},
	}
}

// IssuePatchRequestBodyUpdate represntes the `labels` slice in the `update` property
// of an issue update request.
type IssuePatchRequestBodyUpdate struct {
	Labels []IssuePatchRequestBodyUpdateLabel `json:"labels,omitempty"`
}

// IssuePatchRequestBodyUpdateLabel represents a specific label operation in the `update` property
// of an issue update request.
type IssuePatchRequestBodyUpdateLabel struct {
	// cannot have both
	Add    *string `json:"add,omitempty"`
	Remove *string `json:"remove,omitempty"`
}

// Validate ensures that the `add` and `remove` propererties cannot both be set
// at the same time.
func (body IssuePatchRequestBody) Validate() error {
	if body.Update != nil {
		if len(body.Update.Labels) > 0 {
			for _, l := range body.Update.Labels {
				if l.Add != nil && l.Remove != nil {
					return errors.New("label update cannot have both add and remove")
				}
			}
		}
	}
	return nil
}

// FieldPatchRequestObject can be used IssuePatchRequestBody.Fields
type IssuePatchRequestBodyField struct {
	Value string                      `json:"value"`
	Child *IssuePatchRequestBodyField `json:"child,omitempty"`
}

// IssuePatch updates fields for an issue. See more here:
// https://community.developer.atlassian.com/t/update-issue-custom-field-value-via-api-without-going-forge/71161
func (svc *IssueService) IssuePatch(ctx context.Context, issueKeyOrID string, issueUpdateRequestBody IssuePatchRequestBody) (*http.Response, error) {
	if err := issueUpdateRequestBody.Validate(); err != nil {
		return nil, err
	} else if issueKeyOrID = strings.TrimSpace(issueKeyOrID); issueKeyOrID == "" {
		return nil, errors.New("issue key or id must be provided")
	} else if issueUpdateRequestBody.Update == nil && len(issueUpdateRequestBody.Fields) == 0 {
		return nil, errors.New("issue `update` or `fields` must be provided")
	} else {
		return svc.Client.simpleClient.Do(ctx, httpsimple.Request{
			Method:   http.MethodPut, // This only updates certain fields but uses a PUT http method.
			URL:      urlutil.JoinAbsolute(APIV3URLIssue, issueKeyOrID),
			Body:     issueUpdateRequestBody,
			BodyType: httpsimple.BodyTypeJSON})
	}
}

// IssuesPatch updates fields for multiple issues. See more here:
// https://community.developer.atlassian.com/t/update-issue-custom-field-value-via-api-without-going-forge/71161
func (svc *IssueService) IssuesPatch(ctx context.Context, issueKeyOrID []string, issueUpdateRequestBody IssuePatchRequestBody) error {
	issueKeyOrID = stringsutil.SliceCondenseSpace(issueKeyOrID, true, true)
	if len(issueKeyOrID) == 0 {
		return errors.New("ids cannot be empty")
	}
	for _, id := range issueKeyOrID {
		_, err := svc.IssuePatch(ctx, id, issueUpdateRequestBody)
		if err != nil {
			return err
		}
	}
	return nil
}

// BulkUpdateResult contains results for a bulk update operation.
type BulkUpdateResult struct {
	Successful   []UpdateResult `json:"successful"`
	Failed       []UpdateError  `json:"failed"`
	Total        int            `json:"total"`
	SuccessCount int            `json:"success_count"`
	FailCount    int            `json:"fail_count"`
}

// UpdateResult represents a successful update.
type UpdateResult struct {
	Key string `json:"key"`
}

// UpdateError represents a failed update attempt.
type UpdateError struct {
	Key   string `json:"key"`
	Error string `json:"error"`
}

// BulkUpdateOptions configures a bulk update operation.
type BulkUpdateOptions struct {
	// Fields to set (key is field name or ID, value is the new value)
	Fields map[string]any `json:"fields,omitempty"`
	// Labels to add
	AddLabels []string `json:"add_labels,omitempty"`
	// Labels to remove
	RemoveLabels []string `json:"remove_labels,omitempty"`
	// Comment to add to each issue
	Comment string `json:"comment,omitempty"`
}

// BulkUpdateIssues updates multiple issues with the same field changes.
// It continues processing even if some updates fail.
func (svc *IssueService) BulkUpdateIssues(ctx context.Context, issueKeys []string, opts *BulkUpdateOptions) (*BulkUpdateResult, error) {
	if svc.Client == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}
	if opts == nil {
		return nil, errors.New("update options cannot be nil")
	}

	issueKeys = stringsutil.SliceCondenseSpace(issueKeys, true, true)
	result := &BulkUpdateResult{
		Successful: []UpdateResult{},
		Failed:     []UpdateError{},
		Total:      len(issueKeys),
	}

	for _, key := range issueKeys {
		err := svc.updateSingleIssue(ctx, key, opts)
		if err != nil {
			result.Failed = append(result.Failed, UpdateError{
				Key:   key,
				Error: err.Error(),
			})
			result.FailCount++
		} else {
			result.Successful = append(result.Successful, UpdateResult{Key: key})
			result.SuccessCount++
		}
	}

	return result, nil
}

// updateSingleIssue updates a single issue with the given options.
func (svc *IssueService) updateSingleIssue(ctx context.Context, issueKey string, opts *BulkUpdateOptions) error {
	// Build the update body
	body := IssuePatchRequestBody{}

	// Handle field updates
	if len(opts.Fields) > 0 {
		body.Fields = make(map[string]IssuePatchRequestBodyField)
		for k, v := range opts.Fields {
			if strVal, ok := v.(string); ok {
				body.Fields[k] = IssuePatchRequestBodyField{Value: strVal}
			}
		}
	}

	// Handle label updates
	if len(opts.AddLabels) > 0 || len(opts.RemoveLabels) > 0 {
		body.Update = &IssuePatchRequestBodyUpdate{
			Labels: []IssuePatchRequestBodyUpdateLabel{},
		}
		for _, label := range opts.AddLabels {
			body.Update.Labels = append(body.Update.Labels, IssuePatchRequestBodyUpdateLabel{
				Add: pointer.Pointer(label),
			})
		}
		for _, label := range opts.RemoveLabels {
			body.Update.Labels = append(body.Update.Labels, IssuePatchRequestBodyUpdateLabel{
				Remove: pointer.Pointer(label),
			})
		}
	}

	// Execute the update if there are fields or labels to update
	if body.Update != nil || len(body.Fields) > 0 {
		resp, err := svc.IssuePatch(ctx, issueKey, body)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 300 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("update failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		}
	}

	// Add comment if provided
	if opts.Comment != "" {
		comment := &gojira.Comment{Body: opts.Comment}
		_, resp, err := svc.Client.JiraClient.Issue.AddCommentWithContext(ctx, issueKey, comment)
		if err != nil {
			return fmt.Errorf("failed to add comment: %w", err)
		}
		if resp.StatusCode >= 300 {
			return fmt.Errorf("failed to add comment: status %d", resp.StatusCode)
		}
	}

	return nil
}

// IssuePatchCustomFieldRecursive updates an issue, and optionally child issues, with a
// custom field value.
func (svc *IssueService) IssuePatchCustomFieldRecursive(ctx context.Context, issueKeyOrID string, iss *gojira.Issue, customFieldLabel, customFieldValue string, processChildren bool, processChildrenTypes []string, skipUpdate bool) (int, error) {
	count := 0
	customFieldLabel = strings.TrimSpace(customFieldLabel)
	if customFieldLabel == "" {
		return count, ErrCustomFieldLabelRequired
	}
	issueKeyOrID = strings.TrimSpace(issueKeyOrID)
	if issueKeyOrID == "" && iss == nil {
		return count, ErrIssueOrIssueKeyOrIssueIDRequired
	}
	processChildrenTypes = stringsutil.SliceCondenseSpace(processChildrenTypes, true, true)
	if iss == nil {
		if issGet, err := svc.Issue(ctx, issueKeyOrID, nil); err != nil {
			return 0, err
		} else {
			iss = issGet
		}
	}
	im := NewIssueMore(iss)
	if issueKeyOrID == "" && iss != nil {
		issueKeyOrID = im.Key()
	}

	curVal, err := im.CustomFieldString(customFieldLabel)
	if err != nil {
		return count, err
	}
	svc.Client.LogOrNotAny(
		ctx,
		slog.LevelDebug,
		"processing issue custom field update: current value",
		"issueKey", issueKeyOrID,
		"issueType", im.Type(),
		"customFieldLabel", customFieldLabel,
		"customFieldValue", customFieldValue,
		"customFieldValueCurrent", curVal)
	if curVal != customFieldValue {
		if skipUpdate {
			count++
		} else if !skipUpdate {
			svc.Client.LogOrNotAny(
				ctx,
				slog.LevelDebug,
				"updating issue custom field",
				"issueKey", issueKeyOrID,
				"issueType", im.Type(),
				"customFieldLabel", customFieldLabel,
				"customFieldValue", customFieldValue)
			reqBody := NewIssuePatchRequestBodyCustomField(customFieldLabel, customFieldValue)
			if resp, err := svc.IssuePatch(ctx, issueKeyOrID, reqBody); err != nil {
				return 0, nil
			} else if resp.StatusCode >= 300 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					body = []byte(err.Error())
				}
				return 0, fmt.Errorf("key (%s) status code (%d) bodyOrErr (%s)", im.Key(), resp.StatusCode, string(body))
			} else {
				count++
			}
		}
	}
	if processChildren {
		ii, err := svc.SearchChildrenIssues([]string{issueKeyOrID})
		if err != nil {
			return count, err
		}
		is, err := ii.IssuesSet(nil)
		if err != nil {
			return count, err
		}
		if len(processChildrenTypes) > 0 {
			is, err = is.FilterByType(processChildrenTypes...)
			if err != nil {
				return count, err
			}
		}
		cIssKeys := is.Keys()
		for _, cIssKey := range cIssKeys {
			svc.Client.LogOrNotAny(
				ctx,
				slog.LevelInfo,
				"processing issue custom field update children count",
				"issueKey", cIssKey,
				"issueType", im.Type(),
				"customFieldLabel", customFieldLabel,
				"customFieldValue", customFieldValue,
				"parentChildrenCount", is.Len())
			cISS, err := is.Issue(cIssKey)
			if err != nil {
				return count, err
			}
			// cIM := NewIssueMore(&cISS)
			if countChildren, err := svc.IssuePatchCustomFieldRecursive(ctx, "", &cISS, customFieldLabel, customFieldValue, processChildren, processChildrenTypes, skipUpdate); err != nil {
				return count, err
			} else {
				count += countChildren
			}
		}
	}

	return count, nil
}

func (svc *IssueService) IssuesPatchAddLabel(ctx context.Context, issues []gojira.Issue, label string, skipUpdate bool) (int, error) {
	n := 0
	for _, iss := range issues {
		im := NewIssueMore(&iss)
		if ni, err := svc.IssuePatchLabelRecursive(ctx, im.Key(), &iss, label, false, false, []string{}, skipUpdate); err != nil {
			return n, err
		} else {
			n += ni
		}
	}
	return n, nil
}

// IssuePatchLabelRecursive updates fields for an issue. See more here:
// https://community.developer.atlassian.com/t/update-issue-custom-field-value-via-api-without-going-forge/71161
func (svc *IssueService) IssuePatchLabelRecursive(ctx context.Context, issueKeyOrID string, iss *gojira.Issue, label string, removeLabel, processChildren bool, processChildrenTypes []string, skipUpdate bool) (int, error) {
	count := 0
	issueKeyOrID = strings.TrimSpace(issueKeyOrID)
	processChildrenTypes = stringsutil.SliceCondenseSpace(processChildrenTypes, true, true)
	var labelOperation string // for logging
	if removeLabel {
		labelOperation = OperationRemove
	} else {
		labelOperation = OperationAdd
	}
	if issueKeyOrID == "" {
		return count, ErrIssueOrIssueKeyOrIssueIDRequired
	}
	if iss == nil {
		if issGet, err := svc.Issue(ctx, issueKeyOrID, nil); err != nil {
			return 0, err
		} else {
			iss = issGet
		}
	}
	im := NewIssueMore(iss)

	var labelExists bool
	if im.LabelExists(label, false) {
		labelExists = true
	} else {
		labelExists = false
	}

	svc.Client.LogOrNotAny(
		ctx,
		slog.LevelDebug,
		"validating issue labels",
		"issueKey", issueKeyOrID,
		"issueType", im.Type(),
		"labelAction", labelOperation,
		"label", label,
		"labelExists", labelExists)
	if (removeLabel && im.LabelExists(label, false)) || (!removeLabel && !im.LabelExists(label, false)) {
		if skipUpdate {
			count++
		} else {
			svc.Client.LogOrNotAny(
				ctx,
				slog.LevelDebug,
				"updating issue labels",
				"issueKey", issueKeyOrID,
				"issueType", im.Type(),
				"labelAction", labelOperation,
				"label", label)
			reqBody := NewIssuePatchRequestBodyLabelAddRemove(label, removeLabel)
			if resp, err := svc.IssuePatch(ctx, issueKeyOrID, reqBody); err != nil {
				return count, err
			} else if resp.StatusCode >= 300 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					body = []byte(err.Error())
				}
				return count, fmt.Errorf("key (%s) status code (%d) bodyOrErr (%s)", im.Key(), resp.StatusCode, string(body))
			} else {
				count++
			}
		}
	}
	if processChildren {
		ii, err := svc.SearchChildrenIssues([]string{issueKeyOrID})
		if err != nil {
			return count, err
		}
		is, err := ii.IssuesSet(nil)
		if err != nil {
			return count, err
		}
		if len(processChildrenTypes) > 0 {
			is, err = is.FilterByType(processChildrenTypes...)
			if err != nil {
				return count, err
			}
		}
		svc.Client.LogOrNotAny(
			ctx,
			slog.LevelInfo,
			"processing label update children count",
			"issueKey", issueKeyOrID,
			"issueType", im.Type(),
			"labelAction", labelOperation,
			"label", label,
			"childrenCount", is.Len())
		issKeys := is.Keys()
		for _, issKey := range issKeys {
			cISS, err := is.Issue(issKey)
			if err != nil {
				return count, err
			}
			im := NewIssueMore(&cISS)
			if countChildren, err := svc.IssuePatchLabelRecursive(ctx, im.Key(), &cISS, label, removeLabel, processChildren, processChildrenTypes, skipUpdate); err != nil {
				return count, err
			} else {
				count += countChildren
			}
		}
	}
	return count, nil
}
