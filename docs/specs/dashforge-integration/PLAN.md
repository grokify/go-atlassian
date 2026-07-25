# Dashforge Integration - Implementation Plan

## Phase Overview

| Phase | Focus | Deliverables |
|-------|-------|--------------|
| 1 | Core Infrastructure | Report package, definition types, parser |
| 2 | Section Processors | Markdown, JQL, Velocity, Burndown processors |
| 3 | Dashforge IR | Transformer, IR types, data source mapping |
| 4 | Export | HTML, Markdown, PDF export |
| 5 | CLI Integration | Report commands, templates |
| 6 | Interactive Mode | Notebook server, live updates |

## Phase 1: Core Infrastructure

### Tasks

- [ ] **1.1** Create `report/` package structure
- [ ] **1.2** Define `Definition` and related types
- [ ] **1.3** Create YAML/JSON parser with validation
- [ ] **1.4** Create JSON Schema for report definitions
- [ ] **1.5** Add unit tests for parser

### Implementation Details

```
report/
├── definition.go      # Core types (Definition, Section, Variable)
├── parser.go          # YAML/JSON parsing
├── schema.go          # JSON Schema validation
├── schema.json        # Embedded schema
└── definition_test.go # Parser tests
```

### Acceptance Criteria

- [x] Can parse valid YAML report definition
- [x] Validates against JSON Schema
- [x] Returns clear errors for invalid definitions
- [x] Supports all section types

---

## Phase 2: Section Processors

### Tasks

- [ ] **2.1** Create `SectionProcessor` interface
- [ ] **2.2** Implement `MarkdownProcessor`
- [ ] **2.3** Implement `JQLProcessor`
- [ ] **2.4** Implement `VelocityProcessor`
- [ ] **2.5** Implement `BurndownProcessor`
- [ ] **2.6** Implement `WorklogProcessor`
- [ ] **2.7** Implement `CycleTimeProcessor`
- [ ] **2.8** Implement `MetricProcessor`
- [ ] **2.9** Add unit tests for each processor

### Implementation Details

```go
// Section processor interface
type SectionProcessor interface {
    Process(ctx *ExecutionContext, section *Section) (*SectionResult, error)
}

// Each processor in report/sections/
// - markdown.go: Simple pass-through to text widget
// - jql.go: Execute JQL, transform to chart/table
// - velocity.go: Fetch velocity data, create bar chart
// - burndown.go: Fetch sprint data, create line chart
// - worklog.go: Aggregate worklogs, create chart/table
// - cycletime.go: Calculate cycle times, create histogram
// - metric.go: Calculate single metric value
```

### Acceptance Criteria

- [ ] Each processor produces valid WidgetIR
- [ ] Processors handle errors gracefully
- [ ] Data is correctly transformed from Jira format
- [ ] Tests cover edge cases

---

## Phase 3: Dashforge IR Transformation

### Tasks

- [ ] **3.1** Define Dashforge IR types (matching Dashforge schema)
- [ ] **3.2** Create transformer from Section results to Dashboard IR
- [ ] **3.3** Implement data source mapping (inline data)
- [ ] **3.4** Implement chart configuration mapping
- [ ] **3.5** Add layout/positioning logic
- [ ] **3.6** Add theme conversion
- [ ] **3.7** Validate output against Dashforge schema

### Implementation Details

```go
// report/transformer.go
func (e *Engine) Execute(ctx context.Context, def *Definition, vars map[string]any) (*DashboardIR, error) {
    // 1. Resolve variables
    // 2. Process each section
    // 3. Collect widgets and data sources
    // 4. Apply layout
    // 5. Apply theme
    // 6. Return DashboardIR
}
```

### Acceptance Criteria

- [ ] Output matches Dashforge Dashboard schema
- [ ] Data sources correctly embedded as inline
- [ ] Widgets positioned correctly in grid
- [ ] Themes applied to all widgets

---

## Phase 4: Export Formats

### Tasks

- [ ] **4.1** Create HTML exporter (self-contained)
- [ ] **4.2** Embed Dashforge viewer assets
- [ ] **4.3** Create Markdown exporter
- [ ] **4.4** Create PDF exporter (via headless browser)
- [ ] **4.5** Add export tests

### Implementation Details

```
report/export/
├── exporter.go        # Exporter interface
├── html.go            # HTML export (embeds viewer)
├── markdown.go        # Markdown tables/text
├── pdf.go             # PDF via chromedp
└── viewer/            # Embedded Dashforge viewer assets
    ├── template.html
    ├── viewer.js
    └── styles.css
```

### Acceptance Criteria

- [ ] HTML export works offline (self-contained)
- [ ] Markdown export readable as plain text
- [ ] PDF export renders charts correctly
- [ ] All exports include title and metadata

---

## Phase 5: CLI Integration

### Tasks

- [ ] **5.1** Add `gojira report` command group
- [ ] **5.2** Implement `gojira report validate`
- [ ] **5.3** Implement `gojira report render`
- [ ] **5.4** Implement `gojira report export`
- [ ] **5.5** Add built-in templates
- [ ] **5.6** Implement `gojira report list`
- [ ] **5.7** Add help documentation

### Implementation Details

```bash
# Validate a report definition
gojira report validate sprint-report.yaml

# Export to HTML
gojira report export sprint-report.yaml -f html -o report.html

# Export to PDF
gojira report export sprint-report.yaml -f pdf -o report.pdf

# Render interactively (opens browser)
gojira report render sprint-report.yaml

# List built-in templates
gojira report list

# Generate from template
gojira report export --template sprint-report --board-id 123 -o report.html
```

### Acceptance Criteria

- [ ] All commands have help text
- [ ] Errors are clear and actionable
- [ ] Templates cover common use cases
- [ ] Variables can be passed via CLI flags

---

## Phase 6: Interactive Mode (Notebook)

### Tasks

- [ ] **6.1** Create notebook server
- [ ] **6.2** Implement WebSocket for live updates
- [ ] **6.3** Add variable controls UI
- [ ] **6.4** Add save/export from browser
- [ ] **6.5** Implement cell execution
- [ ] **6.6** Add session management

### Implementation Details

```
report/notebook/
├── server.go          # HTTP server
├── handler.go         # API handlers
├── websocket.go       # Live updates
└── static/            # Frontend assets
```

### Acceptance Criteria

- [ ] Can modify variables and see live updates
- [ ] Can export from browser
- [ ] Session persists across refreshes
- [ ] Multiple concurrent users supported

---

## File Creation Order

### Batch 1: Core Types
1. `report/definition.go` - Core types
2. `report/schema.json` - JSON Schema
3. `report/parser.go` - Parser
4. `report/definition_test.go` - Tests

### Batch 2: Engine & Processors
5. `report/engine.go` - Report engine
6. `report/sections/processor.go` - Interface
7. `report/sections/markdown.go`
8. `report/sections/jql.go`
9. `report/sections/velocity.go`
10. `report/sections/burndown.go`
11. `report/sections/worklog.go`
12. `report/sections/cycletime.go`
13. `report/sections/metric.go`

### Batch 3: IR & Transformer
14. `report/ir.go` - Dashforge IR types
15. `report/transformer.go` - Transformer
16. `report/transformer_test.go`

### Batch 4: Export
17. `report/export/exporter.go`
18. `report/export/html.go`
19. `report/export/markdown.go`
20. `report/export/viewer/` - Embedded assets

### Batch 5: CLI
21. `cmd/gojira/report.go`
22. `cmd/gojira/report_validate.go`
23. `cmd/gojira/report_export.go`
24. `cmd/gojira/report_render.go`
25. `report/templates/` - Built-in templates

### Batch 6: Notebook (Future)
26. `report/notebook/server.go`
27. `report/notebook/handler.go`

---

## Dependencies

### Go Dependencies

```go
// go.mod additions
require (
    github.com/plexusone/dashforge v0.1.0  // Dashforge IR types
    gopkg.in/yaml.v3 v3.0.1                // YAML parsing (existing)
    github.com/xeipuuv/gojsonschema v1.2.0 // JSON Schema validation
    github.com/chromedp/chromedp v0.9.3    // PDF export (optional)
)
```

### Dashforge Integration

- Import `dashboardir` package for type compatibility
- Use embedded viewer for HTML export
- Follow Dashforge JSON Schema for validation

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Dashforge API changes | Pin to specific version, abstract IR types |
| Large report performance | Add caching, pagination, progress indicators |
| Complex JQL handling | Leverage existing GoJira JQL support |
| PDF rendering issues | Make PDF optional, provide HTML as default |

---

## Success Metrics

1. **Phase 1-3**: Can generate valid Dashforge IR from YAML
2. **Phase 4**: Can export to HTML and view in browser
3. **Phase 5**: CLI commands work end-to-end
4. **Phase 6**: Interactive mode provides value beyond static

---

## Current Status

**Starting Phase 1: Core Infrastructure**

Next steps:
1. Create `report/` package directory
2. Implement `Definition` types
3. Create parser with YAML support
4. Add JSON Schema validation
