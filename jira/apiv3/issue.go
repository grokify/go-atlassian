package apiv3

import (
	"encoding/json"
	"strings"
)

// Issue represents a Jira issue from the V3 API
type Issue struct {
	Expand string  `json:"expand"`
	Fields *Fields `json:"fields"`
	ID     string  `json:"id"`
	Key    string  `json:"key"`
	Self   string  `json:"self"`
}

// Fields represents the fields section of a V3 issue
type Fields struct {
	AggregateProgress        *Progress         `json:"aggregateprogress"`
	AggregateTimeEstimate    *int              `json:"aggregatetimeestimate"`
	AggregateTimeOriginalEst *int              `json:"aggregatetimeoriginalestimate"`
	AggregateTimeSpent       *int              `json:"aggregatetimespent"`
	Assignee                 *User             `json:"assignee"`
	Attachment               []Attachment      `json:"attachment"`
	Comment                  *CommentContainer `json:"comment"`
	Components               []Component       `json:"components"`
	Created                  string            `json:"created"`
	Creator                  *User             `json:"creator"`
	Description              any               `json:"description"` // Can be string or ADF object
	DueDate                  *string           `json:"duedate"`
	Environment              any               `json:"environment"` // Can be string or ADF object
	FixVersions              []Version         `json:"fixVersions"`
	IssueLinks               []IssueLink       `json:"issuelinks"`
	IssueRestriction         *IssueRestriction `json:"issuerestriction"`
	IssueType                *IssueType        `json:"issuetype"`
	Labels                   []string          `json:"labels"`
	LastViewed               *string           `json:"lastViewed"`
	Priority                 *Priority         `json:"priority"`
	Progress                 *Progress         `json:"progress"`
	Project                  *Project          `json:"project"`
	Reporter                 *User             `json:"reporter"`
	Resolution               *Resolution       `json:"resolution"`
	ResolutionDate           *string           `json:"resolutiondate"`
	Security                 *SecurityLevel    `json:"security"`
	Status                   *Status           `json:"status"`
	StatusCategory           *StatusCategory   `json:"statusCategory"`
	StatusCategoryChangeDate string            `json:"statuscategorychangedate"`
	Subtasks                 []Issue           `json:"subtasks"`
	Summary                  string            `json:"summary"`
	TimeEstimate             *int              `json:"timeestimate"`
	TimeOriginalEstimate     *int              `json:"timeoriginalestimate"`
	TimeSpent                *int              `json:"timespent"`
	TimeTracking             *TimeTracking     `json:"timetracking"`
	Updated                  string            `json:"updated"`
	Versions                 []Version         `json:"versions"`
	Votes                    *Votes            `json:"votes"`
	Watches                  *Watches          `json:"watches"`
	Worklog                  *WorklogContainer `json:"worklog"`
	WorkRatio                int               `json:"workratio"`

	// Custom fields - using map for flexibility since field IDs vary
	CustomFields map[string]any `json:"-"` // Will be populated separately
}

// IssuesResponse represents the response from the V3 search/jql API
type IssuesResponse struct {
	Issues        []Issue `json:"issues"`
	Expand        string  `json:"expand"`
	StartAt       int     `json:"startAt"`
	MaxResults    int     `json:"maxResults"`
	Total         int     `json:"total"`
	NextPageToken string  `json:"nextPageToken"`
	IsLast        bool    `json:"isLast"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Fields to capture custom fields
func (f *Fields) UnmarshalJSON(data []byte) error {
	// Create an alias type to avoid infinite recursion
	type Alias Fields
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(f),
	}

	// First unmarshal into the alias to populate known fields
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Now unmarshal into a map to capture all fields including custom ones
	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawFields); err != nil {
		return err
	}

	// Initialize CustomFields if it's nil
	if f.CustomFields == nil {
		f.CustomFields = make(map[string]any)
	}

	// Iterate through all fields and capture custom fields
	for key, value := range rawFields {
		if strings.HasPrefix(key, "customfield_") {
			var customValue any
			if err := json.Unmarshal(value, &customValue); err != nil {
				// If unmarshaling fails, store as raw JSON string
				f.CustomFields[key] = string(value)
			} else {
				f.CustomFields[key] = customValue
			}
		}
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling for Fields to include custom fields
func (f *Fields) MarshalJSON() ([]byte, error) {
	// Create an alias type to avoid infinite recursion
	type Alias Fields
	aux := (*Alias)(f)

	// Marshal the known fields first
	fieldsData, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	// If no custom fields, return the basic marshaling
	if len(f.CustomFields) == 0 {
		return fieldsData, nil
	}

	// Unmarshal the basic fields into a map
	var fieldsMap map[string]any
	if err := json.Unmarshal(fieldsData, &fieldsMap); err != nil {
		return nil, err
	}

	// Add custom fields to the map
	for key, value := range f.CustomFields {
		fieldsMap[key] = value
	}

	// Marshal the combined map
	return json.Marshal(fieldsMap)
}

// GetCustomField retrieves a custom field value by its key (e.g., "customfield_12345")
func (f *Fields) GetCustomField(key string) (any, bool) {
	if f.CustomFields == nil {
		return nil, false
	}
	value, exists := f.CustomFields[key]
	return value, exists
}

// SetCustomField sets a custom field value
func (f *Fields) SetCustomField(key string, value any) {
	if f.CustomFields == nil {
		f.CustomFields = make(map[string]any)
	}
	f.CustomFields[key] = value
}

// ListCustomFields returns a slice of all custom field keys
func (f *Fields) ListCustomFields() []string {
	if f.CustomFields == nil {
		return nil
	}

	keys := make([]string, 0, len(f.CustomFields))
	for key := range f.CustomFields {
		keys = append(keys, key)
	}
	return keys
}
