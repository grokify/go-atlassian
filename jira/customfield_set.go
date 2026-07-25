package jira

import (
	"errors"
	"strings"

	gojira "github.com/andygrunwald/go-jira"
)

type CustomFieldSet struct {
	Data map[string]CustomField
}

func NewCustomFieldSet() *CustomFieldSet {
	return &CustomFieldSet{Data: map[string]CustomField{}}
}

func (set *CustomFieldSet) Init() {
	if set.Data == nil {
		set.Data = map[string]CustomField{}
	}
}

func (set *CustomFieldSet) Add(fields ...CustomField) error {
	set.Init()
	for _, ci := range fields {
		id := strings.TrimSpace(ci.ID)
		if id == "" {
			return errors.New("custom field cannot have empty id")
		}
		set.Data[ci.ID] = ci
	}
	return nil
}

func (set *CustomFieldSet) IDToName(id string) (string, error) {
	if can, ok := IsCustomFieldKey(id); ok {
		if cf, ok := set.Data[can]; ok {
			if name := strings.TrimSpace(cf.Name); name != "" {
				return name, nil
			} else {
				return "", errors.New("customfield has no name")
			}
		} else {
			return "", errors.New("customfieldid not found")
		}
	} else {
		return "", errors.New("cannot parse custom field key")
	}
}

// NameToIDs returns all custom field IDs (e.g. "customfield_12345") that match
// the given display name. This handles the common Jira situation where multiple
// custom fields share the same name (e.g. from copied schemes or reinstalled apps).
// The match is case-insensitive.
func (set *CustomFieldSet) NameToIDs(name string) []string {
	if set == nil || set.Data == nil {
		return nil
	}
	lower := strings.ToLower(strings.TrimSpace(name))
	var ids []string
	for id, cf := range set.Data {
		if strings.ToLower(cf.Name) == lower {
			ids = append(ids, id)
		}
	}
	return ids
}

// NameToFields returns all CustomField entries that match the given display name.
// Case-insensitive.
func (set *CustomFieldSet) NameToFields(name string) []CustomField {
	if set == nil || set.Data == nil {
		return nil
	}
	lower := strings.ToLower(strings.TrimSpace(name))
	var fields []CustomField
	for _, cf := range set.Data {
		if strings.ToLower(cf.Name) == lower {
			fields = append(fields, cf)
		}
	}
	return fields
}

// IssueCustomFieldValue holds a resolved custom field value from an issue, including
// the field metadata and the extracted value.
type IssueCustomFieldValue struct {
	FieldID   string      `json:"field_id"`   // e.g. "customfield_13665"
	FieldName string      `json:"field_name"` // e.g. "Module"
	Value     string      `json:"value"`      // extracted string value
	Populated bool        `json:"populated"`  // true if value is non-empty
	Field     CustomField `json:"field"`      // full field metadata
}

// IssueCustomFieldsByName looks up all custom fields matching `name` in the set,
// then extracts their values from the given issue. Returns one entry per matching
// field ID, regardless of whether it's populated on the issue.
func (set *CustomFieldSet) IssueCustomFieldsByName(iss *gojira.Issue, name string) []IssueCustomFieldValue {
	fields := set.NameToFields(name)
	if len(fields) == 0 {
		return nil
	}

	im := NewIssueMore(iss)
	results := make([]IssueCustomFieldValue, 0, len(fields))
	for _, cf := range fields {
		val := im.CustomFieldStringOrDefault(cf.ID, "")
		results = append(results, IssueCustomFieldValue{
			FieldID:   cf.ID,
			FieldName: cf.Name,
			Value:     val,
			Populated: val != "",
			Field:     cf,
		})
	}
	return results
}

// IssueCustomFieldsByNamePopulated is like IssueCustomFieldsByName but only returns
// fields that have a non-empty value on the issue. This is the recommended way to
// resolve ambiguous field names — the populated one is typically the active one.
func (set *CustomFieldSet) IssueCustomFieldsByNamePopulated(iss *gojira.Issue, name string) []IssueCustomFieldValue {
	all := set.IssueCustomFieldsByName(iss, name)
	populated := make([]IssueCustomFieldValue, 0, len(all))
	for _, v := range all {
		if v.Populated {
			populated = append(populated, v)
		}
	}
	return populated
}

// CreateFuncIssueToMap creates a function to use with `IssuesSet.HistogramMapFunc`.
func (set *CustomFieldSet) CreateFuncIssueToMap(fieldsWithDefaults map[string]string, useCustomFieldDisplayNames bool) FuncIssueToMap {
	return func(iss *gojira.Issue) (map[string]string, error) {
		out := map[string]string{}
		im := NewIssueMore(iss)
		for k, def := range fieldsWithDefaults {
			v := im.ValueOrDefault(k, def)
			if useCustomFieldDisplayNames {
				if can, ok := IsCustomFieldKey(k); !ok {
					out[k] = v
				} else if canName, err := set.IDToName(can); err != nil {
					return out, err
				} else {
					out[canName] = v
				}
			} else {
				out[k] = v
			}
		}
		return out, nil
	}
}
