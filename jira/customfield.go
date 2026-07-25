package jira

import (
	"errors"
	"io"
	"sort"
	"strings"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/gocharts/v2/data/table"
	"github.com/grokify/mogo/encoding/jsonutil"
	"github.com/grokify/mogo/type/slicesutil"
	"github.com/grokify/mogo/type/stringsutil"
	"github.com/olekukonko/tablewriter"
)

const (
	CustomFieldNameEpicLink = "Epic Link"
)

var ErrJiraRESTClientCannotBeNil = errors.New("rest.Client cannot be nil")

type CustomFields []CustomField

type CustomField struct {
	ID               string            `json:"id"` // "customfield_12345"
	Key              string            `json:"key"`
	Name             string            `json:"name"`
	UntranslatedName string            `json:"untranslatedName"`
	Custom           bool              `json:"custom"`
	Orderable        bool              `json:"orderable"`
	Navigable        bool              `json:"navigable"`
	Searchable       bool              `json:"searchable"`
	ClauseNames      []string          `json:"clauseNames"`
	Schema           CustomFieldSchema `json:"schema"`
}

type CustomFieldSchema struct {
	Type     string `json:"type"`
	Custom   string `json:"custom"`
	CustomID int    `json:"customId"`
}

func (cfs CustomFields) SortByName(asc bool) CustomFields {
	if asc {
		sort.Slice(cfs, func(i, j int) bool {
			return cfs[i].Name < cfs[j].Name
		})
	} else {
		sort.Slice(cfs, func(i, j int) bool {
			return cfs[i].Name > cfs[j].Name
		})
	}
	return cfs
}

func (cfs CustomFields) FilterByIDs(ids ...string) CustomFields {
	filtered := CustomFields{}
	if len(ids) == 0 {
		return filtered
	}
	ids = slicesutil.Dedupe(ids)
	idsMap := map[string]int{}
	for _, id := range ids {
		idsMap[id] = 1
	}
	for _, cf := range cfs {
		if _, ok := idsMap[cf.ID]; ok {
			filtered = append(filtered, cf)
			if len(filtered) == len(ids) {
				return filtered
			}
		}
	}
	return filtered
}

func (cfs CustomFields) FilterByNames(names ...string) CustomFields {
	filtered := CustomFields{}
	if len(names) == 0 {
		return filtered
	}
	names = slicesutil.Dedupe(names)
	namesMap := map[string]int{}
	for _, name := range names {
		namesMap[name] = 1
	}
	for _, cf := range cfs {
		if _, ok := namesMap[cf.Name]; ok {
			filtered = append(filtered, cf)
		}
	}
	return filtered
}

// MapNameToIDs returns a map of field name to slice of IDs. This handles the case
// where multiple custom fields share the same display name (e.g. from copied schemes
// or reinstalled apps).
func (cfs CustomFields) MapNameToIDs() map[string][]string {
	result := make(map[string][]string)
	for _, cf := range cfs {
		result[cf.Name] = append(result[cf.Name], cf.ID)
	}
	return result
}

// MapIDToName returns a map of field ID to name.
func (cfs CustomFields) MapIDToName() map[string]string {
	result := make(map[string]string, len(cfs))
	for _, cf := range cfs {
		result[cf.ID] = cf.Name
	}
	return result
}

// DuplicateNames returns names that appear more than once in the custom fields list.
// This is useful for identifying fields that may cause ambiguity when querying by name.
func (cfs CustomFields) DuplicateNames() []string {
	nameCount := make(map[string]int)
	for _, cf := range cfs {
		nameCount[cf.Name]++
	}
	var duplicates []string
	for name, count := range nameCount {
		if count > 1 {
			duplicates = append(duplicates, name)
		}
	}
	sort.Strings(duplicates)
	return duplicates
}

// SuggestSimilar returns field names that are similar to the query.
// It uses case-insensitive prefix matching, substring matching, and word matching.
// The maxResults parameter limits the number of suggestions returned.
func (cfs CustomFields) SuggestSimilar(query string, maxResults int) []FieldSuggestion {
	if maxResults <= 0 {
		maxResults = 5
	}

	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	queryWords := strings.Fields(query)

	type scoredField struct {
		field CustomField
		score int
	}

	var scored []scoredField

	for _, cf := range cfs {
		nameLower := strings.ToLower(cf.Name)
		score := 0

		// Exact match (highest priority)
		if nameLower == query {
			score = 100
		} else if strings.HasPrefix(nameLower, query) {
			// Prefix match
			score = 80
		} else if strings.Contains(nameLower, query) {
			// Substring match
			score = 60
		} else {
			// Word matching - check if any query words appear in the field name
			for _, word := range queryWords {
				if len(word) >= 3 && strings.Contains(nameLower, word) {
					score += 20
				}
			}
			// Also check if field name words appear in query
			fieldWords := strings.Fields(nameLower)
			for _, word := range fieldWords {
				if len(word) >= 3 && strings.Contains(query, word) {
					score += 15
				}
			}
		}

		if score > 0 {
			scored = append(scored, scoredField{field: cf, score: score})
		}
	}

	// Sort by score descending, then by name
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].field.Name < scored[j].field.Name
	})

	// Limit results
	if len(scored) > maxResults {
		scored = scored[:maxResults]
	}

	// Convert to suggestions
	suggestions := make([]FieldSuggestion, len(scored))
	for i, sf := range scored {
		suggestions[i] = FieldSuggestion{
			ID:    sf.field.ID,
			Name:  sf.field.Name,
			Score: sf.score,
		}
	}

	return suggestions
}

// FieldSuggestion represents a suggested field with a relevance score.
type FieldSuggestion struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

// FindByNameWithSuggestions looks up a field by name and returns suggestions if not found.
func (cfs CustomFields) FindByNameWithSuggestions(name string) (*CustomField, []FieldSuggestion) {
	nameLower := strings.ToLower(strings.TrimSpace(name))

	// First try exact match (case-insensitive)
	for i := range cfs {
		if strings.ToLower(cfs[i].Name) == nameLower {
			return &cfs[i], nil
		}
	}

	// Not found, return suggestions
	return nil, cfs.SuggestSimilar(name, 5)
}

func (cfs CustomFields) Table(name string) table.Table {
	if strings.TrimSpace(name) == "" {
		name = "Custom Fields"
	}
	tbl := table.NewTable(name)
	tbl.Columns = []string{"Name", "ID", "Clause Names"}
	for _, cf := range cfs {
		row := []string{
			cf.Name, cf.ID, stringsutil.JoinLiteraryQuote(cf.ClauseNames, `"`, `"`, `, `, ""),
		}
		tbl.Rows = append(tbl.Rows, row)
	}
	return tbl
}

func (cfs CustomFields) WriteTable(w io.Writer) error {
	cfs.SortByName(true)
	tbl := cfs.Table("")
	tw := tablewriter.NewWriter(w)
	tw.Header(tbl.Columns)
	if err := tw.Bulk(tbl.Rows); err != nil {
		return err
	} else {
		return tw.Render()
	}
}

// IssueFieldsCustomFieldString returns a string custom field, e.g "Epic Link"
func IssueFieldsCustomFieldString(fields *gojira.IssueFields, id string) string {
	if fields == nil {
		return ""
	}
	val, err := fields.Unknowns.String(id)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(val)
}

// GetCustomValueString attempts to return a string if either the custom value is a simple string
// or is an `IssueCustomField`, in which case it returns the `value` property.
func GetCustomValueString(iss gojira.Issue, customFieldKey string) (string, error) {
	if iss.Fields == nil {
		return "", nil
	}
	any, ok := iss.Fields.Unknowns[customFieldKey]
	if !ok {
		return "", nil
	}
	if strval, ok := any.(string); ok {
		return strval, nil
	}
	icf := &IssueCustomField{}
	err := GetUnmarshalCustomValue(iss, customFieldKey, icf)
	if err != nil {
		return "", err
	}
	return icf.Value, nil
}

// GetUnmarshalCustomValue can be used to unmarshal a value to `IssueCustomField{}`.
func GetUnmarshalCustomValue(iss gojira.Issue, customFieldKey string, v *IssueCustomField) error {
	if iss.Fields == nil {
		return nil
	} else if key, err := CustomFieldKeyCanonical(customFieldKey); err != nil {
		return err
	} else if unv, ok := iss.Fields.Unknowns[key]; !ok {
		return nil
	} else {
		return jsonutil.UnmarshalAny(unv, v)
	}
}

type IssueCustomField struct {
	ID    string `json:"id"`
	Self  string `json:"self"`
	Value string `json:"value"`
}
