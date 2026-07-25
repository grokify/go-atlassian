package jira

import "errors"

var (
	ErrClientCannotBeNil                = errors.New("client cannot be nil")
	ErrJiraClientCannotBeNil            = errors.New("jira client cannot be nil")
	ErrSimpleClientCannotBeNil          = errors.New("simple client cannot be nil")
	ErrCustomFieldLabelRequired         = errors.New("custom field label is required")
	ErrIssueCannotBeNil                 = errors.New("issue cannot be nil")
	ErrIssueKeyCannotBeEmpty            = errors.New("issue key cannot be empty")
	ErrKeyNotFound                      = errors.New("key not found")
	ErrIssueOrIssueKeyOrIssueIDRequired = errors.New("issue, issue id, or issue key required")
	ErrIssuesSetCannotBeNil             = errors.New("issuesSet cannot be nil")
	ErrFunctionCannotBeNil              = errors.New("function cannot be nil")
	ErrNotFound                         = errors.New("Issue does not exist or you do not have permission to see it.: request failed. Please analyze the request body for more details. Status code: 400")
)
