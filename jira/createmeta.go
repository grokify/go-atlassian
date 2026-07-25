package jira

import "strings"

// CreateMetaIssueType represents an issue type returned by the createmeta API.
type CreateMetaIssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Subtask     bool   `json:"subtask"`
}

// CreateMetaIssueTypesResponse represents the response from
// GET /rest/api/3/issue/createmeta/{projectKey}/issuetypes
type CreateMetaIssueTypesResponse struct {
	MaxResults int                   `json:"maxResults"`
	StartAt    int                   `json:"startAt"`
	Total      int                   `json:"total"`
	IssueTypes []CreateMetaIssueType `json:"issueTypes"`
}

// CreateMetaField represents a field available for issue creation.
type CreateMetaField struct {
	Key             string `json:"key"`             // e.g. "customfield_10001" or "summary"
	Name            string `json:"name"`            // Display name
	Required        bool   `json:"required"`        // Whether the field is required
	HasDefaultValue bool   `json:"hasDefaultValue"` // Whether field has a default value
	FieldID         string `json:"fieldId"`         // Field ID (same as key for most fields)
}

// IsCustomField returns true if this is a custom field (key starts with "customfield_").
func (f CreateMetaField) IsCustomField() bool {
	return strings.HasPrefix(f.Key, "customfield_")
}

// CreateMetaFields is a slice of CreateMetaField.
type CreateMetaFields []CreateMetaField

// CustomOnly returns only custom fields (fields whose key starts with "customfield_").
func (fields CreateMetaFields) CustomOnly() CreateMetaFields {
	var result CreateMetaFields
	for _, f := range fields {
		if f.IsCustomField() {
			result = append(result, f)
		}
	}
	return result
}

// RequiredOnly returns only required fields.
func (fields CreateMetaFields) RequiredOnly() CreateMetaFields {
	var result CreateMetaFields
	for _, f := range fields {
		if f.Required {
			result = append(result, f)
		}
	}
	return result
}

// Keys returns the keys of all fields.
func (fields CreateMetaFields) Keys() []string {
	keys := make([]string, len(fields))
	for i, f := range fields {
		keys[i] = f.Key
	}
	return keys
}

// ByKey returns a map of field key to CreateMetaField.
func (fields CreateMetaFields) ByKey() map[string]CreateMetaField {
	result := make(map[string]CreateMetaField, len(fields))
	for _, f := range fields {
		result[f.Key] = f
	}
	return result
}

// CreateMetaFieldsResponse represents the response from
// GET /rest/api/3/issue/createmeta/{projectKey}/issuetypes/{issueTypeId}
type CreateMetaFieldsResponse struct {
	MaxResults int               `json:"maxResults"`
	StartAt    int               `json:"startAt"`
	Total      int               `json:"total"`
	Values     []CreateMetaField `json:"values"`
}

// RequiredFieldInfo provides information about a required field including
// which issue types require it.
type RequiredFieldInfo struct {
	Field      CreateMetaField `json:"field"`
	IssueTypes []string        `json:"issue_types"` // Issue type names that require this field
}
