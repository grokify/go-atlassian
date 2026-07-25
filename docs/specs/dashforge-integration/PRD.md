# Dashforge Integration - Product Requirements Document

## Executive Summary

Integrate Go-Atlassian with Dashforge to provide a powerful, reusable report generation system that supports both interactive notebook-style exploration and static report export. This integration enables users to create Jira-based reports and dashboards using a JSON-first intermediate representation (IR) that can be rendered interactively or exported to static formats.

## Problem Statement

Currently, Go-Atlassian provides excellent CLI and SDK capabilities for querying Jira data and generating reports (velocity, burndown, worklog, cycle-time). However:

1. **No unified report format** - Each report command outputs independently with no way to combine them
2. **Limited visualization** - Output is text-based (JSON, table, CSV, markdown)
3. **No interactivity** - Reports are static snapshots with no drill-down or filtering
4. **No reusability** - Report definitions cannot be saved, shared, or version-controlled

## Solution Overview

Create a hybrid integration where:

- **Go-Atlassian** handles Jira data fetching, transformation, and business logic
- **Dashforge** provides the visualization IR, rendering, and export capabilities

This produces a **Jira Report IR** that transforms into **Dashforge Dashboard IR** for rendering.

## User Personas

### 1. Engineering Manager
- Wants weekly sprint reports with velocity trends
- Needs to share static PDF reports with stakeholders
- Values consistency and automation

### 2. Scrum Master
- Wants interactive dashboards for sprint planning
- Needs real-time burndown visibility
- Values drill-down into individual issues

### 3. Developer
- Wants to query Jira data programmatically
- Needs to build custom reports for specific needs
- Values code-first, version-controlled definitions

## Use Cases

### UC-1: Generate Sprint Report
```yaml
# sprint-report.yaml
title: "Sprint 42 Report"
sections:
  - type: markdown
    content: |
      # Sprint 42 Summary
      Sprint dates: 2024-01-15 to 2024-01-29
  - type: velocity
    board_id: 123
    sprint_count: 5
  - type: burndown
    sprint_id: 456
  - type: jql
    title: "Completed Issues"
    jql: "sprint = 456 AND status = Done"
    chart_type: pie
    group_by: issuetype
```

**Output modes:**
- `gojira report render sprint-report.yaml` → Interactive HTML
- `gojira report export sprint-report.yaml --format=pdf` → Static PDF
- `gojira report export sprint-report.yaml --format=html` → Self-contained HTML

### UC-2: Interactive Notebook Mode
```bash
# Start notebook server
gojira notebook serve --port 8080

# Opens browser with:
# - Live JQL execution
# - Real-time chart updates
# - Variable/parameter controls
# - Save/export capabilities
```

### UC-3: Programmatic Report Generation
```go
report := &report.Definition{
    Title: "Q4 Analysis",
    Sections: []report.Section{
        {Type: "velocity", Config: report.VelocityConfig{BoardID: 123}},
        {Type: "markdown", Content: "## Analysis\nKey findings..."},
    },
}

// Render to Dashforge IR
dashboard := report.ToDashforgeIR(ctx, client)

// Export
html := dashforge.RenderHTML(dashboard)
pdf := dashforge.RenderPDF(dashboard)
```

## Functional Requirements

### FR-1: Jira Report Definition Schema
- YAML/JSON schema for defining Jira-specific reports
- Support for: JQL queries, velocity, burndown, worklog, cycle-time sections
- Variables/parameters for dynamic filtering
- Markdown sections for commentary

### FR-2: Data Fetching Layer
- Execute JQL queries via Go-Atlassian SDK
- Fetch sprint/board data for reports
- Calculate derived metrics (velocity, cycle time, etc.)
- Cache results for performance

### FR-3: Dashforge IR Transformation
- Convert Jira report sections to Dashforge widgets
- Map Jira data to inline data sources
- Configure chart types and styling
- Support theme customization

### FR-4: Rendering Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| Interactive | Server-based with live data | Exploration, planning |
| Static HTML | Self-contained HTML file | Email, documentation |
| Static PDF | Printable document | Stakeholder reports |
| Markdown | Text-based output | Git-friendly docs |

### FR-5: CLI Commands
- `gojira report validate <file>` - Validate report definition
- `gojira report render <file>` - Render interactive report
- `gojira report export <file> --format=<fmt>` - Export to format
- `gojira report list` - List available report templates
- `gojira notebook serve` - Start notebook server

## Non-Functional Requirements

### NFR-1: Performance
- Report generation < 5 seconds for typical reports
- Incremental data fetching for large datasets
- Caching of Jira API responses

### NFR-2: Compatibility
- Dashforge IR v1.0 compatibility
- Go-Atlassian SDK v0.37+ compatibility
- Works with Jira Cloud and Server

### NFR-3: Extensibility
- Plugin architecture for custom section types
- Custom chart configurations
- Theme customization

## Success Metrics

1. **Adoption**: 50% of Go-Atlassian users try report generation within 3 months
2. **Retention**: Users generate 2+ reports per week on average
3. **Satisfaction**: NPS score > 40 for report features

## Out of Scope (v1)

- Real-time collaborative editing
- Report scheduling/automation
- Custom visualization plugins
- Data source federation (non-Jira)

## Dependencies

- Dashforge v0.1.0+ (Dashboard IR, viewer)
- Go-Atlassian v0.37.0+ (SDK, CLI)
- ECharts (via Dashforge viewer)

## Timeline

See [ROADMAP.md](ROADMAP.md) for detailed timeline.

## Appendix

### A. Example Report Definition

```yaml
id: weekly-sprint-report
title: "Weekly Sprint Report"
description: "Automated weekly sprint status report"
version: "1.0.0"

variables:
  - id: board_id
    type: number
    label: "Board ID"
    default: 123
  - id: sprint_count
    type: number
    label: "Sprints to show"
    default: 4

sections:
  - id: header
    type: markdown
    content: |
      # Weekly Sprint Report
      Generated: {{ .GeneratedAt }}

  - id: velocity-chart
    type: velocity
    title: "Team Velocity"
    board_id: "{{ .Variables.board_id }}"
    sprint_count: "{{ .Variables.sprint_count }}"
    chart_type: bar

  - id: current-sprint
    type: burndown
    title: "Current Sprint Burndown"
    sprint_id: active

  - id: issue-breakdown
    type: jql
    title: "Issue Type Distribution"
    jql: "sprint in openSprints() AND project = PROJ"
    chart_type: pie
    group_by: issuetype

  - id: blockers
    type: jql
    title: "Current Blockers"
    jql: "sprint in openSprints() AND labels = blocker"
    display: table
    fields:
      - key
      - summary
      - assignee
      - status

theme:
  colors:
    primary: "#0052CC"
    success: "#36B37E"
    warning: "#FFAB00"
    danger: "#FF5630"
```

### B. Dashforge IR Output (Simplified)

```json
{
  "id": "weekly-sprint-report",
  "title": "Weekly Sprint Report",
  "layout": {"columns": 12},
  "dataSources": [
    {
      "id": "velocity-data",
      "type": "inline",
      "data": [
        {"sprint": "Sprint 40", "completed": 34, "committed": 40},
        {"sprint": "Sprint 41", "completed": 42, "committed": 45}
      ]
    }
  ],
  "widgets": [
    {
      "id": "velocity-chart",
      "type": "chart",
      "position": {"x": 0, "y": 0, "w": 6, "h": 4},
      "dataSourceId": "velocity-data",
      "config": {
        "type": "bar",
        "xField": "sprint",
        "yFields": ["completed", "committed"]
      }
    }
  ]
}
```
