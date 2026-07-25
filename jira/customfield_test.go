package jira

import (
	"reflect"
	"sort"
	"testing"
)

func TestCustomFields_MapNameToIDs(t *testing.T) {
	tests := []struct {
		name   string
		fields CustomFields
		want   map[string][]string
	}{
		{
			name:   "empty fields",
			fields: CustomFields{},
			want:   map[string][]string{},
		},
		{
			name: "unique names",
			fields: CustomFields{
				{ID: "customfield_10001", Name: "Sprint"},
				{ID: "customfield_10002", Name: "Epic Link"},
			},
			want: map[string][]string{
				"Sprint":    {"customfield_10001"},
				"Epic Link": {"customfield_10002"},
			},
		},
		{
			name: "duplicate names",
			fields: CustomFields{
				{ID: "customfield_10001", Name: "Module"},
				{ID: "customfield_10002", Name: "Sprint"},
				{ID: "customfield_10003", Name: "Module"},
			},
			want: map[string][]string{
				"Module": {"customfield_10001", "customfield_10003"},
				"Sprint": {"customfield_10002"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fields.MapNameToIDs()

			if len(got) != len(tt.want) {
				t.Errorf("MapNameToIDs() returned %d entries, want %d", len(got), len(tt.want))
				return
			}

			for name, wantIDs := range tt.want {
				gotIDs, ok := got[name]
				if !ok {
					t.Errorf("MapNameToIDs() missing key %q", name)
					continue
				}
				// Sort for comparison since map iteration order is random
				sort.Strings(gotIDs)
				sort.Strings(wantIDs)
				if !reflect.DeepEqual(gotIDs, wantIDs) {
					t.Errorf("MapNameToIDs()[%q] = %v, want %v", name, gotIDs, wantIDs)
				}
			}
		})
	}
}

func TestCustomFields_MapIDToName(t *testing.T) {
	tests := []struct {
		name   string
		fields CustomFields
		want   map[string]string
	}{
		{
			name:   "empty fields",
			fields: CustomFields{},
			want:   map[string]string{},
		},
		{
			name: "multiple fields",
			fields: CustomFields{
				{ID: "customfield_10001", Name: "Sprint"},
				{ID: "customfield_10002", Name: "Epic Link"},
				{ID: "customfield_10003", Name: "Story Points"},
			},
			want: map[string]string{
				"customfield_10001": "Sprint",
				"customfield_10002": "Epic Link",
				"customfield_10003": "Story Points",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fields.MapIDToName()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapIDToName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomFields_DuplicateNames(t *testing.T) {
	tests := []struct {
		name   string
		fields CustomFields
		want   []string
	}{
		{
			name:   "empty fields",
			fields: CustomFields{},
			want:   nil,
		},
		{
			name: "no duplicates",
			fields: CustomFields{
				{ID: "customfield_10001", Name: "Sprint"},
				{ID: "customfield_10002", Name: "Epic Link"},
			},
			want: nil,
		},
		{
			name: "one duplicate",
			fields: CustomFields{
				{ID: "customfield_10001", Name: "Module"},
				{ID: "customfield_10002", Name: "Sprint"},
				{ID: "customfield_10003", Name: "Module"},
			},
			want: []string{"Module"},
		},
		{
			name: "multiple duplicates",
			fields: CustomFields{
				{ID: "customfield_10001", Name: "Module"},
				{ID: "customfield_10002", Name: "Sprint"},
				{ID: "customfield_10003", Name: "Module"},
				{ID: "customfield_10004", Name: "Sprint"},
				{ID: "customfield_10005", Name: "Epic Link"},
			},
			want: []string{"Module", "Sprint"},
		},
		{
			name: "triple duplicate",
			fields: CustomFields{
				{ID: "customfield_10001", Name: "Module"},
				{ID: "customfield_10002", Name: "Module"},
				{ID: "customfield_10003", Name: "Module"},
			},
			want: []string{"Module"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fields.DuplicateNames()
			// DuplicateNames returns sorted slice
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DuplicateNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomFields_FilterByIDs(t *testing.T) {
	fields := CustomFields{
		{ID: "customfield_10001", Name: "Sprint"},
		{ID: "customfield_10002", Name: "Epic Link"},
		{ID: "customfield_10003", Name: "Story Points"},
	}

	tests := []struct {
		name    string
		ids     []string
		wantLen int
		wantIDs []string
	}{
		{
			name:    "no IDs",
			ids:     []string{},
			wantLen: 0,
			wantIDs: nil,
		},
		{
			name:    "single ID found",
			ids:     []string{"customfield_10001"},
			wantLen: 1,
			wantIDs: []string{"customfield_10001"},
		},
		{
			name:    "multiple IDs found",
			ids:     []string{"customfield_10001", "customfield_10003"},
			wantLen: 2,
			wantIDs: []string{"customfield_10001", "customfield_10003"},
		},
		{
			name:    "ID not found",
			ids:     []string{"customfield_99999"},
			wantLen: 0,
			wantIDs: nil,
		},
		{
			name:    "mixed found and not found",
			ids:     []string{"customfield_10001", "customfield_99999"},
			wantLen: 1,
			wantIDs: []string{"customfield_10001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fields.FilterByIDs(tt.ids...)
			if len(got) != tt.wantLen {
				t.Errorf("FilterByIDs() returned %d fields, want %d", len(got), tt.wantLen)
				return
			}
			for i, f := range got {
				if tt.wantIDs != nil && f.ID != tt.wantIDs[i] {
					t.Errorf("FilterByIDs()[%d].ID = %q, want %q", i, f.ID, tt.wantIDs[i])
				}
			}
		})
	}
}

func TestCustomFields_FilterByNames(t *testing.T) {
	fields := CustomFields{
		{ID: "customfield_10001", Name: "Sprint"},
		{ID: "customfield_10002", Name: "Epic Link"},
		{ID: "customfield_10003", Name: "Sprint"}, // duplicate name
	}

	tests := []struct {
		name      string
		names     []string
		wantLen   int
		wantNames []string
	}{
		{
			name:      "no names",
			names:     []string{},
			wantLen:   0,
			wantNames: nil,
		},
		{
			name:      "single name found",
			names:     []string{"Epic Link"},
			wantLen:   1,
			wantNames: []string{"Epic Link"},
		},
		{
			name:      "duplicate name returns all",
			names:     []string{"Sprint"},
			wantLen:   2,
			wantNames: []string{"Sprint", "Sprint"},
		},
		{
			name:      "name not found",
			names:     []string{"Not Exist"},
			wantLen:   0,
			wantNames: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fields.FilterByNames(tt.names...)
			if len(got) != tt.wantLen {
				t.Errorf("FilterByNames() returned %d fields, want %d", len(got), tt.wantLen)
				return
			}
			for i, f := range got {
				if tt.wantNames != nil && f.Name != tt.wantNames[i] {
					t.Errorf("FilterByNames()[%d].Name = %q, want %q", i, f.Name, tt.wantNames[i])
				}
			}
		})
	}
}

func TestCustomFields_SortByName(t *testing.T) {
	fields := CustomFields{
		{ID: "customfield_10003", Name: "Zebra"},
		{ID: "customfield_10001", Name: "Apple"},
		{ID: "customfield_10002", Name: "Banana"},
	}

	t.Run("ascending", func(t *testing.T) {
		sorted := fields.SortByName(true)
		if sorted[0].Name != "Apple" || sorted[1].Name != "Banana" || sorted[2].Name != "Zebra" {
			t.Errorf("SortByName(true) order incorrect: %v", []string{sorted[0].Name, sorted[1].Name, sorted[2].Name})
		}
	})

	t.Run("descending", func(t *testing.T) {
		sorted := fields.SortByName(false)
		if sorted[0].Name != "Zebra" || sorted[1].Name != "Banana" || sorted[2].Name != "Apple" {
			t.Errorf("SortByName(false) order incorrect: %v", []string{sorted[0].Name, sorted[1].Name, sorted[2].Name})
		}
	})
}

func TestCustomFields_SuggestSimilar(t *testing.T) {
	fields := CustomFields{
		{ID: "customfield_10001", Name: "Sprint"},
		{ID: "customfield_10002", Name: "Epic Link"},
		{ID: "customfield_10003", Name: "Story Points"},
		{ID: "customfield_10004", Name: "Sprint Goal"},
		{ID: "customfield_10005", Name: "Release Version"},
	}

	tests := []struct {
		name       string
		query      string
		maxResults int
		wantFirst  string
		wantLen    int
	}{
		{
			name:       "exact match",
			query:      "Sprint",
			maxResults: 5,
			wantFirst:  "Sprint",
			wantLen:    2, // Sprint and Sprint Goal
		},
		{
			name:       "prefix match",
			query:      "Spr",
			maxResults: 5,
			wantFirst:  "Sprint",
			wantLen:    2,
		},
		{
			name:       "substring match",
			query:      "Link",
			maxResults: 5,
			wantFirst:  "Epic Link",
			wantLen:    1,
		},
		{
			name:       "word match",
			query:      "story",
			maxResults: 5,
			wantFirst:  "Story Points",
			wantLen:    1,
		},
		{
			name:       "case insensitive",
			query:      "SPRINT",
			maxResults: 5,
			wantFirst:  "Sprint",
			wantLen:    2,
		},
		{
			name:       "no match",
			query:      "xyz123",
			maxResults: 5,
			wantLen:    0,
		},
		{
			name:       "empty query",
			query:      "",
			maxResults: 5,
			wantLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := fields.SuggestSimilar(tt.query, tt.maxResults)

			if len(suggestions) != tt.wantLen {
				t.Errorf("SuggestSimilar(%q) returned %d suggestions, want %d", tt.query, len(suggestions), tt.wantLen)
				return
			}

			if tt.wantLen > 0 && suggestions[0].Name != tt.wantFirst {
				t.Errorf("SuggestSimilar(%q)[0].Name = %q, want %q", tt.query, suggestions[0].Name, tt.wantFirst)
			}
		})
	}
}

func TestCustomFields_FindByNameWithSuggestions(t *testing.T) {
	fields := CustomFields{
		{ID: "customfield_10001", Name: "Sprint"},
		{ID: "customfield_10002", Name: "Epic Link"},
		{ID: "customfield_10003", Name: "Story Points"},
	}

	t.Run("exact match found", func(t *testing.T) {
		field, suggestions := fields.FindByNameWithSuggestions("Sprint")
		if field == nil {
			t.Error("FindByNameWithSuggestions() field = nil, want non-nil")
			return
		}
		if field.Name != "Sprint" {
			t.Errorf("FindByNameWithSuggestions().Name = %q, want %q", field.Name, "Sprint")
		}
		if len(suggestions) != 0 {
			t.Errorf("FindByNameWithSuggestions() suggestions = %d, want 0 when found", len(suggestions))
		}
	})

	t.Run("case insensitive match", func(t *testing.T) {
		field, suggestions := fields.FindByNameWithSuggestions("sprint")
		if field == nil {
			t.Error("FindByNameWithSuggestions() field = nil, want non-nil")
			return
		}
		if field.Name != "Sprint" {
			t.Errorf("FindByNameWithSuggestions().Name = %q, want %q", field.Name, "Sprint")
		}
		if len(suggestions) != 0 {
			t.Errorf("FindByNameWithSuggestions() suggestions = %d, want 0 when found", len(suggestions))
		}
	})

	t.Run("not found with suggestions", func(t *testing.T) {
		field, suggestions := fields.FindByNameWithSuggestions("Sprin")
		if field != nil {
			t.Error("FindByNameWithSuggestions() field = non-nil, want nil for partial match")
		}
		if len(suggestions) == 0 {
			t.Error("FindByNameWithSuggestions() suggestions empty, want suggestions")
		}
		if suggestions[0].Name != "Sprint" {
			t.Errorf("FindByNameWithSuggestions() first suggestion = %q, want %q", suggestions[0].Name, "Sprint")
		}
	})

	t.Run("completely unknown", func(t *testing.T) {
		field, suggestions := fields.FindByNameWithSuggestions("xyz123")
		if field != nil {
			t.Error("FindByNameWithSuggestions() field = non-nil, want nil")
		}
		if len(suggestions) != 0 {
			t.Errorf("FindByNameWithSuggestions() suggestions = %d, want 0 for no match", len(suggestions))
		}
	})
}
