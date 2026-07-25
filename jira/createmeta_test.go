package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// writeJSON is a helper to write JSON responses in tests.
func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("failed to encode JSON: %v", err)
	}
}

// newCreateMetaMockServer creates a mock server for CreateMeta API tests.
func newCreateMetaMockServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Mock createmeta issue types endpoint
	mux.HandleFunc("/rest/api/3/issue/createmeta/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Handle /rest/api/3/issue/createmeta/{projectKey}/issuetypes
		if path == "/rest/api/3/issue/createmeta/TEST/issuetypes" {
			response := CreateMetaIssueTypesResponse{
				MaxResults: 50,
				StartAt:    0,
				Total:      3,
				IssueTypes: []CreateMetaIssueType{
					{ID: "10001", Name: "Bug", Description: "A bug in the system", Subtask: false},
					{ID: "10002", Name: "Story", Description: "A user story", Subtask: false},
					{ID: "10003", Name: "Sub-task", Description: "A subtask", Subtask: true},
				},
			}
			writeJSON(t, w, response)
			return
		}

		// Handle /rest/api/3/issue/createmeta/{projectKey}/issuetypes/{issueTypeId}
		if path == "/rest/api/3/issue/createmeta/TEST/issuetypes/10001" {
			response := CreateMetaFieldsResponse{
				MaxResults: 50,
				StartAt:    0,
				Total:      4,
				Values: []CreateMetaField{
					{Key: "summary", Name: "Summary", Required: true, HasDefaultValue: false},
					{Key: "description", Name: "Description", Required: false, HasDefaultValue: false},
					{Key: "customfield_10001", Name: "Epic Link", Required: false, HasDefaultValue: false},
					{Key: "customfield_10002", Name: "Sprint", Required: true, HasDefaultValue: false},
				},
			}
			writeJSON(t, w, response)
			return
		}

		if path == "/rest/api/3/issue/createmeta/TEST/issuetypes/10002" {
			response := CreateMetaFieldsResponse{
				MaxResults: 50,
				StartAt:    0,
				Total:      3,
				Values: []CreateMetaField{
					{Key: "summary", Name: "Summary", Required: true, HasDefaultValue: false},
					{Key: "customfield_10003", Name: "Story Points", Required: false, HasDefaultValue: false},
					{Key: "customfield_10001", Name: "Epic Link", Required: false, HasDefaultValue: false},
				},
			}
			writeJSON(t, w, response)
			return
		}

		if path == "/rest/api/3/issue/createmeta/TEST/issuetypes/10003" {
			response := CreateMetaFieldsResponse{
				MaxResults: 50,
				StartAt:    0,
				Total:      2,
				Values: []CreateMetaField{
					{Key: "summary", Name: "Summary", Required: true, HasDefaultValue: false},
					{Key: "parent", Name: "Parent", Required: true, HasDefaultValue: false},
				},
			}
			writeJSON(t, w, response)
			return
		}

		// Project not found
		if path == "/rest/api/3/issue/createmeta/NOTFOUND/issuetypes" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	return httptest.NewServer(mux)
}

func TestCreateMetaService_GetIssueTypes(t *testing.T) {
	server := newCreateMetaMockServer(t)
	defer server.Close()

	client, err := NewClientFromBasicAuth(server.URL, "testuser", "testtoken", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("valid project", func(t *testing.T) {
		issueTypes, err := client.CreateMetaAPI.GetIssueTypes(ctx, "TEST")
		if err != nil {
			t.Fatalf("GetIssueTypes() error = %v", err)
		}

		if len(issueTypes) != 3 {
			t.Errorf("GetIssueTypes() returned %d issue types, want 3", len(issueTypes))
		}

		// Check first issue type
		if issueTypes[0].ID != "10001" || issueTypes[0].Name != "Bug" {
			t.Errorf("GetIssueTypes()[0] = %+v, want ID=10001, Name=Bug", issueTypes[0])
		}

		// Check subtask flag
		if issueTypes[2].Subtask != true {
			t.Errorf("GetIssueTypes()[2].Subtask = false, want true")
		}
	})

	t.Run("project not found", func(t *testing.T) {
		_, err := client.CreateMetaAPI.GetIssueTypes(ctx, "NOTFOUND")
		if err == nil {
			t.Error("GetIssueTypes() expected error for non-existent project")
		}
	})
}

func TestCreateMetaService_GetFields(t *testing.T) {
	server := newCreateMetaMockServer(t)
	defer server.Close()

	client, err := NewClientFromBasicAuth(server.URL, "testuser", "testtoken", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("Bug issue type", func(t *testing.T) {
		fields, err := client.CreateMetaAPI.GetFields(ctx, "TEST", "10001")
		if err != nil {
			t.Fatalf("GetFields() error = %v", err)
		}

		if len(fields) != 4 {
			t.Errorf("GetFields() returned %d fields, want 4", len(fields))
		}

		// Check required fields
		required := fields.RequiredOnly()
		if len(required) != 2 {
			t.Errorf("RequiredOnly() returned %d fields, want 2", len(required))
		}

		// Check custom fields
		custom := fields.CustomOnly()
		if len(custom) != 2 {
			t.Errorf("CustomOnly() returned %d fields, want 2", len(custom))
		}
	})

	t.Run("Story issue type", func(t *testing.T) {
		fields, err := client.CreateMetaAPI.GetFields(ctx, "TEST", "10002")
		if err != nil {
			t.Fatalf("GetFields() error = %v", err)
		}

		if len(fields) != 3 {
			t.Errorf("GetFields() returned %d fields, want 3", len(fields))
		}
	})
}

func TestCreateMetaService_GetAllFieldsForProject(t *testing.T) {
	server := newCreateMetaMockServer(t)
	defer server.Close()

	client, err := NewClientFromBasicAuth(server.URL, "testuser", "testtoken", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	fields, err := client.CreateMetaAPI.GetAllFieldsForProject(ctx, "TEST")
	if err != nil {
		t.Fatalf("GetAllFieldsForProject() error = %v", err)
	}

	// Should have unique fields from all issue types
	// Bug: summary, description, customfield_10001, customfield_10002
	// Story: summary, customfield_10003, customfield_10001
	// Sub-task: summary, parent
	// Unique: summary, description, customfield_10001, customfield_10002, customfield_10003, parent = 6

	if len(fields) != 6 {
		t.Errorf("GetAllFieldsForProject() returned %d fields, want 6", len(fields))
	}

	// Verify custom fields
	customFields := fields.CustomOnly()
	if len(customFields) != 3 {
		t.Errorf("CustomOnly() returned %d fields, want 3", len(customFields))
	}
}

func TestCreateMetaFields_CustomOnly(t *testing.T) {
	fields := CreateMetaFields{
		{Key: "summary", Name: "Summary"},
		{Key: "customfield_10001", Name: "Epic Link"},
		{Key: "description", Name: "Description"},
		{Key: "customfield_10002", Name: "Sprint"},
	}

	custom := fields.CustomOnly()
	if len(custom) != 2 {
		t.Errorf("CustomOnly() returned %d fields, want 2", len(custom))
	}

	for _, f := range custom {
		if !f.IsCustomField() {
			t.Errorf("CustomOnly() returned non-custom field: %s", f.Key)
		}
	}
}

func TestCreateMetaFields_RequiredOnly(t *testing.T) {
	fields := CreateMetaFields{
		{Key: "summary", Name: "Summary", Required: true},
		{Key: "customfield_10001", Name: "Epic Link", Required: false},
		{Key: "description", Name: "Description", Required: false},
		{Key: "customfield_10002", Name: "Sprint", Required: true},
	}

	required := fields.RequiredOnly()
	if len(required) != 2 {
		t.Errorf("RequiredOnly() returned %d fields, want 2", len(required))
	}

	for _, f := range required {
		if !f.Required {
			t.Errorf("RequiredOnly() returned non-required field: %s", f.Key)
		}
	}
}

func TestCreateMetaFields_Keys(t *testing.T) {
	fields := CreateMetaFields{
		{Key: "summary", Name: "Summary"},
		{Key: "customfield_10001", Name: "Epic Link"},
	}

	keys := fields.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys() returned %d keys, want 2", len(keys))
	}
	if keys[0] != "summary" || keys[1] != "customfield_10001" {
		t.Errorf("Keys() = %v, want [summary customfield_10001]", keys)
	}
}

func TestCreateMetaFields_ByKey(t *testing.T) {
	fields := CreateMetaFields{
		{Key: "summary", Name: "Summary"},
		{Key: "customfield_10001", Name: "Epic Link"},
	}

	byKey := fields.ByKey()
	if len(byKey) != 2 {
		t.Errorf("ByKey() returned %d entries, want 2", len(byKey))
	}

	if f, ok := byKey["summary"]; !ok || f.Name != "Summary" {
		t.Errorf("ByKey()[summary] not found or incorrect")
	}
	if f, ok := byKey["customfield_10001"]; !ok || f.Name != "Epic Link" {
		t.Errorf("ByKey()[customfield_10001] not found or incorrect")
	}
}

func TestCreateMetaField_IsCustomField(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"customfield_10001", true},
		{"customfield_99999", true},
		{"summary", false},
		{"description", false},
		{"custom_field", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			f := CreateMetaField{Key: tt.key}
			if got := f.IsCustomField(); got != tt.want {
				t.Errorf("IsCustomField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateMetaService_GetRequiredFields(t *testing.T) {
	server := newCreateMetaMockServer(t)
	defer server.Close()

	client, err := NewClientFromBasicAuth(server.URL, "testuser", "testtoken", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("Bug issue type", func(t *testing.T) {
		// Bug has 2 required fields: summary and customfield_10002 (Sprint)
		fields, err := client.CreateMetaAPI.GetRequiredFields(ctx, "TEST", "10001")
		if err != nil {
			t.Fatalf("GetRequiredFields() error = %v", err)
		}

		if len(fields) != 2 {
			t.Errorf("GetRequiredFields() returned %d fields, want 2", len(fields))
		}

		// Verify all returned fields are required
		for _, f := range fields {
			if !f.Required {
				t.Errorf("GetRequiredFields() returned non-required field: %s", f.Key)
			}
		}
	})

	t.Run("Story issue type", func(t *testing.T) {
		// Story has 1 required field: summary
		fields, err := client.CreateMetaAPI.GetRequiredFields(ctx, "TEST", "10002")
		if err != nil {
			t.Fatalf("GetRequiredFields() error = %v", err)
		}

		if len(fields) != 1 {
			t.Errorf("GetRequiredFields() returned %d fields, want 1", len(fields))
		}

		if fields[0].Key != "summary" {
			t.Errorf("GetRequiredFields()[0].Key = %q, want %q", fields[0].Key, "summary")
		}
	})

	t.Run("Sub-task issue type", func(t *testing.T) {
		// Sub-task has 2 required fields: summary and parent
		fields, err := client.CreateMetaAPI.GetRequiredFields(ctx, "TEST", "10003")
		if err != nil {
			t.Fatalf("GetRequiredFields() error = %v", err)
		}

		if len(fields) != 2 {
			t.Errorf("GetRequiredFields() returned %d fields, want 2", len(fields))
		}
	})
}

func TestCreateMetaService_GetRequiredFieldsForProject(t *testing.T) {
	server := newCreateMetaMockServer(t)
	defer server.Close()

	client, err := NewClientFromBasicAuth(server.URL, "testuser", "testtoken", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	fields, err := client.CreateMetaAPI.GetRequiredFieldsForProject(ctx, "TEST")
	if err != nil {
		t.Fatalf("GetRequiredFieldsForProject() error = %v", err)
	}

	// Required fields across all issue types:
	// Bug: summary, customfield_10002 (Sprint)
	// Story: summary
	// Sub-task: summary, parent
	// Unique required: summary, customfield_10002, parent = 3

	if len(fields) != 3 {
		t.Errorf("GetRequiredFieldsForProject() returned %d fields, want 3", len(fields))
	}

	// Check that summary is required by all 3 issue types
	for _, info := range fields {
		if info.Field.Key == "summary" {
			if len(info.IssueTypes) != 3 {
				t.Errorf("summary field required by %d issue types, want 3", len(info.IssueTypes))
			}
		}
		if info.Field.Key == "parent" {
			if len(info.IssueTypes) != 1 {
				t.Errorf("parent field required by %d issue types, want 1", len(info.IssueTypes))
			}
		}
		if info.Field.Key == "customfield_10002" {
			if len(info.IssueTypes) != 1 {
				t.Errorf("customfield_10002 field required by %d issue types, want 1", len(info.IssueTypes))
			}
		}
	}
}
