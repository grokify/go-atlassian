# Dashforge Integration - Technical Requirements Document

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Go-Atlassian Report System                         │
│                                                                      │
│  ┌──────────────────┐    ┌──────────────────┐    ┌───────────────┐  │
│  │  Report Schema   │    │  Report Engine   │    │  CLI Commands │  │
│  │  (YAML/JSON)     │───▶│  (Go Package)    │◀───│  (Cobra)      │  │
│  └──────────────────┘    └────────┬─────────┘    └───────────────┘  │
│                                   │                                  │
│                                   ▼                                  │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Go-Atlassian SDK (jira package)                  │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │   │
│  │  │ Issue   │ │ Board   │ │ Sprint  │ │ Worklog │ │ Custom  │ │   │
│  │  │ Service │ │ Service │ │ Service │ │ Reports │ │ Fields  │ │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                   │                                  │
│                                   ▼                                  │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                 IR Transformer (report package)               │   │
│  │  Jira Report Definition ──▶ Dashforge Dashboard IR            │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                   │                                  │
└───────────────────────────────────┼──────────────────────────────────┘
                                    │
                                    ▼
┌───────────────────────────────────────────────────────────────────────┐
│                        Dashforge (External)                           │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐   │
│  │  Dashboard IR   │───▶│  Viewer/Server  │───▶│  Export Formats │   │
│  │  (JSON Schema)  │    │  (HTML/JS)      │    │  (HTML/PDF/MD)  │   │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘   │
└───────────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
go-atlassian/
├── report/                      # New package for report system
│   ├── definition.go            # Report definition types
│   ├── schema.go                # JSON Schema for validation
│   ├── parser.go                # YAML/JSON parsing
│   ├── engine.go                # Report execution engine
│   ├── transformer.go           # Transform to Dashforge IR
│   ├── sections/                # Section type implementations
│   │   ├── markdown.go
│   │   ├── jql.go
│   │   ├── velocity.go
│   │   ├── burndown.go
│   │   ├── worklog.go
│   │   └── cycletime.go
│   ├── export/                  # Export implementations
│   │   ├── html.go
│   │   ├── pdf.go
│   │   └── markdown.go
│   └── templates/               # Built-in report templates
│       ├── sprint-report.yaml
│       ├── velocity-report.yaml
│       └── team-worklog.yaml
├── cmd/gojira/
│   ├── report.go                # report command group
│   ├── report_render.go         # render subcommand
│   ├── report_export.go         # export subcommand
│   └── notebook.go              # notebook command
└── docs/specs/dashforge-integration/
    ├── PRD.md
    ├── TRD.md                   # This file
    ├── PLAN.md
    └── ROADMAP.md
```

## Core Types

### Report Definition Schema

```go
// report/definition.go

package report

import (
    "encoding/json"
    "time"
)

// Definition represents a complete Jira report definition.
type Definition struct {
    ID          string     `json:"id" yaml:"id"`
    Title       string     `json:"title" yaml:"title"`
    Description string     `json:"description,omitempty" yaml:"description,omitempty"`
    Version     string     `json:"version,omitempty" yaml:"version,omitempty"`

    Variables   []Variable `json:"variables,omitempty" yaml:"variables,omitempty"`
    Sections    []Section  `json:"sections" yaml:"sections"`
    Theme       *Theme     `json:"theme,omitempty" yaml:"theme,omitempty"`

    // Metadata
    Author      string     `json:"author,omitempty" yaml:"author,omitempty"`
    CreatedAt   time.Time  `json:"created_at,omitempty" yaml:"created_at,omitempty"`
    UpdatedAt   time.Time  `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

// Variable represents a configurable parameter in a report.
type Variable struct {
    ID          string         `json:"id" yaml:"id"`
    Type        VariableType   `json:"type" yaml:"type"` // "string", "number", "date", "select"
    Label       string         `json:"label" yaml:"label"`
    Description string         `json:"description,omitempty" yaml:"description,omitempty"`
    Default     any            `json:"default,omitempty" yaml:"default,omitempty"`
    Required    bool           `json:"required,omitempty" yaml:"required,omitempty"`
    Options     []SelectOption `json:"options,omitempty" yaml:"options,omitempty"` // For select type
}

type VariableType string

const (
    VariableTypeString VariableType = "string"
    VariableTypeNumber VariableType = "number"
    VariableTypeDate   VariableType = "date"
    VariableTypeSelect VariableType = "select"
)

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
    Config   json.RawMessage `json:"config,omitempty" yaml:"config,omitempty"`

    // Type-specific fields (convenience, parsed from Config)
    Content  string `json:"content,omitempty" yaml:"content,omitempty"` // For markdown
}

type SectionType string

const (
    SectionTypeMarkdown  SectionType = "markdown"
    SectionTypeJQL       SectionType = "jql"
    SectionTypeVelocity  SectionType = "velocity"
    SectionTypeBurndown  SectionType = "burndown"
    SectionTypeWorklog   SectionType = "worklog"
    SectionTypeCycleTime SectionType = "cycletime"
    SectionTypeMetric    SectionType = "metric"
)

type Position struct {
    X int `json:"x" yaml:"x"`
    Y int `json:"y" yaml:"y"`
    W int `json:"w" yaml:"w"`
    H int `json:"h" yaml:"h"`
}

// Theme defines visual styling for the report.
type Theme struct {
    Colors    *ThemeColors `json:"colors,omitempty" yaml:"colors,omitempty"`
    FontSize  string       `json:"font_size,omitempty" yaml:"font_size,omitempty"`
    DarkMode  bool         `json:"dark_mode,omitempty" yaml:"dark_mode,omitempty"`
}

type ThemeColors struct {
    Primary   string `json:"primary,omitempty" yaml:"primary,omitempty"`
    Secondary string `json:"secondary,omitempty" yaml:"secondary,omitempty"`
    Success   string `json:"success,omitempty" yaml:"success,omitempty"`
    Warning   string `json:"warning,omitempty" yaml:"warning,omitempty"`
    Danger    string `json:"danger,omitempty" yaml:"danger,omitempty"`
}
```

### Section Configurations

```go
// report/sections/configs.go

package sections

// JQLConfig configures a JQL-based section.
type JQLConfig struct {
    JQL       string   `json:"jql" yaml:"jql"`
    Fields    []string `json:"fields,omitempty" yaml:"fields,omitempty"`
    MaxResults int     `json:"max_results,omitempty" yaml:"max_results,omitempty"`

    // Visualization
    Display   DisplayType `json:"display" yaml:"display"` // "table", "chart", "list"
    ChartType string      `json:"chart_type,omitempty" yaml:"chart_type,omitempty"` // "bar", "pie", "line"
    GroupBy   string      `json:"group_by,omitempty" yaml:"group_by,omitempty"`

    // Drill-down
    DrillDown *DrillDownConfig `json:"drill_down,omitempty" yaml:"drill_down,omitempty"`
}

type DisplayType string

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
    ChartType        string `json:"chart_type,omitempty" yaml:"chart_type,omitempty"` // "bar", "line"
}

// BurndownConfig configures a burndown chart section.
type BurndownConfig struct {
    SprintID         int    `json:"sprint_id" yaml:"sprint_id"` // Use 0 or "active" for current sprint
    SprintName       string `json:"sprint_name,omitempty" yaml:"sprint_name,omitempty"` // Alternative to ID
    StoryPointsField string `json:"story_points_field,omitempty" yaml:"story_points_field,omitempty"`
    ShowIdealLine    bool   `json:"show_ideal_line,omitempty" yaml:"show_ideal_line,omitempty"`
}

// WorklogConfig configures a worklog summary section.
type WorklogConfig struct {
    JQL        string `json:"jql,omitempty" yaml:"jql,omitempty"`
    SprintID   int    `json:"sprint_id,omitempty" yaml:"sprint_id,omitempty"`
    GroupBy    string `json:"group_by,omitempty" yaml:"group_by,omitempty"` // "author", "issue", "date"
    Display    DisplayType `json:"display" yaml:"display"`
    ChartType  string `json:"chart_type,omitempty" yaml:"chart_type,omitempty"`
}

// CycleTimeConfig configures a cycle time analysis section.
type CycleTimeConfig struct {
    JQL       string `json:"jql,omitempty" yaml:"jql,omitempty"`
    SprintID  int    `json:"sprint_id,omitempty" yaml:"sprint_id,omitempty"`
    Display   DisplayType `json:"display" yaml:"display"`
    ChartType string `json:"chart_type,omitempty" yaml:"chart_type,omitempty"` // "histogram", "scatter"
}

// MetricConfig configures a single metric display.
type MetricConfig struct {
    JQL          string `json:"jql,omitempty" yaml:"jql,omitempty"`
    Aggregation  string `json:"aggregation" yaml:"aggregation"` // "count", "sum", "avg"
    Field        string `json:"field,omitempty" yaml:"field,omitempty"` // For sum/avg
    Format       string `json:"format,omitempty" yaml:"format,omitempty"` // "number", "percent", "duration"
    Comparison   *MetricComparison `json:"comparison,omitempty" yaml:"comparison,omitempty"`
}

type MetricComparison struct {
    Type      string `json:"type" yaml:"type"` // "previous_sprint", "previous_period", "target"
    TargetValue float64 `json:"target_value,omitempty" yaml:"target_value,omitempty"`
}

// DrillDownConfig defines drill-down behavior.
type DrillDownConfig struct {
    Enabled    bool   `json:"enabled" yaml:"enabled"`
    TargetURL  string `json:"target_url,omitempty" yaml:"target_url,omitempty"`
    TargetJQL  string `json:"target_jql,omitempty" yaml:"target_jql,omitempty"`
}
```

### Report Engine

```go
// report/engine.go

package report

import (
    "context"
    "fmt"

    "github.com/grokify/go-atlassian/jira"
)

// Engine executes report definitions and produces Dashforge IR.
type Engine struct {
    client     *jira.Client
    cache      *Cache
    processors map[SectionType]SectionProcessor
}

// NewEngine creates a new report engine.
func NewEngine(client *jira.Client) *Engine {
    e := &Engine{
        client:     client,
        cache:      NewCache(),
        processors: make(map[SectionType]SectionProcessor),
    }

    // Register built-in processors
    e.RegisterProcessor(SectionTypeMarkdown, &MarkdownProcessor{})
    e.RegisterProcessor(SectionTypeJQL, &JQLProcessor{client: client})
    e.RegisterProcessor(SectionTypeVelocity, &VelocityProcessor{client: client})
    e.RegisterProcessor(SectionTypeBurndown, &BurndownProcessor{client: client})
    e.RegisterProcessor(SectionTypeWorklog, &WorklogProcessor{client: client})
    e.RegisterProcessor(SectionTypeCycleTime, &CycleTimeProcessor{client: client})
    e.RegisterProcessor(SectionTypeMetric, &MetricProcessor{client: client})

    return e
}

// RegisterProcessor registers a custom section processor.
func (e *Engine) RegisterProcessor(sectionType SectionType, processor SectionProcessor) {
    e.processors[sectionType] = processor
}

// Execute runs the report and returns a Dashforge Dashboard IR.
func (e *Engine) Execute(ctx context.Context, def *Definition, vars map[string]any) (*DashboardIR, error) {
    // Resolve variables
    resolvedVars := e.resolveVariables(def.Variables, vars)

    // Create execution context
    execCtx := &ExecutionContext{
        Context:   ctx,
        Variables: resolvedVars,
        Cache:     e.cache,
    }

    // Process sections
    widgets := make([]WidgetIR, 0, len(def.Sections))
    dataSources := make([]DataSourceIR, 0)

    for i, section := range def.Sections {
        processor, ok := e.processors[section.Type]
        if !ok {
            return nil, fmt.Errorf("unknown section type: %s", section.Type)
        }

        result, err := processor.Process(execCtx, &section)
        if err != nil {
            return nil, fmt.Errorf("process section %s: %w", section.ID, err)
        }

        // Add data source
        if result.DataSource != nil {
            dataSources = append(dataSources, *result.DataSource)
        }

        // Add widget with auto-positioning if not specified
        widget := result.Widget
        if widget.Position == nil {
            widget.Position = &PositionIR{
                X: 0,
                Y: i * 4, // Stack vertically
                W: 12,
                H: 4,
            }
        }
        widgets = append(widgets, widget)
    }

    // Build dashboard IR
    dashboard := &DashboardIR{
        ID:          def.ID,
        Title:       def.Title,
        Description: def.Description,
        Version:     def.Version,
        Layout:      LayoutIR{Columns: 12},
        DataSources: dataSources,
        Widgets:     widgets,
        Theme:       e.convertTheme(def.Theme),
    }

    return dashboard, nil
}

// SectionProcessor processes a section and produces widget IR.
type SectionProcessor interface {
    Process(ctx *ExecutionContext, section *Section) (*SectionResult, error)
}

// ExecutionContext provides context for section processing.
type ExecutionContext struct {
    Context   context.Context
    Variables map[string]any
    Cache     *Cache
}

// SectionResult contains the output of processing a section.
type SectionResult struct {
    Widget     WidgetIR
    DataSource *DataSourceIR
}
```

### Dashforge IR Types

```go
// report/transformer.go

package report

import "encoding/json"

// DashboardIR represents a Dashforge Dashboard.
// This matches the Dashforge schema for compatibility.
type DashboardIR struct {
    ID          string         `json:"id"`
    Title       string         `json:"title"`
    Description string         `json:"description,omitempty"`
    Version     string         `json:"version,omitempty"`
    Layout      LayoutIR       `json:"layout"`
    DataSources []DataSourceIR `json:"dataSources"`
    Widgets     []WidgetIR     `json:"widgets"`
    Variables   []VariableIR   `json:"variables,omitempty"`
    Theme       *ThemeIR       `json:"theme,omitempty"`
}

type LayoutIR struct {
    Columns    int `json:"columns"`
    RowHeight  int `json:"rowHeight,omitempty"`
    Responsive bool `json:"responsive,omitempty"`
}

type DataSourceIR struct {
    ID        string          `json:"id"`
    Type      string          `json:"type"` // "inline", "url", "derived"
    Data      json.RawMessage `json:"data,omitempty"`
    URL       string          `json:"url,omitempty"`
    Transform []TransformIR   `json:"transform,omitempty"`
}

type TransformIR struct {
    Type   string          `json:"type"` // "filter", "aggregate", "sort", etc.
    Config json.RawMessage `json:"config"`
}

type WidgetIR struct {
    ID           string          `json:"id"`
    Type         string          `json:"type"` // "chart", "table", "metric", "text"
    Title        string          `json:"title,omitempty"`
    Position     *PositionIR     `json:"position"`
    DataSourceID string          `json:"dataSourceId,omitempty"`
    Config       json.RawMessage `json:"config"`
}

type PositionIR struct {
    X int `json:"x"`
    Y int `json:"y"`
    W int `json:"w"`
    H int `json:"h"`
}

type VariableIR struct {
    ID      string         `json:"id"`
    Type    string         `json:"type"`
    Label   string         `json:"label"`
    Default any            `json:"default,omitempty"`
    Options []SelectOption `json:"options,omitempty"`
}

type ThemeIR struct {
    Mode   string          `json:"mode,omitempty"` // "light", "dark"
    Colors *ThemeColorsIR  `json:"colors,omitempty"`
}

type ThemeColorsIR struct {
    Primary   string `json:"primary,omitempty"`
    Secondary string `json:"secondary,omitempty"`
    Success   string `json:"success,omitempty"`
    Warning   string `json:"warning,omitempty"`
    Danger    string `json:"danger,omitempty"`
}
```

## Integration Points

### Dashforge Dependency

```go
// go.mod addition
require (
    github.com/plexusone/dashforge v0.1.0
)
```

### Using Dashforge Viewer

```go
// report/export/html.go

package export

import (
    "embed"
    "html/template"
    "io"

    "github.com/grokify/go-atlassian/report"
)

//go:embed viewer/*
var viewerFS embed.FS

// HTMLExporter exports reports to self-contained HTML.
type HTMLExporter struct {
    template *template.Template
}

func NewHTMLExporter() (*HTMLExporter, error) {
    tmpl, err := template.ParseFS(viewerFS, "viewer/template.html")
    if err != nil {
        return nil, err
    }
    return &HTMLExporter{template: tmpl}, nil
}

func (e *HTMLExporter) Export(w io.Writer, dashboard *report.DashboardIR) error {
    data := struct {
        Dashboard *report.DashboardIR
        Embedded  bool
    }{
        Dashboard: dashboard,
        Embedded:  true,
    }
    return e.template.Execute(w, data)
}
```

## CLI Integration

```go
// cmd/gojira/report.go

package main

import (
    "github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
    Use:   "report",
    Short: "Generate and export Jira reports",
    Long:  `Generate reports from Jira data using YAML/JSON definitions.`,
}

var reportRenderCmd = &cobra.Command{
    Use:   "render <file>",
    Short: "Render a report interactively",
    Args:  cobra.ExactArgs(1),
    RunE:  runReportRender,
}

var reportExportCmd = &cobra.Command{
    Use:   "export <file>",
    Short: "Export a report to a file",
    Args:  cobra.ExactArgs(1),
    RunE:  runReportExport,
}

var reportValidateCmd = &cobra.Command{
    Use:   "validate <file>",
    Short: "Validate a report definition",
    Args:  cobra.ExactArgs(1),
    RunE:  runReportValidate,
}

func init() {
    reportCmd.AddCommand(reportRenderCmd)
    reportCmd.AddCommand(reportExportCmd)
    reportCmd.AddCommand(reportValidateCmd)

    reportExportCmd.Flags().StringP("format", "f", "html", "Export format (html, pdf, markdown)")
    reportExportCmd.Flags().StringP("output", "o", "", "Output file path")

    reportRenderCmd.Flags().IntP("port", "p", 8080, "Server port for interactive mode")

    rootCmd.AddCommand(reportCmd)
}
```

## Data Flow

```
1. User creates report.yaml
   │
   ▼
2. CLI parses YAML → Definition struct
   │
   ▼
3. Engine resolves variables
   │
   ▼
4. For each section:
   │  a. Processor fetches Jira data (via SDK)
   │  b. Transforms data to Dashforge format
   │  c. Produces WidgetIR + DataSourceIR
   │
   ▼
5. Engine assembles DashboardIR
   │
   ▼
6. Export:
   ├── HTML: Embed in Dashforge viewer template
   ├── PDF: Render HTML via headless browser
   └── Markdown: Convert widgets to markdown tables/text
```

## Testing Strategy

### Unit Tests

```go
// report/engine_test.go

func TestEngineExecute(t *testing.T) {
    // Mock Jira client
    client := &jira.Client{...}
    engine := NewEngine(client)

    def := &Definition{
        ID:    "test-report",
        Title: "Test Report",
        Sections: []Section{
            {ID: "s1", Type: SectionTypeMarkdown, Content: "# Hello"},
        },
    }

    dashboard, err := engine.Execute(context.Background(), def, nil)
    require.NoError(t, err)
    assert.Equal(t, "test-report", dashboard.ID)
    assert.Len(t, dashboard.Widgets, 1)
}
```

### Integration Tests

```go
// report/integration_test.go

func TestFullReportGeneration(t *testing.T) {
    if os.Getenv("JIRA_URL") == "" {
        t.Skip("JIRA_URL not set")
    }

    client := setupTestClient(t)
    engine := NewEngine(client)

    def, err := ParseFile("testdata/sprint-report.yaml")
    require.NoError(t, err)

    dashboard, err := engine.Execute(context.Background(), def, nil)
    require.NoError(t, err)

    // Export to HTML
    var buf bytes.Buffer
    exporter := export.NewHTMLExporter()
    err = exporter.Export(&buf, dashboard)
    require.NoError(t, err)

    assert.Contains(t, buf.String(), "<html>")
}
```

## Performance Considerations

1. **Caching**: Cache Jira API responses during report generation
2. **Parallelism**: Process independent sections concurrently
3. **Pagination**: Handle large JQL result sets with pagination
4. **Incremental**: Support incremental data fetching for interactive mode

## Security Considerations

1. **JQL Injection**: Validate and sanitize user-provided JQL
2. **Variable Validation**: Validate variable values against schema
3. **Authentication**: Use existing Go-Atlassian authentication mechanisms
4. **Export Security**: Sanitize HTML output to prevent XSS
