# Dashforge Integration - Roadmap

## Version Planning

| Version | Milestone | Features |
|---------|-----------|----------|
| v0.38.0 | Core Report IR | Definition types, parser, basic engine |
| v0.39.0 | Section Processors | All section types, data transformation |
| v0.40.0 | Export & CLI | HTML/Markdown export, CLI commands |
| v0.41.0 | Interactive Mode | Notebook server, live updates |

---

## v0.38.0 - Core Report Infrastructure

### Report Package Foundation

| Feature | Priority | Status | Description |
|---------|----------|--------|-------------|
| `report.Definition` type | High | Pending | Core report definition structure |
| YAML/JSON parser | High | Pending | Parse report definitions |
| JSON Schema | High | Pending | Validate report definitions |
| `report.Engine` | High | Pending | Report execution engine |
| Section processor interface | High | Pending | Extensible section handling |

### Initial Section Types

| Section | Priority | Status | Description |
|---------|----------|--------|-------------|
| `markdown` | High | Pending | Static markdown content |
| `jql` | High | Pending | JQL query results |
| `velocity` | High | Pending | Sprint velocity chart |
| `burndown` | High | Pending | Sprint burndown chart |

### Dashforge IR

| Feature | Priority | Status | Description |
|---------|----------|--------|-------------|
| `DashboardIR` types | High | Pending | Match Dashforge schema |
| IR transformer | High | Pending | Convert sections to widgets |
| Data source mapping | High | Pending | Inline data sources |

---

## v0.39.0 - Complete Section Coverage

### Additional Section Types

| Section | Priority | Status | Description |
|---------|----------|--------|-------------|
| `worklog` | Medium | Pending | Time tracking summary |
| `cycletime` | Medium | Pending | Cycle time analysis |
| `metric` | Medium | Pending | Single metric display |
| `table` | Medium | Pending | Generic data table |

### Enhanced Features

| Feature | Priority | Status | Description |
|---------|----------|--------|-------------|
| Variable resolution | Medium | Pending | Template variables support |
| Theme support | Medium | Pending | Color and styling themes |
| Caching layer | Medium | Pending | Cache Jira API responses |
| Parallel processing | Low | Pending | Process sections concurrently |

---

## v0.40.0 - Export & CLI

### Export Formats

| Format | Priority | Status | Description |
|--------|----------|--------|-------------|
| HTML (self-contained) | High | Pending | Embed viewer and data |
| Markdown | Medium | Pending | Text-based output |
| PDF | Low | Pending | Via headless browser |
| JSON (IR) | Low | Pending | Raw Dashforge IR |

### CLI Commands

| Command | Priority | Status | Description |
|---------|----------|--------|-------------|
| `gojira report validate` | High | Pending | Validate definition |
| `gojira report export` | High | Pending | Export to file |
| `gojira report render` | Medium | Pending | Open in browser |
| `gojira report list` | Low | Pending | List templates |

### Built-in Templates

| Template | Priority | Status | Description |
|----------|----------|--------|-------------|
| `sprint-report` | High | Pending | Weekly sprint summary |
| `velocity-report` | Medium | Pending | Team velocity trends |
| `worklog-report` | Medium | Pending | Time tracking summary |
| `issue-breakdown` | Low | Pending | Issue analysis |

---

## v0.41.0 - Interactive Mode

### Notebook Server

| Feature | Priority | Status | Description |
|---------|----------|--------|-------------|
| HTTP server | Medium | Pending | Serve reports interactively |
| WebSocket updates | Medium | Pending | Live data refresh |
| Variable controls | Medium | Pending | UI for parameters |
| Session management | Low | Pending | Persist user state |

### Interactive Features

| Feature | Priority | Status | Description |
|---------|----------|--------|-------------|
| Cell execution | Medium | Pending | Run individual sections |
| Drill-down navigation | Low | Pending | Click to explore data |
| Export from browser | Low | Pending | Download reports |
| Save/load sessions | Low | Pending | Persistent notebooks |

---

## Future Considerations (Post v0.41.0)

### Extended Data Sources

- [ ] Multiple Jira instances
- [ ] External data sources (CSV, APIs)
- [ ] Derived/computed data sources

### Advanced Visualizations

- [ ] Custom chart types
- [ ] Geographic maps
- [ ] Network diagrams
- [ ] Custom widgets

### Collaboration

- [ ] Shared report templates
- [ ] Team dashboards
- [ ] Scheduled reports
- [ ] Email delivery

### Enterprise Features

- [ ] Role-based access
- [ ] Audit logging
- [ ] Report versioning
- [ ] Template marketplace

---

## Dependencies & Prerequisites

### External Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| Dashforge | v0.1.0+ | Dashboard IR and viewer |
| gopkg.in/yaml.v3 | v3.0.1 | YAML parsing |
| gojsonschema | v1.2.0 | JSON Schema validation |
| chromedp | v0.9.3 | PDF export (optional) |

### GoJira Prerequisites

| Feature | Version | Required For |
|---------|---------|--------------|
| BoardService | v0.36.0 | Velocity, burndown |
| VelocityReport | v0.35.0 | Velocity sections |
| BurndownReport | v0.35.0 | Burndown sections |
| WorklogReport | v0.35.0 | Worklog sections |
| CycleTimeReport | v0.35.0 | Cycle time sections |

---

## Release Checklist

### v0.38.0 Release

- [ ] Report package implemented
- [ ] Parser with validation working
- [ ] Basic sections (markdown, jql, velocity, burndown)
- [ ] Dashforge IR generation
- [ ] Unit tests passing
- [ ] Documentation updated

### v0.39.0 Release

- [ ] All section types implemented
- [ ] Variable resolution working
- [ ] Theme support added
- [ ] Caching implemented
- [ ] Integration tests passing

### v0.40.0 Release

- [ ] HTML export working
- [ ] Markdown export working
- [ ] CLI commands implemented
- [ ] Built-in templates added
- [ ] PDF export (optional)
- [ ] User documentation

### v0.41.0 Release

- [ ] Notebook server working
- [ ] Live updates functional
- [ ] Variable controls UI
- [ ] Export from browser
- [ ] Performance optimized

---

## Current Status

**Preparing for v0.38.0 development**

Immediate next steps:
1. Create `report/` package structure
2. Implement core definition types
3. Create YAML parser
4. Implement basic engine
5. Add markdown and JQL processors
