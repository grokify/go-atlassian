# GoJira Tasks

Open items identified from project assessment.

## High Priority

All high priority tasks completed.

## Medium Priority

All medium priority tasks completed.

## Low Priority

- [x] Add webhook/event handling support (WebhookService with event handlers and filters)
- [ ] Improve error messages with more context

## Completed

- [x] Implement gojira CLI (search, get, version commands)
- [x] Add cobra dependency for CLI
- [x] Implement JSON, Table, TOON output formats
- [x] Add flexible authentication (env vars, goauth files, CLI flags)
- [x] Add unit tests for IssueService operations (search, get, patch)
- [x] Fix lint issue in `rest/issue_service.go:88` (ineffectual assignment)
- [x] Unify logging to slog (remove zerolog dependency)
- [x] Rename `SearchIssuesDeprecated` to `SearchIssuesOnPremise` with clear documentation
- [x] Add godoc comments to exported functions in rest/
- [x] Refactor package names: `jirarest/` → `rest/`, `jiraxml/` → `xml/`, `jiraweb/` → `web/`
- [x] Add integration tests with mock Jira server (supports live server via env vars)
- [x] Add architecture documentation (added to README.md)
- [x] Clean up commented-out code blocks
- [x] Add issue creation helpers (`gojira create` with YAML support)
- [x] Add comment management (`gojira comments` command)
- [x] Add `gojira link` command for creating issue links
- [x] Add `--validate` flag to `gojira create` for pre-submission validation
- [x] Add field suggestions for custom fields (fuzzy matching)
- [x] Add `gojira transition` command for transitioning issues
- [x] Add `gojira clone` command for cloning issues
- [x] Add `gojira bulk-transition` command for batch transitions
- [x] Add sprint/board support (Agile API)
  - [x] `gojira boards` - list boards
  - [x] `gojira sprints <board-id>` - list sprints
  - [x] `gojira sprint-issues <sprint-id>` - list issues in sprint
  - [x] `gojira velocity <board-id>` - velocity report
- [x] Add `gojira bulk-update` command for batch updates
- [x] Add MCP server tools for clone and bulk update
- [x] Add MCP server tools for reports (velocity, burndown, worklog, cycle-time)
- [x] Add MCP server tools for agile (boards, sprints, move to sprint)
- [x] Add BoardService unit tests (GetBoards, GetBoard, GetSprints, etc.)
- [x] Add workflow templates (WorkflowService for reusable transition sequences)
- [x] Add reporting commands
  - [x] `gojira burndown <sprint-id>` - sprint burndown data
  - [x] `gojira worklog` - time tracking summary
  - [x] `gojira cycle-time` - cycle time analysis
- [x] Add CSV and Markdown output formats (`--csv`, `--markdown`)
- [x] Add sprint planning: `gojira sprint-move <sprint-id> <issues...>`
