// Package report provides a report generation system for Jira data.
// It defines a report schema that can be transformed into Dashforge Dashboard IR
// for rendering as interactive dashboards or static exports.
package report

import (
	"encoding/json"
	"time"
)

// Definition represents a complete Jira report definition.
type Definition struct {
	ID          string `json:"id" yaml:"id"`
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string `json:"version,omitempty" yaml:"version,omitempty"`

	Variables []Variable `json:"variables,omitempty" yaml:"variables,omitempty"`
	Sections  []Section  `json:"sections" yaml:"sections"`
	Theme     *Theme     `json:"theme,omitempty" yaml:"theme,omitempty"`

	// Metadata
	Author    string    `json:"author,omitempty" yaml:"author,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

// Variable represents a configurable parameter in a report.
type Variable struct {
	ID          string         `json:"id" yaml:"id"`
	Type        VariableType   `json:"type" yaml:"type"`
	Label       string         `json:"label" yaml:"label"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	Default     any            `json:"default,omitempty" yaml:"default,omitempty"`
	Required    bool           `json:"required,omitempty" yaml:"required,omitempty"`
	Options     []SelectOption `json:"options,omitempty" yaml:"options,omitempty"`
}

// VariableType defines the type of a variable.
type VariableType string

// Variable types.
const (
	VariableTypeString VariableType = "string"
	VariableTypeNumber VariableType = "number"
	VariableTypeDate   VariableType = "date"
	VariableTypeSelect VariableType = "select"
	VariableTypeBool   VariableType = "bool"
)

// SelectOption represents an option in a select variable.
type SelectOption struct {
	Value string `json:"value" yaml:"value"`
	Label string `json:"label" yaml:"label"`
}

// Section represents a single section in the report.
type Section struct {
	ID       string          `json:"id" yaml:"id"`
	Type     SectionType     `json:"type" yaml:"type"`
	Title    string          `json:"title,omitempty" yaml:"title,omitempty"`
	Position *Position       `json:"position,omitempty" yaml:"position,omitempty"`
	Config   json.RawMessage `json:"config,omitempty" yaml:"-"`

	// Convenience fields (parsed from Config or set directly)
	Content string `json:"content,omitempty" yaml:"content,omitempty"` // For markdown
}

// sectionYAML is an intermediate type for YAML unmarshaling.
type sectionYAML struct {
	ID       string         `yaml:"id"`
	Type     SectionType    `yaml:"type"`
	Title    string         `yaml:"title,omitempty"`
	Position *Position      `yaml:"position,omitempty"`
	Config   map[string]any `yaml:"config,omitempty"`
	Content  string         `yaml:"content,omitempty"`
}

// UnmarshalYAML implements custom YAML unmarshaling for Section.
func (s *Section) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sy sectionYAML
	if err := unmarshal(&sy); err != nil {
		return err
	}

	s.ID = sy.ID
	s.Type = sy.Type
	s.Title = sy.Title
	s.Position = sy.Position
	s.Content = sy.Content

	// Convert map to JSON for Config
	if sy.Config != nil {
		configJSON, err := json.Marshal(sy.Config)
		if err != nil {
			return err
		}
		s.Config = configJSON
	}

	return nil
}

// SectionType defines the type of a report section.
type SectionType string

// Section types.
const (
	SectionTypeMarkdown  SectionType = "markdown"
	SectionTypeJQL       SectionType = "jql"
	SectionTypeVelocity  SectionType = "velocity"
	SectionTypeBurndown  SectionType = "burndown"
	SectionTypeWorklog   SectionType = "worklog"
	SectionTypeCycleTime SectionType = "cycletime"
	SectionTypeMetric    SectionType = "metric"
	SectionTypeTable     SectionType = "table"
)

// Position defines the grid position of a section.
type Position struct {
	X int `json:"x" yaml:"x"`
	Y int `json:"y" yaml:"y"`
	W int `json:"w" yaml:"w"`
	H int `json:"h" yaml:"h"`
}

// Theme defines visual styling for the report.
type Theme struct {
	Colors   *ThemeColors `json:"colors,omitempty" yaml:"colors,omitempty"`
	FontSize string       `json:"font_size,omitempty" yaml:"font_size,omitempty"`
	DarkMode bool         `json:"dark_mode,omitempty" yaml:"dark_mode,omitempty"`
}

// ThemeColors defines color palette for the theme.
type ThemeColors struct {
	Primary   string `json:"primary,omitempty" yaml:"primary,omitempty"`
	Secondary string `json:"secondary,omitempty" yaml:"secondary,omitempty"`
	Success   string `json:"success,omitempty" yaml:"success,omitempty"`
	Warning   string `json:"warning,omitempty" yaml:"warning,omitempty"`
	Danger    string `json:"danger,omitempty" yaml:"danger,omitempty"`
}

// JQLConfig configures a JQL-based section.
type JQLConfig struct {
	JQL        string      `json:"jql" yaml:"jql"`
	Fields     []string    `json:"fields,omitempty" yaml:"fields,omitempty"`
	MaxResults int         `json:"max_results,omitempty" yaml:"max_results,omitempty"`
	Display    DisplayType `json:"display" yaml:"display"`
	ChartType  string      `json:"chart_type,omitempty" yaml:"chart_type,omitempty"`
	GroupBy    string      `json:"group_by,omitempty" yaml:"group_by,omitempty"`
}

// DisplayType defines how data should be displayed.
type DisplayType string

// Display types.
const (
	DisplayTable DisplayType = "table"
	DisplayChart DisplayType = "chart"
	DisplayList  DisplayType = "list"
)

// VelocityConfig configures a velocity chart section.
type VelocityConfig struct {
	BoardID          int    `json:"board_id" yaml:"board_id"`
	SprintCount      int    `json:"sprint_count,omitempty" yaml:"sprint_count,omitempty"`
	IncludeActive    bool   `json:"include_active,omitempty" yaml:"include_active,omitempty"`
	StoryPointsField string `json:"story_points_field,omitempty" yaml:"story_points_field,omitempty"`
	ChartType        string `json:"chart_type,omitempty" yaml:"chart_type,omitempty"`
}

// BurndownConfig configures a burndown chart section.
type BurndownConfig struct {
	SprintID         int    `json:"sprint_id" yaml:"sprint_id"`
	SprintName       string `json:"sprint_name,omitempty" yaml:"sprint_name,omitempty"`
	StoryPointsField string `json:"story_points_field,omitempty" yaml:"story_points_field,omitempty"`
	ShowIdealLine    bool   `json:"show_ideal_line,omitempty" yaml:"show_ideal_line,omitempty"`
}

// WorklogConfig configures a worklog summary section.
type WorklogConfig struct {
	JQL       string      `json:"jql,omitempty" yaml:"jql,omitempty"`
	SprintID  int         `json:"sprint_id,omitempty" yaml:"sprint_id,omitempty"`
	GroupBy   string      `json:"group_by,omitempty" yaml:"group_by,omitempty"`
	Display   DisplayType `json:"display" yaml:"display"`
	ChartType string      `json:"chart_type,omitempty" yaml:"chart_type,omitempty"`
}

// CycleTimeConfig configures a cycle time analysis section.
type CycleTimeConfig struct {
	JQL       string      `json:"jql,omitempty" yaml:"jql,omitempty"`
	SprintID  int         `json:"sprint_id,omitempty" yaml:"sprint_id,omitempty"`
	Display   DisplayType `json:"display" yaml:"display"`
	ChartType string      `json:"chart_type,omitempty" yaml:"chart_type,omitempty"`
}

// MetricConfig configures a single metric display.
type MetricConfig struct {
	JQL         string            `json:"jql,omitempty" yaml:"jql,omitempty"`
	Aggregation string            `json:"aggregation" yaml:"aggregation"`
	Field       string            `json:"field,omitempty" yaml:"field,omitempty"`
	Format      string            `json:"format,omitempty" yaml:"format,omitempty"`
	Comparison  *MetricComparison `json:"comparison,omitempty" yaml:"comparison,omitempty"`
}

// MetricComparison defines comparison for a metric.
type MetricComparison struct {
	Type        string  `json:"type" yaml:"type"`
	TargetValue float64 `json:"target_value,omitempty" yaml:"target_value,omitempty"`
}

// TableConfig configures a table section.
type TableConfig struct {
	JQL        string   `json:"jql,omitempty" yaml:"jql,omitempty"`
	Fields     []string `json:"fields" yaml:"fields"`
	MaxResults int      `json:"max_results,omitempty" yaml:"max_results,omitempty"`
	Sortable   bool     `json:"sortable,omitempty" yaml:"sortable,omitempty"`
	Paginated  bool     `json:"paginated,omitempty" yaml:"paginated,omitempty"`
}

// Validate validates the report definition.
func (d *Definition) Validate() error {
	if d.ID == "" {
		return &ValidationError{Field: "id", Message: "id is required"}
	}
	if d.Title == "" {
		return &ValidationError{Field: "title", Message: "title is required"}
	}
	if len(d.Sections) == 0 {
		return &ValidationError{Field: "sections", Message: "at least one section is required"}
	}

	for i, section := range d.Sections {
		if err := section.Validate(); err != nil {
			return &ValidationError{
				Field:   "sections",
				Message: "invalid section",
				Index:   i,
				Cause:   err,
			}
		}
	}

	return nil
}

// Validate validates a section.
func (s *Section) Validate() error {
	if s.ID == "" {
		return &ValidationError{Field: "id", Message: "id is required"}
	}
	if s.Type == "" {
		return &ValidationError{Field: "type", Message: "type is required"}
	}

	validTypes := map[SectionType]bool{
		SectionTypeMarkdown:  true,
		SectionTypeJQL:       true,
		SectionTypeVelocity:  true,
		SectionTypeBurndown:  true,
		SectionTypeWorklog:   true,
		SectionTypeCycleTime: true,
		SectionTypeMetric:    true,
		SectionTypeTable:     true,
	}

	if !validTypes[s.Type] {
		return &ValidationError{Field: "type", Message: "invalid section type: " + string(s.Type)}
	}

	return nil
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
	Index   int
	Cause   error
}

func (e *ValidationError) Error() string {
	if e.Cause != nil {
		return e.Field + ": " + e.Message + ": " + e.Cause.Error()
	}
	return e.Field + ": " + e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}
