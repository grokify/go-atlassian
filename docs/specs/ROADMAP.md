# GoJira Roadmap

This document outlines planned features and enhancements for GoJira.

## v0.33.0 - CLI & SDK Enhancements (Completed)

### CLI Commands

| Command | Status | Description |
|---------|--------|-------------|
| `gojira issue-types` | Done | List available issue types for a project |
| `gojira transitions` | Done | Show available workflow transitions for an issue |
| `gojira link` | Done | Create links between issues |
| `gojira create --validate` | Done | Validate YAML against createmeta before creating |

### SDK Features

| Feature | Status | Description |
|---------|--------|-------------|
| `GetRequiredFieldsForProject()` | Done | Return only required fields for issue creation |
| `GetRequiredFields()` | Done | Return required fields for specific issue type |
| Field name suggestions | Done | Suggest similar field names when not found |
| Custom field caching | Done | Client caches fields during initialization |

### Testing

| Item | Status | Description |
|------|--------|-------------|
| CustomFields helper tests | Done | Unit tests for MapNameToIDs, DuplicateNames, etc. |
| CreateMetaService tests | Done | Mock server tests for createmeta API |
| CustomFieldService tests | Done | Tests for lookup methods |

### Documentation

| Item | Status | Description |
|------|--------|-------------|
| Troubleshooting guide | Done | Common errors and solutions |
| Custom fields guide | Done | Deep dive on working with custom fields |

## v0.34.0 - Workflow Automation (Completed)

### CLI Commands

| Command | Status | Description |
|---------|--------|-------------|
| `gojira transition` | Done | Transition an issue to a new status |
| `gojira clone` | Done | Clone an issue with field mapping |
| `gojira bulk-transition` | Done | Transition multiple issues at once |
| `gojira bulk-update` | Done | Update multiple issues with same changes |

### SDK Features

| Feature | Status | Description |
|---------|--------|-------------|
| `TransitionIssue()` | Done | Transition single issue with validation |
| `CloneIssue()` | Done | Clone issue with configurable field mapping |
| `BulkTransitionIssues()` | Done | Transition multiple issues efficiently |
| `BulkUpdateIssues()` | Done | Update multiple issues in batch |
| Workflow templates | Done | Define reusable transition sequences via WorkflowService |

### MCP Server

| Tool | Status | Description |
|------|--------|-------------|
| `jira_transition_issue` | Done | Transition issue via MCP (updated to use SDK) |
| `jira_clone_issue` | Done | Clone issue via MCP |
| `jira_bulk_update` | Done | Bulk operations via MCP |

## v0.35.0 - Reporting & Analytics (Completed)

### CLI Commands

| Command | Status | Description |
|---------|--------|-------------|
| `gojira velocity` | Done | Sprint velocity report |
| `gojira burndown` | Done | Sprint burndown data |
| `gojira worklog` | Done | Time tracking summary |
| `gojira cycle-time` | Done | Issue cycle time analysis |

### SDK Features

| Feature | Status | Description |
|---------|--------|-------------|
| `GetVelocityReport()` | Done | Calculate team velocity from sprints |
| `GetBurndownReport()` | Done | Generate burndown chart data |
| `GetWorklogReport()` | Done | Aggregate time tracking data |
| `GetCycleTimeReport()` | Done | Calculate lead/cycle times |
| Custom report builder | Planned | Flexible report generation |

### Export Formats

| Format | Status | Description |
|--------|--------|-------------|
| CSV | Done | Export reports as CSV (`--csv`) |
| JSON | Done | Structured report data (all commands) |
| Markdown | Done | Report as markdown tables (`--markdown`) |

### MCP Server

| Tool | Status | Description |
|------|--------|-------------|
| `jira_velocity_report` | Done | Velocity report via MCP |
| `jira_burndown_report` | Done | Burndown report via MCP |
| `jira_worklog_report` | Done | Worklog report via MCP |
| `jira_cycle_time_report` | Done | Cycle time report via MCP |

## v0.36.0 - Agile & Sprint Support (Completed)

### CLI Commands

| Command | Status | Description |
|---------|--------|-------------|
| `gojira boards` | Done | List available boards |
| `gojira sprints <board-id>` | Done | List sprints for a board |
| `gojira sprint-issues <sprint-id>` | Done | List issues in a sprint |
| `gojira sprint-move <sprint-id> <issues>` | Done | Move issues to a sprint |
| `gojira backlog` | Existing | View/manage backlog |

### SDK Features

| Feature | Status | Description |
|---------|--------|-------------|
| `BoardService` | Done | Board listing and details |
| `GetBoards()` | Done | List all boards |
| `GetBoard()` | Done | Get single board by ID |
| `GetSprints()` | Done | List sprints for a board |
| `GetSprintIssues()` | Done | Get issues in a sprint |
| `MoveIssuesToSprint()` | Done | Move issues to a sprint |
| Backlog service | Existing | BacklogService already exists |

### MCP Server

| Tool | Status | Description |
|------|--------|-------------|
| `jira_get_boards` | Done | List boards via MCP |
| `jira_get_sprints` | Done | List sprints via MCP |
| `jira_move_to_sprint` | Done | Move issues to sprint via MCP |

## v0.37.0 - Webhooks & Events (In Progress)

### SDK Features

| Feature | Status | Description |
|---------|--------|-------------|
| `WebhookService` | Done | Service for receiving and processing webhook events |
| Event handlers | Done | `On()`, `OnAll()`, `OnIssue()`, `OnComment()`, etc. |
| Event filtering | Done | Filter by project, issue type, or user |
| HTTP handler | Done | `HTTPHandler()` returns http.HandlerFunc |
| Event types | Done | All standard Jira webhook event types defined |

### Remaining Features

| Feature | Priority | Description |
|---------|----------|-------------|
| Webhook registration | Medium | Register webhooks programmatically via API |
| Retry handling | Low | Automatic retry for failed deliveries |

## Contributing

See [PLAN.md](PLAN.md) for implementation details on current work items.
