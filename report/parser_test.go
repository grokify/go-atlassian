package report

import (
	"testing"
)

func TestParseYAMLString(t *testing.T) {
	yaml := `
id: test-report
title: Test Report
sections:
  - id: intro
    type: markdown
    content: "# Hello World"
  - id: issues
    type: jql
    config:
      jql: "project = TEST"
      display: table
`
	def, err := ParseYAMLString(yaml)
	if err != nil {
		t.Fatalf("ParseYAMLString() error = %v", err)
	}

	if def.ID != "test-report" {
		t.Errorf("ID = %q, want %q", def.ID, "test-report")
	}
	if def.Title != "Test Report" {
		t.Errorf("Title = %q, want %q", def.Title, "Test Report")
	}
	if len(def.Sections) != 2 {
		t.Errorf("len(Sections) = %d, want 2", len(def.Sections))
	}

	// Check first section
	if def.Sections[0].Type != SectionTypeMarkdown {
		t.Errorf("Sections[0].Type = %q, want %q", def.Sections[0].Type, SectionTypeMarkdown)
	}
	if def.Sections[0].Content != "# Hello World" {
		t.Errorf("Sections[0].Content = %q, want %q", def.Sections[0].Content, "# Hello World")
	}

	// Check second section
	if def.Sections[1].Type != SectionTypeJQL {
		t.Errorf("Sections[1].Type = %q, want %q", def.Sections[1].Type, SectionTypeJQL)
	}
}

func TestParseJSONString(t *testing.T) {
	json := `{
  "id": "test-report",
  "title": "Test Report",
  "sections": [
    {"id": "intro", "type": "markdown", "content": "# Hello"}
  ]
}`
	def, err := ParseJSONString(json)
	if err != nil {
		t.Fatalf("ParseJSONString() error = %v", err)
	}

	if def.ID != "test-report" {
		t.Errorf("ID = %q, want %q", def.ID, "test-report")
	}
	if len(def.Sections) != 1 {
		t.Errorf("len(Sections) = %d, want 1", len(def.Sections))
	}
}

func TestParseYAMLStringInvalid(t *testing.T) {
	// Missing required id field
	yaml := `
title: Test Report
sections:
  - id: intro
    type: markdown
`
	_, err := ParseYAMLString(yaml)
	if err == nil {
		t.Error("ParseYAMLString() expected error for missing id")
	}
}

func TestParseSectionConfig(t *testing.T) {
	section := Section{
		ID:   "test",
		Type: SectionTypeJQL,
		Config: []byte(`{
			"jql": "project = TEST",
			"display": "table",
			"max_results": 100
		}`),
	}

	config, err := section.GetJQLConfig()
	if err != nil {
		t.Fatalf("GetJQLConfig() error = %v", err)
	}
	if config == nil {
		t.Fatal("GetJQLConfig() returned nil")
	}

	if config.JQL != "project = TEST" {
		t.Errorf("JQL = %q, want %q", config.JQL, "project = TEST")
	}
	if config.Display != DisplayTable {
		t.Errorf("Display = %q, want %q", config.Display, DisplayTable)
	}
	if config.MaxResults != 100 {
		t.Errorf("MaxResults = %d, want 100", config.MaxResults)
	}
}

func TestDefinitionToYAML(t *testing.T) {
	def := Definition{
		ID:    "test",
		Title: "Test",
		Sections: []Section{
			{ID: "s1", Type: SectionTypeMarkdown, Content: "Hello"},
		},
	}

	data, err := def.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("ToYAML() returned empty data")
	}
}

func TestDefinitionToJSON(t *testing.T) {
	def := Definition{
		ID:    "test",
		Title: "Test",
		Sections: []Section{
			{ID: "s1", Type: SectionTypeMarkdown, Content: "Hello"},
		},
	}

	data, err := def.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("ToJSON() returned empty data")
	}
}
