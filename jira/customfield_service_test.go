package jira

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newCustomFieldMockServer creates a mock server for CustomFieldService tests.
func newCustomFieldMockServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Mock custom fields endpoint
	mux.HandleFunc("/rest/api/2/field", func(w http.ResponseWriter, r *http.Request) {
		response := []CustomField{
			{
				ID:   "customfield_10001",
				Name: "Sprint",
				Key:  "customfield_10001",
				Schema: CustomFieldSchema{
					Type:     "array",
					Custom:   "com.pyxis.greenhopper.jira:gh-sprint",
					CustomID: 10001,
				},
			},
			{
				ID:   "customfield_10002",
				Name: "Epic Link",
				Key:  "customfield_10002",
				Schema: CustomFieldSchema{
					Type:     "any",
					Custom:   "com.pyxis.greenhopper.jira:gh-epic-link",
					CustomID: 10002,
				},
			},
			{
				ID:   "customfield_10003",
				Name: "Story Points",
				Key:  "customfield_10003",
				Schema: CustomFieldSchema{
					Type:     "number",
					Custom:   "com.atlassian.jira.plugin.system.customfieldtypes:float",
					CustomID: 10003,
				},
			},
			{
				ID:   "customfield_10004",
				Name: "Sprint", // Duplicate name intentionally
				Key:  "customfield_10004",
				Schema: CustomFieldSchema{
					Type:     "array",
					Custom:   "com.pyxis.greenhopper.jira:gh-sprint",
					CustomID: 10004,
				},
			},
		}
		writeJSON(t, w, response)
	})

	// Mock createmeta endpoints for project-based tests
	mux.HandleFunc("/rest/api/3/issue/createmeta/TEST/issuetypes", func(w http.ResponseWriter, r *http.Request) {
		response := CreateMetaIssueTypesResponse{
			MaxResults: 50,
			StartAt:    0,
			Total:      1,
			IssueTypes: []CreateMetaIssueType{
				{ID: "10001", Name: "Bug", Description: "A bug", Subtask: false},
			},
		}
		writeJSON(t, w, response)
	})

	mux.HandleFunc("/rest/api/3/issue/createmeta/TEST/issuetypes/10001", func(w http.ResponseWriter, r *http.Request) {
		response := CreateMetaFieldsResponse{
			MaxResults: 50,
			StartAt:    0,
			Total:      3,
			Values: []CreateMetaField{
				{Key: "summary", Name: "Summary", Required: true},
				{Key: "customfield_10001", Name: "Sprint", Required: false},
				{Key: "customfield_10003", Name: "Story Points", Required: false},
			},
		}
		writeJSON(t, w, response)
	})

	return httptest.NewServer(mux)
}

func TestCustomFieldService_GetCustomFields(t *testing.T) {
	server := newCustomFieldMockServer(t)
	defer server.Close()

	client, err := NewClientFromBasicAuth(server.URL, "testuser", "testtoken", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	fields, err := client.CustomFieldAPI.GetCustomFields()
	if err != nil {
		t.Fatalf("GetCustomFields() error = %v", err)
	}

	if len(fields) != 4 {
		t.Errorf("GetCustomFields() returned %d fields, want 4", len(fields))
	}
}

func TestCustomFieldService_GetCustomField(t *testing.T) {
	server := newCustomFieldMockServer(t)
	defer server.Close()

	client, err := NewClientFromBasicAuth(server.URL, "testuser", "testtoken", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	t.Run("unique field found", func(t *testing.T) {
		field, err := client.CustomFieldAPI.GetCustomField("Epic Link")
		if err != nil {
			t.Fatalf("GetCustomField() error = %v", err)
		}
		if field.ID != "customfield_10002" {
			t.Errorf("GetCustomField().ID = %q, want %q", field.ID, "customfield_10002")
		}
	})

	t.Run("duplicate field name returns error", func(t *testing.T) {
		_, err := client.CustomFieldAPI.GetCustomField("Sprint")
		if err == nil {
			t.Error("GetCustomField() expected error for duplicate field name")
		}
	})

	t.Run("field not found", func(t *testing.T) {
		_, err := client.CustomFieldAPI.GetCustomField("Not Exist")
		if err == nil {
			t.Error("GetCustomField() expected error for non-existent field")
		}
	})
}

func TestCustomFieldService_GetCustomFieldsByName(t *testing.T) {
	server := newCustomFieldMockServer(t)
	defer server.Close()

	client, err := NewClientFromBasicAuth(server.URL, "testuser", "testtoken", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	t.Run("returns all matching fields", func(t *testing.T) {
		fields, err := client.CustomFieldAPI.GetCustomFieldsByName("Sprint")
		if err != nil {
			t.Fatalf("GetCustomFieldsByName() error = %v", err)
		}
		if len(fields) != 2 {
			t.Errorf("GetCustomFieldsByName() returned %d fields, want 2", len(fields))
		}
	})

	t.Run("returns empty for no match", func(t *testing.T) {
		fields, err := client.CustomFieldAPI.GetCustomFieldsByName("Not Exist")
		if err != nil {
			t.Fatalf("GetCustomFieldsByName() error = %v", err)
		}
		if len(fields) != 0 {
			t.Errorf("GetCustomFieldsByName() returned %d fields, want 0", len(fields))
		}
	})
}

func TestCustomFieldService_GetCustomFieldByID(t *testing.T) {
	server := newCustomFieldMockServer(t)
	defer server.Close()

	client, err := NewClientFromBasicAuth(server.URL, "testuser", "testtoken", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	t.Run("field found", func(t *testing.T) {
		field, err := client.CustomFieldAPI.GetCustomFieldByID("customfield_10001")
		if err != nil {
			t.Fatalf("GetCustomFieldByID() error = %v", err)
		}
		if field.Name != "Sprint" {
			t.Errorf("GetCustomFieldByID().Name = %q, want %q", field.Name, "Sprint")
		}
	})

	t.Run("field not found", func(t *testing.T) {
		_, err := client.CustomFieldAPI.GetCustomFieldByID("customfield_99999")
		if err == nil {
			t.Error("GetCustomFieldByID() expected error for non-existent ID")
		}
	})
}

func TestCustomFieldService_GetCustomFieldSet(t *testing.T) {
	server := newCustomFieldMockServer(t)
	defer server.Close()

	client, err := NewClientFromBasicAuth(server.URL, "testuser", "testtoken", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	set, err := client.CustomFieldAPI.GetCustomFieldSet()
	if err != nil {
		t.Fatalf("GetCustomFieldSet() error = %v", err)
	}

	if set == nil {
		t.Fatal("GetCustomFieldSet() returned nil")
	}
}

func TestCustomFieldService_GetCustomFieldsForProject(t *testing.T) {
	server := newCustomFieldMockServer(t)
	defer server.Close()

	client, err := NewClientFromBasicAuth(server.URL, "testuser", "testtoken", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	fields, err := client.CustomFieldAPI.GetCustomFieldsForProject(ctx, "TEST")
	if err != nil {
		t.Fatalf("GetCustomFieldsForProject() error = %v", err)
	}

	// The createmeta returns customfield_10001 and customfield_10003
	// These should be enriched with full metadata from the field list
	if len(fields) != 2 {
		t.Errorf("GetCustomFieldsForProject() returned %d fields, want 2", len(fields))
	}

	// Verify the fields have full metadata (Schema field populated)
	for _, f := range fields {
		if f.Schema.Type == "" {
			t.Errorf("GetCustomFieldsForProject() field %s missing Schema.Type", f.ID)
		}
	}
}

func TestCustomFieldService_NilClient(t *testing.T) {
	svc := &CustomFieldService{JRClient: nil}

	_, err := svc.GetCustomFields()
	if err == nil {
		t.Error("GetCustomFields() expected error for nil client")
	}
}
