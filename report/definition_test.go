package report

import (
	"strings"
	"testing"
)

func TestDefinitionValidate(t *testing.T) {
	tests := []struct {
		name    string
		def     Definition
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty definition",
			def:     Definition{},
			wantErr: true,
			errMsg:  "id is required",
		},
		{
			name:    "missing title",
			def:     Definition{ID: "test"},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name:    "missing sections",
			def:     Definition{ID: "test", Title: "Test"},
			wantErr: true,
			errMsg:  "at least one section is required",
		},
		{
			name: "valid definition",
			def: Definition{
				ID:    "test",
				Title: "Test Report",
				Sections: []Section{
					{ID: "section1", Type: SectionTypeMarkdown},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid section type",
			def: Definition{
				ID:    "test",
				Title: "Test Report",
				Sections: []Section{
					{ID: "section1", Type: "invalid"},
				},
			},
			wantErr: true,
			errMsg:  "invalid section type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.def.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestSectionValidate(t *testing.T) {
	tests := []struct {
		name    string
		section Section
		wantErr bool
	}{
		{
			name:    "empty section",
			section: Section{},
			wantErr: true,
		},
		{
			name:    "missing type",
			section: Section{ID: "test"},
			wantErr: true,
		},
		{
			name:    "valid markdown section",
			section: Section{ID: "test", Type: SectionTypeMarkdown},
			wantErr: false,
		},
		{
			name:    "valid jql section",
			section: Section{ID: "test", Type: SectionTypeJQL},
			wantErr: false,
		},
		{
			name:    "valid velocity section",
			section: Section{ID: "test", Type: SectionTypeVelocity},
			wantErr: false,
		},
		{
			name:    "valid burndown section",
			section: Section{ID: "test", Type: SectionTypeBurndown},
			wantErr: false,
		},
		{
			name:    "valid worklog section",
			section: Section{ID: "test", Type: SectionTypeWorklog},
			wantErr: false,
		},
		{
			name:    "valid cycletime section",
			section: Section{ID: "test", Type: SectionTypeCycleTime},
			wantErr: false,
		},
		{
			name:    "valid metric section",
			section: Section{ID: "test", Type: SectionTypeMetric},
			wantErr: false,
		},
		{
			name:    "valid table section",
			section: Section{ID: "test", Type: SectionTypeTable},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.section.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationErrorUnwrap(t *testing.T) {
	inner := &ValidationError{Field: "inner", Message: "inner error"}
	outer := &ValidationError{Field: "outer", Message: "outer error", Cause: inner}

	if outer.Unwrap() != inner {
		t.Error("Unwrap() did not return the cause")
	}
}
